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
// en el cuerpo de la solicitud JSON. Usar una struct con etiquetas JSON (`json:"..."`)
// permite una decodificación segura y automática del JSON entrante.
type LoginPayload struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Login es el manejador para la ruta POST /api/login. Orquesta el proceso de autenticación.
// Recibe las credenciales en formato JSON, las valida contra la base de datos,
// y si son correctas, establece una sesión de usuario.
func Login(w http.ResponseWriter, r *http.Request) {
	// Log para depuración: Indica que se ha recibido una solicitud en este endpoint.
	log.Println("--- Intento de Inicio de Sesión Recibido ---")
	var payload LoginPayload

	// 'json.NewDecoder(r.Body).Decode(&payload)' lee el cuerpo de la solicitud HTTP
	// y decodifica el JSON directamente en la struct 'payload'.
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		// Log para depuración: Registra un error si el JSON es inválido.
		log.Printf("[ERROR-LOGIN] Falla al decodificar el cuerpo JSON: %v", err)
		respondWithError(w, http.StatusBadRequest, "Cuerpo de la solicitud inválido")
		return
	}

	// Log para depuración: Muestra los datos recibidos del frontend.
	// Es crucial para verificar que los datos llegan correctamente.
	log.Printf("[DEBUG-LOGIN] Payload recibido: Usuario=[%s], Contraseña=[%s]", payload.Username, payload.Password)

	// Se busca al usuario en la base de datos usando el nombre de usuario proporcionado.
	user, err := database.GetUserByUsername(payload.Username)
	if err != nil {
		// Log para depuración: Registra si el usuario no fue encontrado.
		log.Printf("[ERROR-LOGIN] Error al buscar usuario '%s' en la BD: %v", payload.Username, err)
		// Se devuelve un error genérico para no revelar si el usuario existe o no.
		respondWithError(w, http.StatusUnauthorized, "Credenciales inválidas")
		return
	}

	// Log para depuración: Muestra el hash de la contraseña recuperado de la base de datos.
	log.Printf("[DEBUG-LOGIN] Usuario '%s' encontrado en la BD. Hash de la BD: [%s]", user.Username, user.Password)
	log.Println("[DEBUG-LOGIN] Comparando hash de la BD con la contraseña proporcionada...")

	// 'bcrypt.CompareHashAndPassword' es la función de seguridad clave.
	// Compara el hash almacenado en la base de datos ('user.Password') con la contraseña
	// en texto plano enviada por el usuario ('payload.Password'). Es un proceso lento por diseño
	// para dificultar los ataques de fuerza bruta. Devuelve un error si no coinciden.
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(payload.Password))
	if err != nil {
		// Log para depuración: Este es el log más importante. Si aparece, significa que la contraseña no coincide.
		log.Printf("[ERROR-LOGIN] bcrypt.CompareHashAndPassword falló: %v", err)
		respondWithError(w, http.StatusUnauthorized, "Credenciales inválidas")
		return
	}

	// Log para depuración: Confirma que la contraseña es correcta.
	log.Println("[SUCCESS-LOGIN] ¡La contraseña coincide! Creando sesión...")

	// Si la contraseña es correcta, se procede a crear la sesión.
	session, _ := auth.Store.Get(r, "session-name")
	// Se guardan los datos relevantes en el mapa de valores de la sesión.
	session.Values["authenticated"] = true
	session.Values["userID"] = user.ID
	session.Values["username"] = user.Username
	// 'session.Save' escribe la cookie de sesión en la respuesta HTTP.
	if err := session.Save(r, w); err != nil {
		log.Printf("[ERROR-LOGIN] Falla al guardar la sesión: %v", err)
		respondWithError(w, http.StatusInternalServerError, "No se pudo guardar la sesión")
		return
	}

	// Log para depuración: Confirma que la sesión se creó y se está enviando la respuesta.
	log.Printf("[SUCCESS-LOGIN] Sesión creada para el usuario '%s'. Respondiendo al cliente.", user.Username)
	// Se devuelve una respuesta JSON con los datos del usuario (sin la contraseña).
	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"id":         user.ID,
		"username":   user.Username,
		"firstLogin": user.FirstLogin,
	})
}

// Logout maneja el cierre de sesión del usuario para la ruta POST /api/logout.
func Logout(w http.ResponseWriter, r *http.Request) {
	session, _ := auth.Store.Get(r, "session-name")

	// Se limpia el mapa de valores de la sesión para eliminar los datos del usuario.
	session.Values["authenticated"] = false
	delete(session.Values, "userID")
	delete(session.Values, "username")
	delete(session.Values, "connectionID")

	// Se establece MaxAge a -1 para que el navegador elimine la cookie inmediatamente.
	session.Options.MaxAge = -1
	if err := session.Save(r, w); err != nil {
		respondWithError(w, http.StatusInternalServerError, "No se pudo cerrar la sesión")
		return
	}

	// Se envía una respuesta de éxito.
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
	// Se recupera el ID de usuario desde la sesión para asegurar que el usuario
	// correcto está cambiando su propia contraseña.
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

	// Se realizan validaciones básicas sobre la nueva contraseña.
	if payload.NewPassword == "" || payload.NewPassword != payload.ConfirmPassword {
		respondWithError(w, http.StatusBadRequest, "Las contraseñas no coinciden o están vacías")
		return
	}
	if payload.NewPassword == "admin" {
		respondWithError(w, http.StatusBadRequest, "La nueva contraseña no puede ser la predeterminada")
		return
	}

	// Se llama a la función del paquete 'database' para actualizar la contraseña.
	err := database.UpdateUserPassword(userID, payload.NewPassword)
	if err != nil {
		log.Printf("Error al actualizar la contraseña: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Error al actualizar la contraseña")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"message": "Contraseña actualizada exitosamente"})
}

// ExtendSession refresca la cookie de la sesión para extender su tiempo de vida.
// Responde a una solicitud POST a /api/session/extend.
func ExtendSession(w http.ResponseWriter, r *http.Request) {
	session, err := auth.Store.Get(r, "session-name")
	if err != nil || session.IsNew {
		respondWithError(w, http.StatusUnauthorized, "Sesión no válida o no encontrada")
		return
	}

	// Simplemente volviendo a guardar la sesión, el middleware de gorilla/sessions
	// actualizará la fecha de expiración de la cookie.
	if err := session.Save(r, w); err != nil {
		respondWithError(w, http.StatusInternalServerError, "No se pudo extender la sesión")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"message": "Sesión extendida"})
}
