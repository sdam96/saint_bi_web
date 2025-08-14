package handlers

import (
	"log"
	"net/http"
	"strconv"

	"saintnet.com/m/internal/database"
	"saintnet.com/m/internal/models"
)

// ConnectionsPage muestra la página de gestión de conexiones.
func ConnectionsPage(w http.ResponseWriter, r *http.Request) {
	connections, err := database.GetConnections()
	if err != nil {
		log.Printf("Error obteniendo conexiones: %v", err)
		http.Error(w, "Error del servidor", http.StatusInternalServerError)
		return
	}
	templates.ExecuteTemplate(w, "connections.html", map[string]interface{}{
		"Connections": connections,
	})
}

// AddConnection agrega una nueva conexión.
func AddConnection(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	refresh, _ := strconv.Atoi(r.FormValue("refresh_seconds"))
	configID, _ := strconv.Atoi(r.FormValue("config_id"))

	conn := models.Connection{
		Alias:          r.FormValue("alias"),
		ApiURL:         r.FormValue("api_url"),
		ApiUser:        r.FormValue("api_user"),
		ApiPassword:    r.FormValue("api_password"),
		RefreshSeconds: refresh,
		ConfigID:       configID,
	}

	err := database.AddConnection(conn)
	if err != nil {
		log.Printf("Error agregando conexión: %v", err)
		// Aquí podríamos devolver un error a HTMX
	}

	// Redirigir (o devolver fragmento HTMX) a la lista actualizada
	connections, _ := database.GetConnections()
	templates.ExecuteTemplate(w, "connections.html", map[string]interface{}{
		"Connections": connections,
	})
}

// DeleteConnection elimina una conexión.
func DeleteConnection(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	if err := database.DeleteConnection(id); err != nil {
		log.Printf("Error eliminando conexión: %v", err)
		// Manejar el error apropiadamente para HTMX
	}

	// Devolver una respuesta vacía, HTMX eliminará el elemento del DOM.
	w.WriteHeader(http.StatusOK)
}
