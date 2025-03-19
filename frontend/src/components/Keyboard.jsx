import React, {useLayoutEffect, useRef} from 'react';
import {MathfieldElement} from 'mathlive';
import {defaultKeyboard} from "../utils/keyboard";

// не удалось сделать чтобы клавиатура не закрывалась после действий вне поля клавиатуры
const Keyboard = () => {
    const ref = useRef(null);

    useLayoutEffect(() => {
        const mfe = new MathfieldElement();
        mfe.setAttribute("virtual-keyboard-mode", "manual");
        mfe.setAttribute("virtual-keyboards", "numeric high-school-keyboard");

        window.mathVirtualKeyboard.layouts = defaultKeyboard

        if (ref.current) {
            ref.current.appendChild(mfe);
        }

    }, [])

    return <math-field className={"formula_input"} ref={ref}/>;
};


export default Keyboard;