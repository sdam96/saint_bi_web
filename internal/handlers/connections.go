// internal/handlers/connections.go
package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"saintnet.com/m/internal/auth"
	"saintnet.com/m/internal/database"
	"saintnet.com/m/internal/models"
)

func GetConnections(w http.ResponseWriter, r *http.Request) {
	connections, err := database.GetConnections()
	if err != nil {
		log.Printf("[ERROR] Error obteniendo conexiones: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Error del servidor al obtener conexiones")
		return
	}
	consolidated := models.Connection{
		ID:    0,
		Alias: "Consolidado",
	}
	allConnections := append([]models.Connection{consolidated}, connections...)
	respondWithJSON(w, http.StatusOK, allConnections)
}

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
		log.Printf("[ERROR] Error agregando conexión a la BD: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Error al agregar la conexión")
		return
	}

	session, _ := auth.Store.Get(r, "session-name")
	username := session.Values["username"]
	log.Printf("[INFO] Usuario '%s' agregó una nueva conexión: '%s'", username, conn.Alias)

	connections, _ := database.GetConnections()
	respondWithJSON(w, http.StatusCreated, connections)
}

func DeleteConnection(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "ID de conexión inválido")
		return
	}

	// Obtenemos el alias ANTES de borrar para el log.
	conn, err := database.GetConnectionByID(id)
	if err != nil {
		log.Printf("[WARN] Intento de eliminar conexión no existente con ID %d", id)
		respondWithError(w, http.StatusNotFound, "La conexión no existe")
		return
	}

	if err := database.DeleteConnection(id); err != nil {
		log.Printf("[ERROR] Error eliminando conexión de la BD con ID %d: %v", id, err)
		respondWithError(w, http.StatusInternalServerError, "Error al eliminar la conexión")
		return
	}

	session, _ := auth.Store.Get(r, "session-name")
	username := session.Values["username"]
	log.Printf("[INFO] Usuario '%s' eliminó la conexión: '%s'", username, conn.Alias)

	w.WriteHeader(http.StatusNoContent)
}
