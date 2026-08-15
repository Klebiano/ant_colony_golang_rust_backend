package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"ant_colony_golang_backend/database"
	"ant_colony_golang_backend/handlers"
)

func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func main() {
	execDir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get working directory: %v", err)
	}

	dbPath := filepath.Join(execDir, "sql_app.db")
	sqlPath := filepath.Join(execDir, "database", "database.sql")

	db, err := database.InitDB(dbPath, sqlPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	acHandler := handlers.NewAntColonyHandler(db, execDir)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /ant-colony/get-turbines-map", acHandler.GetTurbinesMap)
	mux.HandleFunc("GET /ant-colony/get-subsystems", acHandler.GetSubsystems)
	mux.HandleFunc("POST /ant-colony/run-route-optimizer", acHandler.RunRouteOptimizer)
	mux.HandleFunc("POST /ant-colony/run-route-optimizer/", acHandler.RunRouteOptimizer)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Starting Golang Ant Colony Backend on port %s ...\n", port)
	log.Fatal(http.ListenAndServe(":"+port, CORSMiddleware(mux)))
}
