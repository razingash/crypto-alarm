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
        // из-за того как работает insertToken условие ^2 заменяется на ^
        let latex = [];
        let wrapperStack = [];

        for (let i = 0; i < tokens.length; i++) {
            const token = tokens[i];
            if (/^[a-zA-Z_]+_[a-zA-Z_]+$/.test(token)) {
                const [firstPart, secondPart] = token.split("_");
                latex.push(`\\text{\\textcolor{orange}{${firstPart}}\\_${secondPart}}`);
            } else if (token === "abs") {
                latex.push("\\left|");
                wrapperStack.push("abs");
            } else if (token === "sqrt") {
                latex.push("\\sqrt{");
                wrapperStack.push("sqrt");
            } else if (token === "^") {
                if (tokens[i-1] === ')') {
                    for (let j = i - 2; j >= 0; j--) {
                        if (tokens[j] === "(") {
                            if (tokens[j-1] === '^') {
                                console.log('BIG WIN')
                                latex.splice(latex.length - 1, 0, tokens[i + 2]);
                            }
                            break;
                        }
                    }
                } else {
                    latex.push("^{");
                    wrapperStack.push("^");
                }
            } else if (token === ")") {
                const lastWrapper = wrapperStack.pop();
                if (lastWrapper === "abs") {
                    latex.push("\\right|");
                } else if (lastWrapper === "sqrt" || lastWrapper === "^") {
                    latex.push("}");
                } else {
                    latex.push(")"); // обычная скобка
                }
            } else if (token === "frac") { // дробь - умная
                console.warn("TODO")
            } else if (token === "matrix") { // дробь
                console.warn("TODO 2")
            } else if (token === "÷") {
                latex.push("\\div");
            } else if (token === "*") {
                latex.push("\\times");
            } else if (token === ">=") {
                latex.push("\\ge");
            } else if (token === "<=") {
                latex.push("\\le");
            } else if (token === "(") {
                const expressions_formula = ['^', 'sqrt', 'abs'];

                const isWrapperCall = expressions_formula.includes(tokens[i-1])

                if (!isWrapperCall) {
                    wrapperStack.push("(");
                    latex.push(token);
                }
            } else {
                latex.push(token);
            }
        }

        // закрытие элемента если он не закрыт
        while (wrapperStack.length > 0) {
            const type = wrapperStack.pop();
            if (type === "abs") {
                latex.push("\\right|");
            } else if (type === "sqrt" || type === "^") {
                latex.push("}");
            } else if (type === "(") {
                latex.push(")");
            }
        }

        console.log('latex', latex)
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