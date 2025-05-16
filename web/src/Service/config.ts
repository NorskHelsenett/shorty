function getEnv(variable: string | undefined, name: string, defaultValue: string): string {
    if (!variable){
        console.warn(`Enveroment variable ${name} is not set; using standard value: ${defaultValue}`)
        return defaultValue;
    }
    return variable;
}


// get environment from terminal
export const AUTH_URL = getEnv(import.meta.env.VITE_AUTH_URL, "VITE_AUTH_URL", "http://localhost:5556");
export const API_URL = getEnv(import.meta.env.VITE_API_URL,"VITE_API_URL", "http://localhost:8880");
export const REDIRECT_URI = getEnv(import.meta.env.VITE_REDIRECT_URI, "VITE_REDIRECT_URI", "http://localhost:5173");


import {TAuthConfig} from "react-oauth2-code-pkce"

export const AUTH_CONFIG: TAuthConfig = {
    clientId: "shortyfront",
    authorizationEndpoint: `${AUTH_URL}/dex/auth`,
    tokenEndpoint: `${AUTH_URL}/dex/token`,
    redirectUri: `${REDIRECT_URI}/admin`,
    scope: "openid profile email groups",
}