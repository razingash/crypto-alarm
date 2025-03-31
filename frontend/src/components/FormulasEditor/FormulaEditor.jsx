import React, {useState} from 'react';
import "../../styles/keyboard.css"
import Keyboard from "./Keyboard";
import FormulaInput from "./FormulaInput";
import {useFetching} from "../../hooks/useFetching";
import TriggersService from "../../API/TriggersService";

/*
- добавить пагинацию для динамической клавиатуры

- оптимизировать эту фигню - сделать чтобы базовый вариант спавнился тут а динамические кэшировались, чтобы клавиатура не лагала
- сделать чтобы клава вылазила когда надо будет

Баги:
некорректный рендер продвинутых выражений по типу abs
*/
const FormulaEditor = () => {
    const [formula, setFormula] = useState([
        "(", "2", "3", "+", "2", "*", "VAR3", ")", "/",
        "(", "1", "7", "+", "abs", "(", "VAR1", ")", ")",
        "≤", "2", "0", "\\textunderscore"
    ]);
    const [cursorIndex, setCursorIndex] = useState(formula.length);

    const [fetchNewFormula, isNewFormulaLoading, newFormulaError] = useFetching(async (formula) => {
        await TriggersService.getKeyboard(formula)
    }, 0, 1000)

    const sendNewFormula = async () => { // улучшить потом
        const validatedFormula = formula.filter(item => item !== "\\textunderscore");
        await fetchNewFormula(validatedFormula);
    }

    console.log(formula)

    const formulaToLatex = (tokens, cursorPos) => {
        let latex = [];
        let absStack = [];
        let sqrtStack = [];
        let isInFraction = false; // Флаг, указывающий, что мы внутри дроби
        let numerator = []; // Буфер для числителя
        let denominator = []; // Буфер для знаменателя
        let isDenominator = false; // Флаг для переключения в знаменатель

        for (let i = 0; i < tokens.length; i++) {
            if (i === cursorPos) latex.push("\\textunderscore");

            const token = tokens[i];

            if (token === "abs") {
                latex.push("\\left|");
                absStack.push(true);
            } else if (token === "sqrt") {
                latex.push("\\sqrt{");
                sqrtStack.push(true);
            } else if (token === ")" && absStack.length > 0) {
                latex.push("\\right|");
                absStack.pop();
            } else if (token === ")" && sqrtStack.length > 0) {
                latex.push("}");
                sqrtStack.pop();
            } else if (token === "/") {
                isInFraction = true;
                isDenominator = false;
                numerator = [...latex]; // сохранение текущего LateX как числителя
                latex = []; // очищение основного массива для знаменателя
            } else if (token === "(" && isInFraction) {
                isDenominator = true; // начало знаменателя
            } else if (token === ")" && isInFraction) {
                isInFraction = false; // закрытие дроби
                denominator = [...latex]; // сохранение знаменателя
                latex = [`\\frac{${numerator.join(" ")}}{${denominator.join(" ")}}`]; // создание дроби
            } else if (token === "÷") {
                latex.push("\\div");
            } else if (token === "*") {
                latex.push("\\times");
            } else if (token === ">=") {
                latex.push("\\ge");
            } else if (token === "<=") {
                latex.push("\\le");
            } else {
                latex.push(token);
            }
        }

        if (cursorPos === tokens.length) latex.push("\\textunderscore");

        // закрытие модуля если он не закрыт
        while (absStack.length > 0) {
            latex.push("\\right|");
            absStack.pop();
        }

        while (sqrtStack.length > 0) {
            latex.push("}");
            sqrtStack.pop();
        }

        return latex;
    };

    const moveCursor = (direction) => {
        const currentIndex = formula.indexOf("\\textunderscore");
        if (currentIndex === -1) return;

        let newIndex = currentIndex + direction;
        if (newIndex < 0 || newIndex >= formula.length) return;

        let newFormula = [...formula];
        newFormula.splice(currentIndex, 1);
        newFormula.splice(newIndex, 0, "\\textunderscore");

        setFormula(newFormula);
        setCursorIndex(newIndex);
    };

    const moveCursorLeft = () => moveCursor(-1);
    const moveCursorRight = () => moveCursor(1);

    const insertToken = (token) => {
        let newFormula = [...formula];
        const cursorIndex = newFormula.indexOf("\\textunderscore");

        if (token === "sqrt" || token === "abs") {
            // Вставляем `sqrt` или `abs` и сразу добавляем скобки
            newFormula.splice(cursorIndex, 0, token, "(", ")");
            setCursorIndex(cursorIndex + 1);
        } else {
            newFormula.splice(cursorIndex, 0, token);
        }

        setFormula(newFormula);
    };

    const deleteToken = () => {
        let newFormula = [...formula];
        const cursorIndex = newFormula.indexOf("\\textunderscore");
        if (cursorIndex > 0) {
            newFormula.splice(cursorIndex - 1, 1);
            setFormula(newFormula);
        }
    };

    const handleKeyPress = (key) => {
        if (key.token === "backspace") {
            deleteToken();
        } else if (key.token === "left") {
            moveCursorLeft();
        } else if (key.token === "right") {
            moveCursorRight();
        } else {
            insertToken(key.token || key.toString());
        }
    };

    return (
        <div className="section__main">
            <FormulaInput latexArray={formulaToLatex(formula)} cursorPos={cursorIndex}/>
            <Keyboard onKeyPress={handleKeyPress}/>
        </div>
    );
};

export default FormulaEditor;