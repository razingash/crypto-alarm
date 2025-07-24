import React, {useEffect} from 'react';
import Keyboard from "./Keyboard";
import FormulaInput from "./FormulaInput";
/*
- оптимизировать эту фигню - сделать чтобы базовый вариант спавнился тут а динамические кэшировались, чтобы клавиатура не лагала
- сделать чтобы клава вылазила когда надо будет
*/
// возможно после добавления редактирования придется изменить компонент
const FormulaEditor = ({
                           rawFormulas, setRawFormula, deleteCondition, activeFormulaIndex, setActiveFormulaIndex,
    isNewVariable
}) => {
    const moveCursor = (direction) => {
        const currentFormula = [...rawFormulas[activeFormulaIndex]];
        const currentIndex = currentFormula.indexOf("\\textunderscore");
        if (currentIndex === -1) return;

        let moveBy = 1;

        if (direction === 1) {
            const nextToken = currentFormula[currentIndex + 1];
            if (["abs", "sqrt", "matrix", "frac", "^", '^2'].includes(nextToken)) {
                moveBy = 2;
            }
        } else if (direction === -1) {
            const tokenTwoLeft = currentFormula[currentIndex - 2];
            if (["abs", "sqrt", "matrix", "frac", "^", '^2'].includes(tokenTwoLeft)) {
                moveBy = 2;
            }
        }

        const newIndex = currentIndex + (direction * moveBy);
        if (newIndex < 0 || newIndex >= currentFormula.length) return;

        const newFormula = [...currentFormula];
        newFormula.splice(currentIndex, 1);
        newFormula.splice(newIndex, 0, "\\textunderscore");
        setRawFormula(activeFormulaIndex, newFormula);
    };

    const moveCursorLeft = () => moveCursor(-1);
    const moveCursorRight = () => moveCursor(1);

    const isInsidePower = (formula, cursorIndex) => { // костыль падла
        if (formula[cursorIndex - 1] !== ")") return false;

        let depth = 0;
        for (let i = cursorIndex - 1; i >= 0; i--) {
            const token = formula[i];

            if (token === ")") {
                depth++;
            } else if (token === "(") {
                depth--;
                if (depth === 0) {
                    return formula[i - 1] === "^";
                }
            }
        }

        return false;
    };

    const insertToken = (token) => {
        let newFormula = [...rawFormulas[activeFormulaIndex]];
        const cursorIndex = newFormula.indexOf("\\textunderscore");

        if (token === "sqrt" || token === "abs") {
            newFormula.splice(cursorIndex, 1);
            newFormula.splice(cursorIndex, 0, token, "(", "\\textunderscore", ")");
        } else if (token === '^') {
            if (isInsidePower(newFormula, cursorIndex)) {
                newFormula.splice(cursorIndex, 1);
                newFormula.splice(cursorIndex - 1, 0, "\\textunderscore");
            } else {
                newFormula.splice(cursorIndex, 1);
                newFormula.splice(cursorIndex, 0, "^", "(", "\\textunderscore", ")");
            }
        } else if (token === "^2") {
            if (isInsidePower(newFormula, cursorIndex)) {
                newFormula.splice(cursorIndex - 1, 0, "2");
            } else {
                newFormula.splice(cursorIndex, 1);
                newFormula.splice(cursorIndex, 0, "^", "(", "2", "\\textunderscore", ")");
            }
        } else if (token === "brackets-l" || token === "brackets-r") {
            newFormula.splice(cursorIndex, 1);
            newFormula.splice(cursorIndex, 0, "(", "\\textunderscore", ")");
        } else if (token === "matrix" || token === "frac") {
            newFormula.splice(cursorIndex, 1);
            newFormula.splice(cursorIndex, 0, 'matrix', '(', '\\textunderscore', ',', ')');
        } else {
            newFormula.splice(cursorIndex, 0, token);
        }

        setRawFormula(activeFormulaIndex, newFormula);
    };

    const deleteToken = () => {
        let newFormula = [...rawFormulas[activeFormulaIndex]];
        const cursorIndex = newFormula.indexOf("\\textunderscore");

        if (cursorIndex === -1) return;

        const tokenBefore = newFormula[cursorIndex - 1];
        const isWrapper = (token) => ["abs", "sqrt", "matrix", "^", "^2"].includes(token);

        if (tokenBefore === ',') {
            let matrixStart = -1;
            for (let i = cursorIndex - 2; i >= 0; i--) {
                if (rawFormulas[i] === 'matrix' && rawFormulas[i + 1] === '(') {
                    matrixStart = i;
                    break;
                }
            }

            if (matrixStart !== -1) {
                const openIndex = matrixStart + 1;
                const closeIndex = rawFormulas.indexOf(')', cursorIndex);
                const outOfBounds = closeIndex === -1;
                const isEmpty = !outOfBounds && isEmptyExpression(openIndex, closeIndex, rawFormulas);

                if (isEmpty) { // удалить всю дробь
                    newFormula.splice(matrixStart, closeIndex - matrixStart + 1);
                    newFormula.splice(matrixStart, 0, "\\textunderscore");
                } else { // просто сдвинуть курсор влево
                    newFormula.splice(cursorIndex, 1);
                    newFormula.splice(cursorIndex - 1, 0, '\\textunderscore');
                }

                setRawFormula(activeFormulaIndex, newFormula);
                return;
            }
        } else if (tokenBefore === ")") {
            let depth = 0;
            for (let i = cursorIndex - 2; i >= 0; i--) {
                if (newFormula[i] === ")") {
                    depth++;
                } else if (newFormula[i] === "(") {
                    if (depth === 0) {
                        const closeIndex = cursorIndex - 1;
                        const isEmpty = isEmptyExpression(i, closeIndex, newFormula);
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
                        const isEmpty = isEmptyExpression(cursorIndex, i, newFormula);
                        const deleteFrom = isWrapper(newFormula[cursorIndex - 2]) ? cursorIndex - 2 : cursorIndex - 1;

                        if (isEmpty) { // если выражение пустое | НЕ МЕНЯТЬ
                            newFormula.splice(cursorIndex, 1);
                            newFormula.splice(deleteFrom, i - deleteFrom);
                            newFormula.splice(deleteFrom, 0, "\\textunderscore");
                        } else { // если выражение не пустое, переместить курсор левее | НЕ МЕНЯТЬ
                            newFormula.splice(cursorIndex, 1);
                            const isWrapperBeforeParen = ["abs", "sqrt", "matrix", "^"].includes(newFormula[cursorIndex - 2]);
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

        setRawFormula(activeFormulaIndex, newFormula);
    };

    const isEmptyExpression = (openIndex, closeIndex, formula) => {
        if (closeIndex - openIndex <= 1) return true;
        const innerTokens = formula.slice(openIndex + 1, closeIndex);
        return innerTokens.every(token => token === ',');
    };

    const handleKeyPress = (key) => {
        // защищает от ошибок при нажатии на кнопки, когда нет выражений
        if (activeFormulaIndex == null || !rawFormulas[activeFormulaIndex]) return;
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

    useEffect(() => { // костыль, помогает не ловить ошибки связаные с индексом при удаление выражений
        if (rawFormulas.length === 0) {
            setActiveFormulaIndex(null);
        } else if (activeFormulaIndex >= rawFormulas.length) {
            setActiveFormulaIndex(rawFormulas.length - 1);
        }
    }, [rawFormulas.length, activeFormulaIndex, setActiveFormulaIndex]);

    return (
        <>
            {rawFormulas.map((rawFormula, index) => (
                <div className={"condition__container"} key={index} onClick={() => setActiveFormulaIndex(index)}>
                    {!isNewVariable && (
                        <svg className={"svg__trash_can"} onClick={() => deleteCondition(index)}>
                            <use xlinkHref={"#icon_trash_can"}></use>
                        </svg>
                    )}
                    <FormulaInput formula={rawFormula}/>
                </div>
            ))}
            <Keyboard onKeyPress={handleKeyPress} isNewVariable={isNewVariable}/>
        </>
    );
};

export default FormulaEditor;