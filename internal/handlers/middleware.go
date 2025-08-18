// internal/handlers/middleware.go
package handlers

import (
	"net/http"

	"saintnet.com/m/internal/auth"
)

// AuthMiddleware protege las rutas de la API que requieren autenticación.
// Ha sido adaptado para funcionar con un cliente API (como una SPA de Vue)
// en lugar de un navegador tradicional que maneja redirecciones.
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, err := auth.Store.Get(r, "session-name")
		if err != nil {
			// Si la cookie de sesión es inválida o no se puede decodificar, el acceso es denegado.
			respondWithError(w, http.StatusUnauthorized, "Sesión inválida o corrupta")
			return
		}

		isAuthenticated, ok := session.Values["authenticated"].(bool)

		// Si el usuario no está autenticado, se devuelve un error 401.
		if !ok || !isAuthenticated {
			// El frontend recibirá este error y sabrá que debe redirigir al usuario a la página de login.
			respondWithError(w, http.StatusUnauthorized, "Acceso no autorizado")
			return
		}

		// Sesión deslizante: se actualiza la cookie de sesión en cada solicitud válida
		// para mantener al usuario conectado mientras esté activo.
		if err := session.Save(r, w); err != nil {
			// Si no se puede guardar la sesión, es un error del servidor.
			respondWithError(w, http.StatusInternalServerError, "No se pudo guardar la sesión")
			return
		}

		// Si la autenticación es exitosa, se pasa la solicitud al siguiente manejador en la cadena.
		next.ServeHTTP(w, r)
	})
}
