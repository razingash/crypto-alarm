import apiClient from "../hooks/useApiInterceptor";

export default class WorkspaceService {
    static async getDiagrams(params) {
        const response = await apiClient.get(`/workspace/diagram/`, {params: params})
        return response.data
    }
    static async createDiagram(name) {
        return await apiClient.post(`/workspace/diagram/`, {name: name})
    }
    static async deleteDiagram(pk) {
        return await apiClient.delete(`/workspace/diagram/${pk}`)
    }
    static async updateDiagram(pk, data) {
        return await apiClient.patch(`/workspace/diagram/${pk}`, data)
    }
    static async updateDiagramNodes(pk, data) {
        return await apiClient.patch(`/workspace/diagram/${pk}/nodes`, data)
    }
}