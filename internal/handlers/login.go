package handlers

import (
	"log"
	"net/http"

	"golang.org/x/crypto/bcrypt"
	"saintnet.com/m/internal/auth"
	"saintnet.com/m/internal/database"
)

// LoginPage muestra la página de inicio de sesión.
func LoginPage(w http.ResponseWriter, r *http.Request) {
	if auth.IsLoggedIn(r) {
		http.Redirect(w, r, "/dashboard", http.StatusFound)
		return
	}
	templates.ExecuteTemplate(w, "login.html", nil)
}

// Login maneja el envío del formulario de inicio de sesión.
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

// Logout cierra la sesión del usuario.
func Logout(w http.ResponseWriter, r *http.Request) {
	session, _ := auth.Store.Get(r, "session-name")
	session.Values["authenticated"] = false
	session.Save(r, w)
	http.Redirect(w, r, "/login", http.StatusFound)
}

// ForcePasswordChangePage muestra la página de cambio de contraseña forzado.
func ForcePasswordChangePage(w http.ResponseWriter, r *http.Request) {
	templates.ExecuteTemplate(w, "force-password-change.html", nil)
}

// ForcePasswordChange maneja el cambio de contraseña.
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
