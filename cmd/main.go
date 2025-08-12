package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/naval1525/Social_Media_Backend/internal/config"
	"github.com/naval1525/Social_Media_Backend/internal/database"
	"github.com/naval1525/Social_Media_Backend/internal/handler"
	"github.com/naval1525/Social_Media_Backend/internal/repository"
	"github.com/naval1525/Social_Media_Backend/internal/service"
)

func main() {
    // Load configuration (env/config file). No defaults inside code; must be provided.
    cfg, err := config.Load()
    if err != nil {
        log.Fatal("Failed to load config:", err)
    }

	// Connect to database
    // Create DB connection from config
    var db *database.DB
    if cfg.Database.URL != "" {
        db, err = database.NewWithParams(database.ConnParams{URL: cfg.Database.URL})
    } else {
        db, err = database.NewWithParams(database.ConnParams{
            Host:     cfg.Database.Host,
            Port:     cfg.Database.Port,
            User:     cfg.Database.User,
            Password: cfg.Database.Password,
            Name:     cfg.Database.Name,
            SSLMode:  cfg.Database.SSLMode,
        })
    }
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Run migrations
	if err := db.Migrate(); err != nil {
		log.Fatal("Failed to run migrations:", err)
	}

	// Initialize repositories
	userRepo := repository.NewUserRepository(db.DB)

	// Initialize services
    jwtSecret := cfg.JWTSecret

	userService := service.NewUserService(userRepo, jwtSecret)

	// Initialize handlers
	userHandler := handler.NewUserHandler(userService)

	// Setup router
	router := setupRouter(userHandler, userService)

	// Start server
    log.Printf("🚀 Server starting on port %s", cfg.Server.Port)
    log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", cfg.Server.Port), router))
}

func setupRouter(userHandler *handler.UserHandler, userService service.UserService) *mux.Router {
	router := mux.NewRouter()

	// Apply global middleware
	router.Use(handler.CORSMiddleware)
	router.Use(handler.LoggingMiddleware)

	// API routes
	api := router.PathPrefix("/api/v1").Subrouter()

	// Health check
    api.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
        // use exported helper
        type payload struct{ Status string `json:"status"` }
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusOK)
        json.NewEncoder(w).Encode(map[string]string{"message": "API is running", "status": "healthy"})
    }).Methods("GET")

	// Auth routes (no authentication required)
	auth := api.PathPrefix("/auth").Subrouter()
	auth.HandleFunc("/register", userHandler.Register).Methods("POST")
	auth.HandleFunc("/login", userHandler.Login).Methods("POST")

	// User routes
	users := api.PathPrefix("/users").Subrouter()
	users.HandleFunc("/{id}", userHandler.GetProfile).Methods("GET")

	// Protected user routes (authentication required)
	protectedUsers := users.PathPrefix("").Subrouter()
	protectedUsers.Use(handler.AuthMiddleware(userService))
	protectedUsers.HandleFunc("/me", userHandler.GetMyProfile).Methods("GET")
	protectedUsers.HandleFunc("/me", userHandler.UpdateProfile).Methods("PUT")

	return router
}
