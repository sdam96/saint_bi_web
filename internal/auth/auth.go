package auth

import (
	"net/http"

	"github.com/gorilla/sessions"
)

var Store *sessions.CookieStore

func InitStore() {
	// IMPORTANTE: Cambia "super-secret-key" por una clave aleatoria y segura en producción.
	Store = sessions.NewCookieStore([]byte("super-secret-key"))
	Store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   1200, // 20 minutos
		HttpOnly: true,
	}
}

// IsLoggedIn verifica si el usuario tiene una sesión activa.
func IsLoggedIn(r *http.Request) bool {
	session, err := Store.Get(r, "session-name")
	if err != nil {
		return false
	}
	auth, ok := session.Values["authenticated"].(bool)
	return ok && auth
}
