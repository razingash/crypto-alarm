import React, {useEffect, useState} from 'react';
import {useFetching} from "../hooks/useFetching";
import {useNavigate, useParams} from "react-router-dom";
import FormulaInput from "../components/Keyboard/FormulaInput";
import "../styles/strategy.css"
import VariablesService from "../API/VariablesService";
import {getModifiedFields} from "../utils/utils";
import ErrorField from "../components/UI/ErrorField";
import AdaptiveLoading from "../components/UI/AdaptiveLoading";

const Variable = () => {
    const navigate = useNavigate();
    const {id} = useParams();
    const [variable, setVariable] = useState(null);
    const [changeMod, setChangeMod] = useState(false);
    const [variableNewData, setVariableNewData] = useState(null); // changed data
    const [fetchVariable, isVariableLoading, variableError] = useFetching(async () => {
        return await VariablesService.getVariables({id: id})
    }, 0, 1000)
    const [removeVariable, isRemovingLoading, removeFormulaError] = useFetching(async () => {
        return await VariablesService.deleteVariable(id)
    }, 0, 1000)
    const [updateVariableData, , ] = useFetching(async (newData) => {
        return await VariablesService.updateVariable(id, newData)
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

    const handleRemoveVariable = async () => {
        const isConfirmed = window.confirm("Are you sure you want to remove this variable?");
        if (!isConfirmed) return;

        const response = await removeVariable();
        if (response?.status === 200) {
            navigate("/variables");
        } else {
            alert(`Error: Failed to remove variable - ${response?.error || 'Unknown error'}`);
        }
    };

    const handleSaveChanges = async () => {
        const changedFields = getModifiedFields(variable, variableNewData);
        if (Object.keys(changedFields).length === 0) {
            alert("No changes to save.");
            return;
        }

        const response = await updateVariableData({...changedFields,});
        if (response && response.status === 200) {
            setVariable(variableNewData);
            setChangeMod(false);
            document.getElementById("variable__checkbox").checked = false;
        } else {
            alert("Failed to save changes.");
        }
    };

    return (
        <div className={"section__main"}>
            {variableError === "Network Error" && variable === null ? (
                <ErrorField/>
            ) : isVariableLoading ? (
                <div className={"loading__center"}>
                    <AdaptiveLoading/>
                </div>
            ) : variable ? (
            <div className={"field__default"}>
                <div className={"variable__container"}>
                    <div className={`variable__row ${variableNewData.symbol !== variable.symbol ? "param__status_unsaved": ""}`}>
                        <div className={"new_variable__meaning"}>Symbol</div>
                        {changeMod ? (
                            <input className={"input__default input__variable_change"}
                                type="text" maxLength={40} placeholder={"input symbol..."} value={variableNewData.symbol}
                                onChange={(e) => setVariableNewData((prev) => ({...prev, symbol: e.target.value,}))}
                            />
                        ) : (
                            <div className={"variable__symbol"}>{variableNewData.symbol || `Nameless formula with id ${variable.id}`}</div>
                        )}
                    </div>
                    <div className={`variable__row ${variableNewData.name !== variable.name ? "param__status_unsaved": ""}`}>
                        <div className={"new_variable__meaning"}>Name</div>
                        {changeMod ? (
                            <input className={`input__default input__variable_change`}
                                type="text" maxLength={255} placeholder={"input short description..."} value={variableNewData.name}
                                onChange={(e) => setVariableNewData((prev) => ({...prev, name: e.target.value,}))}
                            />
                        ) : (
                            <div>{variableNewData.name || `Nameless formula with id ${variable.id}`}</div>
                        )}
                    </div>
                    <div className={`variable__column ${variableNewData.description !== variable.description && "param__status_unsaved"}`}>
                        <div className={"new_variable__meaning"}>Description</div>
                        {changeMod ? (
                            <textarea className={"textarea__default"}
                                maxLength={1500} value={variableNewData.description} placeholder={"input description..."}
                                onChange={(e) => setVariableNewData((prev) => ({...prev, description: e.target.value,}))}
                            />
                        ) : (
                            <div className={"varialbe__description"}>{variableNewData.description}</div>
                        )}
                    </div>
                    <div className={"variable__row"}>
                        <input type="checkbox" id="variable__checkbox" onChange={() => setChangeMod((prev) => !prev)}/>
                        <div className={"button__remove"} onClick={handleRemoveVariable}>remove</div>
                        <div className={"button__save"} onClick={handleSaveChanges}>save</div>
                        <label className={"button__change"} htmlFor="variable__checkbox">change</label>
                        <label className={"button__cancle"} htmlFor="variable__checkbox">cancle</label>
                    </div>
                    <span className={"line-1"}></span>
                    <div className={"condition__container"}>
                        <FormulaInput formula={variable.formula_raw}/>
                    </div>
                </div>
            </div>
            ): (isVariableLoading === false && !variableError) && (
                <ErrorField message={"You don't possess any variables yet"}/>
            )}
        </div>
    );
};

export default Variable;