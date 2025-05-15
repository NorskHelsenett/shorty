# Shorty - URL Shortener and Admin Management System

- Admin user management through email registration. A admin user can delete and modify all paths.
- URL shortening service with QR code generation capabilities

The web application is built using React, Vite, and TypeScript, while the server is developed in Go and uses Redis as its database.

## Features

### Admin Panel

- Add new admin users by email
- Delete existing admin accounts

### URL Shortening Service

- Create shortened URLs for any given link
- Delete existing shortcuts
- Edit and modify saved URLs
- Generate QR codes
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
   npm run dev or npm run build
   ```
### Build containers
```bash
docker compose up
```
### Kubernetes
```bash
Helmcharts that are updated must have Redis and an identity provider.
```
