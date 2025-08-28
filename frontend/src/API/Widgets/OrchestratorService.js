import apiClient from "../../hooks/useApiInterceptor";

export default class OrchestratorService {
    static async create() {
        const response = await apiClient.post(`/workspace/widgets/orchestrator`)
        return response.data
    }
    static async update(pk, params) {
        const response = await apiClient.get(`/workspace/widgets/orchestrator/${pk}`, {params: params})
        return response.data
    }
    static async get(pk) {
        const response = await apiClient.get(`/workspace/widgets/orchestrator/${pk}`)
        return response.data
    }
    static async getParts(searchParams) {
        const response = await apiClient.get(`/workspace/widgets/orchestrator/parts?${searchParams}`)
        return response.data
    }
    static async delete(pk) {
        return await apiClient.delete(`/workspace/widgets/orchestrator/${pk}`)
    }
}