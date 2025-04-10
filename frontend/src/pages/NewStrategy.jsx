import "../styles/strategy.css"
import FormulaEditor from "../components/FormulasEditor/FormulaEditor";
import {useEffect, useState} from "react";
import {useFetching} from "../hooks/useFetching";
import TriggersService from "../API/TriggersService";
import {useNavigate} from "react-router-dom";

const NewStrategy = () => {
    const navigate = useNavigate();
    const [formula, setFormula] = useState(["\\textunderscore"]);
    const [formulaName, setFormulaName] = useState('');
    const [localError, setLocalError] = useState(null);

    const [fetchNewFormula, , newFormulaError] = useFetching(async (formula, name) => {
        return await TriggersService.createFormula(formula, name)
    }, 0, 1000)

    useEffect(() => {
        console.log('formula', formula)
    }, [formula])

    function cleanKatexExpression(expr) { // проверки на правильность тут не будет
        return expr
            .replace(/\\textcolor{[^}]+}{([^}]+)}/g, '$1')
            .replace(/\\text{([^}]+)}/g, '$1')
            .replace(/\\_/g, '_')
            .replace(/\\\\/g, '\\');
    }

    const sendNewFormula = async () => {
        if (formulaName.trim() === '') {
            setLocalError("Name is required");
            setTimeout(() => setLocalError(null), 4000);
            return;
        }

        const cleanedFormula = formula
            .filter(item => item !== "\\textunderscore")
            .map(cleanKatexExpression)
            .join('');

        const response = await fetchNewFormula(cleanedFormula, formulaName);
        if (response && response.status === 200) {
            navigate(`/strategy/${response.data.id}`);
        }
    };

    return (
        <div className={"section__main"}>
            <div className={"field__new_formula"}>
                {<FormulaEditor formula={formula} setFormula={setFormula}/>}
                <div className={"new_formula__core"}>
                    <input className={"strategy__name__input"} placeholder={"input formula name..."}
                       type="text" maxLength={150} onChange={(e) => setFormulaName(e.target.value)}/>
                    <div className="strategy__change__save strategy__create" onClick={sendNewFormula}>apply</div>
                    <div className={"field__new_formula_errors"}>
                        {localError && <div className={"cell__error"}>Notification: {localError}</div>}
                        {newFormulaError && <div className={"cell__error"}>Error: {newFormulaError?.error}</div>}
                    </div>
                </div>
            </div>
        </div>
    );
};

export default NewStrategy;