// Package export μετατρέπει το αποτέλεσμα του ascii-art σε αρχεία προς download.
//
// Κάθε format (txt, rtf, json, ...) δηλώνει έναν Writer και τον καταχωρεί στο
// Registry μέσω init() στο δικό του αρχείο (π.χ. txt.go). Το handler που σερβίρει
// το /export απλά ψάχνει στο Registry με το ζητούμενο format — δεν χρειάζεται να
// ξέρει τίποτα για τη λογική κάθε writer.
package export

// Art είναι τα δεδομένα που χρειάζεται ένας writer για να παράξει το αρχείο.
// Text: το καθαρό ASCII art, χωρίς HTML <span> tags.
// Color: το χρώμα που είχε επιλέξει ο χρήστης στη φόρμα (μπορεί να είναι κενό).
// Ένας writer μπορεί να αγνοήσει όποιο πεδίο δεν τον αφορά (π.χ. το txt αγνοεί το Color).
type Art struct {
	Text  string
	Color string
}

// Writer παράγει τα bytes του αρχείου εξαγωγής για ένα συγκεκριμένο format.
type Writer func(Art) ([]byte, error)

// Registry συνδέει το όνομα ενός format (π.χ. "txt") με τον Writer του.
// Νέα formats προστίθενται με μία γραμμή σε ξεχωριστό αρχείο, χωρίς να
// τροποποιείται αυτό το αρχείο.
var Registry = map[string]Writer{}
