// Dex
import {TAuthConfig} from "react-oauth2-code-pkce"


export const authConfig: TAuthConfig = {
    clientId: "shortyfront",
    authorizationEndpoint: "http://localhost:5556/dex/auth",
    tokenEndpoint: "http://localhost:5556/dex/token",
    redirectUri: "http://localhost:5173",
    scope: "openid profile email groups",
}