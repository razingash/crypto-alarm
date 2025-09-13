import apiClient from "../../hooks/useApiInterceptor";

export default class NotificationTelegramService {
    static async create(searchParams, data) {
        const response = await apiClient.post(`/modules/notification-telegram/?${searchParams}`, data)
        return response.data
    }
    static async update(searchParams, params) {
        const response = await apiClient.patch(`/modules/notification-telegram/?${searchParams}`, {params: params})
        return response.data
    }
    static async get(pk) {
        const response = await apiClient.get(`/modules/notification-telegram/?id=${pk}`)
        return response.data
    }
}