#¡/bin/bash
npm install
BASE_URL=/app APP_URL=http://localhost:8888 npm run build
cp favicon.ico dist/favicon.ico
