import React, {useEffect, useRef, useState} from 'react';
import "../styles/strategy.css"
import {Link} from "react-router-dom";
import FormulaInput from "../components/Keyboard/FormulaInput";
import {useFetching} from "../hooks/useFetching";
import VariablesService from "../API/VariablesService";
import {useObserver} from "../hooks/useObserver";
import AdaptiveLoading from "../components/UI/AdaptiveLoading";
import ErrorField from "../components/UI/ErrorField";

const Variables = () => {
    const [page, setPage] = useState(1);
    const [hasNext, setNext] = useState(false);
    const lastElement = useRef();
    const [variables, setVariables] = useState([]);
    const [fetchVariables, isVariablesLoading, variablesError]  = useFetching(async () => {
        const data = await VariablesService.getVariables({page: page})
        setVariables((prevVariables) => {
            const newVariables = data.data.filter(
                (variable) => !prevVariables.some((obj) => obj.id === variable.id)
            )
            return [...prevVariables, ...newVariables]
        })
        setNext(data.has_next)
    }, 0, 1000)

    useObserver(lastElement, fetchVariables, isVariablesLoading, hasNext, page, setPage);

    useEffect(() => {
        const loadData = async () => {
            await fetchVariables();
        }
        void loadData();
    }, [page])

    return (
        <div className={"section__main"}>
            <div className={"field__default field__scrollable field__variables"}>
                <div className={"variables__header"}>
                    <input className={"input__default"}/>
                    <Link to={"/new-variable"} className={"link__default"}>new variable</Link>
                </div>
                {isVariablesLoading ? (
                <div className={"loading__center"}>
                    <AdaptiveLoading/>
                </div>
                ) : variablesError ? (
                    <ErrorField/>
                ) : variables.length > 0 ? (
                <div className={"variables__list"}>
                    {variables.map((variable) => (
                    <div className={"variable__item"} key={variable.id}>
                        <input type={"checkbox"} id={`variables__checkbox-${variable.id}`} className={"checkbox__variable_formula"}/>
                        <div className={"variable__header"}>
                            <Link to={`/variables/${variable.id}`} className={"variable__symbol"}>{variable.symbol}</Link>
                            <label className={"variable__formula__symbol"} htmlFor={`variables__checkbox-${variable.id}`}>
                                <svg className={"svg__switch_variable"}>
                                    <use xlinkHref={"#icon_switch_on"}></use>
                                </svg>
                                <svg className={"svg__switch_variable"}>
                                    <use xlinkHref={"#icon_switch_off"}></use>
                                </svg>
                            </label>
                        </div>
                        <div className={"variable__name"}>{variable.name}</div>
                        <div className={"varialbe__description"}>{variable.description}</div>
                        <div className={"variable__formula__container"}>
                            <div className={"variable__formula"}>
                                <FormulaInput formula={variable.formula_raw}/>
                            </div>
                        </div>
                    </div>
                    ))}
                </div>
                ) : (isVariablesLoading === false && !variablesError) && (
                <ErrorField message={"You don't possess any variables yet"}/>
                )}
            </div>
        </div>
    );
};

export default Variables;