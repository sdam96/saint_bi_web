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

	// 1. Cargar las plantillas HTML
	templates := ParseTemplates()

	// 2. Inicializar los handlers con las plantillas cargadas
	handlers.InitHandlers(templates)

	auth.InitStore()

	mux := http.NewServeMux()

	// Rutas publicas
	mux.HandleFunc("GET /login", handlers.LoginPage)
	mux.HandleFunc("POST /login", handlers.Login)
	mux.HandleFunc("GET /force-password-change", handlers.ForcePasswordChangePage)
	mux.HandleFunc("POST /force-password-change", handlers.ForcePasswordChange)

	// Rutas protegidas por autenticacion
	mux.Handle("GET /dashboard", handlers.AuthMiddleware(http.HandlerFunc(handlers.DashboardPage)))
	mux.Handle("GET /dashboard/connections", handlers.AuthMiddleware(http.HandlerFunc(handlers.GetConnectionsDropdown)))
	mux.Handle("POST /dashboard/select-connection", handlers.AuthMiddleware(http.HandlerFunc(handlers.SelectConnection)))
	mux.Handle("GET /dashboard/data", handlers.AuthMiddleware(http.HandlerFunc(handlers.GetDashboardData)))

	mux.Handle("GET /connections", handlers.AuthMiddleware(http.HandlerFunc(handlers.ConnectionsPage)))
	mux.Handle("POST /connections", handlers.AuthMiddleware(http.HandlerFunc(handlers.AddConnection)))
	mux.Handle("DELETE /connections/{id}", handlers.AuthMiddleware(http.HandlerFunc(handlers.DeleteConnection)))

	mux.Handle("GET /users", handlers.AuthMiddleware(http.HandlerFunc(handlers.UsersPage)))
	mux.Handle("POST /users", handlers.AuthMiddleware(http.HandlerFunc(handlers.AddUser)))

	mux.Handle("POST /logout", handlers.AuthMiddleware(http.HandlerFunc(handlers.Logout)))

	// Ruta raiz redirige al dashboard si esta autenticado, si no al login
	mux.Handle("/", handlers.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/dashboard", http.StatusFound)
	})))

	log.Println("Servidor iniciado en http://localhost:8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalf("Error al iniciar el servidor: %v", err)
	}
}
