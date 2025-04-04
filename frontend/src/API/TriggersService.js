import apiClient from "../hooks/useApiInterceptor";

export default class TriggersService {
    static async getKeyboard() {
        // получает данные для клавиатуры
        const response = await apiClient.get(`/triggers/keyboard`)
        return response.data
    }
    static async getUserFormulas(page) {
        const response = await apiClient.get('triggers/formula', {params: {page: page}})
        return response.data
    }
    static async createFormula(formula) {
        const response = await apiClient.post('/triggers/formula', {formula})
        return response.data
    }
    static async updateUserFormula(data) { // data - словарь с formula_id и полями которые нужно изменить
        const response = await apiClient.patch('/triggers/formula', data)
        return response.data
    }
    static async deleteUserFormula(formula_id) {
        const response = await apiClient.delete(`/triggers/formula?id=${formula_id}`)
        return response.data
    }
}