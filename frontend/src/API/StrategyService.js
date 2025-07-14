import apiClient from "../hooks/useApiInterceptor";

export default class StrategyService {
    static async getKeyboard() {
        // получает данные для клавиатуры
        const response = await apiClient.get(`/triggers/keyboard`)
        return response.data
    }
    static async getStrategies(params) {
        const response = await apiClient.get('/triggers/strategy', {params: params})
        return response.data
    }
    static async getFormulaHistory(formula_id, page, prevCursor) {
        const response = await apiClient.get(`/triggers/strategy/history/${formula_id}`, {params: {page, prevCursor}})
        return response.data
    }
    static async createFormula(rawFormula, name) {
        const formula = rawFormulaToFormula(rawFormula);
        rawFormula = rawFormula.filter(item => item !== "\\textunderscore").map(cleanKatexExpression).join('');
        return await apiClient.post('/triggers/strategy', {formula, name, 'formula_raw': rawFormula});
    }
    static async updateUserFormula(data) { // data - словарь с formula_id и полями которые нужно изменить
        return await apiClient.patch('/triggers/strategy', data)
    }
    static async deleteUserFormula(strategy_id, formula_id=null) {
        return await apiClient.delete(`/triggers/strategy/${strategy_id}/?formula_id=${formula_id}`)
    }
}

function cleanKatexExpression(expr) { // проверки на правильность тут не будет
    return expr
        .replace(/\\textcolor{[^}]+}{([^}]+)}/g, '$1')
        .replace(/\\text{([^}]+)}/g, '$1')
        .replace(/\\_/g, '_')
        .replace(/\\\\/g, '\\');
}

function rawFormulaToFormula(tokens) {
    const result = [];
    const stack = [];
    let i = 0;

    while (i < tokens.length) {
        const tk = tokens[i];
        if (tk === "\\textunderscore") {
            i++;
            continue
        }
        if (tk === 'matrix' && tokens[i + 1] === '(') {
            stack.push('matrix');
            result.push('(', '(');
            i += 2;
            continue;
        }

        if (stack.length > 0 && tk === ',') {
            result.push(')', '/', '(');
            i++;
            continue;
        }

        if (stack.length > 0 && tk === ')') {
            result.push(')', ')');
            stack.pop();
            i += 1;
            continue;
        }

        result.push(tk);
        i++;
    }

    return result.map(cleanKatexExpression).join('');
}
