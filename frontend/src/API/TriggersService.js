import apiClient from "../hooks/useApiInterceptor";

export default class TriggersService {
    static async getKeyboard() {
        // получает данные для клавиатуры
        const response = await apiClient.get(`/triggers/keyboard`)
        return response.data
    }
    static async getUserFormulas(params) {
        const response = await apiClient.get('/triggers/formula', {params: params})
        return response.data
    }
    static async getFormulaHistory(formula_id, page) {
        const response = await apiClient.get(`/triggers/formula/history/${formula_id}`, {params: {page}})
        return response.data
    }
    static async createFormula(formula, name) {
        return await apiClient.post('/triggers/formula', {formula, name});
    }
    static async updateUserFormula(data) { // data - словарь с formula_id и полями которые нужно изменить
        return await apiClient.patch('/triggers/formula', data)
    }
    static async deleteUserFormula(formula_id) {
        return await apiClient.delete(`/triggers/formula?formula_id=${formula_id}`)
    }
}