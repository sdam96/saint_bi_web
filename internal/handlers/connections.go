// internal/handlers/connections.go
package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"saintnet.com/m/internal/database"
	"saintnet.com/m/internal/models"
)

// GetConnections es el manejador para obtener la lista de todas las conexiones.
// Responde a una solicitud GET.
func GetConnections(w http.ResponseWriter, r *http.Request) {
	// Llama a la base de datos para obtener todas las conexiones.
	connections, err := database.GetConnections()
	if err != nil {
		log.Printf("Error obteniendo conexiones: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Error del servidor al obtener conexiones")
		return
	}

	// Responde con la lista de conexiones en formato JSON.
	respondWithJSON(w, http.StatusOK, connections)
}

// AddConnection es el manejador para agregar una nueva conexión.
// Responde a una solicitud POST con un cuerpo JSON.
func AddConnection(w http.ResponseWriter, r *http.Request) {
	var conn models.Connection

	// Decodifica el cuerpo de la solicitud JSON en la struct de conexión.
	if err := json.NewDecoder(r.Body).Decode(&conn); err != nil {
		respondWithError(w, http.StatusBadRequest, "Cuerpo de la solicitud inválido")
		return
	}

	// Validación básica.
	if conn.Alias == "" || conn.ApiURL == "" || conn.ApiUser == "" || conn.ApiPassword == "" {
		respondWithError(w, http.StatusBadRequest, "Los campos alias, api_url, api_user y api_password son requeridos")
		return
	}

	// Llama a la base de datos para guardar la nueva conexión.
	if err := database.AddConnection(conn); err != nil {
		log.Printf("Error agregando conexión: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Error al agregar la conexión")
		return
	}

	// Devuelve la lista actualizada de conexiones para que el frontend pueda refrescar su estado.
	connections, _ := database.GetConnections()
	respondWithJSON(w, http.StatusCreated, connections)
}

// DeleteConnection es el manejador para eliminar una conexión existente.
// Responde a una solicitud DELETE.
func DeleteConnection(w http.ResponseWriter, r *http.Request) {
	// Extrae el ID de la ruta de la URL (ej. /api/connections/123).
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "ID de conexión inválido")
		return
	}

	// Llama a la base de datos para eliminar la conexión.
	if err := database.DeleteConnection(id); err != nil {
		log.Printf("Error eliminando conexión: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Error al eliminar la conexión")
		return
	}

	// Responde con un código 204 No Content, que es la respuesta estándar
	// para una eliminación exitosa sin necesidad de devolver un cuerpo.
	w.WriteHeader(http.StatusNoContent)
}
