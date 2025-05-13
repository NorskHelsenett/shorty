
export const AUTH_URL = import.meta.env.VITE_AUTH_URL;
export const API_URL = import.meta.env.VITE_API_URL;


import {TAuthConfig} from "react-oauth2-code-pkce"

export const AUTH_CONFIG: TAuthConfig = {
    clientId: "shortyfront",
    authorizationEndpoint: `${AUTH_URL}/dex/auth`,
    tokenEndpoint: `${AUTH_URL}/dex/token`,
    redirectUri: "http://localhost:5173",
    scope: "openid profile email groups",
}