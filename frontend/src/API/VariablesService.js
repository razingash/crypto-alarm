import apiClient from "../hooks/useApiInterceptor";

export default class VariablesService {
    static async getVariables(params) {
        const response = await apiClient.get(`/variable/`, {params: params})
        return response.data
    }
    static async createVariable() {
        const response = await apiClient.post(`/variable/`)
        return response.data
    }
    static async deleteVariable(pk) {
        const response = await apiClient.delete(`/variable/${pk}`)
        return response.data
    }
    static async updateVariable(pk) {
        const response = await apiClient.patch(`/variable/${pk}`)
        return response.data
    }
}