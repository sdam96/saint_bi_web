// La declaración 'package database' define que este archivo pertenece al paquete 'database'.
// Este paquete encapsula toda la lógica de interacción con la base de datos,
// proporcionando una capa de abstracción para el resto de la aplicación.
package database

import (
	// "database/sql" es el paquete estándar de Go que proporciona una interfaz genérica
	// para bases de datos SQL. No es un driver de base de datos en sí mismo, sino una
	// capa de abstracción que permite escribir código SQL independiente del motor de base de datos subyacente.
	"database/sql"
	// "log" implementa un paquete de logging simple, útil para imprimir mensajes
	// de estado o errores, como la creación del usuario administrador por defecto.
	"log"

	// La importación con un guion bajo '_' como alias ('_ "github.com/mattn/go-sqlite3"')
	// es una forma de importar un paquete únicamente por sus efectos secundarios. En este caso,
	// el efecto secundario es que el driver de go-sqlite3 se registra a sí mismo con el paquete
	// 'database/sql'. Esto permite que 'sql.Open' pueda usar el driver "sqlite3".
	_ "github.com/mattn/go-sqlite3"
	// 'golang.org/x/crypto/bcrypt' proporciona una implementación del algoritmo de hashing de contraseñas bcrypt.
	// Es la forma recomendada y segura de almacenar contraseñas, ya que es computacionalmente costoso
	// y previene ataques de fuerza bruta y de tablas arcoíris.
	"golang.org/x/crypto/bcrypt"
	// Importamos el paquete 'models' que contiene las definiciones de las estructuras de datos
	// como 'User' y 'Connection', permitiendo un tipado fuerte y claro en nuestro código.
	"saintnet.com/m/internal/models"
)

// 'var DB *sql.DB' declara una variable global a nivel de paquete para la conexión a la base de datos.
// Usar un puntero a 'sql.DB' es la práctica estándar. Es importante destacar que 'sql.DB' no es una
// única conexión, sino un pool de conexiones a la base de datos, gestionado de forma concurrente y segura por el paquete 'sql'.
var DB *sql.DB

// InitDB inicializa la conexión a la base de datos, crea las tablas si no existen,
// y se asegura de que haya un usuario administrador por defecto.
// Devuelve el pool de conexiones (*sql.DB) y un posible error.
func InitDB() (*sql.DB, error) {
	var err error
	// 'sql.Open' abre una conexión a la base de datos.
	// El primer argumento ("sqlite3") es el nombre del driver que registramos con la importación '_'.
	// El segundo argumento ("./data.db") es la cadena de conexión (DSN), que en el caso de SQLite
	// es simplemente la ruta al archivo de la base de datos.
	// Importante: 'sql.Open' no crea la conexión inmediatamente ni verifica que sea válida,
	// solo prepara el objeto 'sql.DB'. La conexión real se establece de forma perezosa (lazy) cuando es necesaria.
	DB, err = sql.Open("sqlite3", "./data.db")
	if err != nil {
		return nil, err
	}

	// 'createTables' es una cadena de texto (string) multi-línea que contiene las sentencias SQL
	// para crear las tablas 'users' y 'connections' si es que no existen ya.
	// `IF NOT EXISTS` previene errores si la aplicación se reinicia.
	createTables := `
		CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT NOT NULL UNIQUE,
			password TEXT NOT NULL,
			first_login BOOLEAN NOT NULL DEFAULT TRUE
		);
		CREATE TABLE IF NOT EXISTS connections (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			alias TEXT NOT NULL UNIQUE,
			api_url TEXT NOT NULL,
			api_user TEXT NOT NULL,
			api_password TEXT NOT NULL,
			refresh_seconds INTEGER NOT NULL,
			config_id INTEGER NOT NULL
		);
	`
	// 'DB.Exec' ejecuta una consulta SQL que no devuelve filas, como CREATE, INSERT, UPDATE o DELETE.
	// Los `?` en las consultas posteriores son marcadores de posición para evitar inyecciones SQL.
	_, err = DB.Exec(createTables)
	if err != nil {
		return nil, err
	}

	// --- Lógica para crear el usuario 'admin' por defecto ---
	var count int
	// 'DB.QueryRow' ejecuta una consulta que se espera que devuelva como máximo una fila.
	// Es ideal para consultas como 'SELECT COUNT(*)'.
	// '.Scan(&count)' copia el valor de la columna resultante en la variable 'count'.
	// Se pasa un puntero a 'count' para que Scan pueda modificar su valor.
	err = DB.QueryRow("SELECT COUNT(*) FROM users WHERE username = 'admin'").Scan(&count)
	if err != nil {
		return nil, err
	}

	// Si el conteo es 0, el usuario 'admin' no existe y debe ser creado.
	if count == 0 {
		// 'bcrypt.GenerateFromPassword' hashea la contraseña.
		// El primer argumento es la contraseña en un slice de bytes.
		// El segundo es el "costo", que determina cuánto trabajo computacional se requiere.
		// 'bcrypt.DefaultCost' (actualmente 10) es un valor seguro por defecto.
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("admin"), bcrypt.DefaultCost)
		// Se inserta el nuevo usuario 'admin' en la base de datos.
		_, err = DB.Exec("INSERT INTO users (username, password, first_login) VALUES (?, ?, ?)", "admin", string(hashedPassword), true)
		if err != nil {
			return nil, err
		}
		log.Println("Usuario 'admin' con clave 'admin' creado")
	}
	// Devuelve el pool de conexiones y un error nulo para indicar éxito.
	return DB, nil
}

// GetUserByUsername busca un usuario por su nombre de usuario.
func GetUserByUsername(username string) (*models.User, error) {
	// Se crea un puntero a una instancia vacía de 'models.User' para almacenar el resultado.
	user := &models.User{}
	// Se usa QueryRow para obtener la fila del usuario. Luego, Scan mapea las columnas
	// del resultado (id, username, password, first_login) a los campos de la struct 'user'.
	// El orden de los argumentos en Scan DEBE coincidir con el orden de las columnas en el SELECT.
	err := DB.QueryRow("SELECT id, username, password, first_login FROM users WHERE username = ?", username).Scan(&user.ID, &user.Username, &user.Password, &user.FirstLogin)
	if err != nil {
		// Si 'QueryRow' no encuentra ninguna fila, '.Scan' devolverá un error 'sql.ErrNoRows'.
		// Esto es útil para saber si un usuario no existe.
		return nil, err
	}
	return user, nil
}

// UpdateUserPassword actualiza la contraseña de un usuario y marca first_login como false.
func UpdateUserPassword(userID int, newPassword string) error {
	// Se hashea la nueva contraseña antes de guardarla.
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	// Se ejecuta la sentencia UPDATE, pasando los valores como parámetros para seguridad.
	// Se convierte el 'hashedPassword' (un []byte) a string para almacenarlo.
	_, err = DB.Exec("UPDATE users SET password = ?, first_login = ? WHERE id = ?", string(hashedPassword), false, userID)
	return err
}

// AddUser agrega un nuevo usuario a la base de datos.
func AddUser(username, password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	_, err = DB.Exec("INSERT INTO users (username, password) VALUES (?, ?)", username, string(hashedPassword))
	return err
}

// GetUsers devuelve todos los usuarios de la aplicación (solo ID y username por seguridad).
func GetUsers() ([]models.User, error) {
	// 'DB.Query' se usa para consultas que pueden devolver múltiples filas.
	rows, err := DB.Query("SELECT id, username FROM users")
	if err != nil {
		return nil, err
	}
	// 'defer rows.Close()' es crucial. Asegura que el conjunto de resultados se cierre
	// al final de la función, liberando la conexión a la base de datos para que otros la usen.
	defer rows.Close()

	// Se declara un slice de 'models.User' para acumular los resultados.
	var users []models.User
	// 'for rows.Next()' itera sobre cada una de las filas del resultado.
	// Devuelve 'false' cuando no hay más filas o si ocurre un error durante la iteración.
	for rows.Next() {
		var u models.User
		// 'rows.Scan' escanea los valores de la fila actual en la struct 'u'.
		if err := rows.Scan(&u.ID, &u.Username); err != nil {
			return nil, err
		}
		// 'append' agrega el usuario recién escaneado al slice de usuarios.
		users = append(users, u)
	}
	// Si todo va bien, se devuelve el slice de usuarios y un error 'nil'.
	return users, nil
}

// AddConnection agrega una nueva conexión a la API.
func AddConnection(conn models.Connection) error {
	// El método Exec se usa para la inserción, pasando los campos de la struct 'conn' como
	// parámetros para la consulta parametrizada.
	_, err := DB.Exec("INSERT INTO connections (alias, api_url, api_user, api_password, refresh_seconds, config_id) VALUES (?, ?, ?, ?, ?, ?)",
		conn.Alias, conn.ApiURL, conn.ApiUser, conn.ApiPassword, conn.RefreshSeconds, conn.ConfigID)
	return err
}

// GetConnections devuelve todas las conexiones guardadas.
func GetConnections() ([]models.Connection, error) {
	rows, err := DB.Query("SELECT id, alias, api_url, api_user, api_password, refresh_seconds, config_id FROM connections ORDER BY alias")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var connections []models.Connection
	// El bucle 'for rows.Next()' es el patrón estándar en Go para procesar resultados de múltiples filas.
	for rows.Next() {
		var c models.Connection
		if err := rows.Scan(&c.ID, &c.Alias, &c.ApiURL, &c.ApiUser, &c.ApiPassword, &c.RefreshSeconds, &c.ConfigID); err != nil {
			return nil, err
		}
		connections = append(connections, c)
	}
	return connections, nil
}

// GetConnectionByID busca una conexión por su ID.
func GetConnectionByID(id int) (*models.Connection, error) {
	conn := &models.Connection{}
	err := DB.QueryRow("SELECT id, alias, api_url, api_user, api_password, refresh_seconds, config_id FROM connections WHERE id = ?", id).Scan(
		&conn.ID, &conn.Alias, &conn.ApiURL, &conn.ApiUser, &conn.ApiPassword, &conn.RefreshSeconds, &conn.ConfigID)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

// DeleteConnection elimina una conexión por su ID.
func DeleteConnection(id int) error {
	_, err := DB.Exec("DELETE FROM connections WHERE id = ?", id)
	return err
}
