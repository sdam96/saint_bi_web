// internal/handlers/logging.go
package handlers

import (
	// "log" implementa un paquete de logging simple, que usaremos para imprimir
	// información estandarizada de cada solicitud en la terminal.
	"log"
	// "net/http" proporciona las herramientas para construir el servidor web, incluyendo
	// los tipos http.Handler, http.ResponseWriter y *http.Request.
	"net/http"
	// "time" nos permite medir el tiempo para calcular la latencia de cada solicitud.
	"time"
)

// LoggingMiddleware es una función que implementa el patrón de middleware en Go.
// Un middleware es una pieza de código que se ejecuta ANTES y/o DESPUÉS del manejador principal de una ruta.
// En este caso, su propósito es interceptar cada solicitud HTTP, registrar sus detalles y medir su rendimiento.
//
// SINTAXIS de Go:
//   - `func LoggingMiddleware(next http.Handler) http.Handler`: Define una función que acepta un argumento `next`
//     de tipo `http.Handler` y devuelve un valor también de tipo `http.Handler`. Este es el patrón estándar
//     para encadenar middlewares. `next` representa el siguiente manejador en la cadena.
func LoggingMiddleware(next http.Handler) http.Handler {
	// `http.HandlerFunc(...)` es un adaptador que nos permite usar una función anónima
	// (una función sin nombre) como si fuera un `http.Handler`.
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// --- CÓDIGO QUE SE EJECUTA ANTES DEL HANDLER PRINCIPAL ---

		// SINTAXIS de Go: `start := time.Now()` usa el operador de declaración corta para declarar
		// e inicializar la variable `start` con la hora actual.
		start := time.Now()

		// LÓGICA: Para capturar el código de estado HTTP (ej. 200, 404, 500), que se escribe
		// en el `ResponseWriter` original, necesitamos "envolverlo". Creamos una instancia de
		// nuestro `loggingResponseWriter` personalizado, pasándole el `ResponseWriter` original.
		lrw := &loggingResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		// `next.ServeHTTP(lrw, r)` es el punto crucial. Aquí, pasamos el control al siguiente
		// manejador en la cadena (que podría ser otro middleware o el handler final de la ruta).
		// Es importante notar que le pasamos nuestro `ResponseWriter` envuelto (`lrw`).
		next.ServeHTTP(lrw, r)

		// --- CÓDIGO QUE SE EJECUTA DESPUÉS DEL HANDLER PRINCIPAL ---

		// LÓGICA: Una vez que el handler principal ha terminado y la respuesta ha sido enviada,
		// el control vuelve a este punto.
		// SINTAXIS de Go: `time.Since(start)` calcula el tiempo transcurrido desde la variable `start`.
		latency := time.Since(start)

		// LÓGICA: Se formatea e imprime el log final en la terminal.
		// SINTAXIS de Go: `log.Printf` imprime un string formateado. Los verbos (`%s`, `%d`, `%v`) son
		// reemplazados por los argumentos correspondientes.
		// - `start.Format("2006-01-02 15:04:05")`: Go usa esta fecha de referencia específica para definir
		//   el formato deseado, en lugar de símbolos como "YYYY-MM-DD".
		// - `%d`: Formatea un entero (código de estado).
		// - `%s`: Formatea un string.
		// - `%12v`: Formatea un valor (la latencia) alineado a la derecha en un espacio de 12 caracteres.
		// - `%15s`: Formatea un string (la IP) alineado a la derecha en un espacio de 15 caracteres.
		log.Printf("[HTTP] %s | %d %s | %12v | %15s | %s %s",
			start.Format("2006-01-02 15:04:05"),
			lrw.statusCode,
			http.StatusText(lrw.statusCode), // Convierte un código (ej. 200) a su texto ("OK").
			latency,
			r.RemoteAddr, // La dirección IP del cliente.
			r.Method,     // El método HTTP (GET, POST, etc.).
			r.URL.Path,   // La ruta solicitada (ej. /api/dashboard/data).
		)
	})
}

// loggingResponseWriter es una `struct` que "envuelve" al `http.ResponseWriter` original.
// LÓGICA: Su propósito es interceptar la llamada al método `WriteHeader` para poder
// guardar el código de estado HTTP antes de que se envíe al cliente.
//
// SINTAXIS de Go:
//   - `type loggingResponseWriter struct { ... }` define nuestra nueva struct.
//   - `http.ResponseWriter`: Esto es "embedding" (incrustación). Al incluir un tipo sin un nombre de campo,
//     nuestra struct `loggingResponseWriter` hereda automáticamente todos los métodos de `http.ResponseWriter`.
//     Esto nos permite pasarla a `next.ServeHTTP` porque satisface la misma interfaz.
//   - `statusCode int`: Este es el campo adicional que usamos para almacenar el código de estado.
type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader es un método que sobreescribe el `WriteHeader` del `http.ResponseWriter` incrustado.
//
// SINTAXIS de Go:
//   - `func (lrw *loggingResponseWriter) WriteHeader(code int)` define un método llamado `WriteHeader`
//     que opera sobre un puntero a `loggingResponseWriter` (el "receptor" `lrw`).
func (lrw *loggingResponseWriter) WriteHeader(code int) {
	// LÓGICA: Antes de que la respuesta se escriba, guardamos el código de estado en nuestro campo.
	lrw.statusCode = code
	// LÓGICA: Después de guardar el código, llamamos al método `WriteHeader` original del
	// `ResponseWriter` que envolvimos para que la respuesta se envíe al cliente como de costumbre.
	lrw.ResponseWriter.WriteHeader(code)
}
