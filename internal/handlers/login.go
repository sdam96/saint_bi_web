// El paquete 'handlers' contiene todos los manejadores de solicitudes HTTP,
// que actúan como controladores en una arquitectura MVC, conectando las rutas,
// la lógica de negocio y las respuestas.
package handlers

import (
	// "encoding/json" es necesario para decodificar (unmarshal) el cuerpo de la solicitud JSON
	// que enviará el frontend de Vue.
	"encoding/json"
	// "log" para registrar información de depuración y errores en la terminal del servidor.
	"log"
	// "net/http" proporciona las herramientas para construir el servidor web.
	"net/http"

	// "golang.org/x/crypto/bcrypt" para la comparación segura de contraseñas hasheadas.
	"golang.org/x/crypto/bcrypt"
	// Se importan nuestros paquetes internos para la lógica de sesión y base de datos.
	"saintnet.com/m/internal/auth"
	"saintnet.com/m/internal/database"
)

// LoginPayload define la estructura de datos que el manejador 'Login' espera recibir
// en el cuerpo de la solicitud JSON.
//
// SINTAXIS de Go:
//   - `type LoginPayload struct { ... }` declara un nuevo tipo de dato llamado 'LoginPayload' que es una estructura (struct).
//     Una struct es una colección de campos con nombre, similar a un objeto en otros lenguajes.
//   - `Username string` define un campo llamado 'Username' de tipo 'string'.
//   - “ `json:"username"` “ es una "etiqueta de campo de struct". Le indica al paquete 'json' que cuando
//     decodifique un JSON, el campo 'username' del JSON debe mapearse a este campo de la struct.
type LoginPayload struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Login es el manejador para la ruta POST /api/login. Orquesta el proceso de autenticación.
//
// LÓGICA:
// 1. Recibe las credenciales (usuario/clave) en formato JSON.
// 2. Valida las credenciales contra el hash almacenado en la base de datos usando bcrypt.
// 3. Si son correctas, establece una cookie de sesión segura en el navegador del usuario.
//
// PARÁMETROS:
//   - w (http.ResponseWriter): Es una interfaz que permite al handler escribir la respuesta HTTP
//     (cabeceras, código de estado y cuerpo) que se enviará al cliente.
//   - r (*http.Request): Es un puntero a una struct que representa la solicitud HTTP entrante,
//     incluyendo la URL, el método, las cabeceras y el cuerpo.
func Login(w http.ResponseWriter, r *http.Request) {
	// SINTAXIS de Go: `var payload LoginPayload` declara una variable llamada 'payload' del tipo 'LoginPayload'.
	// Se inicializará con los valores cero de sus campos (strings vacíos en este caso).
	var payload LoginPayload

	// LÓGICA: Se intenta decodificar el cuerpo JSON de la solicitud en la variable 'payload'.
	// SINTAXIS de Go: `err := json.NewDecoder(r.Body).Decode(&payload)`
	// - `r.Body`: Es el cuerpo de la solicitud, que es un stream de datos (io.ReadCloser).
	// - `json.NewDecoder(...)`: Crea un decodificador que lee desde ese stream.
	// - `.Decode(&payload)`: Intenta leer el JSON y volcar los datos en la variable 'payload'. Se pasa un
	//   puntero `&payload` para que la función Decode pueda modificar la variable original.
	// - `err := ...`: Esta es la declaración corta de variables. Declara 'err' y le asigna el valor de retorno.
	err := json.NewDecoder(r.Body).Decode(&payload)

	// PATRÓN COMÚN en Go: Manejo de errores inmediato.
	// Si la variable 'err' no es 'nil', significa que ocurrió un error.
	if err != nil {
		// LOGGING: Se registra un aviso (WARN) con el error específico para diagnóstico.
		log.Printf("[WARN] Falla al decodificar el cuerpo JSON en Login: %v", err)
		respondWithError(w, http.StatusBadRequest, "Cuerpo de la solicitud inválido")
		// `return` detiene la ejecución de la función para que no continúe con datos inválidos.
		return
	}

	// LOGGING: Se registra un evento informativo (INFO) para auditoría y seguimiento en tiempo real.
	// r.RemoteAddr contiene la dirección IP del cliente que realiza la solicitud.
	log.Printf("[INFO] Intento de inicio de sesión para el usuario '%s' desde %s", payload.Username, r.RemoteAddr)

	// LÓGICA: Se busca al usuario en la base de datos.
	user, err := database.GetUserByUsername(payload.Username)
	if err != nil {
		// LOGGING: Se registra un aviso (WARN) si el usuario no se encuentra. Es un aviso y no un error
		// porque es un fallo de autenticación esperado, no un fallo del sistema.
		log.Printf("[WARN] Fallo de autenticación para '%s': usuario no encontrado.", payload.Username)
		// Se devuelve un error 401 Unauthorized genérico para no revelar si el nombre de usuario existe o no.
		respondWithError(w, http.StatusUnauthorized, "Credenciales inválidas")
		return
	}

	// LÓGICA: Se compara de forma segura la contraseña proporcionada con el hash almacenado.
	// SINTAXIS de Go: `[]byte(...)` es una conversión de tipo (casting) que convierte un string a un slice de bytes,
	// que es el formato que requiere la función de bcrypt.
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(payload.Password))
	if err != nil {
		// LOGGING: Se registra un aviso (WARN) si la contraseña no coincide.
		log.Printf("[WARN] Fallo de autenticación para '%s': contraseña incorrecta.", payload.Username)
		respondWithError(w, http.StatusUnauthorized, "Credenciales inválidas")
		return
	}

	// LÓGICA: Si la autenticación es exitosa, se crea y guarda la sesión.
	session, _ := auth.Store.Get(r, "session-name")
	session.Values["authenticated"] = true
	session.Values["userID"] = user.ID
	session.Values["username"] = user.Username
	if err := session.Save(r, w); err != nil {
		// LOGGING: Se registra un error crítico (ERROR) si no se puede guardar la sesión.
		log.Printf("[ERROR] Falla al guardar la sesión para '%s': %v", payload.Username, err)
		respondWithError(w, http.StatusInternalServerError, "No se pudo guardar la sesión")
		return
	}

	// LOGGING: Se registra un evento de éxito (SUCCESS) que confirma la autenticación.
	log.Printf("[SUCCESS] Usuario '%s' autenticado exitosamente.", user.Username)

	// LÓGICA: Se envía una respuesta JSON al cliente con los datos del usuario.
	// SINTAXIS de Go: `map[string]interface{}{...}` crea un mapa con claves de tipo string y valores de
	// cualquier tipo (`interface{}`), que es perfecto para construir respuestas JSON dinámicas.
	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"id":         user.ID,
		"username":   user.Username,
		"firstLogin": user.FirstLogin,
	})
}

// Logout maneja el cierre de sesión del usuario para la ruta POST /api/logout.
func Logout(w http.ResponseWriter, r *http.Request) {
	session, _ := auth.Store.Get(r, "session-name")

	// LÓGICA: Se intenta obtener el nombre de usuario de la sesión para el log.
	// SINTAXIS de Go: `username, ok := session.Values["username"].(string)` es una "aserción de tipo".
	// Intenta convertir el valor a string. Si tiene éxito, 'ok' será true.
	username, ok := session.Values["username"].(string)
	if !ok {
		username = "desconocido"
	}

	// LÓGICA: Se invalidan los datos de la sesión.
	session.Values["authenticated"] = false
	delete(session.Values, "userID")
	delete(session.Values, "username")
	delete(session.Values, "connectionID")
	session.Options.MaxAge = -1 // Esto le indica al navegador que borre la cookie.

	if err := session.Save(r, w); err != nil {
		// LOGGING: Se registra un error si falla el guardado de la sesión invalidada.
		log.Printf("[ERROR] Falla al cerrar la sesión para '%s': %v", username, err)
		respondWithError(w, http.StatusInternalServerError, "No se pudo cerrar la sesión")
		return
	}

	// LOGGING: Se registra un evento informativo del cierre de sesión.
	log.Printf("[INFO] Sesión cerrada para el usuario '%s'.", username)
	respondWithJSON(w, http.StatusOK, map[string]string{"message": "Sesión cerrada exitosamente"})
}

// PasswordChangePayload define la estructura para la solicitud de cambio de clave.
type PasswordChangePayload struct {
	NewPassword     string `json:"newPassword"`
	ConfirmPassword string `json:"confirmPassword"`
}

// ForcePasswordChange maneja el cambio de contraseña para la ruta POST /api/force-password-change.
func ForcePasswordChange(w http.ResponseWriter, r *http.Request) {
	session, _ := auth.Store.Get(r, "session-name")
	userID, ok := session.Values["userID"].(int)
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "Sesión inválida")
		return
	}

	var payload PasswordChangePayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		respondWithError(w, http.StatusBadRequest, "Cuerpo de la solicitud inválido")
		return
	}

	if payload.NewPassword == "" || payload.NewPassword != payload.ConfirmPassword {
		respondWithError(w, http.StatusBadRequest, "Las contraseñas no coinciden o están vacías")
		return
	}
	if payload.NewPassword == "admin" {
		respondWithError(w, http.StatusBadRequest, "La nueva contraseña no puede ser la predeterminada")
		return
	}

	err := database.UpdateUserPassword(userID, payload.NewPassword)
	if err != nil {
		log.Printf("Error al actualizar la contraseña para el usuario ID %d: %v", userID, err)
		respondWithError(w, http.StatusInternalServerError, "Error al actualizar la contraseña")
		return
	}

	username := session.Values["username"]
	log.Printf("[INFO] Usuario '%s' (ID: %d) ha cambiado su contraseña.", username, userID)

	respondWithJSON(w, http.StatusOK, map[string]string{"message": "Contraseña actualizada exitosamente"})
}

// ExtendSession refresca la cookie de la sesión para extender su tiempo de vida.
func ExtendSession(w http.ResponseWriter, r *http.Request) {
	session, err := auth.Store.Get(r, "session-name")
	if err != nil || session.IsNew {
		respondWithError(w, http.StatusUnauthorized, "Sesión no válida o no encontrada")
		return
	}
	if err := session.Save(r, w); err != nil {
		respondWithError(w, http.StatusInternalServerError, "No se pudo extender la sesión")
		return
	}
	respondWithJSON(w, http.StatusOK, map[string]string{"message": "Sesión extendida"})
}
