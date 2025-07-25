import React, {useEffect, useState} from 'react';
import {useFetching} from "../hooks/useFetching";
import StrategyService from "../API/StrategyService";
import {useParams} from "react-router-dom";
import FormulaInput from "../components/Keyboard/FormulaInput";
import "../styles/variables.css"

const Variable = () => {
    const {id} = useParams();
    const [variable, setVariable] = useState(null);
    const [changeMod, setChangeMod] = useState(false);
    const [variableNewData, setVariableNewData] = useState(null); // changed data
    const [fetchVariable, isVariableLoading, variableError] = useFetching(async () => {
        return await StrategyService.getStrategies({id: id})
    }, 0, 1000)

    useEffect(() => {
        const loadData = async () => {
            if (!isVariableLoading && variable === null && !variableError){
                const data = await fetchVariable();
                if (data) {
                    setVariable(data.data);
                    setVariableNewData(data.data);
                }
            }
        }
        void loadData();
    }, [isVariableLoading])

    return (
        <div className={"section__main"}>
            <div className={"field__default"}>
                <div className={"variable__container"}>
                    <div className={"variable__row"}>
                        <div className={"new_variable__meaning"}>RSI</div>
                        <input className={"input__default input__variable_change"} placeholder={"input symbol..."} maxLength={40}/>
                    </div>
                    <div className={"variable__row"}>
                        <div className={"new_variable__meaning"}>Name</div>
                        <input className={"input__default input__variable_change"} placeholder={"input short description..."} maxLength={255}/>
                    </div>
                    <div className={"variable__column"}>
                        <div className={"new_variable__meaning"}>Description</div>
                        <textarea className={"textarea__default"} placeholder={"input description..."}/>
                    </div>
                    <div className={"variable__column"}>
                        <div className={"new_variable__meaning"}>Formula</div>
                        <FormulaInput formula={"(a+b)/(2*(c-11.5))"}/>
                    </div>
                </div>
            </div>
        </div>
    );
};

export default Variable;