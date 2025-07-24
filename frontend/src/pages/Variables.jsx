import React from 'react';
import "../styles/variables.css"
import {Link} from "react-router-dom";
import FormulaInput from "../components/Keyboard/FormulaInput";

const Variables = () => {
    return (
        <div className={"section__main"}>
            <div className={"field__default field__variables"}>
                <div className={"variables__header"}>
                    <input className={"input__default"}/>
                    <Link to={"/new-variable"} className={"link__default"}>new variable</Link>
                </div>
                <div className={"variables__list"}>
                    <div className={"variable__item"}>
                        <input type={"checkbox"} id={"variables__checkbox-1"} className={"checkbox__variable_formula"}/>
                        <div className={"variable__header"}>
                            <div className={"variable__symbol"}>RCI</div>
                            <label className={"variable__formula__symbol"} htmlFor={"variables__checkbox-1"}>
                                <svg className={"svg__switch_variable"}>
                                    <use xlinkHref={"#icon_switch_on"}></use>
                                </svg>
                                <svg className={"svg__switch_variable"}>
                                    <use xlinkHref={"#icon_switch_off"}></use>
                                </svg>
                            </label>
                        </div>
                        <div className={"variable__name"}>Short description</div>
                        <div className={"varialbe__description"}>Huge description</div>
                        <div className={"variable__formula__container"}>
                            <div className={"variable__formula"}>
                                <FormulaInput formula={"(a+b)/(2)"}/>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    );
};

export default Variables;