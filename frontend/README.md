# CargoPrener frontend

Landing page for CargoPrener, a group-buying platform for small businesses. Other pages are title-only routes so navigation can be wired up later.

Run these commands from the `frontend` folder.

## Run locally

```bash
npm install
npm run dev
```

Open http://localhost:5173

## Run with Docker

```bash
docker build -t cargoprener .
docker run --rm -p 8080:80 cargoprener
```

Open http://localhost:8080
