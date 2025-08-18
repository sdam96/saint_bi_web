// internal/handlers/users.go
package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"saintnet.com/m/internal/database"
)

// GetUsers es el manejador que obtiene la lista de todos los usuarios.
// Responde a una solicitud GET.
func GetUsers(w http.ResponseWriter, r *http.Request) {
	// Llama a la base de datos para obtener la lista de usuarios.
	// La función de base de datos ya omite las contraseñas por seguridad.
	users, err := database.GetUsers()
	if err != nil {
		log.Printf("Error obteniendo usuarios: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Error del servidor al obtener usuarios")
		return
	}

	// Responde con la lista de usuarios en formato JSON.
	respondWithJSON(w, http.StatusOK, users)
}

// AddUserPayload define la estructura para crear un nuevo usuario.
type AddUserPayload struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// AddUser es el manejador que procesa la solicitud para agregar un nuevo usuario.
// Responde a una solicitud POST con un cuerpo JSON.
func AddUser(w http.ResponseWriter, r *http.Request) {
	var payload AddUserPayload

	// Decodifica el cuerpo de la solicitud JSON en la struct.
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		respondWithError(w, http.StatusBadRequest, "Cuerpo de la solicitud inválido")
		return
	}

	// Realiza una validación simple.
	if payload.Username == "" || payload.Password == "" {
		respondWithError(w, http.StatusBadRequest, "El usuario y la clave son requeridos")
		return
	}

	// Llama a la base de datos para crear el nuevo usuario.
	// La lógica de hashear la contraseña está encapsulada en la capa de base de datos.
	err := database.AddUser(payload.Username, payload.Password)
	if err != nil {
		log.Printf("Error agregando usuario: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Error al agregar el usuario (posiblemente ya existe)")
		return
	}

	// Después de agregar el usuario, devuelve la lista actualizada.
	// Esto es útil para que el frontend refresque su estado sin hacer otra llamada.
	users, _ := database.GetUsers()
	respondWithJSON(w, http.StatusCreated, users)
}
