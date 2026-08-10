# Σημείωμα παράδοσης — Docker B

**Από:** Άτομο 1 · **Προς:** Docker C (Άτομο 3, για το Dockerfile/`.dockerignore`) **και** Stylize C (Άτομο 1)

---

## Τι άλλαξε

- [`Dockerfile`](Dockerfile) — έγινε **multi-stage**: Stage 1 (`golang:1.25.5 AS builder`) χτίζει το binary με `CGO_ENABLED=0`, Stage 2 (`alpine:latest`) αντιγράφει μόνο `server` + `templates/` + `banners/` + `static/`. Τελικό image: **~33MB** (ήταν ~1.4GB).
- [`Dockerfile`](Dockerfile) — προστέθηκε `LABEL` (maintainer/description/version).
- [`.dockerignore`](.dockerignore) — επεκτάθηκε: `.git/`, `*.md`, `.gitignore`, `.dockerignore`, `exports/`.
- [`templates/layout.html`](templates/layout.html) — **νέο** αρχείο, το κοινό "κέλυφος" όλων των error pages, με `{{define "layout"}}` + `{{template "content" .}}`.
- [`templates/400.html`](templates/400.html), [`404.html`](templates/404.html), [`405.html`](templates/405.html), [`500.html`](templates/500.html) — μειώθηκαν από ολόκληρες HTML σελίδες σε μόνο `{{define "content"}}<p>...μήνυμα...</p>{{end}}`.
- [`handlers/handlers.go`](handlers/handlers.go) — το `statusTemplate` πλέον συναρμολογεί layout + content, και υπολογίζει το title/h1 αυτόματα από το status code (`http.StatusText`), αντί να το επαναλαμβάνει σε κάθε αρχείο.
- [`README.md`](README.md) — ενημερώθηκε η ενότητα "Implementation notes" ώστε να περιγράφει το multi-stage build.

---

## Τι πρέπει να ξέρεις — αν παίρνεις το **Docker C**

**1. Το `CGO_ENABLED=0` είναι κρίσιμο, μην το αφαιρέσεις.** Το Stage 1 (Debian-based) και το Stage 2 (Alpine-based) έχουν διαφορετική C library (glibc vs musl). Χωρίς αυτό, το binary χτίζεται δεμένο στο glibc του Stage 1 και **δεν τρέχει καθόλου** στο Alpine του Stage 2 — σκάει με κρυπτικό σφάλμα, όχι με κάτι προφανές.

**2. Αν προσθέσεις κάτι στο `COPY --from=builder`, θυμήσου τα 4 πράγματα που ήδη χρειάζονται:** `server`, `templates/`, `banners/`, `static/`. Το Stage 2 ξεκινά από άδειο filesystem (ίδια παγίδα που είχε επισημάνει και ο Docker A για το δικό του multi-stage warning).

**3. Το `.dockerignore` καλύπτει τα βασικά** — αν προσθέσεις `build_run.sh` ή κάτι άλλο, σκέψου αν χρειάζεται να μπει και αυτό εκεί (π.χ. αν παράγει προσωρινά αρχεία).

---

## Τι πρέπει να ξέρεις — αν κάνεις styling στα error pages (**Stylize C**)

**1. Η δομή άλλαξε εντελώς.** Δεν υπάρχουν πια 4 ανεξάρτητες HTML σελίδες — υπάρχει **ένα** κοινό `layout.html` (εκεί είναι όλο το `<head>`, το `<div id="container">`, το link επιστροφής) και 4 μικροσκοπικά αρχεία που ορίζουν μόνο το δικό τους μήνυμα.

**2. Αν θες να αλλάξεις κοινό styling** (π.χ. χρώμα container, οτιδήποτε ίδιο σε όλες τις σελίδες σφάλματος) — άλλαξέ το **μόνο** στο `layout.html`. Αν το βάλεις μέσα σε ένα από τα 4 μικρά αρχεία, θα εμφανιστεί μόνο σε αυτή τη μία σελίδα, όχι στις άλλες 3.

**3. Το title/h1 ΔΕΝ είναι hardcoded πουθενά στα templates.** Υπολογίζεται στο `handlers.go` από το status code (`fmt.Sprintf("%d %s", status, http.StatusText(status))`). Αν θες διαφορετικό κείμενο εκεί, θα χρειαστεί να αλλάξεις τη λογική στο Go, όχι το HTML.

**4. Πρόσεχε να μην ξαναδημιουργήσεις το παλιό bug.** Πριν το Docker B, το `405.html` είχε σπασμένα `<body>`/`<div>` tags επειδή κάποιος το επεξεργάστηκε χειροκίνητα, ξεχωριστά από τα άλλα 3. Τώρα αυτό δεν μπορεί να ξανασυμβεί όσο η δομή μένει μόνο στο `layout.html` — μην τη διπλασιάσεις ξανά μέσα στα content αρχεία "για σιγουριά".

---

## Αποφάσεις που πήρα — μη τις γυρίσεις πίσω κατά λάθος

**Επέλεξα το layout+content pattern** (`{{template "content" .}}` μέσα στο layout) αντί για ένα μονολιθικό αρχείο με 4 ξεχωριστά `{{define}}` blocks. Το PRD ζητούσε ρητά `{{define}}` **και** `{{template}}` μαζί — αυτό το pattern τα χρησιμοποιεί και τα δύο, ενώ ένα μονολιθικό αρχείο θα χρησιμοποιούσε μόνο `{{define}}` + `ExecuteTemplate` από Go, όχι composition μέσα στο template.

**Επέλεξα Alpine (όχι Debian-slim) για το Stage 2**, για το μέγιστο μικρό μέγεθος — αυτό είναι που έκανε αναγκαίο το `CGO_ENABLED=0`.

---

## Τι ΔΕΝ έκανα επίτηδες

| | Ανήκει σε |
|---|---|
| `build_run.sh`, garbage collection unused objects | Docker C |
| Ενότητα Docker usage στο README (ήδη υπάρχει βασική, από τον Docker A) | Docker C, αν χρειάζεται επέκταση |
| Styling των error pages (χρώματα, layout polish πέρα από τη δομή) | Stylize C |

---

## Πώς επαληθεύεις ότι δουλεύει

```bash
docker build -t ascii-art-web .
docker run --rm -p 8080:8080 --name ascii ascii-art-web
```
Από δεύτερο τερματικό:
```bash
curl -i localhost:8080/                                          # 200
curl -i -X POST localhost:8080/ascii-art -d "text=hello" -d "banner=standard"  # 200
curl -i localhost:8080/bogus                                     # 404, νεο layout
curl -i -X POST localhost:8080/                                  # 405, νεο layout — πρωην σπασμενο
curl -i -X POST localhost:8080/export -d "text=" -d "banner=standard"          # 400, νεο layout

docker exec ascii ls -la /app        # server, templates/, banners/, static/ — τιποτα αλλο
docker image inspect --format='{{json .Config.Labels}}' ascii-art-web   # LABEL metadata
```
