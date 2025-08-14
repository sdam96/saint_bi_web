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

// activeClients funciona como una caché de clientes de API por ID de usuario.
var activeClients = make(map[int]*api.SaintClient)

// DashboardPage muestra la página principal del dashboard.
func DashboardPage(w http.ResponseWriter, r *http.Request) {
	connections, err := database.GetConnections()
	if err != nil {
		http.Error(w, "Error al cargar conexiones", http.StatusInternalServerError)
		return
	}

	session, _ := auth.Store.Get(r, "session-name")
	userID := session.Values["userID"].(int)

	// Se combinan los datos para la plantilla.
	data := map[string]interface{}{
		"Connections": connections,
		"Summary":     nil,
		"ShowNavbar":  true,
		"Template":    "dashboard", // Se indica a base.html que renderice el contenido del dashboard.
	}

	// **LÓGICA RESTAURADA**: Si ya hay una conexión seleccionada, se intenta precargar los datos.
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

	// **MÉTODO CORRECTO**: Se ejecuta la plantilla base, que sabe qué contenido mostrar.
	if err := templates.ExecuteTemplate(w, "base.html", data); err != nil {
		log.Printf("Error al ejecutar plantilla de dashboard: %v", err)
	}
}

// SelectConnection maneja la selección de una nueva conexión de API.
// (Esta función es correcta y no necesita cambios).
func SelectConnection(w http.ResponseWriter, r *http.Request) {
	connID, err := strconv.Atoi(r.FormValue("connection_id"))
	if err != nil {
		http.Error(w, "ID de conexión inválido", http.StatusBadRequest)
		return
	}

	session, _ := auth.Store.Get(r, "session-name")
	session.Values["connectionID"] = connID
	session.Save(r, w)

	GetDashboardData(w, r)
}

// GetDashboardData obtiene los datos del resumen para HTMX.
// (Esta función es correcta y no necesita cambios).
func GetDashboardData(w http.ResponseWriter, r *http.Request) {
	session, _ := auth.Store.Get(r, "session-name")
	userID, ok := session.Values["userID"].(int)
	if !ok {
		http.Error(w, "Sesión de usuario inválida", http.StatusUnauthorized)
		return
	}

	connID, ok := session.Values["connectionID"].(int)
	if !ok {
		connID, _ = strconv.Atoi(r.URL.Query().Get("connection_id"))
		if connID == 0 {
			w.Write([]byte("<div>Seleccione una conexión para ver los datos.</div>"))
			return
		}
	}

	conn, err := database.GetConnectionByID(connID)
	if err != nil {
		log.Printf("Error obteniendo conexión: %v", err)
		http.Error(w, "Conexión no encontrada", http.StatusNotFound)
		return
	}

	client, exists := activeClients[userID]
	if !exists {
		client = api.NewSaintClient(conn.ApiURL)
		err := client.Login(conn.ApiUser, conn.ApiPassword, "B5D31933-C996-476C-B116-EF212A41479A", "1093")
		if err != nil {
			log.Printf("Error al conectar con la API: %v", err)
			templates.ExecuteTemplate(w, "summary-data", map[string]interface{}{
				"Error": "Error al conectar con la API: " + err.Error(),
			})
			return
		}
		activeClients[userID] = client
	}

	summary, err := services.CalculateManagementSummary(client)
	data := map[string]interface{}{
		"Summary":              summary,
		"Error":                nil,
		"Refresh":              conn.RefreshSeconds,
		"SelectedConnectionID": connID,
	}
	if err != nil {
		log.Printf("Error calculando el resumen: %v", err)
		data["Error"] = err.Error()
		delete(activeClients, userID)
	}

	templates.ExecuteTemplate(w, "summary-data", data)
}
