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

	// Route "/export" χειρίζεται το κουμπί "Export as .txt" (POST)
	http.HandleFunc("/export", handlers.Export)

	// Route "/static/" σερβίρει τα στατικά αρχεία (CSS) από τον φάκελο static/
	// StripPrefix: ο FileServer ψάχνει το URL path αυτούσιο μέσα στον φάκελο —
	// χωρίς αυτό θα έψαχνε το "static/static/style.css" και θα γύριζε 404
	// Δεν χρειάζεται εξαίρεση στον Home: ο ServeMux διαλέγει το μακρύτερο pattern
	// που ταιριάζει, οπότε το "/static/" κερδίζει το "/"
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	log.Println("Server running at http://localhost:8080")

	// ListenAndServe μπλοκάρει — αν αποτύχει (π.χ. port κατειλημμένο), log.Fatal σταματά το πρόγραμμα
	log.Fatal(http.ListenAndServe(":8080", nil))
}
