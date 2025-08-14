package handlers

import (
	"log"
	"net/http"
	"strconv"

	"saintnet.com/m/internal/database"
	"saintnet.com/m/internal/models"
)

func ConnectionsPage(w http.ResponseWriter, r *http.Request) {
	connections, err := database.GetConnections()
	if err != nil {
		log.Printf("Error obteniendo conexiones: %v", err)
		http.Error(w, "Error del servidor", http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"Connections": connections,
		"ShowNavbar":  true,
		"Template":    "connections", // Indica a base.html qué contenido mostrar
	}

	if err := templates.ExecuteTemplate(w, "base.html", data); err != nil {
		log.Printf("Error al ejecutar plantilla de conexiones: %v", err)
	}
}

// El resto del archivo (AddConnection, DeleteConnection) no necesita cambios.
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

	if err := database.AddConnection(conn); err != nil {
		log.Printf("Error agregando conexión: %v", err)
		http.Error(w, "Error al agregar conexión", http.StatusInternalServerError)
		return
	}

	connections, _ := database.GetConnections()
	templates.ExecuteTemplate(w, "connection-list", map[string]interface{}{
		"Connections": connections,
	})
}

func DeleteConnection(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	if err := database.DeleteConnection(id); err != nil {
		log.Printf("Error eliminando conexión: %v", err)
		http.Error(w, "Error al eliminar conexión", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
