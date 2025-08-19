// internal/handlers/middleware.go
package handlers

import (
	// "context" es un paquete estándar de Go que permite pasar datos a través de la cadena de handlers
	// para una solicitud específica. Es la forma idiomática de compartir información como el cliente API.
	"context"
	"net/http"

	"saintnet.com/m/internal/api"
	"saintnet.com/m/internal/auth"
	"saintnet.com/m/internal/database"
)

// contextKey es un tipo local que usaremos como clave para almacenar valores en el contexto de la solicitud.
// Usar un tipo local en lugar de un string previene colisiones con claves de otros paquetes.
// Sintaxis: type <NombreTipo> <TipoSubyacente>
type contextKey string

// clientContextKey es la clave específica que usaremos para guardar y recuperar el *api.SaintClient.
// Al ser una constante, evitamos errores de tipeo al usarla.
const clientContextKey = contextKey("apiClient")

// AuthMiddleware es una función que recibe un 'http.Handler' (el siguiente manejador en la cadena)
// y devuelve un nuevo 'http.Handler'. Este patrón es la base del middleware en Go.
// Su función es interceptar una solicitud, realizar acciones (como la autenticación) y luego,
// si todo es correcto, pasar la solicitud al siguiente manejador.
func AuthMiddleware(next http.Handler) http.Handler {
	// 'http.HandlerFunc' es un adaptador que nos permite usar una función ordinaria
	// (en este caso, una función anónima) como si fuera un 'http.Handler'.
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// --- 1. Verificación de Sesión ---
		session, err := auth.Store.Get(r, "session-name")
		if err != nil {
			respondWithError(w, http.StatusUnauthorized, "Sesión inválida o corrupta")
			return // 'return' detiene la ejecución del handler, la solicitud no continúa.
		}

		isAuthenticated, ok := session.Values["authenticated"].(bool)
		if !ok || !isAuthenticated {
			respondWithError(w, http.StatusUnauthorized, "Acceso no autorizado")
			return
		}

		// --- 2. Inyección del Cliente API en el Contexto ---
		userID, ok := session.Values["userID"].(int)
		if !ok {
			respondWithError(w, http.StatusUnauthorized, "Sesión de usuario inválida")
			return
		}

		connID, ok := session.Values["connectionID"].(int)
		if !ok || connID == 0 {
			// Si no hay una conexión seleccionada en la sesión, no podemos crear el cliente API.
			// Pasamos la solicitud al siguiente handler sin inyectar nada.
			// Esto permite que rutas como '/api/connections' (que no necesitan un cliente) funcionen.
			next.ServeHTTP(w, r)
			return
		}

		// --- 3. Lógica de Caché ---
		// Buscamos un cliente API ya autenticado en nuestra caché 'activeClients' (definida en handlers.go).
		client, exists := activeClients[userID]
		if !exists {
			// Si no existe en la caché, lo creamos.
			conn, err := database.GetConnectionByID(connID)
			if err != nil {
				respondWithError(w, http.StatusInternalServerError, "No se pudo recuperar la conexión de la sesión")
				return
			}
			client = api.NewSaintClient(conn.ApiURL)
			if err := client.Login(conn.ApiUser, conn.ApiPassword, "B5D31933-C996-476C-B116-EF212A41479A", "1093"); err != nil {
				respondWithError(w, http.StatusInternalServerError, "Error al conectar con la API externa: "+err.Error())
				return
			}
			// Una vez creado y autenticado, lo guardamos en la caché para futuras solicitudes.
			activeClients[userID] = client
		}

		// --- 4. Inyección en el Contexto ---
		// 'context.WithValue' crea un nuevo contexto a partir del contexto existente de la solicitud (r.Context())
		// y le añade nuestro cliente API bajo la clave que definimos.
		ctx := context.WithValue(r.Context(), clientContextKey, client)

		// 'next.ServeHTTP' pasa el control al siguiente handler en la cadena.
		// 'r.WithContext(ctx)' es crucial: reemplaza el contexto de la solicitud original
		// con nuestro nuevo contexto que ahora contiene el cliente API.
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// getClientFromContext es una función auxiliar para ser usada por otros handlers.
// Su propósito es extraer de forma segura y tipada el cliente del contexto de la solicitud.
//
// Parámetros:
//   - r (*http.Request): La solicitud HTTP de la cual se extraerá el contexto.
//
// Retorna:
//   - *api.SaintClient: Un puntero al cliente de la API si se encuentra, o 'nil' si no.
func getClientFromContext(r *http.Request) *api.SaintClient {
	// 'r.Context().Value()' recupera un valor del contexto usando la clave proporcionada.
	// Devuelve un 'interface{}', por lo que necesitamos convertirlo de vuelta a nuestro tipo.
	//
	// La sintaxis 'variable, ok := valor.(Tipo)' es una "aserción de tipo con comprobación".
	// Intenta convertir el valor a '*api.SaintClient'. Si la conversión es exitosa,
	// 'ok' será 'true' y 'client' tendrá el valor. Si no, 'ok' será 'false' y 'client' será 'nil'.
	// Este es el método seguro y idiomático en Go para trabajar con interfaces.
	client, ok := r.Context().Value(clientContextKey).(*api.SaintClient)
	if !ok {
		// Si 'ok' es falso, significa que el cliente no estaba en el contexto o no era del tipo correcto.
		return nil
	}
	return client
}
