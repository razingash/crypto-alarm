import axios from "axios";
import apiClient, {baseURL} from "../hooks/useApiInterceptor";

export default class AuthService {
    static async register(username, password) {
        const response = await axios.post(`${baseURL}/auth/register/`, {username, password},
        { headers: { "Content-Type": "application/json" } })
        return response.data
    }
    static async login(username, password) {
        const response = await axios.post(`${baseURL}/auth/token/`, {username, password},
        { headers: { "Content-Type": "application/json" } })
        return response.data
    }
    static async verifyToken(token) { // both access and refresh tokens
        const response = await axios.post(`${baseURL}/auth/token/verify/`, {token},
        { headers: { "Content-Type": "application/json" } })
        return response.data
    }
    static async refreshAccessToken(refreshToken) {
        const response = await axios.post(`${baseURL}/auth/token/refresh/`, {token: refreshToken},
        { headers: { "Content-Type": "application/json" } })
        return response.data
    }
    static async logout(refreshToken) {
        const response = await apiClient.post('/auth/logout/', {token: refreshToken})
        return response.data
    }
}