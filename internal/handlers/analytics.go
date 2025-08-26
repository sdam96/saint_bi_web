// internal/handlers/analytics.go
package handlers

import (
	"log"
	"net/http"
	"time"

	"saintnet.com/m/internal/auth"
	"saintnet.com/m/internal/database"
	"saintnet.com/m/internal/services"
)

func GetSalesForecastHandler(w http.ResponseWriter, r *http.Request) {
	// Obtenemos la sesión para leer el ID de la conexión seleccionada.
	session, _ := auth.Store.Get(r, "session-name")
	connID, _ := session.Values["connectionID"].(int)

	// La lógica de fechas no cambia.
	queryParams := r.URL.Query()
	layout := "2006-01-02"
	endDate, err := time.Parse(layout, queryParams.Get("endDate"))
	if err != nil {
		endDate = time.Now()
	}
	startDate, err := time.Parse(layout, queryParams.Get("startDate"))
	if err != nil {
		startDate = endDate.AddDate(0, 0, -90)
	}
	endDate = endDate.Add(23*time.Hour + 59*time.Minute)

	var forecast *services.SalesForecast

	// **LÓGICA DE DECISIÓN**
	if connID == 0 {
		// Si el ID es 0, usamos la vista consolidada.
		connections, dbErr := database.GetConnections()
		if dbErr != nil {
			respondWithError(w, http.StatusInternalServerError, "Error al obtener conexiones para consolidar")
			return
		}
		forecast, err = services.GetConsolidatedSalesForecast(connections, startDate, endDate)
	} else {
		// Para cualquier otro ID, usamos la lógica de conexión individual.
		client := getClientFromContext(r)
		if client == nil {
			respondWithError(w, http.StatusUnauthorized, "Cliente API no disponible para esta conexión")
			return
		}
		forecast, err = services.CalculateSalesForecast(client, startDate, endDate)
	}

	// El manejo de errores ahora funciona para ambas rutas.
	if err != nil {
		log.Printf("Error calculando la proyección de ventas: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Error al calcular la proyección: "+err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, forecast)
}

func GetMarketBasketHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := auth.Store.Get(r, "session-name")
	connID, _ := session.Values["connectionID"].(int)

	var results []services.MarketBasketResult
	var err error

	if connID == 0 {
		// Vista Consolidada
		connections, dbErr := database.GetConnections()
		if dbErr != nil {
			respondWithError(w, http.StatusInternalServerError, "Error al obtener conexiones")
			return
		}
		results, err = services.GetConsolidatedMarketBasket(connections)
	} else {
		// Vista Individual
		client := getClientFromContext(r)
		if client == nil {
			respondWithError(w, http.StatusUnauthorized, "Cliente API no disponible")
			return
		}
		results, err = services.CalculateMarketBasket(client)
	}

	if err != nil {
		log.Printf("Error calculando el análisis de canasta: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Error al realizar el análisis: "+err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, results)
}
