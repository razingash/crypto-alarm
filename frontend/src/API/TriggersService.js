import apiClient from "../hooks/useApiInterceptor";

export default class TriggersService {
    static async getKeyboard() {
        // получает данные для клавиатуры
        const response = await apiClient.get(`/triggers/keyboard`)
        return response.data
    }
    static async registerFormula(formula) {
        const response = await apiClient.post('/triggers/formula', {formula})
        return response.data
    }
    static async getUserFormulas() {
        const response = await apiClient('triggers/formulas')
        return response.data
    }
}