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

//go:embed all:frontend/dist
var embeddedFiles embed.FS

// La función main es donde todo comienza.
func main() {
	// --- 1. Inicialización de Dependencias ---
	db, err := database.InitDB()
	if err != nil {
		log.Fatalf("[FATAL] Error al inicializar la base de datos: %v", err)
	}
	defer db.Close()

	auth.InitStore()

	// --- 2. Preparación del Frontend Embebido ---
	distFS, err := fs.Sub(embeddedFiles, "frontend/dist")
	if err != nil {
		log.Fatalf("[FATAL] Error: No se pudo encontrar el directorio 'frontend/dist'. Asegúrate de haber compilado el frontend con 'npm run build'. %v", err)
	}

	// --- 3. Configuración del Enrutador Principal ---
	// http.NewServeMux() crea un nuevo enrutador de solicitudes (multiplexer).
	mux := http.NewServeMux()

	// --- Rutas de la API (Públicas) ---
	// La sintaxis "METHOD /path" (de Go 1.22+) registra un handler para un método y ruta específicos.
	mux.HandleFunc("POST /api/login", handlers.Login)
	mux.HandleFunc("POST /api/logout", handlers.Logout)
	mux.HandleFunc("POST /api/force-password-change", handlers.ForcePasswordChange)
	mux.Handle("POST /api/session/extend", handlers.AuthMiddleware(http.HandlerFunc(handlers.ExtendSession)))

	mux.Handle("GET /api/analytics/sales-forecast", handlers.AuthMiddleware(http.HandlerFunc(handlers.GetSalesForecastHandler)))
	mux.Handle("GET /api/analytics/market-basket", handlers.AuthMiddleware(http.HandlerFunc(handlers.GetMarketBasketHandler)))

	// --- Rutas de la API (Protegidas) ---
	// Cada una de estas rutas se envuelve con nuestro 'AuthMiddleware'.
	// 'http.HandlerFunc()' es un adaptador que nos permite usar una función ordinaria
	// (como 'handlers.GetDashboardData') como un 'http.Handler'.

	// Rutas del Dashboard
	mux.Handle("GET /api/dashboard/data", handlers.AuthMiddleware(http.HandlerFunc(handlers.GetDashboardData)))
	mux.Handle("POST /api/dashboard/select-connection", handlers.AuthMiddleware(http.HandlerFunc(handlers.SelectConnection)))

	// --- NUEVAS RUTAS PARA LA FUNCIONALIDAD DE DRILLDOWN ---
	// Se registra la ruta para obtener listas de transacciones. Cualquier solicitud GET a /api/transactions
	// pasará primero por AuthMiddleware y luego será manejada por GetTransactionsList.
	mux.Handle("GET /api/transactions", handlers.AuthMiddleware(http.HandlerFunc(handlers.GetTransactionsList)))

	// Se registra la ruta para obtener el detalle de una transacción.
	// Los segmentos {type} y {id} son comodines (wildcards) que capturan valores de la URL.
	// Por ejemplo, en una solicitud a /api/transaction/invoice/123, el handler podrá
	// acceder a "invoice" como 'type' y a "123" como 'id'.
	mux.Handle("GET /api/transaction/{type}/{id}", handlers.AuthMiddleware(http.HandlerFunc(handlers.GetTransactionDetail)))

	// Ruta para obtener detalles de entidades genericas
	// Capturara solamente solicitudes como /api/entity/customer/123
	mux.Handle("GET /api/entity/{type}/{id}", handlers.AuthMiddleware(http.HandlerFunc(handlers.GetEntityDetail)))

	mux.Handle("GET /api/settings", handlers.AuthMiddleware(http.HandlerFunc(handlers.GetSettings)))
	mux.Handle("POST /api/settings", handlers.AuthMiddleware(http.HandlerFunc(handlers.UpdateSettings)))

	// Rutas de Administración
	mux.Handle("GET /api/connections", handlers.AuthMiddleware(http.HandlerFunc(handlers.GetConnections)))
	mux.Handle("POST /api/connections", handlers.AuthMiddleware(http.HandlerFunc(handlers.AddConnection)))
	mux.Handle("DELETE /api/connections/{id}", handlers.AuthMiddleware(http.HandlerFunc(handlers.DeleteConnection)))
	mux.Handle("GET /api/users", handlers.AuthMiddleware(http.HandlerFunc(handlers.GetUsers)))
	mux.Handle("POST /api/users", handlers.AuthMiddleware(http.HandlerFunc(handlers.AddUser)))

	// --- Ruta para el Frontend (SPA) ---
	// mux.Handle("/", ...) actúa como una ruta "catch-all" (atrapa todo).
	// Cualquier solicitud que NO coincida con una ruta de API más específica será
	// manejada por el 'FrontendHandler', que sirve la aplicación de Vue.
	mux.Handle("/", handlers.FrontendHandler(distFS))

	// --- 4. Inicio del Servidor ---
	port := "8080"
	log.Printf("=====================================")
	log.Printf("SAINT B.I.")
	log.Printf("Version: 1.0.0")
	log.Printf("=====================================")
	log.Printf("[INFO] Estado del servidor: En linea")
	log.Printf("[INFO] Conexión a base de datos: Exitosa")
	log.Printf("Servidor iniciado en puerto: %s", port)

	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("[FATAL] Error al iniciar el servidor: %v", err)
	}
}
