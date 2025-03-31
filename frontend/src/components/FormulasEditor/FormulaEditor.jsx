import React, {useState} from 'react';
import "../../styles/keyboard.css"
import Keyboard from "./Keyboard";
import FormulaInput from "./FormulaInput";

/*
- добавить пагинацию для динамической клавиатуры
- сделать чтобы шрифт уменьшался с экраном - как вообще без понятия

- оптимизировать эту фигню - сделать чтобы базовый вариант спавнился тут а динамические кэшировались, чтобы клавиатура не лагала
- сделать чтобы клава вылазила когда надо будет

Баги:
нельзя выйти за модуль если он в конце(хз пока как решать эту проблему, надо что-то не стандартное придумать)
*/
const FormulaEditor = () => {
    const [formula, setFormula] = useState([
        "(", "2", "3", "+", "2", "*", "VAR3", ")", "/",
        "(", "1", "7", "+", "abs", "(", "VAR1", ")", ")",
        "≤", "2", "0", "\\textunderscore"
    ]);
    const [cursorIndex, setCursorIndex] = useState(formula.length);

    const formulaToLatex = (tokens, cursorPos) => {
        let latex = [];
        let absStack = [];
        let fracStack = [];

        for (let i = 0; i < tokens.length; i++) {
            if (i === cursorPos) latex.push("\\textunderscore");

            const token = tokens[i];

            if (token === "abs") {
                latex.push("\\left|");
                absStack.push(true);
            } else if (token === "sqrt") {
                latex.push("sqrt1");
            } else if (token === ")" && absStack.length > 0) {
                latex.push("\\right|");
                absStack.pop();
            } else if (token === "/") {
                latex.push("\\frac{");
                fracStack.push(true);
            } else if (token === "matrix") {
                latex.push("matrix1");
            } else if (token === "(" && fracStack.length > 0) {
                latex.push("");
            } else if (token === ")" && fracStack.length > 0) {
                latex.push("}");
                fracStack.pop();
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

        // Закрываем открытые `|` и `{` (если дробь не была закрыта)
        while (absStack.length > 0) {
            latex.push("\\right|");
            absStack.pop();
        }

        while (fracStack.length > 0) {
            latex.push("}");
            fracStack.pop();
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
        newFormula.splice(cursorIndex, 0, token);
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