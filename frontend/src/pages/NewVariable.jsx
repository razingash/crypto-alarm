import React, {useState} from 'react';
import "../styles/variables.css"
import FormulaEditor from "../components/Keyboard/FormulaEditor";

const NewVariable = () => {
    const [rawFormula, setRawFormula] = useState([["\\textunderscore"]]);
    const [activeFormulaIndex, setActiveFormulaIndex] = useState(0);


    return (
        <div className={"section__main"}>
            <div className={"field__default"}>
                <div className={"new_variable__list"}>
                    <div className={"variable__item"}>
                        <div className={"new_variable__meaning"}>Symbol</div>
                        <div className={"new_varialbe__tip"}>This symbol will be displayed in the formulas.</div>
                        <input className={"input__default"} placeholder={"input symbol..."} maxLength={40}/>
                    </div>
                    <div className={"variable__item"}>
                        <div className={"new_variable__meaning"}>Name</div>
                        <div className={"new_varialbe__tip"}>A brief description of the formula, up to 255 characters.</div>
                        <input className={"input__default"} placeholder={"input short description..."} maxLength={255}/>
                    </div>
                    <div className={"variable__item"}>
                        <div className={"new_variable__meaning"}>Description</div>
                        <div className={"new_varialbe__tip"}>Full description of the formula, the characters are not limited.</div>
                        <textarea className={"textarea__default"} placeholder={"input description..."}/>
                    </div>
                    <div className={"variable__item"}>
                        <div className={"new_variable__meaning"}>Formula</div>
                        <div className={"new_varialbe__tip"}>The formula that will use data from specific endpoints. The same data can be stored in different Binance endpoints.</div>
                        <FormulaEditor rawFormulas={rawFormula} activeFormulaIndex={activeFormulaIndex}
                            setActiveFormulaIndex={setActiveFormulaIndex}
                            setRawFormula={(index, newFormula) => {
                                setRawFormula(prev => prev.map((f, i) => i === index ? newFormula : f));
                            }}
                            deleteCondition={(index) => {
                                setRawFormula(prev => prev.filter((_, i) => i !== index));
                            }}
                            isNewVariable={true}
                        />
                    </div>
                </div>
            </div>
        </div>
    );
};

export default NewVariable;