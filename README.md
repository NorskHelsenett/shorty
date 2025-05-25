# Shorty - URL Shortener and Admin Management System

Shorty is a system for URL shortening with an admin panel for user management. The web application is built using React, Vite, and TypeScript, while the server is developed in Go and utilizes Redis as its database.


## Features

### Admin Panel

- Add new admin users by email
- Delete existing admin accounts

### URL Shortening Service

- Create shortened URLs for any given link
- Delete existing shortcuts
- Edit and modify saved URLs
- Generate QR codes for short path
- Download QR codes as images

## BUILD

Make sure that you have installed Go, Node, and Swagger on your machine before proceeding with the next processes.

1. Clone the repository:

   ```bash
   git clone https://github.com/NorskHelsenett/shorty.git

   ```
### Server

1. Set version ```export SHORTY_VERSION=[version]```

2. Build executable
```bash
go mod tidy
swag init -g cmd/shorty/main.go --parseDependency --output internal/docs --parseInternal
```
### Build backend containers
```bash
docker compose up
```

### Web
1. Navigate to the project web-directory:
   ```bash
   cd web
   ```
2. Install the frontend dependencies:
   ```bash
   npm install
   ```
3. Start the development server:
   ```bash
   npm run dev
   ```
   Or build the production version:
   ```bash
   npm run build
   ```
   Alternatively, you can build and run the web container:
   ```bash
   docker build -t shorty-web:latest .
   ```
   ```bash
   docker run -p 5173 -t shorty-web:latest 
   ```

### Windows Setup Tip
If you're setting up the project on Windows, make sure to convert 'docker-entrypoint.sh' to use **LF** (Unix-style) line endings instead of **CRLF**, to avoid execution issues in Linux containers.

### Kubernetes

- Helmcharts that are updated must have Redis and an identity provider.
- Remember to set yours environments variables.
