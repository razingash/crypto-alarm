import React, {useState} from 'react';
import Keyboard from "./Keyboard";
import FormulaInput from "./FormulaInput";
/*
- добавить пагинацию для динамической клавиатуры

- оптимизировать эту фигню - сделать чтобы базовый вариант спавнился тут а динамические кэшировались, чтобы клавиатура не лагала
- сделать чтобы клава вылазила когда надо будет
*/
const FormulaEditor = ({formula, setFormula}) => {
    const [cursorIndex, setCursorIndex] = useState(formula.length);

    const moveCursor = (direction) => {
        const currentIndex = formula.indexOf("\\textunderscore");
        if (currentIndex === -1) return;

        let moveBy = 1;

        if (direction === 1) {
            const nextToken = formula[currentIndex + 1];
            if (["abs", "sqrt", "^", '^2'].includes(nextToken)) {
                moveBy = 2;
            }
        } else if (direction === -1) {
            const tokenTwoLeft = formula[currentIndex - 2];
            if (["abs", "sqrt", "^", '^2'].includes(tokenTwoLeft)) {
                moveBy = 2;
            }
        }

        const newIndex = currentIndex + (direction * moveBy);
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

        if (token === "sqrt" || token === "abs" || token === '^') {
            newFormula.splice(cursorIndex, 1);
            newFormula.splice(cursorIndex, 0, token, "(", "\\textunderscore", ")");
            setCursorIndex(cursorIndex + 2);
        } else if (token === "^2") {
            newFormula.splice(cursorIndex, 1);
            newFormula.splice(cursorIndex, 0, "^", "(",  "2", "\\textunderscore", ")");
            setCursorIndex(cursorIndex + 2);
        } else {
            newFormula.splice(cursorIndex, 0, token);
        }

        setFormula(newFormula);
    };

    const deleteToken = () => {
        let newFormula = [...formula];
        const cursorIndex = newFormula.indexOf("\\textunderscore");

        if (cursorIndex === -1) return;

        const tokenBefore = newFormula[cursorIndex - 1];
        const isWrapper = (token) => token === "abs" || token === "sqrt" || token === '^' || token === '^2';

        if (tokenBefore === ")") {
            let depth = 0;
            for (let i = cursorIndex - 2; i >= 0; i--) {
                if (newFormula[i] === ")") depth++;
                else if (newFormula[i] === "(") {
                    if (depth === 0) {
                        const isEmpty = i === cursorIndex - 2;
                        const deleteFrom = isWrapper(newFormula[i - 1]) ? i - 1 : i;
                        if (isEmpty) { // если выражение пустое
                            newFormula.splice(deleteFrom, cursorIndex - deleteFrom);
                        } else { // если выражение не пустое, переместить курсор левее(пока что просто на 1 часть в лево)
                            newFormula.splice(cursorIndex, 1); // два сплайса...
                            newFormula.splice(cursorIndex - 1, 0, "\\textunderscore");
                        }
                        break;
                    }
                    depth--;
                }
            }
        } else if (tokenBefore === "(") {
            let depth = 0;
            for (let i = cursorIndex; i < newFormula.length; i++) {
                if (newFormula[i] === "(") {
                    depth++;
                } else if (newFormula[i] === ")") {
                    if (depth === 0) {
                        const isEmpty = i === cursorIndex + 1;
                        const deleteFrom = isWrapper(newFormula[cursorIndex - 2]) ? cursorIndex - 2 : cursorIndex - 1;

                        if (isEmpty) { // если выражение пустое | НЕ МЕНЯТЬ
                            newFormula.splice(cursorIndex, 1);
                            newFormula.splice(deleteFrom, i - deleteFrom);
                            newFormula.splice(deleteFrom, 0, "\\textunderscore");
                        } else { // если выражение не пустое, переместить курсор левее | НЕ МЕНЯТЬ
                            newFormula.splice(cursorIndex, 1);
                            const isWrapperBeforeParen = ["abs", "sqrt", "^", '^2'].includes(newFormula[cursorIndex - 2]);
                            const moveLeftBy = isWrapperBeforeParen ? 2 : 1;
                            newFormula.splice(cursorIndex - moveLeftBy, 0, "\\textunderscore");
                        }
                        break;
                    }
                    depth--;
                }
            }
        } else if (cursorIndex > 0) { // дефолт - удаление одного элемента
            newFormula.splice(cursorIndex - 1, 1);
        }

        setFormula(newFormula);
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
        <>
            <FormulaInput formula={formula} cursorPos={cursorIndex}/>
            <Keyboard onKeyPress={handleKeyPress}/>
        </>
    );
};

export default FormulaEditor;