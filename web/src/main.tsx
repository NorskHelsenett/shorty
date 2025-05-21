import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { AuthProvider } from "react-oauth2-code-pkce";
import { AdminProvider } from "./Hooks/authAdminContext.tsx";
import { Route, Routes, BrowserRouter } from "react-router-dom";
import "./App.css";
import "./Index.css";
import AdminPage from "./Pages/AdminPage.tsx";
import HomePage from "./Pages/HomePage.tsx";
import { AUTH_CONFIG } from "./Service/config.ts";

createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <AuthProvider authConfig={AUTH_CONFIG}>
      <AdminProvider>
        <BrowserRouter basename="/admin">
          <Routes>
            <Route path="/" element={<HomePage />} />
            <Route path="/user" element={<AdminPage />} />
          </Routes>
        </BrowserRouter>
      </AdminProvider>
    </AuthProvider>
  </StrictMode>
);
