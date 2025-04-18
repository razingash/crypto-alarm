import apiClient, {baseURL} from "../hooks/useApiInterceptor";
import axios from "axios";

export default class NotificationService {
    static async getVapidKey() {
        const response = await axios.get(`${baseURL}/notifications/vapid-key`);
        return response.data.vapidPublicKey;
    }
    static async subscribeToPushNotifications(endpoint, p256dh, auth) {
        const response = await apiClient.post('/notifications/subscribe', {endpoint, p256dh, auth})
        return response.data
    }
}
