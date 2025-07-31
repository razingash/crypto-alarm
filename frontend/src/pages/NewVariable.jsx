import React, {useState} from 'react';
import "../styles/strategy.css"
import EditorVariable from "../components/Keyboard/editors/EditorVariable";
import {useFetching} from "../hooks/useFetching";
import VariablesService from "../API/VariablesService";
import {useNavigate} from "react-router-dom";
import {cleanKatexExpression, rawFormulaToFormula} from "../components/Keyboard/editors/editor";

const NewVariable = () => {
    const navigate = useNavigate();
    const [rawFormula, setRawFormula] = useState([["\\textunderscore"]]);
    const [activeFormulaIndex, setActiveFormulaIndex] = useState(0);
    const [symbol, setSymbol] = useState('');
    const [name, setName] = useState('');
    const [description, setDescription] = useState(null);
    const [fetchNewVariable, , newVariableError] = useFetching(async (data) => {
        return await VariablesService.createVariable(data)
    }, 0, 1000)

    const sendNewVariable = async (e) => {
        e.preventDefault()
        const response = await fetchNewVariable({
            "symbol": symbol,
            "name": name,
            "description": description,
            "formula": rawFormulaToFormula(rawFormula[0]),
            "formula_raw": rawFormula[0].filter(item => item !== "\\textunderscore").map(cleanKatexExpression).join('')
        });

        if (response && response?.status === 200) {
            navigate(`/variables/${response.data.variable_id}`);
        }
    };

    return (
        <div className={"section__main field__scrollable section__with_keyboard"}>
            <div className={"field__default field__keyboard_input"}>
                <form className={"new_variable__list"} onSubmit={sendNewVariable}>
                    <div className={"variable__item"}>
                        <div className={"new_variable__meaning"}>Symbol</div>
                        <div className={"new_varialbe__tip"}>This symbol will be displayed in the formulas. The symbol should contain only numbers and letters.</div>
                        <input className={"input__default input__variable"} placeholder={"input symbol..."} pattern={"[A-Za-z0-9]+"}
                               type={"text"} maxLength={40} minLength={1} onChange={(e) => setSymbol(e.target.value)}
                               title="Only letters and digits are allowed"/>
                    </div>
                    <div className={"variable__item"}>
                        <div className={"new_variable__meaning"}>Name</div>
                        <div className={"new_varialbe__tip"}>A brief description of the formula, up to 255 characters.</div>
                        <input className={"input__default input__variable"} placeholder={"input short description..."}
                               maxLength={255} minLength={5} onChange={(e) => setName(e.target.value)}/>
                    </div>
                    <div className={"variable__item"}>
                        <div className={"new_variable__meaning"}>Description</div>
                        <div className={"new_varialbe__tip"}>Full description of the formula, the characters are not limited.</div>
                        <textarea className={"textarea__default"} placeholder={"input description..."}
                            onChange={(e) => setDescription(e.target.value)}/>
                    </div>
                    <button className="button__save strategy__create" type="submit">apply</button>
                    <div className={"variable__item"}>
                        <div className={"new_variable__meaning"}>Formula</div>
                        <div className={"new_varialbe__tip"}>The formula that will use data from specific endpoints. The same data can be stored in different Binance endpoints.</div>
                        {newVariableError && <div className={"cell__error"}>{newVariableError?.error}</div>}
                        <EditorVariable rawFormulas={rawFormula} activeFormulaIndex={activeFormulaIndex}
                                setActiveFormulaIndex={setActiveFormulaIndex} setRawFormula={(index, newFormula) => {
                                setRawFormula(prev => prev.map((f, i) => i === index ? newFormula : f));
                            }}
                                deleteCondition={(index) => {
                                setRawFormula(prev => prev.filter((_, i) => i !== index));
                            }}
                        />
                    </div>
                </form>
            </div>
        </div>
    );
};

export default NewVariable;