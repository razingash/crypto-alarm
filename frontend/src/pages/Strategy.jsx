import React, {useEffect, useState} from 'react';
import {useParams} from "react-router-dom";
import FormulaInput from "../components/FormulasEditor/FormulaInput";
import { useNavigate } from "react-router-dom";
import {useFetching} from "../hooks/useFetching";
import StrategyService from "../API/StrategyService";
import '../styles/strategy.css'
import ChartLinear from "../components/UI/ChartLinear";
import {transformData} from "../utils/utils";
import ErrorField from "../components/UI/ErrorField";
import AdaptiveLoading from "../components/UI/AdaptiveLoading";

const Strategy = () => {
    const navigate = useNavigate();
    const {id} = useParams();
    const [strategy, setStrategy] = useState(null);
    const [changeMod, setChangeMod] = useState(false);
    const [strategyNewData, setStrategyNewData] = useState(null); // changed data
    const [hasNext, setHasNext] = useState(null)
    const [historyData, setHistoryData] = useState([]);
    const [prevCursor, setPrevCursor] = useState(0); // by timestamp

    const [fetchStrategy, isStrategyLoading, strategyError] = useFetching(async () => {
        return await StrategyService.getStrategies({id: id})
    }, 0, 1000)
    const [updateStrategyData, , ] = useFetching(async (newData) => {
        console.log(newData)
        return await StrategyService.updateStrategy(newData)
    }, 0, 1000)
    const [removeStrategy, , ] = useFetching(async (conditionID=null) => {
        return await StrategyService.deleteStrategyOrCondition(id, conditionID)
    }, 0, 1000)
    const [fetchStrategyHistory, isFormulaHistoryLoading, ] = useFetching(async () => {
        return await StrategyService.getStrategyHistory(id, 1, prevCursor)
    }, 1000, 1000)

    const loadPrevHistory = async () => {
        const data = await fetchStrategyHistory()
        if (data?.data) {
            let newItems = transformData(data.data).reverse();
            if (prevCursor === 0) {
                setHistoryData(newItems);
            } else {
                setHistoryData(prev => [...newItems, ...prev]);
            }
            setHasNext(data.has_next)
            setPrevCursor(data.data[data.data.length - 1].timestamp)
        }
    }

    useEffect(() => {
        const loadData = async () => {
            if (!isStrategyLoading && strategy === null && !strategyError){
                const data = await fetchStrategy();
                if (data) {
                    setStrategy(data.data);
                    setStrategyNewData(data.data);
                }
            }
        }
        void loadData();
    }, [isStrategyLoading])

    useEffect(() => {
        const loadData = async () => {
            if (strategy?.is_history_on === true && historyData.length === 0) {
                await loadPrevHistory()
            }
        }
        void loadData();
    }, [strategy?.is_history_on, historyData])

    const getModifiedFields = (original, modified) => {
        const changes = {};
        for (const key in original) {
            if (original[key] !== modified[key]) {
                changes[key] = modified[key];
            }
        }
        return changes;
    };

    const handleSaveChanges = async () => {
        const changedFields = getModifiedFields(strategy, strategyNewData);
        if (Object.keys(changedFields).length === 0) {
            alert("No changes to save.");
            return;
        }

        const response = await updateStrategyData({
            strategy_id: strategy.id,
            ...changedFields,
        });

        if (response && response.status === 200) {
            setStrategy(strategyNewData);
            setChangeMod(false);
            document.getElementById("strategy__checkbox").checked = false;
        } else {
            alert("Failed to save changes.");
        }
    };

    const handleRemoveFormula = async () => {
        const isConfirmed = window.confirm("Are you sure you want to remove this formula?");
        if (!isConfirmed) return;

        const response = await removeStrategy();
        if (response && response.status === 200) {
            navigate("/strategies");
        } else {
            alert(`Error: Failed to remove formula ${response}`);
        }
    };

    const handleRemoveCondition = async (formulaId) => {
        const isConfirmed = window.confirm('Are you sure you want to delete this condition?');
        if (!isConfirmed) return;

        const response = await removeStrategy(formulaId);
        if (response && response.status === 200) {
            setStrategy(prev => ({
                ...prev,
                conditions: prev.conditions.filter(cond => cond.formula_id !== formulaId)
            }));
        }
    };

    return (
        <div className={"section__main"}>
            {strategyError === "Network Error" && strategy === null ? (
                <ErrorField/>
            ) : isStrategyLoading ? (
                <div className={"loading__center"}>
                    <AdaptiveLoading/>
                </div>
            ) : strategy ? (
            <div className={"strategy__field"}>
                <div className={"strategy__item__header"}>
                    <div className={"strategy__weight"}>Weight: 80</div>
                    <div className={`strategy__name__blocked ${strategyNewData.name !== strategy.name ? "param__status_unsaved": ""}`}>
                        {changeMod ? (
                            <input className={"strategy__name__input"} type="text" maxLength={150} value={strategyNewData.name}
                                onChange={(e) => setStrategyNewData((prev) => ({...prev, name: e.target.value,}))}
                                placeholder={"input formula name..."}
                            />
                        ) : (
                            strategyNewData.name || `Nameless formula with id ${strategy.id}`
                        )}
                    </div>
                </div>
                <div className={`strategy__description ${strategyNewData.description !== strategy.description && "param__status_unsaved"}`}>
                    {changeMod ? (
                        <textarea className={"strategy__description__textarea"} maxLength={1500} value={strategyNewData.description}
                            onChange={(e) => setStrategyNewData((prev) => ({...prev, description: e.target.value,}))}
                            placeholder={"input formula description..."}
                        />
                    ) : (
                        strategyNewData.description
                    )}
                </div>
                <div className={"strategy__info"}>
                    <div className={"strategy__info__item"}>
                        <div>History</div>
                        {changeMod ? (
                            <label htmlFor={`history_slider${strategy.id}`} className={"checkbox_zipline"}>
                                <span className={"zipline"}></span>
                                <input id={`history_slider${strategy.id}`} type="checkbox" className={"switch"}
                                    checked={strategyNewData.is_history_on}
                                    onChange={(e) =>
                                        setStrategyNewData((prev) => ({
                                            ...prev,
                                            is_history_on: e.target.checked,
                                        }))
                                    }
                                />
                                <span className="slider"></span>
                            </label>
                        ) : (
                             <div className={`${strategyNewData.is_history_on !== strategy.is_history_on
                                ? "param__status_unsaved" : strategyNewData.is_history_on === true
                                ? "param__status_on" : "param__status_off"
                             }`}> {strategyNewData.is_history_on === true ? "On" : "Off"}
                            </div>
                        )}
                    </div>
                    <div className={"strategy__info__item"}>
                        <div>Notifications</div>
                        {changeMod ? (
                        <label htmlFor={`notifications_slider_${strategy.id}`} className={"checkbox_zipline"}>
                            <span className={"zipline"}></span>
                            <input id={`notifications_slider_${strategy.id}`} type="checkbox" className={"switch"}
                                checked={strategyNewData.is_notified}
                                onChange={(e) =>
                                    setStrategyNewData((prev) => ({
                                        ...prev,
                                        is_notified: e.target.checked,
                                    }))
                                }
                            />
                            <span className="slider"></span>
                        </label>
                        ) : (
                            <div className={`${strategyNewData.is_notified !== strategy.is_notified
                                ? "param__status_unsaved" : strategyNewData.is_notified === true
                                ? "param__status_on" : "param__status_off"
                             }`}> {strategyNewData.is_notified === true ? "On" : "Off"}
                            </div>
                        )}
                    </div>
                    <div className={"strategy__info__item"}>
                        <div>Active</div>
                        {changeMod ? (
                            <label htmlFor={`relevance_slider${strategy.id}`} className={"checkbox_zipline"}>
                                <span className={"zipline"}></span>
                                <input id={`relevance_slider${strategy.id}`} type="checkbox" className={"switch"}
                                    checked={strategyNewData.is_active}
                                    onChange={(e) => setStrategyNewData((prev) => ({
                                        ...prev,
                                        is_active: e.target.checked,
                                        }))
                                    }
                                />
                                <span className="slider"></span>
                            </label>
                        ) : (
                            <div className={`${strategyNewData.is_active !== strategy.is_active
                                ? "param__status_unsaved" : strategyNewData.is_active === true
                                ? "param__status_on" : "param__status_off"
                             }`}> {strategyNewData.is_active === true ? "On" : "Off"}
                            </div>
                        )}
                    </div>
                    <div className={"strategy__info__item"}>
                        <div>Cooldown</div>
                        {changeMod ? (
                            <input type="number" min={1} max={604800} className={"input__strategy__cooldown"}
                                value={strategyNewData.cooldown}
                                onChange={(e) => setStrategyNewData((prev) => ({
                                    ...prev,
                                    cooldown: +e.target.value,
                                    }))
                                }
                            />
                        ) : (
                            <div className={`${strategyNewData.cooldown !== strategy.cooldown &&"param__status_unsaved"}`}>
                                {strategyNewData.cooldown}
                            </div>
                        )}
                    </div>
                    <div className={"strategy__info__item"}>
                        <div>Last Triggered</div>
                        <div>{strategy.last_triggered || "Never"}</div>
                    </div>
                </div>
                <div className={"strategy__manipulations"}>
                    <input type="checkbox" id="strategy__checkbox" onChange={() => setChangeMod((prev) => !prev)}/>
                    <div className={"button__remove"} onClick={handleRemoveFormula}>remove</div>
                    <div className={"button__save"} onClick={handleSaveChanges}>save</div>
                    <label className={"button__change"} htmlFor="strategy__checkbox">change</label>
                    <label className={"button__cancle"} htmlFor="strategy__checkbox">cancle</label>
                </div>
                <span className={"line-1"}></span>
                {strategy.conditions.map((condition, index) => (
                    <div className={"condition__container"} key={index}>
                        {changeMod && <svg className={"svg__trash_can"} onClick={() => handleRemoveCondition(condition.formula_id)}>>
                            <use xlinkHref={"#icon_trash_can"}></use>
                        </svg>}
                        <FormulaInput formula={condition.formula_raw}/>
                    </div>
                ))}
                {historyData.length > 0 && (
                    <div className={"area__chart"}>
                        <div className="field__chart chart__strategy_history">
                            <ChartLinear data={historyData} />
                            {!isFormulaHistoryLoading && hasNext && <svg onClick={loadPrevHistory} className="chart_additional_data" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 448 512">
                                <path
                                    d="M9.4 233.4c-12.5 12.5-12.5 32.8 0 45.3l160 160c12.5 12.5 32.8 12.5 45.3 0s12.5-32.8 0-45.3L109.2 288 416 288c17.7 0 32-14.3 32-32s-14.3-32-32-32l-306.7 0L214.6 118.6c12.5-12.5 12.5-32.8 0-45.3s-32.8-12.5-45.3 0l-160 160z">
                                </path>
                            </svg>}
                        </div>
                    </div>
                )}
            </div>
            ) : (isStrategyLoading === false && strategyError) && (
                <ErrorField message={`The formula with ID ${id} does not exist`}/>
            )}
        </div>
    );
};

export default Strategy;