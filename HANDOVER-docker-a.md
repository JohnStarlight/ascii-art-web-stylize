# Σημείωμα παράδοσης — Docker A

**Από:** Άτομο 2 · **Προς:** Άτομο 1 (Docker B) · branch `feature/docker-a-ivogiake`

---

## Τι άλλαξε

- `static/` + route `FileServer` στο [main.go](main.go)
- Το inline `<style>` του [index.html](templates/index.html) μεταφέρθηκε **αυτούσιο** στο `static/style.css`
- `Dockerfile` — single-stage, `WORKDIR /app`
- [README.md](README.md) — ενότητα *Running with Docker*

---

## Τι πρέπει να ξέρεις για να συνεχίσεις

**1. Ο κώδικας διαβάζει αρχεία με σχετικά paths.**
`template.ParseFiles("templates/index.html")` στο [handlers.go](handlers/handlers.go), `filepath.Join("banners", ...)` στο [ascii.go](ascii/ascii.go). Επιλύονται ως προς το **working directory του process** — δηλαδή το `WORKDIR`.

**2. Στο multi-stage θα χρειαστεί ρητό `COPY` για τέσσερα πράγματα:**
`server` (το binary που παράγει το `go build -o server .`), `templates/`, `banners/`, `static/`.
Το δεύτερο stage ξεκινά από άδειο filesystem — τίποτα δεν περνάει αυτόματα.

**3. Αν ξεχάσεις ένα, σκάει ύπουλα.**
Το build περνάει, το container σηκώνεται, τα logs λένε «Server running». Και το request γυρνάει 400 ή 500. **Δεν το πιάνεις από το build.**

**4. Το μόνο test που το πιάνει είναι το POST:**
```bash
curl -i -X POST localhost:8080/ascii-art -d "text=hello" -d "banner=standard"
```
Είναι το μόνο request που αναγκάζει τον server να διαβάσει banner file. Το σκέτο `GET /` περνάει ακόμα κι όταν λείπουν τα banners.

---

## Αποφάσεις που πήρα — μη τις γυρίσεις πίσω κατά λάθος

**Το inline `<script>` έμεινε στο `index.html`.** Δεν είναι ξεχασμένο. Περιέχει `{{.Color}}`, δηλαδή Go template action. Ένα `.js` αρχείο σερβίρεται από τον FileServer **χωρίς** να περάσει από το template engine, οπότε το `{{.Color}}` θα έφτανε στον browser σαν literal string.

**Ο `handlers.Home` δεν πειράχτηκε.** Το `/static/` δουλεύει χωρίς εξαίρεση εκεί μέσα, επειδή ο `ServeMux` διαλέγει το **μακρύτερο pattern** που ταιριάζει, όχι το πρώτο που δηλώθηκε.

**Τα error templates (400/404/405/500) δεν έχουν `<link>` στο style.css.** Δεν είχαν ποτέ inline styles, οπότε δεν υπήρχε τι να μεταφέρω. Το άφησα για το **Stylize C**, ώστε να αποφασίσει μόνος του τη δομή.

**Το CSS μεταφέρθηκε αυτούσιο** — ίδια στοίχιση, ίδια σχόλια, καμία «βελτίωση». Ο Stylize A πρέπει να ξέρει τι παρέλαβε.

---

## Τι ΔΕΝ έκανα επίτηδες

| | Ανήκει σε |
|---|---|
| multi-stage build | Docker B |
| `.dockerignore` | Docker B |
| `LABEL` metadata | Docker B |
| ενοποίηση error templates με `{{define}}` | Docker B |
| `build_run.sh`, garbage collection | Docker C |

**Εξαίρεση:** έγραψα την ενότητα Docker στο README, που τυπικά ήταν του **Docker C**. Το συζητήσαμε — άλλαξε ο τρόπος εκτέλεσης του project και δεν είχε νόημα να μείνει ατεκμηρίωτος. Όποιος πάρει το Docker C: δες τι υπάρχει ήδη και επέκτεινέ το, μην το ξαναγράψεις.

---

## Πώς επαληθεύεις ότι δουλεύει

```bash
docker build -t ascii-art-web .
docker run --rm -p 8080:8080 ascii-art-web
```

Από δεύτερο tab, και τα τέσσερα:

```bash
curl -i localhost:8080/                    # 200 — βρίσκει το template
curl -i -X POST localhost:8080/ascii-art \
     -d "text=hello" -d "banner=standard"  # 200 — βρίσκει τα banners
curl -i localhost:8080/static/style.css    # 200, Content-Type: text/css
curl -i localhost:8080/asdf                # 404 — ο Home δουλεύει ακόμα
```

Επαληθεύτηκαν και τα τέσσερα μέσα σε container.

Αναλυτικός οδηγός με όλα τα «γιατί»: [DOCKER-A.md](DOCKER-A.md)
