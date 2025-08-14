package main

import (
	"log"
	"net/http"

	"saintnet.com/m/internal/auth"
	"saintnet.com/m/internal/database"
	"saintnet.com/m/internal/handlers"
)

func main() {
	db, err := database.InitDB()
	if err != nil {
		log.Fatalf("Error inicializando la base de datos: %v", err)
	}
	defer db.Close()

	// 1. Cargar las plantillas HTML usando tu función ParseTemplates de embed.go
	templates := ParseTemplates()

	// 2. Inicializar los handlers con las plantillas cargadas
	handlers.InitHandlers(templates)

	// 3. Inicializar el almacén de sesiones
	auth.InitStore()

	mux := http.NewServeMux()

	// --- Rutas Públicas (sin autenticación) ---
	mux.HandleFunc("GET /login", handlers.LoginPage)
	mux.HandleFunc("POST /login", handlers.Login)
	mux.HandleFunc("GET /force-password-change", handlers.ForcePasswordChangePage)
	mux.HandleFunc("POST /force-password-change", handlers.ForcePasswordChange)
	mux.HandleFunc("POST /logout", handlers.Logout) // Logout debe ser accesible para cerrar sesión

	// --- Rutas Protegidas (requieren autenticación) ---
	protected := http.NewServeMux()
	protected.HandleFunc("GET /dashboard", handlers.DashboardPage)
	protected.HandleFunc("POST /dashboard/select-connection", handlers.SelectConnection)
	protected.HandleFunc("GET /dashboard/data", handlers.GetDashboardData)

	protected.HandleFunc("GET /connections", handlers.ConnectionsPage)
	protected.HandleFunc("POST /connections", handlers.AddConnection)
	protected.HandleFunc("DELETE /connections/{id}", handlers.DeleteConnection)

	protected.HandleFunc("GET /users", handlers.UsersPage)
	protected.HandleFunc("POST /users", handlers.AddUser)

	// Aplicar el middleware a todas las rutas protegidas
	mux.Handle("/", handlers.AuthMiddleware(protected))

	// --- Iniciar Servidor ---
	log.Println("Servidor iniciado en http://localhost:8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalf("Error al iniciar el servidor: %v", err)
	}
}
