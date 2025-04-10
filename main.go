package main

import (
	"assignment_02/handler"
	"log"
	"net/http"
)

func main() {

	port := "8080"

	log.Println("Firestore client initialized successfully.")

	http.HandleFunc("/dashboard/v1/registrations/", handler.RegistrationHandler)
	http.HandleFunc("/dashboard/v1/dashboards/", handler.HandleDashboard)
	http.HandleFunc("/dashboard/v1/notifications/", handler.NotificationHandler)
	http.HandleFunc("/dashboard/v1/notifications/{id}", handler.NotificationHandler)
	http.HandleFunc("/dashboard/v1/status/", handler.HandleStatus)

	fs := http.FileServer(http.Dir("./handler"))
	http.Handle("/", fs)

	log.Println("Server starting on port " + port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
