# ShortyFront - URL Shortener and Admin Management System

A web application that provides two main functionalities:

- Admin user management through email registration. A admin user can delete and modify all paths.
- URL shortening service with QR code generation capabilities

The application is built with REACT, Vite and TypeScript.
The backend, referred to as **Shorty**.

## Features

### Admin Panel

- Add new admin users by email
- Delete existing admin accounts

### URL Shortening Service

- Create shortened URLs for any given link
- Delete existing shortcuts
- Edit and modify saved URLs
- Generate QR codes for quick access
- Download QR codes as images

## Technical Stack

Frontend: React 18, Vite
Styling: Emotion
Routing: React Router DOM
Form Management: React Hook Form
Authentication: OAuth2 with PKCE flow
API Integration: Axios
QR Code Generation: qrcode.react

## Get startet

### Prerequistes

Ensure you have the following installed:

- Node.js
- npm (Node package manager)
- TypeScript compiler

### Setting Up the Backend

Make sure that the backend (Shorty) is running.

### Running the Frontend Application

1. Clone the repository:

   ```bash
   git clone https://helsegitlab.nhn.no/sdi/prosjekter/sdi/shortygui.git

   ```

2. Navigate to the project directory:
   ```bash
   cd shorty-front2
   ```
3. Install the frontend dependencies:
   ```bash
   npm install
   ```
4. Start the development server:
   ```bash
   npm run dev
   ```

The application should be available at `http://localhost:5173`
