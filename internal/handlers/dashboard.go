// internal/handlers/dashboard.go
package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"saintnet.com/m/internal/api"
	"saintnet.com/m/internal/auth"
	"saintnet.com/m/internal/database"
	"saintnet.com/m/internal/services"
)

// 'activeClients' sigue siendo útil como una caché en memoria para los clientes de la API de Saint,
// evitando inicios de sesión repetidos y mejorando el rendimiento.
var activeClients = make(map[int]*api.SaintClient)

// SelectConnection maneja la selección de una conexión de API.
// Ahora simplemente guarda la selección en la sesión y devuelve un JSON de éxito.
func SelectConnection(w http.ResponseWriter, r *http.Request) {
	// Se espera que el ID de conexión venga en el cuerpo de la solicitud JSON.
	var payload struct {
		ConnectionID int `json:"connection_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		respondWithError(w, http.StatusBadRequest, "Cuerpo de la solicitud inválido. Se esperaba 'connection_id'")
		return
	}

	if payload.ConnectionID == 0 {
		respondWithError(w, http.StatusBadRequest, "El 'connection_id' no puede ser cero")
		return
	}

	// Obtiene la sesión del usuario.
	session, _ := auth.Store.Get(r, "session-name")
	// Almacena el ID de la conexión seleccionada en la sesión del usuario.
	session.Values["connectionID"] = payload.ConnectionID
	// Guarda la sesión para persistir el cambio.
	if err := session.Save(r, w); err != nil {
		respondWithError(w, http.StatusInternalServerError, "No se pudo guardar la sesión")
		return
	}

	// Responde con un mensaje de éxito. El frontend ahora puede solicitar los datos del dashboard.
	respondWithJSON(w, http.StatusOK, map[string]string{"message": "Conexión seleccionada exitosamente"})
}

// GetDashboardData es el endpoint principal que obtiene y devuelve los datos del resumen.
func GetDashboardData(w http.ResponseWriter, r *http.Request) {
	session, _ := auth.Store.Get(r, "session-name")
	userID, ok := session.Values["userID"].(int)
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "Sesión de usuario inválida")
		return
	}

	// Intenta obtener el ID de la conexión desde la sesión.
	connID, ok := session.Values["connectionID"].(int)
	if !ok || connID == 0 {
		// Si no hay una conexión seleccionada en la sesión, se informa al cliente.
		respondWithError(w, http.StatusBadRequest, "No se ha seleccionado ninguna conexión")
		return
	}

	// Con el ID, obtiene los detalles completos de la conexión desde la base de datos.
	conn, err := database.GetConnectionByID(connID)
	if err != nil {
		log.Printf("Error obteniendo conexión: %v", err)
		respondWithError(w, http.StatusNotFound, "Conexión no encontrada")
		return
	}

	// Lógica de caché: comprueba si ya hay un cliente API para este usuario.
	client, exists := activeClients[userID]
	if !exists {
		// Si no existe, se crea uno nuevo y se inicia sesión en la API externa.
		client = api.NewSaintClient(conn.ApiURL)
		err := client.Login(conn.ApiUser, conn.ApiPassword, "B5D31933-C996-476C-B116-EF212A41479A", "1093")
		if err != nil {
			log.Printf("Error al conectar con la API: %v", err)
			respondWithError(w, http.StatusInternalServerError, "Error al conectar con la API: "+err.Error())
			return
		}
		// Si el login es exitoso, se guarda el nuevo cliente en la caché.
		activeClients[userID] = client
	}

	// Con un cliente API válido, se calculan los datos del resumen.
	summary, err := services.CalculateManagementSummary(client)
	if err != nil {
		log.Printf("Error calculando el resumen: %v", err)
		// Si el cálculo falla (ej. token expirado), se elimina el cliente de la caché
		// para forzar un nuevo login en la próxima solicitud.
		delete(activeClients, userID)
		respondWithError(w, http.StatusInternalServerError, "Error al calcular el resumen: "+err.Error())
		return
	}

	// Se envían los datos del resumen como una respuesta JSON exitosa.
	respondWithJSON(w, http.StatusOK, summary)
}

// NOTA: La función 'DashboardPage' ha sido eliminada. El frontend de Vue
// ahora se encarga de mostrar la página y esta solicitará los datos a 'GetDashboardData'.
