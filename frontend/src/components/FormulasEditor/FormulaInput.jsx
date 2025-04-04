import { useEffect, useRef } from "react";
import katex from "katex";
import "katex/dist/katex.min.css";
import "../../styles/keyboard.css"

const renderLatex = (latexArr) => {
    if (!Array.isArray(latexArr)) {
        console.error("latexArr должен быть массивом");
        return '';
    }

    let latexString = latexArr.map(item => {
        if (typeof item === 'object') {
            console.warn("Объект в формуле:", item);
            return '';
        }
        return item;
    }).join(" ");

    return latexString;
};

const FormulaInput = ({ formula, cursorPos }) => {
    const formulaInputRef = useRef(null);

    const formulaToLatexArray = (formula) => {
        if (typeof formula === "object"){ // если список, значит редактор, если нет то строка(строка только для отображения)
            return formulaToLatex(formula)
        }
        const regex = /([a-zA-Z_][a-zA-Z0-9_]*)|(\d+\.\d+|\d+)|([+\-*/^()=<>!]+)|\\textunderscore/g;
        let tokens = [];

        let match;
        while ((match = regex.exec(formula)) !== null) {
            tokens.push(match[0]);
        }

        return formulaToLatex(tokens);
    };

    const formulaToLatex = (tokens) => {
        let latex = [];
        let absStack = [];
        let sqrtStack = [];
        let isInFraction = false; // Флаг, указывающий, что мы внутри дроби
        let numerator = []; // Буфер для числителя
        let denominator = []; // Буфер для знаменателя
        let isDenominator = false; // Флаг для переключения в знаменатель

        for (let i = 0; i < tokens.length; i++) {
            const token = tokens[i];

            if (/^[a-zA-Z_]+_[a-zA-Z_]+$/.test(token)) {
                const [firstPart, secondPart] = token.split("_");
                latex.push(`\\text{\\textcolor{orange}{${firstPart}}\\_${secondPart}}`);
            } else if (token === "abs") {
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

    useEffect(() => {
        if (formulaInputRef.current) {
            try {
                let latexWithCursor = renderLatex(formulaToLatexArray(formula));

                katex.render(latexWithCursor, formulaInputRef.current, {
                    throwOnError: false,
                    displayMode: true,
                });

                const cursorNodes = formulaInputRef.current.querySelectorAll(".mord");
                cursorNodes.forEach((node) => {
                    if (node.textContent.includes("\\textunderscore")) {
                        node.innerHTML = node.innerHTML.replace("\\textunderscore", '<span id="cursor">|</span>');
                    }
                });

                const cursorEl = document.getElementById("cursor");
                if (cursorEl) {
                    cursorEl.scrollIntoView({ behavior: "smooth", block: "nearest", inline: "start" });
                }
            } catch (e) {
                console.error("Ошибка рендеринга KaTeX:", e);
            }
        }
    }, [formula]);

    return <div className="formula__input" ref={formulaInputRef}></div>;
};

export default FormulaInput;