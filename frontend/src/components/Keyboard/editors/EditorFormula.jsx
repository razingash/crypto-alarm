import React, {useEffect} from 'react';
import FormulaInput from "../FormulaInput";
import Keyboard from "../Keyboard";
import {deleteToken, insertToken, moveCursor} from "./editor";

const EditorFormula = ({rawFormulas, setRawFormula, deleteCondition, activeFormulaIndex, setActiveFormulaIndex}) => {
    // пока не переносить, если будет больше эдиторов то вынести в провайдер
    const handleKeyPress = (key) => {
        // защищает от ошибок при нажатии на кнопки, когда нет выражений
        if (activeFormulaIndex == null || !rawFormulas[activeFormulaIndex]) return;
        if (key.token === "backspace") {
            deleteToken(setRawFormula, rawFormulas, activeFormulaIndex);
        } else if (key.token === "left") {
            moveCursor(-1, setRawFormula, activeFormulaIndex, rawFormulas);
        } else if (key.token === "right") {
            moveCursor(1, setRawFormula, activeFormulaIndex, rawFormulas);
        } else {
            insertToken(key.token || key.toString(), setRawFormula, activeFormulaIndex, rawFormulas);
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
            <input type={"checkbox"} id={"editor"} defaultChecked={false}/>
            <label htmlFor={"editor"} className={"editor"}>
            {rawFormulas.map((rawFormula, index) => (
                <div className={"condition__container"} key={index} onClick={() => setActiveFormulaIndex(index)}>
                    <svg className={"svg__trash_can"} onClick={() => deleteCondition(index)}>
                        <use xlinkHref={"#icon_trash_can"}></use>
                    </svg>
                    <FormulaInput formula={rawFormula}/>
                </div>
            ))}
            </label>
            <Keyboard onKeyPress={handleKeyPress}/>
        </>
    );
};

export default EditorFormula;