// internal/handlers/dashboard.go
package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"time" // Se importa el paquete 'time' para manejar las fechas dinámicas.

	"saintnet.com/m/internal/api"
	"saintnet.com/m/internal/auth"
	"saintnet.com/m/internal/database"
	"saintnet.com/m/internal/services"
)

// 'activeClients' sigue siendo útil como una caché en memoria para los clientes de la API de Saint.
// La clave del mapa es el 'userID' y el valor es el cliente de la API ya autenticado.
// Esto evita tener que hacer login contra la API externa en cada solicitud, mejorando drásticamente el rendimiento.
var activeClients = make(map[int]*api.SaintClient)

// SelectConnection maneja la selección de una conexión de API por parte del usuario.
// Su única responsabilidad es guardar el ID de la conexión elegida en la sesión del usuario.
// El frontend hará una solicitud separada a GetDashboardData para obtener la información.
func SelectConnection(w http.ResponseWriter, r *http.Request) {
	// Se define una struct anónima para decodificar el cuerpo de la solicitud JSON.
	// Esto asegura que solo se lea el campo 'connection_id'.
	var payload struct {
		ConnectionID int `json:"connection_id"`
	}

	// Se decodifica el JSON del cuerpo de la solicitud en la struct 'payload'.
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		respondWithError(w, http.StatusBadRequest, "Cuerpo de la solicitud inválido. Se esperaba 'connection_id'")
		return
	}

	// Validación simple para asegurar que se envió un ID válido.
	if payload.ConnectionID == 0 {
		respondWithError(w, http.StatusBadRequest, "El 'connection_id' no puede ser cero")
		return
	}

	// Se obtiene la sesión actual del usuario a partir de la cookie en la solicitud.
	session, _ := auth.Store.Get(r, "session-name")
	// Se almacena el ID de la conexión seleccionada en el mapa de valores de la sesión.
	// Este valor persistirá entre solicitudes para este usuario.
	session.Values["connectionID"] = payload.ConnectionID

	// Se guarda la sesión. Esto envía la cookie actualizada al navegador del cliente.
	if err := session.Save(r, w); err != nil {
		respondWithError(w, http.StatusInternalServerError, "No se pudo guardar la sesión")
		return
	}

	// Se responde con un mensaje de éxito. El frontend usará esta respuesta para saber
	// que puede proceder a solicitar los datos del dashboard.
	respondWithJSON(w, http.StatusOK, map[string]string{"message": "Conexión seleccionada exitosamente"})
}

// GetDashboardData es el endpoint principal que obtiene, procesa y devuelve los datos del resumen comparativo.
func GetDashboardData(w http.ResponseWriter, r *http.Request) {
	// --- 1. Validación de Sesión y Conexión ---
	session, _ := auth.Store.Get(r, "session-name")
	userID, ok := session.Values["userID"].(int)
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "Sesión de usuario inválida")
		return
	}

	// Se recupera el ID de la conexión que el usuario seleccionó previamente y guardó en su sesión.
	connID, ok := session.Values["connectionID"].(int)
	if !ok || connID == 0 {
		respondWithError(w, http.StatusBadRequest, "No se ha seleccionado ninguna conexión")
		return
	}

	// --- 2. Lógica de Fechas para Comparación ---
	// 'r.URL.Query()' obtiene un mapa de todos los parámetros de la URL (ej. ?startDate=2023-01-01).
	queryParams := r.URL.Query()
	// El layout "2006-01-02" es la forma estándar en Go para indicarle a `time.Parse` cómo interpretar un string de fecha.
	layout := "2006-01-02"

	// Se parsea la fecha de fin del período actual. Si no se proporciona o hay un error, se usa la fecha actual.
	endDate, err := time.Parse(layout, queryParams.Get("endDate"))
	if err != nil {
		endDate = time.Now()
	}
	// Se parsea la fecha de inicio. Si no se proporciona, se calcula restando 30 días a la fecha de fin.
	startDate, err := time.Parse(layout, queryParams.Get("startDate"))
	if err != nil {
		startDate = endDate.AddDate(0, 0, -30)
	}
	// Se ajusta la hora de 'endDate' al final del día (23:59:59) para asegurar que se incluyan
	// todas las transacciones de ese día en el rango del filtro.
	endDate = endDate.Add(23*time.Hour + 59*time.Minute + 59*time.Second)

	// Se calcula la duración del período actual (ej. 30 días).
	duration := endDate.Sub(startDate)
	// La fecha de fin del período anterior es un segundo antes de que comience el período actual.
	previousEndDate := startDate.Add(-1 * time.Second)
	// La fecha de inicio del período anterior se calcula restando la misma duración a su fecha de fin.
	previousStartDate := previousEndDate.Add(-duration)

	// --- 3. Obtención y Caché del Cliente API ---
	// Se obtienen los detalles de la conexión (URL, credenciales) desde nuestra base de datos local.
	conn, err := database.GetConnectionByID(connID)
	if err != nil {
		log.Printf("Error obteniendo conexión: %v", err)
		respondWithError(w, http.StatusNotFound, "Conexión no encontrada")
		return
	}

	// Se comprueba si ya existe un cliente API autenticado en nuestra caché en memoria.
	client, exists := activeClients[userID]
	if !exists {
		// Si no existe, se crea uno nuevo.
		client = api.NewSaintClient(conn.ApiURL)
		// Se realiza el login contra la API externa de Saint.
		err := client.Login(conn.ApiUser, conn.ApiPassword, "B5D31933-C996-476C-B116-EF212A41479A", "1093")
		if err != nil {
			log.Printf("Error al conectar con la API: %v", err)
			respondWithError(w, http.StatusInternalServerError, "Error al conectar con la API: "+err.Error())
			return
		}
		// Si el login es exitoso, se guarda el cliente en la caché para futuras solicitudes.
		activeClients[userID] = client
	}

	// --- 4. Cálculo y Respuesta ---
	// Se llama al servicio que orquesta todo el cálculo, pasando el cliente API y los rangos de fechas.
	summary, err := services.GetComparativeSummary(client, startDate, endDate, previousStartDate, previousEndDate)
	if err != nil {
		log.Printf("Error calculando el resumen comparativo: %v", err)
		// Si el cálculo falla (ej. el token de la API externa expiró),
		// se elimina el cliente de la caché para forzar un nuevo login en la próxima solicitud.
		delete(activeClients, userID)
		respondWithError(w, http.StatusInternalServerError, "Error al calcular el resumen: "+err.Error())
		return
	}

	// Si todo es exitoso, se envían los datos del resumen comparativo como una respuesta JSON.
	respondWithJSON(w, http.StatusOK, summary)
}
