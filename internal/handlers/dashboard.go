// internal/handlers/dashboard.go
package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"saintnet.com/m/internal/auth"
	"saintnet.com/m/internal/database"
	"saintnet.com/m/internal/models"
	"saintnet.com/m/internal/services"
)

func SelectConnection(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		ConnectionID int `json:"connection_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		respondWithError(w, http.StatusBadRequest, "Cuerpo de la solicitud inválido")
		return
	}

	session, _ := auth.Store.Get(r, "session-name")
	session.Values["connectionID"] = payload.ConnectionID
	if err := session.Save(r, w); err != nil {
		respondWithError(w, http.StatusInternalServerError, "No se pudo guardar la sesión")
		return
	}

	if userID, ok := session.Values["userID"].(int); ok {
		delete(activeClients, userID)
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"message": "Conexión seleccionada exitosamente"})
}

func GetDashboardData(w http.ResponseWriter, r *http.Request) {
	session, _ := auth.Store.Get(r, "session-name")
	connID, ok := session.Values["connectionID"].(int)
	if !ok {
		// Se usa una comprobación más flexible que también funciona para el ID 0
		connIDValue, connOK := session.Values["connectionID"]
		if !connOK || connIDValue == nil {
			respondWithError(w, http.StatusBadRequest, "ID de conexión no encontrado en la sesión")
			return
		}
		connID = connIDValue.(int)
	}

	queryParams := r.URL.Query()
	layout := "2006-01-02"
	endDate, err := time.Parse(layout, queryParams.Get("endDate"))
	if err != nil {
		endDate = time.Now()
	}
	startDate, err := time.Parse(layout, queryParams.Get("startDate"))
	if err != nil {
		startDate = endDate.AddDate(0, 0, -30)
	}
	endDate = endDate.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
	duration := endDate.Sub(startDate)
	previousEndDate := startDate.Add(-1 * time.Second)
	previousStartDate := previousEndDate.Add(-duration)

	var summary *models.ComparativeSummary

	if connID == 0 {
		connections, dbErr := database.GetConnections()
		if dbErr != nil {
			respondWithError(w, http.StatusInternalServerError, "Error al obtener conexiones para consolidar")
			return
		}
		summary, err = services.GetConsolidatedSummary(connections, startDate, endDate, previousStartDate, previousEndDate)
	} else {
		client := getClientFromContext(r)
		if client == nil {
			respondWithError(w, http.StatusUnauthorized, "No se ha seleccionado una conexión válida para esta operación")
			return
		}
		summary, err = services.GetComparativeSummary(client, startDate, endDate, previousStartDate, previousEndDate)
	}

	if err != nil {
		log.Printf("Error calculando el resumen: %v", err)
		if userID, ok := session.Values["userID"].(int); ok {
			delete(activeClients, userID)
		}
		respondWithError(w, http.StatusInternalServerError, "Error al calcular el resumen: "+err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, summary)
}
