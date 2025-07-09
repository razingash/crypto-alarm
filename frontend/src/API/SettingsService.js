import apiClient from "../hooks/useApiInterceptor";

export default class SettingsService {
    static async getSettings() {
        const response = await apiClient.get("/settings/");
        return response.data;
    }
    static async updateApiSettings(data) {
        const response = await apiClient.patch("/settings/update/", data)
        return response.data
    }
}
