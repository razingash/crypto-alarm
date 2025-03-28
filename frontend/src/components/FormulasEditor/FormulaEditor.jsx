import React, {useState} from 'react';
import "../../styles/keyboard.css"
import KeyboardV2 from "./KeyboardV2";
import FormulaInput from "./FormulaInput";

/*
- добавить пагинацию для динамической клавиатуры
- сделать чтобы шрифт уменьшался с экраном - как вообще без понятия
- сделать кнопки 2 вложенности

- оптимизировать эту фигню - сделать чтобы базовый вариант спавнился тут а динамические кэшировались, чтобы клавиатура не лагала
- сделать чтобы клава вылазила когда надо будет
*/
const FormulaEditor = () => {
    const [latex, setLatex] = useState([]);
    const [cursorPos, setCursorPos] = useState(0);

    // обработчик нажатия клавиш
    const handleKeyPress = (key) => {
        console.log(key)
        setLatex((prevLatex) => {
            let newCursorPos = cursorPos;

            // обработка специальных выражений
            const addLatex = (latexString) => {
                const newLatex = [
                    ...prevLatex.slice(0, cursorPos),
                    { latex: latexString, type: 'expression' },
                    ...prevLatex.slice(cursorPos)
                ];
                newCursorPos += 1;
                setCursorPos(newCursorPos);
                return newLatex;
            };

            if (key.latex) {
                // удаление элемента
                if (key.id === "backspace") {
                    if (cursorPos > 0) {
                        newCursorPos -= 1;
                        setCursorPos(newCursorPos);
                        const newLatex = [...prevLatex];
                        newLatex.splice(cursorPos - 1, 1);
                        return newLatex;
                    }
                    return prevLatex;
                }

                // переход курсора по элементам
                if (key.id === "swl") {
                    newCursorPos = Math.max(0, cursorPos - 1);
                    setCursorPos(newCursorPos);
                    return prevLatex;
                } else if (key.id === "swr") {
                    newCursorPos = Math.min(prevLatex.length, cursorPos + 1);
                    setCursorPos(newCursorPos);
                    return prevLatex;
                }

                // знаки выражений
                if (key.id === 'lt') {
                    return addLatex('\\lt');
                } else if (key.id === 'le') {
                    return addLatex('\\le');
                } else if (key.id === 'gt') {
                    return addLatex('\\gt');
                } else if (key.id === 'ge') {
                    return addLatex('\\ge');
                }

                // умножение и деление
                if (key.id === 'div') {
                    return addLatex('\\div');
                } else if (key.id === 'times') {
                    return addLatex('\\times');
                }

                // НИЖЕ НЕ РАБОТАЕТ

                // модуль
                if (key.id === 'mo') {
                    return addLatex('\\left\\lvert {{▢}} \\right\\rvert');
                }

                // Дробь
                if (key.id === 'frac') {
                    return addLatex('\\frac{#@}{#0}');
                }

                // Матрица
                if (key.id === 'matrix') {
                    return addLatex('\\begin{pmatrix}#0\\\\#0\\end{pmatrix}');
                }

                // Корень
                if (key.id === 'sq') {
                    return addLatex('\\sqrt{#@}');
                }

                // Степень
                if (key.id === 'square2') {
                    return addLatex('#@^{#?}');
                }

                // Квадрат
                if (key.id === 'square') {
                    return addLatex('#@^2');
                }
            }

            // для остальных клавиш, просто символ
            const newLatex = [
                ...prevLatex.slice(0, cursorPos),
                { latex: key, type: 'text' },
                ...prevLatex.slice(cursorPos)
            ];
            newCursorPos += 1;
            setCursorPos(newCursorPos);

            return newLatex;
        });
    };

    return (
        <div className={"section__main"}>
            <FormulaInput latex={latex} onUpdateLatex={setLatex} cursorPos={cursorPos} />
            <KeyboardV2 onKeyPress={handleKeyPress} />
        </div>
    );
};


export default FormulaEditor;