package main

import (
	"log"
	"net/http"
	"os"

	"github.com/pakesson/cfs-server-go/handlers"

	"github.com/gorilla/mux"
)

func main() {
	if os.Getenv("AWS_REGION") == "" ||
		os.Getenv("S3_BUCKET") == "" ||
		os.Getenv("AWS_ACCESS_KEY_ID") == "" ||
		os.Getenv("AWS_SECRET_ACCESS_KEY") == "" {
		log.Fatal("Missing configuration environment variable(s)")
	}

	log.Println("Starting server")
	router := mux.NewRouter()

	router.HandleFunc("/", handlers.WebUploadHandler).Methods("GET")
	router.HandleFunc("/", handlers.WebUploadFileHandler).Methods("POST")

	router.HandleFunc("/download/{key}", handlers.WebDownloadHandler).Methods("GET")
	router.HandleFunc("/download/{key}", handlers.WebDownloadFileHandler).Methods("POST")

	router.HandleFunc("/api/download", handlers.ApiDownloadHandler)
	router.HandleFunc("/api/upload", handlers.ApiUploadHandler)

	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	http.Handle("/", router)
	log.Fatal(http.ListenAndServe(":5000", nil))
}
