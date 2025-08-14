package handlers

import (
	"log"
	"net/http"

	"saintnet.com/m/internal/database"
)

// UsersPage muestra la página de gestión de usuarios.
func UsersPage(w http.ResponseWriter, r *http.Request) {
	users, err := database.GetUsers()
	if err != nil {
		log.Printf("Error obteniendo usuarios: %v", err)
		http.Error(w, "Error del servidor", http.StatusInternalServerError)
		return
	}
	templates.ExecuteTemplate(w, "users.html", map[string]interface{}{
		"Users": users,
	})
}

// AddUser agrega un nuevo usuario de la aplicación.
func AddUser(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	username := r.FormValue("username")
	password := r.FormValue("password")

	if username == "" || password == "" {
		// Manejar error
	}

	err := database.AddUser(username, password)
	if err != nil {
		log.Printf("Error agregando usuario: %v", err)
		// Manejar error
	}

	users, _ := database.GetUsers()
	templates.ExecuteTemplate(w, "users.html", map[string]interface{}{
		"Users": users,
	})
}
