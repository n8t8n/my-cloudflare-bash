package main

import (
	"log"
	"net/http"
	"os"

	"cf-manager/handlers"
	"cf-manager/middleware"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Validate required environment variables
	requiredVars := []string{"CF_API_TOKEN", "CF_ZONE_ID", "CF_DOMAIN"}
	for _, v := range requiredVars {
		if os.Getenv(v) == "" {
			log.Fatalf("Required environment variable %s is not set", v)
		}
	}

	r := mux.NewRouter()

	// Public routes
	r.HandleFunc("/", handlers.IndexHandler).Methods("GET")
	r.HandleFunc("/login", handlers.LoginHandler).Methods("POST")
	r.HandleFunc("/logout", handlers.LogoutHandler).Methods("GET")

	// Protected routes
	protected := r.PathPrefix("/").Subrouter()
	protected.Use(middleware.AuthMiddleware)
	protected.HandleFunc("/dashboard", handlers.DashboardHandler).Methods("GET")

	// DNS Management routes
	protected.HandleFunc("/dns/records", handlers.ListDNSRecordsHandler).Methods("GET")
	protected.HandleFunc("/dns/records", handlers.CreateDNSRecordHandler).Methods("POST")
	protected.HandleFunc("/dns/records/{id}", handlers.UpdateDNSRecordHandler).Methods("PUT")
	protected.HandleFunc("/dns/records/{id}", handlers.DeleteDNSRecordHandler).Methods("DELETE")

	// Tunnel Management routes
	protected.HandleFunc("/tunnels", handlers.ListTunnelsHandler).Methods("GET")
	protected.HandleFunc("/tunnels", handlers.CreateTunnelHandler).Methods("POST")
	protected.HandleFunc("/tunnels/{name}", handlers.DeleteTunnelHandler).Methods("DELETE")
	protected.HandleFunc("/tunnels/{name}/start", handlers.StartTunnelHandler).Methods("POST")
	protected.HandleFunc("/tunnels/{name}/stop", handlers.StopTunnelHandler).Methods("POST")
	protected.HandleFunc("/tunnels/{name}/status", handlers.GetTunnelStatusHandler).Methods("GET")
	protected.HandleFunc("/tunnels/{name}/config", handlers.EditTunnelConfigHandler).Methods("PUT")

	// System routes
	protected.HandleFunc("/system/status", handlers.SystemStatusHandler).Methods("GET")

	// Start the server
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	log.Printf("Cloudflare Manager running on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
