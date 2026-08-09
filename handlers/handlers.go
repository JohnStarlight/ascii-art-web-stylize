package handlers

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"ascii-art-web/ascii"
	"ascii-art-web/export"
)

// validBanners περιέχει τα 3 έγκυρα banner styles.
// Κοινό var ώστε το AsciiArt και το Export να μην κρατούν δύο ξεχωριστά αντίγραφα.
var validBanners = map[string]bool{"standard": true, "shadow": true, "thinkertoy": true}

// PageData είναι το struct που περνάμε στο HTML template
// Χρησιμοποιείται για να εμφανιστούν τα δεδομένα του χρήστη, το αποτέλεσμα και τα errors
type PageData struct {
	Text    string        // κείμενο που έγραψε ο χρήστης
	Banner  string        // banner style που διάλεξε (standard/shadow/thinkertoy)
	Color   string        // χρώμα για το ASCII art (π.χ. red, #ff0000, rgb(255,0,0))
	Letters string        // substring που θα χρωματιστεί — αν κενό, χρωματίζεται όλο
	Result  template.HTML // ASCII art αποτέλεσμα — HTML ώστε να μην escape-αρουν τα <span> tags
	Error   string        // μήνυμα λάθους αν κάτι πάει στραβά
}

// Home χειρίζεται το GET "/" — επιστρέφει την αρχική σελίδα με τη φόρμα
func Home(w http.ResponseWriter, r *http.Request) {
	// Το HandleFunc("/") πιάνει ΟΛΑ τα άγνωστα paths, οπότε ελέγχουμε εδώ για 404
	if r.URL.Path != "/" {
		statusTemplate(w, http.StatusNotFound)
		return
	}

	// Η αρχική σελίδα δέχεται μόνο GET
	if r.Method != http.MethodGet {
		statusTemplate(w, http.StatusMethodNotAllowed)
		return
	}

	tmpl, err := template.ParseFiles("templates/index.html")
	if err != nil {
		http.Error(w, "500 - Internal Server Error", http.StatusInternalServerError)
		return
	}

	// nil γιατί στην αρχική σελίδα δεν έχουμε δεδομένα να περάσουμε στο template
	tmpl.Execute(w, nil)
}

// AsciiArt χειρίζεται το POST "/ascii-art" — παίρνει το κείμενο και το banner,
// καλεί την ascii.Generate() και εμφανίζει το αποτέλεσμα
func AsciiArt(w http.ResponseWriter, r *http.Request) {
	// Αυτό το route δέχεται μόνο POST
	if r.Method != http.MethodPost {
		statusTemplate(w, http.StatusMethodNotAllowed)
		return
	}

	// Διαβάζουμε τα δεδομένα από τη φόρμα
	text := r.FormValue("text")
	banner := r.FormValue("banner")
	color := r.FormValue("color")
	letters := r.FormValue("letters")

	// Αρχικοποιούμε το PageData με τα δεδομένα του χρήστη
	// ώστε να παραμένουν στη φόρμα μετά το submit
	data := PageData{
		Text:    text,
		Banner:  banner,
		Color:   color,
		Letters: letters,
	}

	// Validation: κενό κείμενο → 400
	if text == "" {
		data.Error = "Please enter some text"
		renderTemplate(w, data, http.StatusBadRequest)
		return
	}

	// Validation: έλεγχος ότι το banner είναι ένα από τα 3 έγκυρα styles
	if !validBanners[banner] {
		data.Error = "Invalid banner style"
		renderTemplate(w, data, http.StatusBadRequest)
		return
	}

	// Κλήση της ascii.Generate() από το package του Ατόμου 2
	result, err := ascii.Generate(text, banner, color, letters)
	if err != nil {
		data.Error = err.Error()
		renderTemplate(w, data, http.StatusBadRequest)
		return
	}

	// template.HTML λέει στο Go template "μην κάνεις escape αυτό το string"
	// χρειάζεται γιατί το αποτέλεσμα περιέχει <span> tags για τα χρώματα
	data.Result = template.HTML(result)
	renderTemplate(w, data, http.StatusOK)
}

// Export χειρίζεται το POST "/export" — παίρνει τα ίδια πεδία φόρμας με το
// /ascii-art, παράγει το ASCII art χωρίς χρώμα (καθαρό κείμενο, χωρίς <span>),
// το γράφει σε αρχείο στον server με δικαιώματα rw για τον χρήστη, και το
// στέλνει στον client ως download με τα 3 υποχρεωτικά headers.
func Export(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		statusTemplate(w, http.StatusMethodNotAllowed)
		return
	}

	text := r.FormValue("text")
	banner := r.FormValue("banner")
	color := r.FormValue("color")

	if text == "" {
		statusTemplate(w, http.StatusBadRequest)
		return
	}
	if !validBanners[banner] {
		statusTemplate(w, http.StatusBadRequest)
		return
	}

	// color="" και letters="" -> ascii.Generate επιστρέφει καθαρό κείμενο,
	// χωρίς τα <span style="color: ..."> tags που χρειάζεται μόνο η σελίδα.
	plain, err := ascii.Generate(text, banner, "", "")
	if err != nil {
		statusTemplate(w, http.StatusBadRequest)
		return
	}

	// Μοναδικό διαθέσιμο format προς το παρόν. Το dropdown επιλογής format
	// είναι δουλειά του Export B — τότε το "txt" θα έρχεται από τη φόρμα.
	const format = "txt"
	writeFn, ok := export.Registry[format]
	if !ok {
		statusTemplate(w, http.StatusInternalServerError)
		return
	}

	data, err := writeFn(export.Art{Text: plain, Color: color})
	if err != nil {
		statusTemplate(w, http.StatusInternalServerError)
		return
	}

	// Γράφουμε το αρχείο στον δίσκο του server με δικαιώματα -rw-r--r--
	// (owner: read+write, όλοι οι άλλοι: μόνο read) — αυτό ζητάει η εκφώνηση:
	// "the file must be exported with the right permissions (read and write)".
	if err := os.MkdirAll("exports", 0755); err != nil {
		statusTemplate(w, http.StatusInternalServerError)
		return
	}
	filename := "ascii-art." + format
	path := filepath.Join("exports", filename)
	if err := os.WriteFile(path, data, 0644); err != nil {
		statusTemplate(w, http.StatusInternalServerError)
		return
	}
	// os.WriteFile εφαρμόζει το 0644 μόνο όταν δημιουργεί νέο αρχείο· αν το
	// path υπήρχε ήδη, κρατάει τα παλιά permissions. Το Chmod το εγγυάται
	// και στις δύο περιπτώσεις.
	if err := os.Chmod(path, 0644); err != nil {
		statusTemplate(w, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Content-Disposition", `attachment; filename="`+filename+`"`)
	w.Header().Set("Content-Length", strconv.Itoa(len(data)))
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

// renderTemplate φορτώνει και εκτελεί το HTML template με τα δεδομένα και τον HTTP status code
func renderTemplate(w http.ResponseWriter, data PageData, status int) {
	tmpl, err := template.ParseFiles("templates/index.html")
	if err != nil {
		http.Error(w, "500 - Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Γράφουμε πρώτα σε buffer — αν αποτύχει το Execute αφού έχουμε ήδη
	// γράψει στο w, δεν μπορούμε να αλλάξουμε τον status code
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		http.Error(w, "500 - Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(status)
	buf.WriteTo(w)
}

// statusTemplate συναρμολογεί το κοινό layout.html με το ειδικό περιεχόμενο
// του κάθε status code (400, 404, 405, 500) και το εκτελεί.
func statusTemplate(w http.ResponseWriter, status int) {
	var contentFile string
	switch status {
	case http.StatusBadRequest:
		contentFile = "templates/400.html"
	case http.StatusNotFound:
		contentFile = "templates/404.html"
	case http.StatusMethodNotAllowed:
		contentFile = "templates/405.html"
	case http.StatusInternalServerError:
		contentFile = "templates/500.html"
	default:
		http.Error(w, "500 - Internal Server Error", http.StatusInternalServerError)
		return
	}

	// ParseFiles φορτώνει layout.html + το ειδικό content αρχείο μαζί, στο ίδιο
	// template set — έτσι το {{template "content" .}} μέσα στο layout βρίσκει
	// το σωστό {{define "content"}} block.
	tmpl, err := template.ParseFiles("templates/layout.html", contentFile)
	if err != nil {
		http.Error(w, "500 - Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Το title/h1 είναι ίδιο κείμενο σε όλες τις σελίδες (π.χ. "400 Bad Request"),
	// οπότε το υπολογίζουμε από το status code αντί να το επαναλαμβάνουμε.
	data := struct{ Title string }{
		Title: fmt.Sprintf("%d %s", status, http.StatusText(status)),
	}

	w.WriteHeader(status)
	tmpl.ExecuteTemplate(w, "layout", data)
}
