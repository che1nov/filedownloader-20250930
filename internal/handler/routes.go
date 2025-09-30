package handler

import (
	"net/http"

	"github.com/gorilla/mux"
)

// SetupRoutes configures HTTP API routes
func SetupRoutes(th *TaskHandler) *mux.Router {
	r := mux.NewRouter()
	api := r.PathPrefix("/api/v1").Subrouter()
	api.HandleFunc("/tasks", th.CreateTask).Methods("POST")
	api.HandleFunc("/tasks/{id}/status", th.GetTaskStatus).Methods("GET")
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods("GET")
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("File Downloader API"))
	}).Methods("GET")

	return r
}
