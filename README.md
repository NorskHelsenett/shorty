# Shorty - URL Shortener and Admin Management System

- Admin user management through email registration. A admin user can delete and modify all paths.
- URL shortening service with QR code generation capabilities

The web application is built with REACT, Vite and TypeScript.
The server is built with GO.

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
CGO_ENABLED=0 go build -ldflags "-w -extldflags '-static' -X shorty/main.Version=$SHORTY_VERSION" -o "dist/kort" main.go
```
3. Build and upload docker image
```bash
docker build . -t ncr.sky.nhn.no/nhn/kort:$SHORTY_VERSION
docker push ncr.sky.nhn.no/nhn/kort:$SHORTY_VERSION
```
4. Cleanup
```bash
rm dist/kort
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
   npm run dev or docker build
   ```
