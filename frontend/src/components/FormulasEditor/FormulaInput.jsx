import React, {useEffect, useRef} from 'react';
import katex from 'katex';
import "katex/dist/katex.min.css";


const renderLatex = (latexArr, cursorPos) => {
    if (!Array.isArray(latexArr)) {
        console.error("latexArr должен быть массивом");
        return '';
    }

    let result = latexArr.map((item, index) => {
        if (typeof item === "string") {
            return item;
        }
        return katex.renderToString(item.latex.replace('#@', ''), { throwOnError: false }); // эту чушь изменить
    });

    result.splice(cursorPos, 0, '<span id="cursor">|</span>');

    return result.join('');
};

const FormulaInput = ({latex, onUpdateLatex, cursorPos}) => {
    const formulaInputRef = useRef(null);

    useEffect(() => {
        if (formulaInputRef.current) {
            const latexWithCursor = renderLatex(latex, cursorPos);
            formulaInputRef.current.innerHTML = latexWithCursor;

            const cursorEl = document.getElementById("cursor");
            if (cursorEl) {
                cursorEl.scrollIntoView({ behavior: "smooth", block: "nearest", inline: "start" });
            }
        }
    }, [latex, cursorPos]);

    return (
        <div className={"formula__input"} id={"formula__input"}
             ref={formulaInputRef}
             contentEditable={"false"}
             onInput={(e) => onUpdateLatex(e.target.innerText)}
        >
        </div>
    );
};

export default FormulaInput;