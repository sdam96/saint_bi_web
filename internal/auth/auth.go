// La declaración 'package auth' indica que este archivo pertenece al paquete 'auth'.
// Este paquete agrupa toda la lógica relacionada con la autenticación y gestión de sesiones
// de usuario, manteniendo el código organizado y modular.
package auth

import (
	// 'net/http' proporciona implementaciones de cliente y servidor HTTP.
	// Lo necesitamos aquí para trabajar con objetos de solicitud HTTP (http.Request).
	"net/http"

	// 'github.com/gorilla/sessions' es un paquete de terceros muy popular en el ecosistema de Go
	// para gestionar sesiones. Las sesiones permiten mantener el estado del usuario a través
	// de múltiples solicitudes HTTP, lo cual es esencial para saber si un usuario ha iniciado sesión.
	"github.com/gorilla/sessions"
)

// 'var Store *sessions.CookieStore' declara una variable a nivel de paquete llamada 'Store'.
// Será accesible por cualquier función dentro del paquete 'auth'.
// Es un puntero a un 'sessions.CookieStore', lo que significa que almacenará la referencia
// en memoria a nuestro almacén de cookies de sesión. 'CookieStore' guarda los datos de la sesión
// en cookies seguras firmadas en el lado del cliente.
var Store *sessions.CookieStore

// InitStore inicializa el almacén de sesiones. Esta función debe ser llamada una sola vez
// al iniciar la aplicación (por ejemplo, en la función main) para configurar el mecanismo de sesión.
func InitStore() {
	// IMPORTANTE: Cambia "super-secret-key" por una clave aleatoria y segura en producción.
	// 'sessions.NewCookieStore' crea una instancia de CookieStore. Requiere una clave de autenticación
	// como un slice de bytes ([]byte). Esta clave se usa para firmar (y opcionalmente cifrar) las cookies,
	// evitando que sean manipuladas por el cliente. La seguridad de las sesiones depende de que esta clave sea secreta.
	Store = sessions.NewCookieStore([]byte("super-secret-key"))

	// 'Store.Options' configura el comportamiento predeterminado para todas las cookies de sesión creadas.
	// Es un puntero a una struct de tipo 'sessions.Options'.
	Store.Options = &sessions.Options{
		// Path especifica la ruta de la URL para la cual la cookie es válida. "/" significa que
		// será válida para todo el sitio web.
		Path: "/",
		// MaxAge especifica el tiempo de vida de la cookie en segundos. 1200 segundos son 20 minutos.
		// Después de este tiempo, la cookie expirará en el navegador del usuario.
		MaxAge: 1200, // 20 minutos
		// HttpOnly es una bandera de seguridad importante. Si es 'true', impide que el JavaScript
		// del lado del cliente pueda acceder a la cookie, mitigando ataques de tipo Cross-Site Scripting (XSS).
		HttpOnly: true,
	}
}

// IsLoggedIn verifica si el usuario tiene una sesión activa y autenticada.
// Es una función auxiliar que se puede usar en los 'handlers' (manejadores de rutas) para proteger endpoints.
// Recibe un puntero a 'http.Request' (la solicitud entrante) del cual extraerá la cookie de sesión.
// Devuelve un booleano: 'true' si el usuario está autenticado, 'false' en caso contrario.
func IsLoggedIn(r *http.Request) bool {
	// 'Store.Get' intenta recuperar una sesión del almacén usando la solicitud 'r'.
	// Busca una cookie con el nombre "session-name". Si la cookie no existe, crea una nueva sesión vacía.
	session, err := Store.Get(r, "session-name")
	// Aunque Get puede crear una sesión, puede devolver un error si la cookie existe pero está corrupta
	// (p. ej., la firma no coincide), lo que podría indicar una manipulación.
	if err != nil {
		// Si hay un error al obtener la sesión, se asume que el usuario no está autenticado.
		return false
	}

	// 'session.Values' es un mapa (map[interface{}]interface{}) que contiene los datos de la sesión.
	// Intentamos obtener el valor asociado a la clave "authenticated".
	// La sintaxis 'variable, ok := valor.(tipo)' es una "aserción de tipo con comprobación".
	// Intenta convertir 'session.Values["authenticated"]' a un tipo 'bool'.
	// 'auth' contendrá el valor booleano si la conversión es exitosa.
	// 'ok' será 'true' si el valor existía y era del tipo 'bool', y 'false' en caso contrario.
	// Este es el método seguro y idiomático en Go para verificar tipos de datos de interfaces.
	auth, ok := session.Values["authenticated"].(bool)

	// La función devuelve 'true' solo si ambas condiciones se cumplen:
	// 1. 'ok' es 'true', lo que significa que el valor "authenticated" existe en la sesión y es un booleano.
	// 2. 'auth' es 'true', lo que significa que el valor de la bandera de autenticación es efectivamente verdadero.
	return ok && auth
}
