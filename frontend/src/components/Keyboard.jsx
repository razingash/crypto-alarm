import React, {useEffect, useLayoutEffect, useRef, useState} from 'react';
import {MathfieldElement} from 'mathlive';
import {defaultKeyboard} from "../utils/keyboard";
import {useFetching} from "../hooks/useFetching";
import TriggersService from "../API/TriggersService";

// не удалось сделать чтобы клавиатура не закрывалась после действий вне поля клавиатуры
const Keyboard = () => {
    const ref = useRef(null);
    const [keyboard, setKeyboard] = useState([]);

    const [fetchKeyboard, isKeyboardLoading, keyBoardError] = useFetching(async () => {
        return await TriggersService.getKeyboard()
    }, 0, 3)

    useEffect(() => {
        const loadData = async () => {
            if (!isKeyboardLoading && keyboard.length === 0) {
                const data = await fetchKeyboard();
                data && setKeyboard(data);
            }
        }
        void loadData();
    }, [keyBoardError])

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