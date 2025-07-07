import apiClient from "../hooks/useApiInterceptor";

export default class SettingsService {
    static async getSettings() {
        const response = await apiClient.get("/settings/");
        return response.data;
    }
    static async getLogs() {
        const response = await apiClient.get("/settings/logs/");
        return response.data
    }
    static async updateApiCooldown(id, cooldown) {
        const response = await apiClient.patch("/settings/update/", {id: id, cooldown: cooldown})
        return response.data
    }
}
