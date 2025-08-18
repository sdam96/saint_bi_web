// La declaración 'package main' define este como el paquete principal.
// La función 'main' dentro de este paquete es el punto de inicio de la ejecución del programa.
package main

import (
	// "embed" permite incrustar archivos (como el frontend de Vue) en el binario final de Go.
	"embed"
	// "io/fs" define la interfaz para sistemas de archivos, usada para trabajar con los archivos embebidos.
	"io/fs"
	// "log" para registrar mensajes de estado y errores fatales.
	"log"
	// "net/http" para todo lo relacionado con el servidor web y el enrutamiento.
	"net/http"

	// Se importan los paquetes internos que hemos creado, cada uno con su propia responsabilidad.
	"saintnet.com/m/internal/auth"
	"saintnet.com/m/internal/database"
	"saintnet.com/m/internal/handlers"
)

// Esta directiva incrusta el contenido del directorio 'frontend/dist' en la variable 'embeddedFiles'.
//
//go:embed all:frontend/dist
var embeddedFiles embed.FS

// La función main es donde todo comienza.
func main() {
	// --- 1. Inicialización de Dependencias ---
	db, err := database.InitDB()
	if err != nil {
		log.Fatalf("Error al inicializar la base de datos: %v", err)
	}
	defer db.Close()

	auth.InitStore()

	// --- 2. Preparación del Frontend Embebido ---
	distFS, err := fs.Sub(embeddedFiles, "frontend/dist")
	if err != nil {
		log.Fatalf("Error: No se pudo encontrar el directorio 'frontend/dist'. Asegúrate de haber compilado el frontend con 'npm run build'. %v", err)
	}

	// --- 3. Configuración del Enrutador Principal ---
	mux := http.NewServeMux()

	// --- Rutas de la API (Públicas) ---
	// Estas rutas NO están protegidas por el middleware. Son accesibles para todos.
	// La sintaxis "METHOD /path" es específica de Go 1.22+ y asegura que solo ese método HTTP sea aceptado.
	mux.HandleFunc("POST /api/login", handlers.Login)
	mux.HandleFunc("POST /api/logout", handlers.Logout)
	mux.HandleFunc("POST /api/force-password-change", handlers.ForcePasswordChange)

	// --- Rutas de la API (Protegidas) ---
	// Para cada ruta protegida, envolvemos su manejador explícitamente con el middleware.
	// Esto evita el error de que una ruta genérica intercepte una específica.
	// 'http.HandlerFunc()' convierte una función como 'handlers.GetDashboardData' en un 'http.Handler'.
	mux.Handle("GET /api/dashboard/data", handlers.AuthMiddleware(http.HandlerFunc(handlers.GetDashboardData)))
	mux.Handle("POST /api/dashboard/select-connection", handlers.AuthMiddleware(http.HandlerFunc(handlers.SelectConnection)))

	mux.Handle("GET /api/connections", handlers.AuthMiddleware(http.HandlerFunc(handlers.GetConnections)))
	mux.Handle("POST /api/connections", handlers.AuthMiddleware(http.HandlerFunc(handlers.AddConnection)))
	mux.Handle("DELETE /api/connections/{id}", handlers.AuthMiddleware(http.HandlerFunc(handlers.DeleteConnection)))

	mux.Handle("GET /api/users", handlers.AuthMiddleware(http.HandlerFunc(handlers.GetUsers)))
	mux.Handle("POST /api/users", handlers.AuthMiddleware(http.HandlerFunc(handlers.AddUser)))

	// --- Ruta para el Frontend (SPA) ---
	// 'mux.Handle("/", ...)' actúa como una ruta "catch-all" (atrapa todo).
	// Cualquier solicitud que NO coincida con una ruta de API más específica (como las de arriba)
	// será manejada por el 'FrontendHandler', que sirve la aplicación de Vue.
	mux.Handle("/", handlers.FrontendHandler(distFS))

	// --- 4. Inicio del Servidor ---
	log.Println("Servidor Go listo en http://localhost:8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalf("Error al iniciar el servidor: %v", err)
	}
}
