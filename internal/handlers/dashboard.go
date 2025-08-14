package handlers

import (
	"log"
	"net/http"
	"strconv"

	"saintnet.com/m/internal/api"
	"saintnet.com/m/internal/auth"
	"saintnet.com/m/internal/database"
	"saintnet.com/m/internal/services"
)

var activeClients = make(map[int]*api.SaintClient) // Caché de clientes de API por ID de sesión

// DashboardPage muestra la página principal del dashboard.
func DashboardPage(w http.ResponseWriter, r *http.Request) {
	connections, err := database.GetConnections()
	if err != nil {
		http.Error(w, "Error al cargar conexiones", http.StatusInternalServerError)
		return
	}

	session, _ := auth.Store.Get(r, "session-name")
	userID := session.Values["userID"].(int)

	data := map[string]interface{}{
		"Connections": connections,
		"Summary":     nil, // Inicialmente sin datos
	}

	// Si ya hay una conexión seleccionada en la sesión, la usamos
	if connID, ok := session.Values["connectionID"].(int); ok {
		data["SelectedConnectionID"] = connID
		client, clientExists := activeClients[userID]
		if clientExists {
			summary, err := services.CalculateManagementSummary(client)
			if err != nil {
				data["Error"] = err.Error()
			} else {
				data["Summary"] = summary
			}
		}
	}

	templates.ExecuteTemplate(w, "dashboard.html", data)
}

// GetConnectionsDropdown devuelve el fragmento HTML del selector de conexiones.
func GetConnectionsDropdown(w http.ResponseWriter, r *http.Request) {
	connections, _ := database.GetConnections()
	session, _ := auth.Store.Get(r, "session-name")
	connID, _ := session.Values["connectionID"].(int)

	data := map[string]interface{}{
		"Connections":          connections,
		"SelectedConnectionID": connID,
	}
	templates.ExecuteTemplate(w, "connections-dropdown", data)
}

// SelectConnection maneja la selección de una nueva conexión de API.
func SelectConnection(w http.ResponseWriter, r *http.Request) {
	connID, err := strconv.Atoi(r.FormValue("connection_id"))
	if err != nil {
		http.Error(w, "ID de conexión inválido", http.StatusBadRequest)
		return
	}

	session, _ := auth.Store.Get(r, "session-name")
	session.Values["connectionID"] = connID
	session.Save(r, w)

	// Actualizar el contenido del dashboard
	GetDashboardData(w, r)
}

// GetDashboardData obtiene los datos del resumen y devuelve el fragmento HTML.
func GetDashboardData(w http.ResponseWriter, r *http.Request) {
	session, _ := auth.Store.Get(r, "session-name")
	userID := session.Values["userID"].(int)
	connID, ok := session.Values["connectionID"].(int)
	if !ok {
		// No hay conexión seleccionada, no devolvemos nada o un mensaje
		w.Write([]byte("<div>Seleccione una conexión para ver los datos.</div>"))
		return
	}

	conn, err := database.GetConnectionByID(connID)
	if err != nil {
		log.Printf("Error: %v", err)
		http.Error(w, "Conexión no encontrada", http.StatusNotFound)
		return
	}

	// Usar cliente cacheado o crear uno nuevo
	client, exists := activeClients[userID]
	if !exists {
		client = api.NewSaintClient(conn.ApiURL)
		// El API ID y API Key no están en el modelo de conexión, los hardcodeamos por ahora.
		// En una app real, estos deberían venir de la configuración.
		err := client.Login(conn.ApiUser, conn.ApiPassword, "B5D31933-C996-476C-B116-EF212A41479A", "1093")
		if err != nil {
			log.Printf("Error al conectar con la API: %v", err)
			http.Error(w, "Error al conectar con la API: "+err.Error(), http.StatusInternalServerError)
			return
		}
		activeClients[userID] = client
	}

	summary, err := services.CalculateManagementSummary(client)
	data := map[string]interface{}{
		"Summary": summary,
		"Error":   nil,
		"Refresh": conn.RefreshSeconds,
	}
	if err != nil {
		log.Printf("Error calculando el resumen: %v", err)
		data["Error"] = err.Error()
		// Si hay error de API, puede ser que el token expiró. Limpiamos el cliente.
		delete(activeClients, userID)
	}

	templates.ExecuteTemplate(w, "summary-data", data)
}
