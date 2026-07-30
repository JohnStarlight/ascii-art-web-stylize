# ascii-art-web

A Go web application that converts user-submitted text into ASCII art, rendered in the browser with optional coloring. Built entirely with Go's standard library (no external packages, no frontend frameworks, no CDN dependencies).

This project is the web version of the `ascii-art` exercise: instead of a CLI, it exposes a simple HTML form where a user types text, picks a banner style and an optional color, and gets back the ASCII art rendering in the page.

## Features

- Convert any printable ASCII text (character codes 32–126) into large ASCII-art banners
- Three banner styles: `standard`, `shadow`, `thinkertoy`
- Multi-line input support (newlines in the textarea produce stacked ASCII-art blocks)
- Optional coloring of the output via a color dropdown
- Optional partial coloring: color only the occurrences of a specific substring instead of the whole output
- Form values (text, banner, color, letters) are preserved after submission
- Proper HTTP status codes and custom error pages for bad input, unknown routes, disallowed methods, and server errors
- The server never crashes — invalid input always results in a handled error response

## How it works

1. **`main.go`** starts an HTTP server on port `8080` and registers two routes:
   - `GET /` — serves the home page with the input form
   - `POST /ascii-art` — receives the form submission and returns the rendered result

2. **`handlers/handlers.go`** handles routing logic:
   - Validates the HTTP method and path, returning `404` for unknown paths and `405` for disallowed methods
   - Reads the form fields (`text`, `banner`, `color`, `letters`)
   - Validates that text isn't empty and that the banner style is one of the three supported values, returning `400` on failure
   - Calls `ascii.Generate(...)` to build the ASCII art
   - Renders `templates/index.html` with the result (or the error message), returning `200`, `400`, or `500` as appropriate

3. **`ascii/ascii.go`** contains the actual ASCII art generation logic:
   - Loads the requested banner file from `banners/` (each character in a banner file is represented as 8 lines of glyph data)
   - For each line of input text, looks up the glyph lines for every character and assembles them row by row into the final ASCII art block
   - Rejects any character outside the printable ASCII range (32–126) with an error
   - If a color is given, wraps the relevant output in `<span style="color: ...">` tags — either the whole block (no `letters` given) or just the parts matching the `letters` substring (overlapping matches included)

4. **`templates/index.html`** is the single Go template used for both the empty form and the result page. It reads a `PageData` struct (`Text`, `Banner`, `Color`, `Letters`, `Result`, `Error`) so that previously entered values stay filled in after a submission, and displays the ASCII art result inside a `<pre>` block (using `template.HTML` so the `<span>` color tags aren't escaped).

5. **`templates/400.html`, `404.html`, `405.html`, `500.html`** are static error pages served for their respective HTTP status codes.

## Project structure

```
ascii-art-web/
├── Dockerfile           # container image definition
├── main.go              # server startup and routing
├── handlers/
│   └── handlers.go      # HTTP handlers, validation, template rendering
├── ascii/
│   └── ascii.go         # ASCII art generation logic
├── banners/
│   ├── standard.txt
│   ├── shadow.txt
│   └── thinkertoy.txt
└── templates/
    ├── index.html       # main form + result page
    ├── 400.html
    ├── 404.html
    ├── 405.html
    └── 500.html
```

## Requirements

- Go 1.25.5 or later (see [go.mod](go.mod))
- No external dependencies — standard library only

Alternatively, run it with Docker and skip the Go installation entirely — see [Running with Docker](#running-with-docker).

## Usage

Run the server from the project root:

```bash
go run .
```

The server starts on port `8080`. Open your browser at:

```
http://localhost:8080
```

### Using the app

1. Type or paste text into the textarea.
2. Choose a banner style: `standard`, `shadow`, or `thinkertoy`.
3. (Optional) Pick a color from the dropdown to color the ASCII art output.
4. (Optional) Enter a substring in the "letters" field to color only the matching characters instead of the entire output.
5. Submit the form to see the rendered ASCII art below.

### Notes on input

- Only printable ASCII characters (codes 32–126) are supported; anything else returns a `400 Bad Request` with an error message.
- Newlines in the textarea produce separate stacked ASCII-art blocks, one per line of input.
- Leaving text empty, or choosing an invalid banner, also returns a `400 Bad Request`.

## Running with Docker

The project ships with a `Dockerfile`, so it can be built and run without installing Go locally.

**Requirements:** Docker installed and the Docker daemon running (on macOS/Windows, start Docker Desktop first).

### Build the image

From the project root:

```bash
docker build -t ascii-art-web .
```

This compiles the server inside the image. The first build downloads the `golang` base image and takes a while; later builds reuse the cache and are much faster.

### Run the container

```bash
docker run --rm -p 8080:8080 ascii-art-web
```

Then open <http://localhost:8080> as usual.

- `-p 8080:8080` maps port `8080` on your machine to port `8080` inside the container. The right-hand side must stay `8080` because that is where the server listens (see [main.go](main.go)); the left-hand side can be any free port, e.g. `-p 3000:8080`.
- `--rm` removes the container once it stops, so stopped containers don't pile up.

The terminal stays attached and shows the server log. **Press `Ctrl+C` to stop it.**

To run it in the background instead:

```bash
docker run -d --rm -p 8080:8080 --name ascii ascii-art-web
docker logs -f ascii      # follow the log
docker stop ascii         # stop it
```

### Verifying it works

Starting up is not proof that it works — the server can start fine and still fail on every request if the files it reads at runtime aren't where it expects. Check all three:

```bash
curl -i localhost:8080/
curl -i -X POST localhost:8080/ascii-art -d "text=hello" -d "banner=standard"
curl -i localhost:8080/static/style.css
```

The second one matters most: it is the only request that forces the server to read a banner file, which is what proves the container's working directory is correct.

### Implementation notes

- The server reads `templates/` and `banners/` using **paths relative to its working directory**, so the image sets `WORKDIR /app` and copies the project there. Changing the working directory without moving those folders makes every request fail with `500`, even though the container starts normally.
- The whole project directory is copied into the image, so anything the build needs must be committed to git — otherwise the image builds locally but breaks for everyone else.

## HTTP status codes

| Code | When |
|------|------|
| `200 OK` | Successful page render or ASCII art generation |
| `400 Bad Request` | Empty text, invalid banner, or unsupported characters |
| `404 Not Found` | Unknown path |
| `405 Method Not Allowed` | Wrong HTTP method for a route |
| `500 Internal Server Error` | Template loading/rendering failure |
