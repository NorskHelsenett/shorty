import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { AuthProvider } from "react-oauth2-code-pkce";
import { authConfig } from "./Service/Auth/authService.ts";
import { AdminProvider } from "./Hooks/authAdminContext.tsx";
import { BrowserRouter as Router, Route, Routes } from "react-router-dom";
import "./App.css";
import "./Index.css";
import AdminPage from "./Pages/AdminPage.tsx";
import HomePage from "./Pages/HomePage.tsx";

createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <AuthProvider authConfig={authConfig}>
      <AdminProvider>
        <Router>
          <Routes>
            <Route path="/" element={<HomePage />} />
            <Route path="/admin" element={<AdminPage />} />
          </Routes>
        </Router>
      </AdminProvider>
    </AuthProvider>
  </StrictMode>
);
