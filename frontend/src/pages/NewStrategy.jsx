import "../styles/strategy.css"
import FormulaEditor from "../components/Keyboard/FormulaEditor";
import {useEffect, useState} from "react";
import {useFetching} from "../hooks/useFetching";
import StrategyService from "../API/StrategyService";
import {useNavigate} from "react-router-dom";

const NewStrategy = () => {
    const navigate = useNavigate();
    const [rawFormula, setRawFormula] = useState([["\\textunderscore"]]);
    const [activeFormulaIndex, setActiveFormulaIndex] = useState(0);
    const [formulaName, setFormulaName] = useState('');
    const [localError, setLocalError] = useState(null);

    const [fetchNewFormula, , newFormulaError] = useFetching(async (rawFormula, name) => {
        return await StrategyService.createStrategy(rawFormula, name)
    }, 0, 1000)

    useEffect(() => {
        console.log('rawFormula', rawFormula)
    }, [rawFormula])

    const sendNewFormula = async () => {
        if (formulaName.trim() === '') {
            setLocalError("Name is required");
            setTimeout(() => setLocalError(null), 4000);
            return;
        }

        const response = await fetchNewFormula(rawFormula, formulaName);
        if (response && response.status === 200) {
            navigate(`/strategy/${response.data.id}`);
        }
    };

    const addNewCondition = () => {
        setRawFormula(prev => [...prev, ["\\textunderscore"]]);
        setActiveFormulaIndex(rawFormula.length);
    };

    return (
        <div className={"section__main"}>
            <div className={"field__new_formula"}>
                <div className={"container__new_formula"}>
                    <input className={"strategy__name__input"} placeholder={"input formula name..."}
                       type="text" maxLength={150} onChange={(e) => setFormulaName(e.target.value)}/>
                </div>
                <FormulaEditor rawFormulas={rawFormula} activeFormulaIndex={activeFormulaIndex}
                    setActiveFormulaIndex={setActiveFormulaIndex}
                    setRawFormula={(index, newFormula) => {
                        setRawFormula(prev => prev.map((f, i) => i === index ? newFormula : f));
                    }}
                    deleteCondition={(index) => {
                        setRawFormula(prev => prev.filter((_, i) => i !== index));
                    }}
                />
                <div className={"container__new_formula"}>
                    <div className={"formula__changes"}>
                        <div className="button__save strategy__create" onClick={addNewCondition}>add condition</div>
                        <div className="button__save strategy__create" onClick={sendNewFormula}>apply</div>
                    </div>
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