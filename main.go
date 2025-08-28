// La declaración 'package main' define este como el paquete principal.
// En Go, un programa ejecutable siempre debe tener un paquete 'main' y una función 'main'.
// Es el punto de inicio de la ejecución del programa.
package main

import (
	// "embed" es un paquete especial de Go que permite incrustar archivos y directorios
	// directamente en el binario ejecutable de la aplicación. Lo usamos para empaquetar
	// todo el frontend de Vue compilado.
	"embed"
	// "io/fs" define la interfaz para sistemas de archivos. La usamos para trabajar con los
	// archivos que hemos incrustado con el paquete 'embed'.
	"io/fs"
	// "log" proporciona funciones para un logging simple, ideal para registrar mensajes de estado
	// y errores fatales que detienen la aplicación.
	"log"
	// "net/http" es la librería estándar de Go para todo lo relacionado con el protocolo HTTP.
	// La usamos para crear el servidor web, registrar las rutas (endpoints) y manejar las solicitudes.
	"net/http"

	// Se importan los paquetes internos que hemos creado, cada uno con su propia responsabilidad,
	// manteniendo el código organizado y modular.
	"saintnet.com/m/internal/auth"     // Maneja la lógica de sesiones y autenticación.
	"saintnet.com/m/internal/database" // Maneja la interacción con la base de datos.
	"saintnet.com/m/internal/handlers" // Contiene los manejadores para cada ruta de la API.
)

// SINTAXIS de Go: La directiva `//go:embed all:frontend/dist` es una instrucción para el compilador de Go.
//   - `//go:embed`: Es la directiva que le dice a Go que incruste archivos.
//   - `all:frontend/dist`: Le indica que debe incrustar de forma recursiva (`all:`) todo el contenido
//     del directorio `frontend/dist`, que es donde Vite coloca los archivos compilados del frontend.
//
// LÓGICA: El contenido de ese directorio se cargará en la variable `embeddedFiles`.
// var embeddedFiles embed.FS
var embeddedFiles embed.FS

// La función main es el punto de entrada de la aplicación. La ejecución del programa comienza aquí.
func main() {
	// --- 1. Inicialización de Dependencias ---
	// Este bloque se asegura de que todos los servicios necesarios (como la base de datos y las sesiones)
	// estén configurados y listos antes de que el servidor comience a aceptar solicitudes.

	// SINTAXIS de Go: `db, err := database.InitDB()` es una declaración corta de variables.
	// Llama a la función `InitDB`, que devuelve dos valores: la conexión a la base de datos (`db`)
	// y un posible error (`err`).
	db, err := database.InitDB()
	if err != nil {
		// `log.Fatalf` imprime un mensaje de error en la terminal y detiene la ejecución del programa
		// inmediatamente con un código de estado 1. Se usa para errores que impiden que la aplicación funcione.
		log.Fatalf("[FATAL] Error al inicializar la base de datos: %v", err)
	}
	// SINTAXIS de Go: `defer db.Close()` pospone la ejecución de `db.Close()` hasta que la función `main`
	// esté a punto de terminar. Es la forma idiomática y segura en Go para garantizar que los recursos
	// (como las conexiones a la base de datos) se liberen correctamente, sin importar cómo termine la función.
	defer db.Close()

	// Se inicializa el almacén de cookies para la gestión de sesiones.
	auth.InitStore()

	// --- 2. Preparación del Frontend Embebido ---
	// LÓGICA: Extraemos el subdirectorio `frontend/dist` de nuestros archivos embebidos para
	// poder servirlo como un sistema de archivos independiente.
	distFS, err := fs.Sub(embeddedFiles, "frontend/dist")
	if err != nil {
		log.Fatalf("[FATAL] Error: No se pudo encontrar el directorio 'frontend/dist'. Asegúrate de haber compilado el frontend con 'npm run build'. %v", err)
	}

	// --- 3. Configuración del Enrutador Principal ---
	// SINTAXIS de Go: `http.NewServeMux()` crea una nueva instancia de un enrutador de solicitudes (multiplexer),
	// que se encarga de dirigir las solicitudes HTTP entrantes al manejador correcto según la URL.
	mux := http.NewServeMux()

	// LÓGICA: Se crea un enrutador separado para las rutas de API para poder aplicar el middleware de logging
	// de forma agrupada a todas ellas.
	apiRouter := http.NewServeMux()

	// --- Rutas de la API (Públicas y Protegidas) ---
	// SINTAXIS de Go (a partir de 1.22): `apiRouter.HandleFunc("POST /api/login", handlers.Login)`
	// registra la función `handlers.Login` para que maneje las solicitudes POST a la ruta `/api/login`.
	// SINTAXIS de Go: `apiRouter.Handle("GET /...", ...)` se usa cuando el manejador es una `http.Handler` (como nuestro middleware).
	// `handlers.AuthMiddleware(...)` envuelve nuestro manejador final, asegurando que solo los usuarios autenticados puedan acceder.
	// `http.HandlerFunc(...)` es un adaptador que convierte una función normal en un `http.Handler`.

	// Rutas Públicas (no requieren autenticación completa)
	apiRouter.HandleFunc("POST /api/login", handlers.Login)
	apiRouter.HandleFunc("POST /api/logout", handlers.Logout)
	apiRouter.HandleFunc("POST /api/force-password-change", handlers.ForcePasswordChange)

	// Rutas Protegidas (requieren autenticación)
	apiRouter.Handle("POST /api/session/extend", handlers.AuthMiddleware(http.HandlerFunc(handlers.ExtendSession)))
	apiRouter.Handle("GET /api/analytics/sales-forecast", handlers.AuthMiddleware(http.HandlerFunc(handlers.GetSalesForecastHandler)))
	apiRouter.Handle("GET /api/analytics/market-basket", handlers.AuthMiddleware(http.HandlerFunc(handlers.GetMarketBasketHandler)))
	apiRouter.Handle("GET /api/dashboard/data", handlers.AuthMiddleware(http.HandlerFunc(handlers.GetDashboardData)))
	apiRouter.Handle("POST /api/dashboard/select-connection", handlers.AuthMiddleware(http.HandlerFunc(handlers.SelectConnection)))
	apiRouter.Handle("GET /api/transactions", handlers.AuthMiddleware(http.HandlerFunc(handlers.GetTransactionsList)))
	apiRouter.Handle("GET /api/transaction/{type}/{id}", handlers.AuthMiddleware(http.HandlerFunc(handlers.GetTransactionDetail)))
	apiRouter.Handle("GET /api/entity/{type}/{id}", handlers.AuthMiddleware(http.HandlerFunc(handlers.GetEntityDetail)))
	apiRouter.Handle("GET /api/settings", handlers.AuthMiddleware(http.HandlerFunc(handlers.GetSettings)))
	apiRouter.Handle("POST /api/settings", handlers.AuthMiddleware(http.HandlerFunc(handlers.UpdateSettings)))
	apiRouter.Handle("GET /api/connections", handlers.AuthMiddleware(http.HandlerFunc(handlers.GetConnections)))
	apiRouter.Handle("POST /api/connections", handlers.AuthMiddleware(http.HandlerFunc(handlers.AddConnection)))
	apiRouter.Handle("DELETE /api/connections/{id}", handlers.AuthMiddleware(http.HandlerFunc(handlers.DeleteConnection)))
	apiRouter.Handle("GET /api/users", handlers.AuthMiddleware(http.HandlerFunc(handlers.GetUsers)))
	apiRouter.Handle("POST /api/users", handlers.AuthMiddleware(http.HandlerFunc(handlers.AddUser)))

	// LÓGICA: Se registra el enrutador de la API bajo el prefijo "/api/".
	// Todas las solicitudes que comiencen con "/api/" serán envueltas primero por el LoggingMiddleware
	// y luego dirigidas al `apiRouter` para que encuentre el manejador específico.
	mux.Handle("/api/", handlers.LoggingMiddleware(apiRouter))

	// --- Ruta para el Frontend (SPA) ---
	// LÓGICA: `mux.Handle("/", ...)` actúa como una ruta "catch-all" (atrapa todo).
	// Cualquier solicitud que NO coincida con "/api/" será manejada por el `FrontendHandler`,
	// que se encarga de servir la aplicación de Vue.
	mux.Handle("/", handlers.FrontendHandler(distFS))

	// --- 4. Inicio del Servidor ---
	port := "8080"
	log.Println("==============================================")
	log.Println("  SAINT BI - Business Intelligence Dashboard")
	log.Println("  Versión: 1.0.0")
	log.Println("==============================================")
	log.Println("[INFO] Estado del Servidor: En línea")
	log.Printf("[INFO] Conexión a Base de Datos (data.db): Exitosa")
	log.Printf("🚀 Interfaz de Usuario disponible en: http://localhost:%s", port)

	// SINTAXIS de Go: `http.ListenAndServe(":"+port, mux)` inicia el servidor HTTP.
	// - `":"+port`: Es la dirección en la que escuchará el servidor (ej. ":8080").
	// - `mux`: Es el enrutador que manejará todas las solicitudes entrantes.
	// Esta función es bloqueante; detendrá la ejecución del programa en esta línea
	// mientras el servidor esté en funcionamiento.
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		// Este código solo se ejecuta si ListenAndServe devuelve un error (ej. el puerto ya está en uso).
		log.Fatalf("[FATAL] Error al iniciar el servidor: %v", err)
	}
}
