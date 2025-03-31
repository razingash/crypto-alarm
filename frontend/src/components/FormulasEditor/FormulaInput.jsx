import { useEffect, useRef } from "react";
import katex from "katex";
import "katex/dist/katex.min.css";


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

const FormulaInput = ({ latexArray }) => {
    const formulaInputRef = useRef(null);

    useEffect(() => {
        if (formulaInputRef.current) {
            try {
                let latexWithCursor = renderLatex(latexArray);

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
    }, [latexArray]);

    return <div className="formula__input" ref={formulaInputRef}></div>;
};

export default FormulaInput;