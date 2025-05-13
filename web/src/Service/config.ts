export const AUTH_URL = "https://auth.sky.nhn.no"; // not NHN use http://localhost:5556
export const API_URL ="https://k.test.nhn.no" // not NHN use http://localhost:8880


// Dex
import {TAuthConfig} from "react-oauth2-code-pkce"



export const AUTH_CONFIG: TAuthConfig = {
    clientId: "shortyfront",
    authorizationEndpoint: `${AUTH_URL}/dex/auth`,
    tokenEndpoint: `${AUTH_URL}/dex/token`,
    redirectUri: "http://localhost:5173",
    scope: "openid profile email groups",
}