
export const AUTH_URL = import.meta.env.VITE_AUTH_URL;
export const API_URL = import.meta.env.VITE_API_URL;
export const REDIRECT_URI = import.meta.env.VITE_REDIRECT_URI;


import {TAuthConfig} from "react-oauth2-code-pkce"

export const AUTH_CONFIG: TAuthConfig = {
    clientId: "shortyfront",
    authorizationEndpoint: `${AUTH_URL}/dex/auth`,
    tokenEndpoint: `${AUTH_URL}/dex/token`,
    redirectUri: `${REDIRECT_URI}/admin`,
    scope: "openid profile email groups",
}