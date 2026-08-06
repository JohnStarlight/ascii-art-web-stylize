package export

// init καταχωρεί τον εαυτό του στο Registry μόλις φορτωθεί το package —
// έτσι το handlers package απλά ζητάει Registry["txt"] χωρίς να γνωρίζει
// τίποτα για αυτό το αρχείο.
func init() {
	Registry["txt"] = writeTxt
}

// writeTxt επιστρέφει το ASCII art ως απλό κείμενο.
// Το Color δεν χρησιμοποιείται — το .txt δεν έχει τρόπο να κρατήσει χρώμα.
func writeTxt(a Art) ([]byte, error) {
	return []byte(a.Text), nil
}
