package handlers

import (
	"log"
	"net/http"

	"saintnet.com/m/internal/database"
)

func UsersPage(w http.ResponseWriter, r *http.Request) {
	users, err := database.GetUsers()
	if err != nil {
		log.Printf("Error obteniendo usuarios: %v", err)
		http.Error(w, "Error del servidor", http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"Users":      users,
		"ShowNavbar": true,
		"Template":   "users", // Indica a base.html qué contenido mostrar
	}

	if err := templates.ExecuteTemplate(w, "base.html", data); err != nil {
		log.Printf("Error al ejecutar plantilla de usuarios: %v", err)
	}
}

// La función AddUser no necesita cambios
func AddUser(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	username := r.FormValue("username")
	password := r.FormValue("password")

	if username == "" || password == "" {
		http.Error(w, "Usuario y clave son requeridos", http.StatusBadRequest)
		return
	}

	err := database.AddUser(username, password)
	if err != nil {
		log.Printf("Error agregando usuario: %v", err)
		http.Error(w, "Error al agregar el usuario", http.StatusInternalServerError)
		return
	}

	users, _ := database.GetUsers()
	templates.ExecuteTemplate(w, "user-list", map[string]interface{}{
		"Users": users,
	})
}
