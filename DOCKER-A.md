# Docker A — από το μηδέν μέχρι το merge

**Άτομο 2 · Γύρος 1 · branch `feature/docker-a-ivogiake`**

Ένας οδηγός, τέσσερα μέρη, δεκατέσσερα βήματα στη σειρά. Δεν χρειάζεται να ξέρεις Docker για να ξεκινήσεις.

---

## Η λίστα προόδου

Τσέκαρε καθώς προχωράς. Αν χαθείς, γύρνα εδώ.

**Μέρος 1 — ο κώδικας** *(χωρίς Docker, ξεμπλοκάρει την ομάδα)*
- [ ] 1. `static/` + route FileServer
- [ ] 2. Μεταφορά του CSS

**Μέρος 2 — μαθαίνοντας Docker** *(~45 λεπτά στο εργαστήριο)*
- [ ] 3. Ο daemon
- [ ] 4. Image vs container
- [ ] 5. Το εγγράψιμο στρώμα
- [ ] 6. Λίστες & πώς σταματάς
- [ ] 7. Το πρώτο σου Dockerfile
- [ ] 8. COPY & WORKDIR
- [ ] 9. RUN vs CMD
- [ ] 10. Layers & cache
- [ ] 11. Ports

**Μέρος 3 — το πραγματικό Dockerfile**
- [ ] 12. Γράψ' το μόνος σου
- [ ] 13. Build, run, τα τρία tests

**Μέρος 4 — παράδοση**
- [ ] 14. Σημείωμα παράδοσης
- [ ] Checklist πριν το merge

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
---

# ΜΕΡΟΣ 1 — Ο κώδικας

Αυτά τα δύο βήματα **δεν χρειάζονται Docker**. Κάν' τα πρώτα: είναι αυτά που ξεμπλοκάρουν τους άλλους, και το barrier του §6 σημαίνει ότι αν αργήσεις, κάθονται και οι τρεις.

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
- **`http.Handle`** (όχι `HandleFunc`) — γιατί ο FileServer επιστρέφει `http.Handler`, όχι συνάρτηση. Το `HandleFunc` δέχεται `func(w, r)`.

**Γιατί χρειάζεται το StripPrefix:** χωρίς αυτό, ένα request για `/static/style.css` φτάνει στον FileServer με path `/static/style.css`, και ο FileServer ψάχνει το `static/static/style.css`. 404. Με το StripPrefix το path γίνεται `style.css` και βρίσκει το `static/style.css`.

### Το θέμα με το `"/"`

Το [handlers.go](handlers/handlers.go) έχει ήδη `http.HandleFunc("/", handlers.Home)`, που πιάνει **όλα** τα paths. Λογικά θα έπρεπε να πιάσει και το `/static/style.css` και να επιστρέψει 404 από τον έλεγχο `if r.URL.Path != "/"`.

Δεν συμβαίνει: ο `ServeMux` της Go διαλέγει το **μακρύτερο pattern που ταιριάζει**, όχι το πρώτο που δηλώθηκε. Το `/static/` κερδίζει το `/`.

Συνέπεια: **δεν χρειάζεται να αγγίξεις τον `handlers.Home`.** Αν σε πιάσει ο πειρασμός να προσθέσεις εξαίρεση για το `/static/` εκεί μέσα, σταμάτα — είναι περιττό και δείχνει ότι δεν κατάλαβες το routing.

### Επαλήθευση

```bash
go run .
```
Σε δεύτερο tab:
```bash
curl -i localhost:8080/static/style.css   # 200, Content-Type: text/css
curl -i localhost:8080/asdf               # 404 — ο Home δουλεύει ακόμα σωστά
curl -i localhost:8080/                   # 200
```

---

## Βήμα 2 — Extract το CSS

### Τι μεταφέρεται

Το `<style>` block στο [templates/index.html](templates/index.html) φεύγει **ολόκληρο** στο `static/style.css`. Στη θέση του:

```html
<link rel="stylesheet" href="/static/style.css">
```

Προσοχή στο leading `/` — χωρίς αυτό είναι relative path και θα σπάσει σε άλλα routes.

### ⚠️ Το `<script>` ΜΕΝΕΙ

Το script στο τέλος του [index.html](templates/index.html) **δεν μεταφέρεται** σε στατικό αρχείο. Περιέχει:

```js
document.getElementById("preview").style.color = "{{.Color}}";
```

Το `{{.Color}}` είναι Go template action. Ένα `.js` αρχείο σερβίρεται από τον FileServer **χωρίς να περάσει από το template engine** — το `{{.Color}}` θα έφτανε στον browser αυτούσιο σαν string και θα έσπαγε.

Άφησέ το εκεί. **Γράψ' το στο σημείωμα παράδοσης** — είναι ακριβώς το είδος της απόφασης που ο επόμενος πρέπει να ξέρει, αλλιώς θα νομίζει ότι το ξέχασες.

### Τα error templates

Τα `400/404/405/500.html` **δεν έχουν** inline styles — δεν υπάρχει τίποτα να μεταφέρεις.

Θα σε πιάσει ο πειρασμός να τους προσθέσεις κι αυτών το `<link>` «για consistency». **Μην.** Το styling των error pages είναι Stylize C, και το άτομο που θα το κάνει μπορεί να θέλει άλλη δομή. Άφησέ του την απόφαση και ανάφερέ το στο σημείωμα.

### Επαλήθευση

Άνοιξε τη σελίδα στον browser. Πρέπει να δείχνει **ακριβώς** όπως πριν. Αν κάτι άλλαξε οπτικά, δεν έκανες μεταφορά — έκανες edit.

Στο DevTools → Network, το `style.css` πρέπει να είναι 200 και όχι 404.

---
---

# ΜΕΡΟΣ 2 — Μαθαίνοντας Docker

Από δω και κάτω δουλεύεις σε **ξεχωριστό φάκελο πειραμάτων**, όχι στο project. Θα γυρίσουμε στο project στο βήμα 12.

**Πώς να το διαβάσεις:** κάθε βήμα έχει *Τρέξε → Παρατήρησε → Γιατί*. Μην προσπεράσεις το «Παρατήρησε». Το νόημα του Docker δεν είναι στη σύνταξη, είναι στο **πότε συμβαίνει τι** — και αυτό φαίνεται μόνο τρέχοντάς το.

---

## Βήμα 3 — Ξεκίνα τον daemon

Το `docker` στο terminal είναι απλώς ένας **client**. Στέλνει εντολές σε ένα background process (τον daemon) που κάνει τη δουλειά. Χωρίς αυτόν, το CLI υπάρχει αλλά δεν κάνει τίποτα.

Άνοιξε το **Docker Desktop** από το Launchpad ή με **Cmd+Space** → `Docker`. Στη γραμμή μενού (πάνω δεξιά στην οθόνη) εμφανίζεται εικονίδιο **φάλαινας**:

- **φάλαινα που κινείται** → ακόμα ξεκινάει, περίμενε
- **σταθερή φάλαινα** → έτοιμο

Την πρώτη φορά θέλει 30–60 δευτερόλεπτα. Μετά:

```bash
docker info
```

Αν βγάλει πληροφορίες, είσαι έτοιμος. Αν βγάλει `Cannot connect to the Docker daemon`, δεν έχει σηκωθεί ακόμα — περίμενε και ξαναδοκίμασε.

> **Μην κλείσεις το Docker Desktop** όσο δουλεύεις.

---

## Βήμα 4 — Το μοντέλο: image vs container

Αυτή είναι η μία έννοια που αν την καταλάβεις, όλα τα υπόλοιπα είναι λεπτομέρειες.

- **Image** = ένα παγωμένο, read-only στιγμιότυπο ενός filesystem + η εντολή που πρέπει να τρέξει. Δεν εκτελείται. Είναι αρχείο.
- **Container** = μια *εκτέλεση* ενός image. Παίρνει το filesystem του image και του βάζει από πάνω ένα εγγράψιμο στρώμα.

Σχέση: **image → container** όπως **class → object**, ή **εκτελέσιμο αρχείο → process**. Ένα image, πολλά containers.

### Τρέξε

```bash
docker run --rm -it alpine sh
```

### 🧭 Προσανατολισμός — διάβασέ το πριν πανικοβληθείς

Το prompt σου άλλαξε σε κάτι σαν αυτό:

```
/ #
│ │
│ └── # = είσαι root (μέσα σε container είσαι πάντα root)
└──── / = ο τρέχων φάκελος — εδώ, η ρίζα του filesystem
```

**Δεν κόλλησε τίποτα. Είσαι μέσα στο container.** Το `/ #` είναι απλώς το prompt του.

Από δω και πέρα, **ό,τι πληκτρολογείς εκτελείται μέσα στο container** — όχι στο μηχάνημά σου. Είναι σαν να άνοιξες terminal σε άλλον υπολογιστή.

| Θέλω να... | Εντολή |
|---|---|
| **βγω έξω** | `exit` (ή `Ctrl+D`) |
| δω πού είμαι | `pwd` |
| δω ποιος είμαι | `whoami` |
| δω τι υπάρχει | `ls /` |

Το `exit` σταματάει το container· το `--rm` το σβήνει· γυρνάς στο κανονικό σου prompt. **Δεν μένει τίποτα πίσω και δεν χαλάει τίποτα.**

#### Τι κατέβασες

Το image **alpine** — μινιμαλιστικό Linux, ~8 MB — από το Docker Hub, το δημόσιο registry. Κατέβηκε αυτόματα επειδή δεν το είχες· την επόμενη φορά θα είναι ακαριαίο.

**Δεν εγκαταστάθηκε τίποτα στο λειτουργικό σου.** Κάθεται σαν αρχείο στον χώρο του Docker. Το βλέπεις με `docker images` και το σβήνεις με `docker rmi alpine`.

### Κοίτα γύρω σου

```sh
pwd                    # πού είμαι
whoami                 # ποιος είμαι → root
ls /                   # τι υπάρχει → bin, etc, usr... ένα ολόκληρο Linux
cat /etc/os-release    # τι λειτουργικό είναι → Alpine, όχι macOS
ps aux                 # ποια processes τρέχουν → σχεδόν κανένα
```

Το `ps aux` είναι το πιο διαφωτιστικό. Στο μηχάνημά σου τρέχουν εκατοντάδες processes· εδώ βλέπεις **το `sh` σου και τίποτε άλλο**. Αυτό ακριβώς σημαίνει «απομονωμένο».

### Το πείραμα

```sh
echo "γεια" > /test.txt
ls -la /test.txt
exit
```

Τώρα ξανατρέξε **την ίδια ακριβώς εντολή** και ψάξε το αρχείο σου:

```bash
docker run --rm -it alpine sh
```
```sh
ls /test.txt
```

### Παρατήρησε

Το `/test.txt` **δεν υπάρχει**. Χάθηκε.

### Γιατί

Έγραψες στο εγγράψιμο στρώμα του *πρώτου* container. Το `--rm` το έσβησε στο exit. Το δεύτερο `docker run` έφτιαξε **καινούργιο** container από το ίδιο αμετάβλητο image.

> **Το image δεν αλλάζει ποτέ από τη χρήση.** Ό,τι γράφει ένα container ζει και πεθαίνει μαζί του. Αν θες κάτι μόνιμα μέσα στο image, μπαίνει τη στιγμή του **build**, όχι τη στιγμή του run.

Κράτα αυτή τη φράση. Είναι η ρίζα του βήματος 9.

### Τα flags

| Flag | Τι κάνει |
|---|---|
| `--rm` | σβήσε το container όταν σταματήσει (αλλιώς μαζεύονται) |
| `-it` | interactive + terminal — χρειάζεται για να «μπεις μέσα» |
| `alpine` | το image |
| `sh` | override της default εντολής του image |

---

## Βήμα 5 — Τι ακριβώς είναι το «εγγράψιμο στρώμα»

### Το image είναι στοίβα, όχι αρχείο

Ένα layer ανά εντολή του Dockerfile, **όλα read-only**:

```
┌─────────────────────────────┐
│ layer 3:  CMD metadata      │  ← read-only
├─────────────────────────────┤
│ layer 2:  RUN go build →    │  ← read-only
│           /app/server       │
├─────────────────────────────┤
│ layer 1:  COPY . . →        │  ← read-only
│           /app/*            │
├─────────────────────────────┤
│ layer 0:  FROM golang       │  ← read-only
│           /bin /usr /etc... │
└─────────────────────────────┘
```

Το container προσθέτει **ένα** στρώμα — το μόνο εγγράψιμο:

```
┌─────────────────────────────┐
│ container layer    ✏️ RW    │  ← ζει όσο το container
├═════════════════════════════┤
│         image (RO)          │
└─────────────────────────────┘
```

### Δεν το κάνει «εκτελέσιμο»

Συνηθισμένη παρανόηση. Το image δεν του λείπει κάτι για να τρέξει — του λείπει ένα μέρος για να **γράψει**.

Την εκτέλεση την κάνει ο **Linux kernel**: το Docker ξεκινάει κανονικό process, του δίνει αυτό το filesystem ως ρίζα (`/`), και το απομονώνει με namespaces + cgroups. Το εγγράψιμο στρώμα αφορά **αποθήκευση**, όχι εκτέλεση.

### Union filesystem & copy-on-write

Το process δεν βλέπει στοίβα. Βλέπει **ένα ενιαίο filesystem**: το overlayfs συγχωνεύει τα layers και, όταν ζητάς ένα αρχείο, σου δίνει την πρώτη εκδοχή που θα βρει ψάχνοντας από πάνω προς τα κάτω.

| Ενέργεια | Τι συμβαίνει πραγματικά |
|---|---|
| Διαβάζεις αρχείο | Σερβίρεται κατευθείαν από το layer του image. Καμία αντιγραφή. |
| Γράφεις **νέο** αρχείο | Δημιουργείται στο container layer |
| Τροποποιείς **υπάρχον** αρχείο του image | **Αντιγράφεται πρώτα προς τα πάνω**, μετά αλλάζει το αντίγραφο. Το πρωτότυπο μένει άθικτο. |
| Σβήνεις αρχείο του image | Μπαίνει *whiteout* marker που το κρύβει. **Το αρχείο εξακολουθεί να υπάρχει** από κάτω. |

Η κοκκομετρία είναι το **αρχείο**, όχι το byte: αλλάζεις 1 byte σε αρχείο 500 MB, αντιγράφονται και τα 500 MB.

### Τρέξε — «κατάστρεψε» ένα αρχείο συστήματος

```bash
docker run --rm -it alpine sh
```
```sh
cat /etc/os-release          # υπάρχον αρχείο του image
echo "χαλασμένο" > /etc/os-release
cat /etc/os-release          # άλλαξε
exit
```

Ξανατρέξε το ίδιο και κάνε `cat /etc/os-release`.

### Παρατήρησε

**Ανέπαφο.** Πέταξες ένα αρχείο συστήματος και το image δεν το πήρε χαμπάρι.

### Γιατί σε αφορά

- **Το image δεν μπορεί να μολυνθεί από τη χρήση** — κυριολεκτικά δεν γράφεται. Γι' αυτό είναι αναπαραγώγιμο.
- **Δέκα containers μοιράζονται τα ίδια layers** στον δίσκο. Γι' αυτό το `docker run` είναι τόσο φθηνό.
- **Στο project σου:** ο `server` και τα `banners/` ζουν σε read-only layers και διαβάζονται χωρίς αντιγραφή. Αν κάποτε η εφαρμογή γράψει αρχείο, θα πάει στο εφήμερο στρώμα και θα **εξαφανιστεί** στο restart. Αν πρέπει να επιβιώσει, θέλει *volume* — άλλο κεφάλαιο.

---

## Βήμα 6 — Λίστες, και πώς σταματάς

```bash
docker images     # τα images (τα «παγωμένα»)
docker ps         # containers που τρέχουν ΤΩΡΑ
docker ps -a      # + όσα σταμάτησαν και δεν σβήστηκαν
```

Δύο ξεχωριστές λίστες γιατί είναι δύο ξεχωριστά πράγματα. Όταν κάτι δεν δουλεύει, το πρώτο ερώτημα είναι πάντα: *έσκασε το build (image) ή έσκασε το run (container);*

### Πώς διαβάζεται το `docker ps`

```
CONTAINER ID   IMAGE   COMMAND   STATUS         PORTS                    NAMES
a3f9c2e1b8d4   lab     "sh"      Up 2 minutes   0.0.0.0:8080->8080/tcp   nifty_curie
```

- **CONTAINER ID** — όπου ζητείται `<id>`, **αρκούν τα 3-4 πρώτα ψηφία**
- **NAMES** — τυχαίο όνομα που δίνει το Docker· δουλεύει παντού όπου δουλεύει και το ID
- **STATUS** — `Up ...` τρέχει, `Exited (0)` τερμάτισε κανονικά, `Exited (1)` με σφάλμα
- **PORTS** — σκέτο `8080/tcp` σημαίνει **μόνο `EXPOSE`**, καμία γέφυρα. Το βέλος `0.0.0.0:8080->8080/tcp` σημαίνει ότι έδωσες `-p`.

### 🛑 Πώς σταματάς ένα container

**Αν τρέχει στο προσκήνιο** (το terminal σου δείχνει logs): **Ctrl+C**

**Αν τρέχει αλλού:**
```bash
docker ps                  # βρες το CONTAINER ID
docker stop a3f9
```

**Αν αργεί ή δεν σταματάει:** `docker kill a3f9`

Γιατί μερικές φορές αργεί 10 δευτερόλεπτα: ο Linux kernel παραδίδει σήμα στο **PID 1** μόνο αν το process έχει εγκαταστήσει ρητά handler. Προγράμματα χωρίς handler (π.χ. ο `httpd` του busybox) αγνοούν το `Ctrl+C` και το `docker stop`· το Docker περιμένει 10 δευτερόλεπτα και μετά στέλνει `SIGKILL`, που δεν αγνοείται.

**Ο Go server σου ανταποκρίνεται κανονικά** — το Go runtime εγκαθιστά handlers μόνο του.

**Σταμάτα τα πάντα:** `docker stop $(docker ps -q)`

---

## Βήμα 7 — Το πρώτο σου Dockerfile

Φάκελος πειραμάτων, **εκτός** του project:

```bash
mkdir -p ~/docker-lab
cd ~/docker-lab
pwd                        # /Users/<εσύ>/docker-lab
```

Θα φτιάξεις αρχείο με όνομα **`Dockerfile`** — κεφαλαίο D, **χωρίς κατάληξη**, χωρίς τελεία πουθενά.

### 📝 Πώς φτιάχνεις το αρχείο

#### Τρόπος Α — VSCode

```bash
code ~/docker-lab
```

1. Στο **Explorer** (αριστερή στήλη), δίπλα στο όνομα του φακέλου, το εικονίδιο **New File** — σελίδα με `+`
2. Γράψε `Dockerfile` και **Enter**
3. Γράψε το περιεχόμενο
4. **Cmd+S**

**Σημάδι ότι το πέτυχες:** κάτω δεξιά στη status bar γράφει *Dockerfile* (όχι *Plain Text*), και το `FROM`/`CMD` χρωματίζονται.

Αν η εντολή `code` δεν βρίσκεται: **Cmd+Shift+P** → `shell command` → *Install 'code' command in PATH*.

#### Τρόπος Β — Terminal, μία εντολή

```bash
cat > Dockerfile << 'EOF'
FROM alpine
CMD ["echo", "γεια από το container"]
EOF
```

Το `cat > Dockerfile` σημαίνει «γράψε στο αρχείο». Το `<< 'EOF'` σημαίνει «πάρε ό,τι ακολουθεί μέχρι να δεις τη λέξη EOF». Γράφεις τις γραμμές, μετά **`EOF` μόνη της σε δική της γραμμή**.

#### Τρόπος Γ — nano

```bash
nano Dockerfile
```
**Ctrl+O** → Enter (αποθήκευση), **Ctrl+X** (έξοδος).

### ⚠️ Η παγίδα του macOS — μη χρησιμοποιήσεις TextEdit

**Rich text:** σώζει αόρατη μορφοποίηση, το Docker βλέπει σκουπίδια.

**Κρυφή κατάληξη:** το macOS κρύβει τις καταλήξεις. Σώζεις «Dockerfile» και γράφεται **`Dockerfile.txt`**. Στο Finder φαίνεται σωστό, και το `docker build` λέει:

```
ERROR: failed to read dockerfile: open Dockerfile: no such file or directory
```

### Επαλήθευση

```bash
ls -la              # ψάξε γραμμή που τελειώνει σε ' Dockerfile'
cat Dockerfile      # πρέπει να δεις ακριβώς ό,τι έγραψες
file Dockerfile     # πρέπει να λέει 'ASCII text' ή 'Unicode text'
```

Αν το όνομα βγήκε λάθος: `mv Dockerfile.txt Dockerfile`

### Το περιεχόμενο

```dockerfile
FROM alpine
CMD ["echo", "γεια από το container"]
```

### Τρέξε

```bash
docker build -t lab .
docker run --rm lab
```

### Παρατήρησε

Δύο ξεχωριστές φάσεις. Το `build` παρήγαγε image (`docker images` → θα δεις το `lab`). Το `run` έφτιαξε container από αυτό.

### Γιατί

- **`FROM`** — από ποιο υπάρχον image ξεκινάς. Πάντα η πρώτη εντολή.
- **`CMD`** — η **default εντολή** που τρέχει όταν ξεκινά container. Δεν εκτελείται στο build.
- **`-t lab`** — δίνει όνομα (tag) στο image. Χωρίς αυτό παίρνει μόνο hash.
- **`.`** στο τέλος — το **build context**: ο φάκελος που στέλνεται στον daemon. Το χρειάζεται το `COPY`.

### Δοκίμασε

```bash
docker run --rm lab echo "κάτι άλλο"
```

Ό,τι γράψεις μετά το όνομα του image **αντικαθιστά** το `CMD`. Γι' αυτό λέγεται *default*.

> ℹ️ Με το containerd image store (η νεότερη προεπιλογή), κάποιες εντολές θέλουν **ρητό tag**: το `docker inspect lab` αποτυγχάνει ενώ το `docker inspect lab:latest` δουλεύει. Συνήθισε να γράφεις το tag.

---

## Βήμα 8 — COPY και WORKDIR

🔴 **Το πιο σημαντικό βήμα.** Εδώ κρύβεται το bug που θα σε φάει στο βήμα 13.

```bash
cd ~/docker-lab
mkdir data
echo "περιεχόμενο" > data/file.txt
```

> 📌 **«Άλλαξε το Dockerfile» σημαίνει: αντικατέστησε ΟΛΟ το περιεχόμενο.**
> Δουλεύουμε πάντα στο **ίδιο** αρχείο, σβήνοντας ό,τι είχε πριν. Μην προσθέτεις από κάτω — θα καταλήξεις με πολλαπλά `FROM`.
> Στο VSCode: **Cmd+A** → γράψε από πάνω → **Cmd+S**.

```dockerfile
FROM alpine
COPY . .
CMD ["ls", "-la"]
```

```bash
docker build -t lab . && docker run --rm lab
```

### Παρατήρησε

Τα αρχεία σου προσγειώθηκαν στο `/` (τη ρίζα), ανακατεμένα με `bin`, `etc`, `usr`. Δεν όρισες πού, οπότε πήγαν στο default working directory.

### Τώρα πρόσθεσε WORKDIR

```dockerfile
FROM alpine
WORKDIR /app
COPY . .
CMD ["ls", "-la"]
```

```bash
docker build -t lab . && docker run --rm lab
```

Τώρα βλέπεις **μόνο** τα δικά σου αρχεία. Καθαρά.

### Γιατί — και εδώ είναι το κρίσιμο

Το `WORKDIR` κάνει **δύο** πράγματα, όχι ένα:

1. **Στο build:** ορίζει πού προσγειώνονται τα `COPY` και πού τρέχουν τα `RUN`
2. **Στο run:** ορίζει το **working directory του process**

Το #2 είναι που σε αφορά. Απόδειξέ το:

```dockerfile
FROM alpine
WORKDIR /app
COPY . .
CMD ["cat", "data/file.txt"]
```

```bash
docker build -t lab . && docker run --rm lab
```

Δουλεύει — το `data/file.txt` είναι σχετικό path και επιλύεται ως προς το `/app`.

#### ⚠️ Προσοχή: δεν αρκεί να αλλάξεις το WORKDIR

Ίσως σκεφτείς «βάζω `/somewhere-else` αντί για `/app` και θα σκάσει». **Δεν θα σκάσει** — και ο λόγος είναι το ίδιο το μάθημα.

Αλλάζοντας το ένα `WORKDIR` μετακινείς **και τα δύο** πράγματα ταυτόχρονα: ο προορισμός του `COPY` (η δεύτερη τελεία) είναι σχετικός ως προς το WORKDIR, οπότε **τα αρχεία ακολουθούν**. Μετακίνησες και το σπίτι και τον ένοικο.

Για να δεις την αποτυχία πρέπει να **χωρίσεις** τις δύο λειτουργίες — δεύτερο `WORKDIR` **μετά** το `COPY`:

```dockerfile
FROM alpine
WORKDIR /app
COPY . .
WORKDIR /somewhere-else
CMD ["cat", "data/file.txt"]
```

```
cat: can't open 'data/file.txt': No such file or directory
```

Τα αρχεία προσγειώθηκαν στο `/app` — το WORKDIR που ίσχυε **τη στιγμή του `COPY`**. Το process όμως ξεκινά στο `/somewhere-else`.

```bash
docker run --rm -it lab sh
```
```sh
pwd                    # /somewhere-else — άδειο
ls /app/data           # file.txt — εδώ ήταν όλη την ώρα
exit
```

Ο κώδικας δεν άλλαξε. Το αρχείο **υπάρχει** στο image. Απλώς κανείς δεν το ψάχνει εκεί.

> **Αυτό ακριβώς παθαίνει η εφαρμογή σου.** Το [handlers.go](handlers/handlers.go) κάνει `template.ParseFiles("templates/index.html")` και το [ascii.go](ascii/ascii.go) κάνει `filepath.Join("banners", ...)`. Σχετικά paths, και τα δύο.

### `COPY . .` — τι σημαίνουν οι δύο τελείες

Πρώτη = πηγή, μέσα στο build context (host). Δεύτερη = προορισμός, μέσα στο image, **σχετικά με το WORKDIR** — γι' αυτό το `WORKDIR` πρέπει να προηγείται.

### Το working directory δεν είναι έννοια του shell

Ο kernel κρατάει για **κάθε process** ένα cwd, όπως κρατάει PID και χρήστη. Το `cd` δεν είναι μαγικό — καλεί `chdir()`. Το Docker κάνει ακριβώς το ίδιο, χωρίς shell:

```c
chdir("/app")                              // ⟵ το WORKDIR
execve("/app/server", ["./server"])        // ⟵ το CMD
```

Το cwd **επιβιώνει του `execve`**, οπότε το πρόγραμμα ξεκινάει ήδη «μέσα» στο `/app`.

**Το όρισμα δεν μετακινεί το πρόγραμμα.** Το `cat /etc/hosts` δεν «ανοίγει το cat στο /etc» — το cat μένει όπου είναι και ζητάει ένα path. Απόδειξη:

```bash
cd /tmp && touch marker.txt
ls marker.txt /etc/hosts     # και τα δύο βρίσκονται
```

Αν το `ls` «άνοιγε στο /etc», το σχετικό `marker.txt` θα αποτύγχανε.

**Απόλυτο path = ανοσία στο WORKDIR. Σχετικό path = εξάρτηση από το WORKDIR.**

---

## Βήμα 9 — RUN vs CMD

```dockerfile
FROM alpine
WORKDIR /app
RUN echo "ΕΓΩ ΤΡΕΧΩ ΣΤΟ BUILD" > /app/proof.txt
CMD ["cat", "/app/proof.txt"]
```

```bash
docker build -t lab .          # κοίτα προσεκτικά το output
docker run --rm lab
docker run --rm lab            # και δεύτερη φορά
```

### Παρατήρησε

Το `RUN` εκτελέστηκε **μία φορά**, στο build. Το `proof.txt` υπάρχει τώρα **μέσα στο image**, οπότε κάθε container το βρίσκει έτοιμο. Το `CMD` έτρεξε δύο φορές — μία ανά container.

Πρόσεξε επίσης: το build αριθμεί **3 βήματα**, όχι 4. Το `CMD` δεν εκτελεί τίποτα — γράφεται μόνο ως μεταδεδομένο. Δες το:

```bash
docker inspect lab:latest | jq '.[0].Config.Cmd'
docker history lab:latest        # η γραμμή του CMD είναι 0B
```

### Γιατί

| | Πότε | Το αποτέλεσμα |
|---|---|---|
| `RUN` | build | ψήνεται μόνιμα στο image |
| `CMD` | run | ζει όσο το container |

Θυμήσου το βήμα 4: *ό,τι θες μόνιμα, μπαίνει στο build.*

### Η ασυμμετρία

```dockerfile
RUN echo "α"         # ✅ τρέχει
RUN echo "β"         # ✅ τρέχει — όσα RUN θέλεις, σωρευτικά

CMD ["echo", "α"]    # αγνοείται
CMD ["echo", "β"]    # ⟵ ΜΟΝΟ το τελευταίο ισχύει
```

Το `RUN` είναι **ενέργεια** — όλες γίνονται. Το `CMD` είναι **ιδιότητα** — έχει μία τιμή, και η τελευταία ανάθεση κερδίζει.

### Πολλές εντολές σε ένα CMD

Η μορφή με αγκύλες **δεν είναι γραμμή εντολών** — είναι `[πρόγραμμα, όρισμα, όρισμα]`. Το `&&` δεν είναι εντολή· είναι σύμβολο που καταλαβαίνει **το shell**, και εκεί δεν υπάρχει shell.

```dockerfile
CMD cat a.txt && ls -la                    # shell form — το Docker προσθέτει sh -c
CMD ["sh", "-c", "cat a.txt && ls -la"]    # το ίδιο, ρητά
```

Το `sh` είναι **πρόγραμμα** — αυτό που τρέχει όταν ανοίγεις Terminal. Το `-c` του λέει «να η γραμμή, τρέξ' την και βγες». Το χρειάζεσαι μόνο για `&&`, `;`, `|`, `>`, `*`, `$VAR`.

Στο build δεν ισχύει τίποτα από αυτά: το `RUN` είναι shell form **από προεπιλογή**, γι' αυτό το `&&` δουλεύει εκεί χωρίς να γράψεις τίποτα.

### Η εφαρμογή στο project σου

Η μεταγλώττιση Go τη θέλεις **μία φορά** → `RUN`. Την εκκίνηση του server τη θέλεις **σε κάθε container** → `CMD`.

Αν βάλεις το `go build` σε `CMD`, θα μεταγλωττίζει από την αρχή σε κάθε εκκίνηση. Δουλεύει — και είναι λάθος.

---

## Βήμα 10 — Layers και cache

Κάθε εντολή φτιάχνει ένα **layer**. Το Docker τα κρατά cached και ξαναχρησιμοποιεί όσα δεν άλλαξαν.

Κράτα το Dockerfile του βήματος 9 και κάνε **τρία** builds:

**1ο — χωρίς αλλαγή:**
```bash
docker build -t lab .
```
`CACHED` σχεδόν παντού, ακαριαίο.

**2ο — άλλαξε μόνο την τελευταία γραμμή:**
```dockerfile
CMD ["echo", "άλλαξα μόνο το CMD"]
```
Cached όλα εκτός από την τελευταία.

**3ο — άλλαξε την 3η γραμμή:**
```dockerfile
RUN echo "ΔΙΑΦΟΡΕΤΙΚΟ ΚΕΙΜΕΝΟ" > /app/proof.txt
```

### Παρατήρησε

Στο 3ο ξαναέτρεξε το `RUN` **και όλα όσα ήταν από κάτω** — παρόλο που το `CMD` δεν το άγγιξες.

### Γιατί

Το cache ακυρώνεται **αλυσιδωτά**: κάθε layer χτίζεται πάνω στο προηγούμενο, δεν μπορεί να μείνει έγκυρο πάνω σε άλλη βάση.

**Πρακτικός κανόνας:** ό,τι αλλάζει σπάνια, πάνω. Ό,τι αλλάζει συχνά, κάτω.

### ⚠️ Το κλασικό λάθος

```dockerfile
RUN apk add git          # +40 MB
RUN apk del git          # ⟵ ΔΕΝ μικραίνει το image
```

Θυμήσου το βήμα 5: σβήνοντας αρχείο κατώτερου layer μπαίνει *whiteout marker*. Το αρχείο **παραμένει** από κάτω. Το image έχει και τα 40 MB **και** τους markers.

Για να μετρήσει, εγκατάσταση και καθάρισμα στο **ίδιο** layer:
```dockerfile
RUN apk add git && ...χρήση... && apk del git
```

---

## Βήμα 11 — Ports

Εδώ χρειάζεσαι **δύο terminal tabs**: στο ένα τρέχει ο server, στο άλλο δοκιμάζεις. Νέο tab με **Cmd+T**, ή το `+` στο panel του VSCode.

> Ένας server δεν τερματίζει — μένει να ακούει. Το πρώτο tab θα «κολλήσει» δείχνοντας logs. **Αυτό είναι το σωστό.**

```dockerfile
FROM alpine
RUN apk add --no-cache busybox-extras
RUN mkdir /www && echo "<h1>δουλευει</h1>" > /www/index.html
EXPOSE 9999
CMD ["httpd", "-f", "-p", "9999", "-h", "/www"]
```

- **`apk add busybox-extras`** — το `apk` είναι ο package manager του Alpine. Το σκέτο `alpine` **δεν** περιλαμβάνει τον `httpd`. Χωρίς αυτή τη γραμμή: `exec: "httpd": executable file not found in $PATH`. Πρόσεξε ότι είναι `RUN` — εγκατάσταση μία φορά, στο build.
- **`--no-cache`** — μην κρατήσεις τον κατάλογο πακέτων μέσα στο image.
- **`-f`** — *foreground*. Χωρίς αυτό ο httpd φεύγει στο παρασκήνιο, το PID 1 τερματίζει, και το container κλείνει αμέσως.

### Πρώτα ΧΩΡΙΣ `-p`

**Tab 1:**
```bash
docker build -t lab .
docker run --rm lab
```

**Tab 2:**
```bash
docker ps            # PORTS → σκέτο 9999/tcp, χωρίς βέλος
curl localhost:9999
```
```
curl: (7) Failed to connect to localhost port 9999: Connection refused
```

### Τώρα ΜΕ `-p`

Tab 1: **Ctrl+C** (ή `docker kill` — ο httpd αγνοεί τα σήματα, βλ. βήμα 6), μετά:
```bash
docker run --rm -p 9999:9999 lab
```

Tab 2:
```bash
docker ps            # PORTS → 0.0.0.0:9999->9999/tcp
curl localhost:9999  # <h1>δουλευει</h1>
```

### Παρατήρησε

Το `EXPOSE 9999` υπήρχε **και στις δύο** περιπτώσεις. Το image ήταν το ίδιο. Μόνο του δεν έκανε απολύτως τίποτα.

### Γιατί

- **`EXPOSE`** = τεκμηρίωση. **Δεν ανοίγει τίποτα.** Στο `docker ps` εμφανίζεται ως σκέτο `9999/tcp`.
- **`-p host:container`** = το πραγματικό mapping. Εμφανίζεται με **βέλος**.

Το container έχει δικό του δίκτυο. Όταν ο server σου λέει *«Server running at localhost:8080»*, αυτό το `localhost` είναι **του container**.

Η δεξιά πλευρά του `-p` πρέπει να ταιριάζει με το πού ακούει ο κώδικας. Η αριστερή είναι δική σου επιλογή: `-p 3000:8080` σημαίνει ότι ανοίγεις **http://localhost:3000**.

---
---

# ΜΕΡΟΣ 3 — Το πραγματικό Dockerfile

## Βήμα 12 — Γράψ' το μόνος σου

Τέλος τα πειράματα. Πίσω στο project:

```bash
cd ~/Git/Zone01/ascii-art-web-stylize
ls
```

Το `ls` πρέπει να δείξει `main.go`, `go.mod`, `handlers/`, `ascii/`, `templates/`, `banners/`, `static/`. Το `Dockerfile` μπαίνει **δίπλα στο `main.go`**, στη ρίζα.

Απάντησε τις ερωτήσεις με τη σειρά — **κάθε απάντηση είναι μία γραμμή**.

**1. Από ποιο image ξεκινάς;**
Χρειάζεσαι τον Go compiler *μέσα* στο image. Το [go.mod](go.mod) λέει την έκδοση· το επίσημο image λέγεται `golang` και τα tags του είναι εκδόσεις.
→ *Καρφωμένη έκδοση (`1.25.5`) και όχι `latest`: ο auditor θα χτίσει σε άλλη στιγμή, σε άλλο μηχάνημα, και πρέπει να πάρει τον ίδιο compiler.*
→ *Μη βάλεις `-alpine`. Η μείωση μεγέθους είναι παραδοτέο του Docker B.*

**2. Πού μέσα στο image θα ζήσει το project;**
→ *Βήμα 8: ορίζει και το working directory του process. Μην αφήσεις τα αρχεία στη ρίζα ανακατεμένα με `bin`, `etc`, `usr`. Το `/app` είναι σύμβαση.*

**3. Πώς μπαίνουν τα αρχεία μέσα;**
→ *Στο slice A αρκεί το χοντρικό. Το selective και το `.dockerignore` είναι δουλειά του Docker B.*

**4. Πότε γίνεται η μεταγλώττιση;**
Μία φορά ή σε κάθε εκκίνηση; → *βήμα 9.* Η εντολή είναι `go build -o server .`
→ *Το `-o server` ονομάζει το binary. Χωρίς αυτό θα λεγόταν `ascii-art-web`, από το module name.*

**5. Πώς δηλώνεις το port;**
→ *βήμα 11. Αντιγράφει ό,τι κάνει ο κώδικας — δες το [main.go](main.go). Δεν το αποφασίζει.*

**6. Τι τρέχει όταν ξεκινά το container;**
→ *Το binary του βήματος 4, ως JSON array. Το `./` δεν είναι διακοσμητικό: σκέτο όνομα ψάχνεται στο `$PATH`, όπου το `/app` δεν ανήκει. Με κάθετο, αντιμετωπίζεται ως διαδρομή και επιλύεται ως προς το cwd.*

---

## Βήμα 13 — Build, run, τα τρία tests

Πάλι **δύο tabs**.

**Tab 1:**
```bash
docker build -t ascii-art-web .
docker run --rm -p 8080:8080 ascii-art-web
```

Το πρώτο build αργεί — κατεβάζει το `golang` image. Όταν δεις `Server running at http://localhost:8080`, το terminal μένει κρεμασμένο. **Σωστό.**

### 🔴 Το κρίσιμο σημείο — σίγουρη ερώτηση audit

Ο κώδικας διαβάζει αρχεία με **σχετικά paths**:

- [handlers.go](handlers/handlers.go) → `template.ParseFiles("templates/index.html")`
- [ascii.go](ascii/ascii.go) → `filepath.Join("banners", banner+".txt")`

Επιλύονται ως προς το **working directory του process**, όχι ως προς το πού είναι το binary. Άρα:

> Το `WORKDIR` πρέπει να είναι ο φάκελος όπου κάθονται τα `templates/`, `banners/`, `static/`.

**Ο τρόπος που αποτυγχάνει είναι ύπουλος:** το container ξεκινάει κανονικά, τα logs δείχνουν «Server running», και κάθε request γυρνάει 500. Δεν το πιάνεις από το build — **μόνο αν κάνεις πραγματικό request**.

### Τα τρία tests — από το Tab 2

```bash
curl -i localhost:8080/
```
```bash
curl -i -X POST localhost:8080/ascii-art -d "text=hello" -d "banner=standard"
```
```bash
curl -i localhost:8080/static/style.css
```

| Test | Τι αποδεικνύει |
|---|---|
| 1ο | βρίσκει το `templates/index.html` |
| 2ο | βρίσκει τα `banners/*.txt` |
| 3ο | δουλεύει ο FileServer με το `static/` |

**Το δεύτερο είναι το σημαντικό** — το μόνο που αναγκάζει τον server να διαβάσει αρχείο με σχετικό path. Αν πάρεις `400` με *"could not open banner file"*, το `WORKDIR` ή το `COPY` σου είναι λάθος.

> Το `-i` δείχνει και τα HTTP headers. Η πρώτη γραμμή (`HTTP/1.1 200 OK`) είναι αυτό που θες να δεις.

Όταν τελειώσεις, **Ctrl+C** στο Tab 1.

### Debugging

**Μπες μέσα και κοίτα με τα μάτια σου:**
```bash
docker run --rm -it ascii-art-web sh
```
```sh
pwd          # πού είμαι; ⟵ αυτό είναι το WORKDIR σου
ls -la       # υπάρχουν templates/ banners/ static/ server;
```

Το `sh` παρακάμπτει το `CMD`. Το `pwd` + `ls` απαντάει το 90% των προβλημάτων.

**Δες τι έγινε στο build:**
```bash
docker build --no-cache -t ascii-art-web .    # αγνόησε το cache
docker history ascii-art-web:latest           # τα layers με τα μεγέθη τους
```

> ℹ️ Θα δεις αναφορές στο `docker exec` (μπαίνει σε container που **ήδη τρέχει**). Χρήσιμο για debugging, αλλά η *επαλήθευση filesystem με `docker exec`* είναι παραδοτέο του **Docker B**. Χρησιμοποίησέ το για να λύσεις προβλήματα· μην το γράψεις στο σημείωμα ως δική σου δουλειά.

---
---

# ΜΕΡΟΣ 4 — Παράδοση

## Βήμα 14 — Σημείωμα παράδοσης

Ο κανόνας §6: **όχι «τι έκανα», αλλά «τι πρέπει να ξέρει ο επόμενος»**. Ο επόμενος στο Docker είναι το Άτομο 1 (slice B: multi-stage).

1. **Τα paths είναι σχετικά ως προς το WORKDIR.** Στο multi-stage θα πρέπει να κάνει `COPY` ρητά τα `server`, `static/`, `templates/`, `banners/` στο runtime stage — αν ξεχάσει ένα, σκάει στο runtime, όχι στο build.
2. **Το inline `<script>` έμεινε στο `index.html`** επίτηδες, γιατί χρησιμοποιεί `{{.Color}}`. Δεν είναι ξεχασμένο.
3. **Τα error templates δεν έχουν `<link>` στο style.css** — αφέθηκε για το Stylize C.
4. Το route `/static/` δουλεύει **χωρίς αλλαγή στον `handlers.Home`**, λόγω longest-prefix matching του ServeMux.

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
- [ ] Ό,τι χρειάζεται το build είναι **committed** — αλλιώς χτίζει σε σένα και σπάει στους άλλους
- [ ] Σημείωμα παράδοσης γραμμένο
- [ ] `git merge main` πριν το push, conflicts λυμένα στο branch σου

---

## Καθάρισμα

```bash
docker images                 # δες τι μαζεύτηκε
docker system df              # πόσο χώρο πιάνει τι
docker rmi lab:latest         # σβήσε το πειραματικό
rm -rf ~/docker-lab
docker builder prune          # καθαρίζει μόνο το build cache
```

> ⚠️ Το `docker system prune` σβήνει **όλα** τα αχρησιμοποίητα objects — και άλλων project. Το garbage collection ως θέμα είναι παραδοτέο του **Docker C**· άφησέ το.

Τα images ζουν μέσα στο Linux VM του Docker Desktop, σε ένα αρχείο `~/Library/Containers/com.docker.docker/Data/vms/0/data/Docker.raw`. Δεν το πειράζεις — τα `docker` commands είναι ο τρόπος πρόσβασης.

---
---

# Παράρτημα Α — Οι εντολές του Dockerfile

## Ο διαχωρισμός που εξηγεί τα πάντα

| | **Εκτελούν** | **Δηλώνουν** |
|---|---|---|
| Τι κάνουν | αλλάζουν το filesystem | γράφουν στο config |
| Πότε | στο **build** | ισχύουν στο **run** |
| Παράγουν layer; | ναι | όχι (0 B) |
| Μετρούν ως βήματα; | ✅ | ❌ |
| Ποιες | `FROM` `RUN` `COPY` `ADD` `WORKDIR` | `CMD` `ENTRYPOINT` `ENV` `EXPOSE` `LABEL` `USER` `VOLUME` `ARG` `HEALTHCHECK` `STOPSIGNAL` `SHELL` |

Γι' αυτό το build δείχνει `3/3` ενώ το Dockerfile έχει 4 γραμμές.

## Οι έξι που θα χρησιμοποιήσεις

**`FROM golang:1.25.5`** — πάντα η πρώτη εντολή. Σε multi-stage εμφανίζεται πολλές φορές (`FROM x AS builder`) — δουλειά του Docker B.

**`WORKDIR /app`** — δύο ρόλοι: πού προσγειώνονται τα `COPY`/τρέχουν τα `RUN`, και το working directory του process. Δημιουργεί τον φάκελο αν λείπει. Μπορεί να εμφανιστεί πολλές φορές.

**`COPY . .`** — πηγή μέσα στο build context, προορισμός σχετικά με το WORKDIR. Δεν βλέπει αρχεία εκτός context.

**`RUN go build -o server .`** — όσα θέλεις, με τη σειρά. Κάθε ένα = layer. Shell form από προεπιλογή, άρα το `&&` δουλεύει.

**`CMD ["./server"]`** — **μόνο ένα ισχύει**, το τελευταίο. Exec form για μακρόβια processes ώστε να περνάνε τα σήματα. Παρακάμπτεται με εντολή μετά το όνομα του image στο `docker run`.

**`EXPOSE 8080`** — τεκμηρίωση. **Δεν ανοίγει τίποτα.**

## Χρήσιμες, αλλά όχι για το Docker A

| Εντολή | Τι κάνει |
|---|---|
| `ENV PORT=8080` | μεταβλητή **και στο build και στο runtime**, μένει στο image |
| `ARG VERSION=1.0` | μεταβλητή **μόνο στο build**, δεν επιβιώνει |
| `ENTRYPOINT ["./server"]` | το «σταθερό» μέρος· όταν υπάρχουν και τα δύο, το `CMD` γίνεται **ορίσματά** του |
| `LABEL maintainer="..."` | μεταδεδομένα — 👉 **παραδοτέο Docker B** |
| `USER appuser` | από προεπιλογή τα containers τρέχουν ως **root** |
| `ADD` | σαν `COPY` αλλά κατεβάζει URLs και αποσυμπιέζει tar — **προτίμα πάντα `COPY`** |

## Σπάνιες

| Εντολή | Τι κάνει |
|---|---|
| `VOLUME /data` | δηλώνει σημείο όπου τα δεδομένα πρέπει να **επιβιώνουν** |
| `HEALTHCHECK CMD ...` | περιοδικός έλεγχος· το `docker ps` δείχνει `healthy`/`unhealthy` |
| `STOPSIGNAL SIGINT` | ποιο σήμα στέλνει το `docker stop` (προεπιλογή `SIGTERM`) |
| `SHELL ["/bin/bash","-c"]` | αλλάζει ποιο shell χρησιμοποιεί η shell form |
| `ONBUILD RUN ...` | ενεργοποιείται όταν **άλλος** κάνει `FROM` το image σου |

## Οι τρεις κλασικές συγχύσεις

**`COPY` vs `ADD`** → Χρησιμοποίησε `COPY`. Το `ADD` κάνει σιωπηλά πράγματα που σπάνια θέλεις.

**`CMD` vs `ENTRYPOINT`** → Το `CMD` είναι *πρόταση* (παρακάμπτεται). Το `ENTRYPOINT` είναι *απόφαση*. Με ένα πρόγραμμα, το `CMD` αρκεί.

**`ARG` vs `ENV`** → Το `ARG` ζει μόνο στο build. Το `ENV` μένει στο image. Μυστικά σε κανένα από τα δύο — φαίνονται στο `docker history`.

---

# Παράρτημα Β — Ερωτήσεις audit

**«Γιατί δεν πιάνει το `/` το request για το `/static/style.css`;»**
Ο ServeMux διαλέγει το μακρύτερο pattern που ταιριάζει, όχι το πρώτο που δηλώθηκε.

**«Τι κάνει το StripPrefix και τι θα γινόταν χωρίς αυτό;»**
Ο FileServer ψάχνει το URL path αυτούσιο μέσα στον φάκελο. Χωρίς StripPrefix θα έψαχνε το `static/static/style.css` → 404.

**«Πού ψάχνει το πρόγραμμα τα banner files μέσα στο container;»**
Σχετικά με το working directory του process, δηλαδή το `WORKDIR`. Όχι σχετικά με το πού είναι το binary.

**«Τι διαφορά έχει το RUN από το CMD;»**
`RUN` εκτελείται στο build και ψήνεται στο image. `CMD` εκτελείται όταν ξεκινά container. Γι' αυτό μπορείς να έχεις πολλά `RUN` αλλά μόνο ένα ενεργό `CMD`.

**«Το EXPOSE ανοίγει το port;»**
Όχι. Είναι τεκμηρίωση. Το mapping το κάνει το `-p` στο `docker run`.

**«Γιατί το script δεν μπήκε σε `.js` αρχείο;»**
Γιατί περιέχει `{{.Color}}`, Go template action. Τα στατικά αρχεία δεν περνούν από το template engine.

**«Γιατί `CMD ["./server"]` και όχι `CMD ["server"]`;»**
Σκέτο όνομα ψάχνεται στο `$PATH`, όπου το `/app` δεν ανήκει. Με κάθετο, αντιμετωπίζεται ως διαδρομή σχετική με το cwd.

**«Γιατί exec form και όχι shell form;»**
Με shell form το `sh` γίνεται PID 1. Ο kernel παραδίδει σήματα στο PID 1 μόνο αν υπάρχει ρητός handler — το `sh` δεν προωθεί, οπότε το `docker stop` περιμένει 10 δευτερόλεπτα και σκοτώνει βίαια.

**«Γιατί καρφωμένη έκδοση Go και όχι `latest`;»**
Αναπαραγωγιμότητα. Το build πρέπει να δίνει το ίδιο αποτέλεσμα σε άλλο μηχάνημα, άλλη στιγμή.

---

# Παράρτημα Γ — Παγίδες

| Παγίδα | Σύμπτωμα |
|---|---|
| Ξέχασες το `StripPrefix` | 404 στο `/static/style.css` |
| `href="static/style.css"` χωρίς leading `/` | Δουλεύει στο `/`, σπάει αλλού |
| Λάθος `WORKDIR` ή ελλιπές `COPY` | Container σηκώνεται κανονικά, requests γυρνάνε 500 |
| `CMD ["server"]` χωρίς `./` | `executable file not found in $PATH` |
| Μετέφερες το `<script>` σε `.js` | Το `{{.Color}}` φτάνει σαν literal string στον browser |
| Πείραξες το CSS «λιγάκι» | Ο Stylize A δουλεύει πάνω σε κάτι που δεν περίμενε |
| Έκανες multi-stage γιατί «είναι καλύτερο» | Πήρες τη δουλειά του Ατόμου 1 |
| Αρχείο του build δεν έγινε commit | Χτίζει σε σένα, σπάει στους άλλους |
| `Dockerfile.txt` από TextEdit | `failed to read dockerfile: no such file` |
