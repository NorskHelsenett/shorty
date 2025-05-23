// Define the runtime config interface
interface RuntimeConfig {
    AUTH_URL: string;
    API_URL: string;
    REDIRECT_URI: string;
}

// Declare the window.RUNTIME_CONFIG property
declare global {
    interface Window {
        RUNTIME_CONFIG: RuntimeConfig;
    }
}

function getEnv(runtimeVar: string | undefined, envVar: string | undefined, name: string, defaultValue: string): string {
    // First try runtime config (highest priority)
    if (runtimeVar) {
        return runtimeVar;
    }

    // Then try build-time env variable
    if (envVar) {
        return envVar;
    }

    // Fall back to default
    console.warn(`Environment variable ${name} is not set; using standard value: ${defaultValue}`);
    return defaultValue;
}

// First check runtime config, then fall back to build-time env variables
export const AUTH_URL = getEnv(
    window.RUNTIME_CONFIG?.AUTH_URL,
    import.meta.env.VITE_AUTH_URL as string | undefined,
    "VITE_AUTH_URL",
    "http://localhost:5556/dex"
);
export const API_URL = getEnv(
    window.RUNTIME_CONFIG?.API_URL,
    import.meta.env.VITE_API_URL as string | undefined,
    "VITE_API_URL",
    "http://localhost:8880"
);
export const REDIRECT_URI = getEnv(
    window.RUNTIME_CONFIG?.REDIRECT_URI,
    import.meta.env.VITE_REDIRECT_URI as string | undefined,
    "VITE_REDIRECT_URI",
    "http://localhost:5173"
);


import { TAuthConfig } from "react-oauth2-code-pkce"

export const AUTH_CONFIG: TAuthConfig = {
    clientId: "shortyfront",
    authorizationEndpoint: `${AUTH_URL}/auth`,
    tokenEndpoint: `${AUTH_URL}/token`,
    redirectUri: `${REDIRECT_URI}/admin/`,
    scope: "openid profile email groups",
}