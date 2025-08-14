package handlers

import (
	"net/http"

	"saintnet.com/m/internal/auth"
)

// AuthMiddleware protege las rutas que requieren autenticación.
// Ahora también maneja la redirección para HTMX y el reseteo del tiempo de sesión.
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, err := auth.Store.Get(r, "session-name")
		if err != nil {
			// Si hay error al obtener la sesión (ej. cookie inválida), redirigir al login.
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		// Verificar si el usuario está autenticado en la sesión.
		isAuthenticated, ok := session.Values["authenticated"].(bool)
		if !ok || !isAuthenticated {
			// **SOLUCIÓN 1: Redirección completa para HTMX**
			// Si la petición viene de HTMX (header "HX-Request" existe),
			// enviamos un header especial para que el frontend redirija.
			if r.Header.Get("HX-Request") == "true" {
				w.Header().Set("HX-Redirect", "/login")
				w.WriteHeader(http.StatusOK) // HTMX necesita un 200 para procesar el header
				return
			}
			// Para peticiones normales, usamos la redirección estándar.
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		// **SOLUCIÓN 2: Extender la sesión en cada petición (20 min de inactividad)**
		// Al guardar la sesión en cada solicitud válida, el MaxAge se "reinicia".
		if err := session.Save(r, w); err != nil {
			// Manejar el error si no se puede guardar la sesión.
			http.Error(w, "Error al guardar la sesión", http.StatusInternalServerError)
			return
		}

		// Si todo está bien, continuar con la siguiente función del handler.
		next.ServeHTTP(w, r)
	})
}
