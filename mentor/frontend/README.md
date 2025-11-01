# Mentor Frontend

A Vite + React app for monitoring devices via the Mentor backend.

## Development

1) Use the repository root `.env` to control ports and API URLs (no hardcoded ports in code):

```bash
cp .env.example .env
# Edit .env to set:
# - MENTOR_FRONTEND_PORT
# - MENTOR_BACKEND_PORT
# The start script will export VITE_MENTOR_API_URL for the app.
```

2) Start via the top-level helper script (recommended):

```bash
./scripts/start.sh
```

Or run just the frontend:

```bash
cd mentor/frontend
npm install
# The dev server reads VITE_MENTOR_FRONTEND_PORT from the environment
# Example: set to any available port (e.g., 5000, 3000, 8080)
VITE_MENTOR_FRONTEND_PORT=5000 npm run dev
```

## API

The app expects a Mentor backend with these endpoints:
- GET /devices
- GET /devices/:id/metrics
- GET /devices/:id/processes
- GET /devices/:id/activities
- GET /devices/:id/alerts
- GET /devices/:id/screenshots
- POST /devices/:id/commands

If a section has no data, the UI shows an explicit empty-state message.
