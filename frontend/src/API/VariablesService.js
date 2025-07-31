import apiClient from "../hooks/useApiInterceptor";

export default class VariablesService {
    static async getVariables(params) {
        const response = await apiClient.get(`/variable/`, {params: params})
        return response.data
    }
    static async createVariable(data) {
        return await apiClient.post(`/variable/`, data)
    }
    static async deleteVariable(pk) {
        return await apiClient.delete(`/variable/${pk}`)
    }
    static async updateVariable(pk, data) {
        return await apiClient.patch(`/variable/${pk}`, data)
    }
}