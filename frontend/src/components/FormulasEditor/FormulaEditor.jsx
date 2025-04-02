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
            <FormulaInput formula={formula} cursorPos={cursorIndex} type={2}/>
            <Keyboard onKeyPress={handleKeyPress}/>
        </div>
    );
};

export default FormulaEditor;