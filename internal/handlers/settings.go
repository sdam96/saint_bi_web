package handlers

import (
	"encoding/json"
	"net/http"

	"saintnet.com/m/internal/auth"
	"saintnet.com/m/internal/database"
	"saintnet.com/m/internal/models"
)

func GetSettings(w http.ResponseWriter, r *http.Request) {
	session, _ := auth.Store.Get(r, "session-name")
	connID, ok := session.Values["connectionID"].(int)

	if !ok || connID == 0 {
		respondWithError(w, http.StatusBadRequest, "No hay conexion seleccionada")
		return
	}

	connection, err := database.GetConnectionByID(connID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Conexion no encontrada")
		return
	}

	respondWithJSON(w, http.StatusOK, connection)
}

func UpdateSettings(w http.ResponseWriter, r *http.Request) {
	session, _ := auth.Store.Get(r, "session-name")
	connID, ok := session.Values["connectionID"].(int)
	if !ok || connID == 0 {
		respondWithError(w, http.StatusBadRequest, "No hay una conexión seleccionada")
		return
	}

	var payload models.Connection
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		respondWithError(w, http.StatusBadRequest, "Cuerpo de la solicitud inválido")
		return
	}

	// Llama a una nueva función en la base de datos para actualizar.
	err := database.UpdateConnectionSettings(connID, payload.CurrencyISO, payload.LocaleFormat)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error al actualizar la configuración")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"message": "Configuración actualizada exitosamente"})
}
