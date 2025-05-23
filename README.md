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

1. Clone the repository:

   ```bash
   git clone https://github.com/NorskHelsenett/shorty.git

   ```
### Server
1. Set version ```export SHORTY_VERSION=[version]```

2. Build executable
```bash
go mod tidy
swag init
```
### Build backend containers
```bash
docker compose up
```

### Web
1. Navigate to the project directory:
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
   Alternatively, you can build the web container:
   ```bash
   docker build . -t shorty-web:latest -p 5173
   ```

### Kubernetes
```bash
Helmcharts that are updated must have Redis and an identity provider.
```
Remember to set yours environments variables.
