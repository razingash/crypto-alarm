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

    const handleKeyPress = (key) => {
        setLatex((prevLatex) => {
            let newCursorPos = cursorPos;

            if (key === "backspace") {
                if (cursorPos > 0) {
                    newCursorPos -= 1;
                    setCursorPos(newCursorPos);
                    const newLatex = [...prevLatex];
                    newLatex.splice(cursorPos - 1, 1);
                    return newLatex;
                }
                return prevLatex;
            }

            if (key === "\\left") {
                newCursorPos = Math.max(0, cursorPos - 1);
                setCursorPos(newCursorPos);
                return prevLatex;
            }

            if (key === "\\right") {
                newCursorPos = Math.min(prevLatex.length, cursorPos + 1);
                setCursorPos(newCursorPos);
                return prevLatex;
            }

            const newLatex = [
                ...prevLatex.slice(0, cursorPos),
                { latex: key },
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