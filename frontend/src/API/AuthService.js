import axios from "axios";
import apiClient, {baseURL} from "../hooks/useApiInterceptor";

export default class AuthService {
    static async register(username, password) {
        const response = await axios.post(`${baseURL}/registration/`, {username, password})
        return response.data
    }
    static async login(username, password) {
        const response = await axios.post(`${baseURL}/token/`, {username, password})
        return response.data
    }
    static async verifyToken(token) { // both access and refresh tokens
        const response = await axios.post(`${baseURL}/token/verify/`, {token})
        return response.data
    }
    static async refreshAccessToken(refreshToken) {
        const response = await axios.post(`${baseURL}/token/refresh/`, {refresh: refreshToken})
        return response.data
    }
    static async logout(refreshToken) {
        const response = await apiClient.post('/logout/', {refresh_token: refreshToken})
        return response.data
    }
}