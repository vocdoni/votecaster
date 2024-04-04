#¡/bin/bash
APP_URL=${APP_URL:-http://localhost:8888}
npm install
BASE_URL=/app npm run build
cp favicon.ico dist/favicon.ico
