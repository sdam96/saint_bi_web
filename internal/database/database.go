package database

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
	"saintnet.com/m/internal/models"
)

var DB *sql.DB

func InitDB() (*sql.DB, error) {
	var err error
	DB, err = sql.Open("sqlite3", "./data.db")
	if err != nil {
		return nil, err
	}

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
	_, err = DB.Exec(createTables)
	if err != nil {
		return nil, err
	}

	// Crear usuario admin por defecto
	var count int
	err = DB.QueryRow("SELECT COUNT(*) FROM users WHERE username = 'admin'").Scan(&count)
	if err != nil {
		return nil, err
	}

	if count == 0 {
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("admin"), bcrypt.DefaultCost)
		_, err = DB.Exec("INSERT INTO users (username, password, first_login) VALUES (?, ?, ?)", "admin", string(hashedPassword), true)
		if err != nil {
			return nil, err
		}
		log.Println("Usuario 'admin' con clave 'admin' creado")
	}
	return DB, nil
}

// GetUserByUsername busca un usuario por su nombre de usuario.
func GetUserByUsername(username string) (*models.User, error) {
	user := &models.User{}
	err := DB.QueryRow("SELECT id, username, password, first_login FROM users WHERE username = ?", username).Scan(&user.ID, &user.Username, &user.Password, &user.FirstLogin)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// UpdateUserPassword actualiza la contraseña de un usuario.
func UpdateUserPassword(userID int, newPassword string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
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

// GetUsers devuelve todos los usuarios de la aplicación.
func GetUsers() ([]models.User, error) {
	rows, err := DB.Query("SELECT id, username FROM users")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.ID, &u.Username); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}

// AddConnection agrega una nueva conexión a la API.
func AddConnection(conn models.Connection) error {
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
