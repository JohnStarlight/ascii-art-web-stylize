# Docker A — Οδηγός υλοποίησης

**Άτομο 2 · Γύρος 1 · branch `feature/docker-a-ivogiake`**

---

## Τι περιλαμβάνει το slice

Σύμφωνα με το [PRD.md](PRD.md) §4:

> φάκελος `static/` + route `FileServer` στο `main.go` · μεταφορά του inline `<style>` σε `static/style.css` · βασικό Dockerfile (προσοχή στο `COPY`: χρειάζονται **server, static, templates, banners**) · build & run, επιβεβαίωση ότι σηκώνεται

## Τι ΔΕΝ περιλαμβάνει

Ο κανόνας §6 «δεν υπερβάλλουμε σε ζήλο» δεν είναι ευγένεια — αν κάνεις τη δουλειά του επόμενου, δεν τη μαθαίνει.

| Μην το κάνεις | Ανήκει σε |
|---|---|
| Multi-stage build | Docker B |
| `.dockerignore` | Docker B |
| `LABEL` metadata | Docker B |
| Ενοποίηση των error templates με `{{define}}` | Docker B |
| `build_run.sh`, README ενότητα Docker | Docker C |
| Οποιαδήποτε αισθητική απόφαση στο CSS | Stylize A/B/C |

Στο βήμα 2 **μεταφέρεις** το CSS. Δεν το βελτιώνεις, δεν το καθαρίζεις, δεν αλλάζεις χρώματα. Copy-paste και τέλος.

> **Απόφαση ομάδας:** δουλεύουμε σε ένα codebase και στο τέλος το ανεβάζουμε και στα τρία repos. Δεν χωρίζεται η δουλειά ανά repo.

---

## Βήμα 1 — `static/` + FileServer

### Ο κώδικας

Στο [main.go](main.go), δίπλα στα υπάρχοντα routes:

```go
http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
```

### Τι κάνει το καθένα

- **`http.Dir("static")`** — ορίζει τον φάκελο του filesystem που εκτίθεται. Σχετικό path, ως προς το working directory του process.
- **`http.FileServer(...)`** — handler που σερβίρει αρχεία από αυτόν τον φάκελο. Παίρνει το `r.URL.Path` και το ψάχνει *αυτούσιο* μέσα στον φάκελο.
- **`http.StripPrefix("/static/", ...)`** — αφαιρεί το `/static/` από το path πριν φτάσει στον FileServer.

**Γιατί χρειάζεται το StripPrefix:** χωρίς αυτό, ένα request για `/static/style.css` φτάνει στον FileServer με path `/static/style.css`, και ο FileServer ψάχνει το `static/static/style.css`. 404. Με το StripPrefix το path γίνεται `style.css` και βρίσκει το `static/style.css`.

- **`http.Handle`** (όχι `HandleFunc`) — γιατί ο FileServer επιστρέφει `http.Handler`, όχι συνάρτηση. Το `HandleFunc` δέχεται `func(w, r)`.

### Το θέμα με το `"/"`

Το [handlers.go](handlers/handlers.go) έχει ήδη `http.HandleFunc("/", handlers.Home)`, που πιάνει **όλα** τα paths. Λογικά θα έπρεπε να πιάσει και το `/static/style.css` και να επιστρέψει 404 από τον έλεγχο `if r.URL.Path != "/"`.

Δεν συμβαίνει: ο `ServeMux` της Go διαλέγει το **μακρύτερο pattern που ταιριάζει**, όχι το πρώτο που δηλώθηκε. Το `/static/` (8 χαρακτήρες) κερδίζει το `/` (1 χαρακτήρας).

Συνέπεια: **δεν χρειάζεται να αγγίξεις τον `handlers.Home`.** Αν σε πιάσει ο πειρασμός να προσθέσεις εξαίρεση για το `/static/` εκεί μέσα, σταμάτα — είναι περιττό και δείχνει ότι δεν κατάλαβες το routing.

### Επαλήθευση

```bash
go run .
curl -i localhost:8080/static/style.css   # 200, Content-Type: text/css
curl -i localhost:8080/asdf               # 404 — ο Home δουλεύει ακόμα σωστά
curl -i localhost:8080/                   # 200
```

---

## Βήμα 2 — Extract το CSS

### Τι μεταφέρεται

Το `<style>` block στο [templates/index.html](templates/index.html) (γραμμές 8–48) φεύγει **ολόκληρο** στο `static/style.css`. Στη θέση του:

```html
<link rel="stylesheet" href="/static/style.css">
```

Προσοχή στο leading `/` — χωρίς αυτό είναι relative path και θα σπάσει σε άλλα routes.

### ⚠️ Το `<script>` ΜΕΝΕΙ

Το script στο τέλος του [index.html](templates/index.html) (γραμμές 95–110) **δεν μεταφέρεται** σε στατικό αρχείο. Περιέχει:

```js
document.getElementById("preview").style.color = "{{.Color}}";
```

Το `{{.Color}}` είναι Go template action. Ένα `.js` αρχείο σερβίρεται από τον FileServer χωρίς να περάσει από το template engine — το `{{.Color}}` θα έφτανε στον browser αυτούσιο σαν string και θα έσπαγε.

Άφησέ το εκεί. **Γράψ' το στο σημείωμα παράδοσης** — είναι ακριβώς το είδος της απόφασης που ο επόμενος πρέπει να ξέρει, αλλιώς θα νομίζει ότι το ξέχασες.

### Τα error templates

Τα `400/404/405/500.html` **δεν έχουν** inline styles — δεν υπάρχει τίποτα να μεταφέρεις.

Θα σε πιάσει ο πειρασμός να τους προσθέσεις κι αυτών το `<link>` «για consistency». **Μην.** Το styling των error pages είναι Stylize C, και το άτομο που θα το κάνει μπορεί να θέλει άλλη δομή (ξεχωριστό αρχείο; κοινό;). Άφησέ του την απόφαση και ανάφερέ το στο σημείωμα.

### Επαλήθευση

Άνοιξε τη σελίδα στον browser. Πρέπει να δείχνει **ακριβώς** όπως πριν. Αν κάτι άλλαξε οπτικά, δεν έκανες μεταφορά — έκανες edit.

Στο DevTools → Network, το `style.css` πρέπει να είναι 200 και όχι 404.

---

## Βήμα 3 — Dockerfile

### Το αρχείο

`Dockerfile` στο root:

```dockerfile
FROM golang:1.25

WORKDIR /app

COPY . .

RUN go build -o server .

EXPOSE 8080

CMD ["./server"]
```

### Γραμμή προς γραμμή

- **`FROM golang:1.25`** — το [go.mod](go.mod) ζητάει `go 1.25.5`. Το official image έχει τον compiler και όλο το toolchain.
- **`WORKDIR /app`** — ορίζει τον φάκελο εργασίας *και* για τα επόμενα βήματα του build *και* για το `CMD` στο runtime. Κρίσιμο, βλ. παρακάτω.
- **`COPY . .`** — αντιγράφει τα πάντα από το build context στο `/app`. Απλό, αλλά χοντρό — το slice B το κάνει selective.
- **`RUN go build -o server .`** — γίνεται τη στιγμή του **build**, το αποτέλεσμα ψήνεται στο image.
- **`EXPOSE 8080`** — τεκμηρίωση, δεν ανοίγει port από μόνο του. Το πραγματικό mapping γίνεται με `-p` στο `docker run`.
- **`CMD ["./server"]`** — τρέχει τη στιγμή του **run**. Exec form (JSON array), όχι shell form.

### 🔴 Το κρίσιμο σημείο — σίγουρη ερώτηση audit

Ο κώδικας διαβάζει αρχεία με **σχετικά paths**:

- [handlers.go](handlers/handlers.go) → `template.ParseFiles("templates/index.html")`
- [ascii.go](ascii/ascii.go) → `filepath.Join("banners", banner+".txt")`

Αυτά επιλύονται ως προς το **working directory του process**, όχι ως προς το πού βρίσκεται το binary. Άρα:

> Το `WORKDIR` του container πρέπει να είναι ο φάκελος όπου κάθονται τα `templates/`, `banners/`, `static/`.

**Ο τρόπος που αποτυγχάνει είναι ύπουλος:** το container ξεκινάει κανονικά, το `docker run` δεν βγάζει error, τα logs δείχνουν «Server running at http://localhost:8080». Και κάθε request γυρνάει 500 ή σπασμένη σελίδα. Δεν το πιάνεις από το build — **μόνο αν κάνεις πραγματικό request**.

Αυτός είναι και ο λόγος που το PRD γράφει «προσοχή στο `COPY`: χρειάζονται server, static, templates, banners». Στο δικό σου `COPY . .` έρχονται όλα μαζί — αλλά ο επόμενος (Docker B, multi-stage) θα πρέπει να τα διαλέξει ένα-ένα, και αν ξεχάσει ένα, σκάει έτσι.

---

## Βήμα 4 — Build, run, επαλήθευση

```bash
docker build -t ascii-art-web .
docker run --rm -p 8080:8080 ascii-art-web
```

Το `-p 8080:8080` = `host:container`. Το `--rm` σβήνει το container όταν σταματήσει.

### Τα τρία tests που πρέπει να περάσουν

Το «σηκώνεται» δεν αρκεί. Χρειάζονται και τα τρία, γιατί το καθένα αποδεικνύει διαφορετικό πράγμα:

```bash
# 1. Το template βρίσκεται
curl -i localhost:8080/

# 2. Τα banners βρίσκονται — αυτό είναι το test που πιάνει το WORKDIR πρόβλημα
curl -i -X POST localhost:8080/ascii-art \
  -d "text=hello" -d "banner=standard" -d "color=black"

# 3. Ο FileServer δουλεύει μέσα στο container
curl -i localhost:8080/static/style.css
```

Το #2 είναι το σημαντικό. Αν επιστρέψει 400 με μήνυμα *"could not open banner file"*, το WORKDIR ή το COPY είναι λάθος.

---

## Βήμα 5 — Σημείωμα παράδοσης

Ο κανόνας §6: **όχι «τι έκανα», αλλά «τι πρέπει να ξέρει ο επόμενος»**. Ο επόμενος στο Docker είναι το Άτομο 1 (slice B: multi-stage).

Τα σημεία που πρέπει οπωσδήποτε να μεταφέρεις:

1. **Τα paths είναι σχετικά ως προς το WORKDIR.** Στο multi-stage θα πρέπει να κάνει `COPY` ρητά τα `server`, `static/`, `templates/`, `banners/` στο runtime stage — αν ξεχάσει ένα, σκάει στο runtime όχι στο build.
2. **Το inline `<script>` έμεινε στο `index.html`** επίτηδες, γιατί χρησιμοποιεί `{{.Color}}`. Δεν είναι ξεχασμένο.
3. **Τα error templates δεν έχουν `<link>` στο style.css** — αφέθηκε για το Stylize C.
4. Το route `/static/` δουλεύει χωρίς αλλαγή στον `handlers.Home`, λόγω longest-prefix matching του ServeMux.

---

## Checklist πριν το merge

- [ ] `static/style.css` υπάρχει και περιέχει **αυτούσιο** το παλιό `<style>` block
- [ ] Το `<style>` block έχει αφαιρεθεί από το `index.html`
- [ ] `<link rel="stylesheet" href="/static/style.css">` με leading slash
- [ ] Το `<script>` παραμένει στο `index.html`
- [ ] Route FileServer στο `main.go` με `StripPrefix`
- [ ] Ο `handlers.Home` **δεν** έχει τροποποιηθεί
- [ ] Η σελίδα δείχνει οπτικά ίδια με πριν (local `go run .`)
- [ ] `Dockerfile` υπάρχει, single-stage
- [ ] **Δεν** υπάρχει `.dockerignore` (slice B)
- [ ] `docker build` περνάει
- [ ] Και τα 3 curl tests περνάνε **μέσα στο container**
- [ ] Σημείωμα παράδοσης γραμμένο
- [ ] `git merge main` πριν το push, conflicts λυμένα στο branch σου

---

## Ερωτήσεις audit — να μπορείς να τις απαντήσεις

**«Γιατί δεν πιάνει το `/` το request για το `/static/style.css`;»**
Ο ServeMux διαλέγει το μακρύτερο pattern που ταιριάζει, όχι το πρώτο που δηλώθηκε.

**«Τι κάνει το StripPrefix και τι θα γινόταν χωρίς αυτό;»**
Ο FileServer ψάχνει το URL path αυτούσιο μέσα στον φάκελο. Χωρίς StripPrefix θα έψαχνε το `static/static/style.css` → 404.

**«Πού ψάχνει το πρόγραμμα τα banner files μέσα στο container;»**
Σχετικά με το working directory του process, δηλαδή το `WORKDIR`. Όχι σχετικά με το πού είναι το binary.

**«Τι διαφορά έχει το RUN από το CMD;»**
`RUN` εκτελείται στο build και το αποτέλεσμα ψήνεται στο image. `CMD` εκτελείται όταν ξεκινά το container.

**«Το EXPOSE ανοίγει το port;»**
Όχι. Είναι τεκμηρίωση. Το πραγματικό mapping το κάνει το `-p` στο `docker run`.

**«Γιατί το script δεν μπήκε σε `.js` αρχείο;»**
Γιατί περιέχει `{{.Color}}`, Go template action. Τα στατικά αρχεία δεν περνούν από το template engine.

---

## Παγίδες

| Παγίδα | Σύμπτωμα |
|---|---|
| Ξέχασες το `StripPrefix` | 404 στο `/static/style.css` |
| `href="static/style.css"` χωρίς leading `/` | Δουλεύει στο `/`, σπάει αλλού |
| Λάθος `WORKDIR` ή ελλιπές `COPY` | Container σηκώνεται κανονικά, requests γυρνάνε 500 |
| Μετέφερες το `<script>` σε `.js` | Το `{{.Color}}` φτάνει σαν literal string στον browser |
| Πείραξες το CSS «λιγάκι» | Ο Stylize A δουλεύει πάνω σε κάτι που δεν περίμενε |
| Έκανες multi-stage γιατί «είναι καλύτερο» | Πήρες τη δουλειά του Ατόμου 1 |
