import apiClient from "../hooks/useApiInterceptor";
import axios from "axios";

export default class StrategyService {
    static async getKeyboard() { // получает данные для клавиатуры
        const response = await apiClient.get(`/triggers/keyboard`)
        return response.data
    }
    static async getStrategies(params) {
        const response = await apiClient.get('/triggers/strategy', {params: params})
        return response.data
    }
    static async getStrategyHistory(formula_id, page, prevCursor) {
        const response = await apiClient.get(`/triggers/strategy/history/${formula_id}`, {params: {page, prevCursor}})
        return response.data
    }
    static async createStrategy(rawFormulas, name) {
        const conditions = rawFormulas.map(rawFormula => {
            const formula = rawFormulaToFormula(rawFormula);
            const formula_raw = rawFormula.filter(item => item !== "\\textunderscore").map(cleanKatexExpression).join('');
            return {formula, formula_raw};
        });
        return await apiClient.post('/triggers/strategy', {name, conditions});
    }
    static async updateStrategy(data) { // data - словарь с formula_id и полями которые нужно изменить
        return await apiClient.patch('/triggers/strategy', data)
    }
    static async deleteStrategyOrCondition(strategy_id, conditionID) {
        // если указан formula_id то будет удалено только выражение, а не вся стратегия
        return await apiClient.delete(`/triggers/strategy/${strategy_id}/${conditionID ? `?formula_id=${conditionID}` : ''}`)
    }
    static async getBinanceKlines(symbol, interval) {
        // Open time | Open | High | Low | Close | Volume | Close time | Quote asset volume | Number of trades | Taker buy base asset volume | Taker buy quote asset volume | Ignore
        return await axios.get(`https://api.binance.com/api/v3/klines?symbol=${symbol}&interval=${interval}`)
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
