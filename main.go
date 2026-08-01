package main

import (
	"log"
	"net/http"

	"ascii-art-web/handlers"
)

func main() {
	// Route "/" χειρίζεται την αρχική σελίδα (GET)
	// Προσοχή: το "/" πιάνει και άγνωστα paths — το 404 γίνεται μέσα στον handler
	http.HandleFunc("/", handlers.Home)

	// Route "/ascii-art" χειρίζεται την υποβολή της φόρμας (POST)
	http.HandleFunc("/ascii-art", handlers.AsciiArt)

	// Σερβίρει τα static αρχεία (CSS, εικόνες)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	log.Println("Server running at http://localhost:8080")

	// ListenAndServe μπλοκάρει — αν αποτύχει (π.χ. port κατειλημμένο), log.Fatal σταματά το πρόγραμμα
	log.Fatal(http.ListenAndServe(":8080", nil))
}
