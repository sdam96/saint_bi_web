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
	connections, err := database.GetConnections()
	if err != nil {
		log.Printf("Error obteniendo conexiones: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Error del servidor al obtener conexiones")
		return
	}

	// --- AÑADIR ESTE BLOQUE ---
	// Crear la conexión consolidada virtual
	consolidated := models.Connection{
		ID:    0, // ID especial para identificarla
		Alias: "Consolidado",
	}
	allConnections := append([]models.Connection{consolidated}, connections...)

	// Se responde con la lista modificada
	respondWithJSON(w, http.StatusOK, allConnections)
}

// AddConnection (sin cambios)
func AddConnection(w http.ResponseWriter, r *http.Request) {
	var conn models.Connection

	if err := json.NewDecoder(r.Body).Decode(&conn); err != nil {
		respondWithError(w, http.StatusBadRequest, "Cuerpo de la solicitud inválido")
		return
	}

	if conn.Alias == "" || conn.ApiURL == "" || conn.ApiUser == "" || conn.ApiPassword == "" {
		respondWithError(w, http.StatusBadRequest, "Los campos alias, api_url, api_user y api_password son requeridos")
		return
	}

	if err := database.AddConnection(conn); err != nil {
		log.Printf("Error agregando conexión: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Error al agregar la conexión")
		return
	}

	connections, _ := database.GetConnections()
	respondWithJSON(w, http.StatusCreated, connections)
}

// DeleteConnection (sin cambios)
func DeleteConnection(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "ID de conexión inválido")
		return
	}

	if err := database.DeleteConnection(id); err != nil {
		log.Printf("Error eliminando conexión: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Error al eliminar la conexión")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
