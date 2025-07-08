import axios from "axios";

export const baseURL = process.env.REACT_APP_BASE_URL || 'http://localhost:8001/api/v1';
export const websocketBaseURL = process.env.REACT_APP_WEBSOCKET_URL || 'ws://localhost:8001/api/v1';

const apiClient = axios.create ({
    baseURL,
    headers: {
        'Content-Type': 'application/json',
    }
})

export default apiClient;
