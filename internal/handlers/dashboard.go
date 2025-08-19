// internal/handlers/dashboard.go
package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"saintnet.com/m/internal/auth"
	"saintnet.com/m/internal/services"
)

// SelectConnection maneja la selección de la conexión.
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

	// Al cambiar de conexión, es crucial eliminar el cliente antiguo de la caché
	// para forzar una nueva autenticación con las credenciales correctas.
	if userID, ok := session.Values["userID"].(int); ok {
		delete(activeClients, userID)
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"message": "Conexión seleccionada exitosamente"})
}

// GetDashboardData utiliza el cliente API preparado por el middleware.
func GetDashboardData(w http.ResponseWriter, r *http.Request) {
	client := getClientFromContext(r)
	if client == nil {
		respondWithError(w, http.StatusUnauthorized, "No se ha seleccionado una conexión válida para esta operación")
		return
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

	summary, err := services.GetComparativeSummary(client, startDate, endDate, previousStartDate, previousEndDate)
	if err != nil {
		log.Printf("Error calculando el resumen comparativo: %v", err)

		session, _ := auth.Store.Get(r, "session-name")
		if userID, ok := session.Values["userID"].(int); ok {
			delete(activeClients, userID)
		}

		respondWithError(w, http.StatusInternalServerError, "Error al calcular el resumen: "+err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, summary)
}
