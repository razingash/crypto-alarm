import apiClient from "../hooks/useApiInterceptor";

export default class VariablesService {
    static async getVariables() {
        const response = await apiClient.get(`/variables/`)
        return response.data
    }
    static async createVariable() {
        const response = await apiClient.post(`/variables/`)
        return response.data
    }
    static async deleteVariable(pk) {
        const response = await apiClient.delete(`/variables/${pk}`)
        return response.data
    }
    static async updateVariable(pk) {
        const response = await apiClient.patch(`/variables/${pk}`)
        return response.data
    }
}