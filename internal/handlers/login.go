package handlers

import (
	"log"
	"net/http"

	"golang.org/x/crypto/bcrypt"
	"saintnet.com/m/internal/auth"
	"saintnet.com/m/internal/database"
)

func LoginPage(w http.ResponseWriter, r *http.Request) {
	if auth.IsLoggedIn(r) {
		http.Redirect(w, r, "/dashboard", http.StatusFound)
		return
	}

	data := map[string]interface{}{
		"ShowNavbar": false,
		"Template":   "login", // Indica a base.html qué contenido mostrar
	}

	if err := templates.ExecuteTemplate(w, "base.html", data); err != nil {
		log.Printf("Error al ejecutar plantilla de login: %v", err)
	}
}

func ForcePasswordChangePage(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{
		"ShowNavbar": false,
		"Template":   "force-password-change", // Indica a base.html qué contenido mostrar
	}

	if err := templates.ExecuteTemplate(w, "base.html", data); err != nil {
		log.Printf("Error al ejecutar plantilla de cambio de clave: %v", err)
	}
}

// El resto del archivo (Login, Logout, ForcePasswordChange) no necesita cambios.
func Login(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	username := r.FormValue("username")
	password := r.FormValue("password")

	user, err := database.GetUserByUsername(username)
	if err != nil {
		log.Printf("Error al buscar usuario: %v", err)
		http.Redirect(w, r, "/login?error=credenciales", http.StatusFound)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		http.Redirect(w, r, "/login?error=credenciales", http.StatusFound)
		return
	}

	session, _ := auth.Store.Get(r, "session-name")
	session.Values["authenticated"] = true
	session.Values["userID"] = user.ID
	session.Values["username"] = user.Username
	session.Save(r, w)

	if user.FirstLogin {
		http.Redirect(w, r, "/force-password-change", http.StatusFound)
	} else {
		http.Redirect(w, r, "/dashboard", http.StatusFound)
	}
}

func Logout(w http.ResponseWriter, r *http.Request) {
	session, _ := auth.Store.Get(r, "session-name")
	session.Values["authenticated"] = false
	delete(session.Values, "userID")
	delete(session.Values, "username")
	delete(session.Values, "connectionID")
	session.Options.MaxAge = -1 // Borrar la cookie
	session.Save(r, w)
	http.Redirect(w, r, "/login", http.StatusFound)
}

func ForcePasswordChange(w http.ResponseWriter, r *http.Request) {
	session, _ := auth.Store.Get(r, "session-name")
	userID, ok := session.Values["userID"].(int)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	r.ParseForm()
	newPassword := r.FormValue("new_password")
	confirmPassword := r.FormValue("confirm_password")

	if newPassword == "" || newPassword != confirmPassword {
		http.Redirect(w, r, "/force-password-change?error=match", http.StatusFound)
		return
	}

	if newPassword == "admin" {
		http.Redirect(w, r, "/force-password-change?error=default", http.StatusFound)
		return
	}

	err := database.UpdateUserPassword(userID, newPassword)
	if err != nil {
		log.Printf("Error al actualizar la contraseña: %v", err)
		http.Redirect(w, r, "/force-password-change?error=server", http.StatusFound)
		return
	}

	http.Redirect(w, r, "/dashboard", http.StatusFound)
}
