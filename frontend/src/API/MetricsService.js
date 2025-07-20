import apiClient from "../hooks/useApiInterceptor";

export default class MetricsService {
    static async getAvailabilityLogs() {
        const response = await apiClient.get("/metrics/availability/");
        return response.data
    }
    static async getDetailedLogs(logsType) {
        const response = await apiClient.get(`/metrics/detailed/?type=${logsType}`);
        return response.data
    }
    static async getBasicLogs() {
        const response = await apiClient.get("/metrics/basic/");
        return response.data
    }
    static async getStaticMetrics() {
        const response = await apiClient.get("/metrics/info/");
        return response.data
    }
    static async getBinanceApiWeight() {
        const response = await apiClient.get("/metrics/binance-api-weight-history/");
        return response.data
    }
}
