import apiClient from "../hooks/useApiInterceptor";

export default class MetricsService {
    static async getAvailabilityLogs() {
        const response = await apiClient.get("/metrics/availability/");
        return response.data
    }
    static async getCriticalErrorLogs() {
        const response = await apiClient.get("/metrics/errors/");
        return response.data
    }
    static async getStaticMetrics() {
        const response = await apiClient.get("/metrics/info/");
        return response.data
    }
}
