import React, {useEffect, useState} from 'react';
import {useParams} from "react-router-dom";
import FormulaInput from "../components/FormulasEditor/FormulaInput";
import { useNavigate } from "react-router-dom";
import {useFetching} from "../hooks/useFetching";
import TriggersService from "../API/TriggersService";
import '../styles/strategy.css'
import Chart from "../components/UI/Chart";
import {transformData} from "../utils/utils";

const Strategy = () => {
    const navigate = useNavigate();
    const {id} = useParams();
    const [formula, setFormula] = useState(null);
    const [changeMod, setChangeMod] = useState(false);
    const [formulaNewData, setFormulaNewData] = useState(null); // changed data
    const [hasNext, setHasNext] = useState(null)
    const [historyData, setHistoryData] = useState([]);
    const [prevCursor, setPrevCursor] = useState(0); // by timestamp

    const [fetchFormula, isFormulaLoading, ] = useFetching(async () => {
        return await TriggersService.getUserFormulas({id: id})
    }, 0, 1000)
    const [updateFormulaData, , ] = useFetching(async (newData) => {
        return await TriggersService.updateUserFormula(newData)
    }, 0, 1000)
    const [removeFormula, , ] = useFetching(async () => {
        return await TriggersService.deleteUserFormula(id)
    }, 0, 1000)
    const [fetchFormulaHistory, isFormulaHistoryLoading, ] = useFetching(async () => {
        return await TriggersService.getFormulaHistory(id, 1, prevCursor)
    }, 1000, 1000)

    const loadPrevHistory = async () => {
        const data = await fetchFormulaHistory()
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
        console.log(hasNext)
    }, [hasNext])

    useEffect(() => {
        const loadData = async () => {
            if (!isFormulaLoading && formula === null){
                const data = await fetchFormula();
                if (data) {
                    setFormula(data.data);
                    setFormulaNewData(data.data);
                }
            }
        }
        void loadData();
    }, [isFormulaLoading])

    useEffect(() => {
        const loadData = async () => {
            if (formula?.is_history_on === true && historyData.length === 0) {
                await loadPrevHistory()
            }
        }
        void loadData();
    }, [formula?.is_history_on, historyData])

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
        const changedFields = getModifiedFields(formula, formulaNewData);
        if (Object.keys(changedFields).length === 0) {
            alert("No changes to save.");
            return;
        }

        const response = await updateFormulaData({
            formula_id: formula.id,
            ...changedFields,
        });

        if (response && response.status === 200) {
            setFormula(formulaNewData);
            setChangeMod(false);
            document.getElementById("strategy__checkbox").checked = false;
        } else {
            alert("Failed to save changes.");
        }
    };

    const handleRemoveFormula = async () => {
        const isConfirmed = window.confirm("Are you sure you want to remove this formula?");
        if (!isConfirmed) return;

        const response = await removeFormula();
        if (response && response.status === 200) {
            navigate("/strategies");
        } else {
            alert(`Error: Failed to remove formula ${response}`);
        }
    };

    return (
        <div className={"section__main"}>
            {formula && (
            <div className={"formula__field"}>
                <div className={"strategy__item__header"}>
                    <div className={"strategy__weight"}>Weight: 80</div>
                    <div className={`strategy__name__blocked ${formulaNewData.name !== formula.name ? "param__status_unsaved": ""}`}>
                        {changeMod ? (
                            <input className={"strategy__name__input"} type="text" maxLength={150} value={formulaNewData.name}
                                onChange={(e) => setFormulaNewData((prev) => ({...prev, name: e.target.value,}))}
                                placeholder={"input formula name..."}
                            />
                        ) : (
                            formulaNewData.name || `Nameless formula with id ${formula.id}`
                        )}
                    </div>
                </div>
                <div className={`strategy__description ${formulaNewData.description !== formula.description && "param__status_unsaved"}`}>
                    {changeMod ? (
                        <textarea className={"strategy__description__textarea"} maxLength={1500} value={formulaNewData.description}
                            onChange={(e) => setFormulaNewData((prev) => ({...prev, description: e.target.value,}))}
                            placeholder={"input formula description..."}
                        />
                    ) : (
                        formulaNewData.description
                    )}
                </div>
                <div className={"strategy__info"}>
                    <div className={"strategy__info__item"}>
                        <div>History</div>
                        {changeMod ? (
                            <label htmlFor={`history_slider${formula.id}`} className={"checkbox_zipline"}>
                                <span className={"zipline"}></span>
                                <input id={`history_slider${formula.id}`} type="checkbox" className={"switch"}
                                    checked={formulaNewData.is_history_on}
                                    onChange={(e) =>
                                        setFormulaNewData((prev) => ({
                                            ...prev,
                                            is_history_on: e.target.checked,
                                        }))
                                    }
                                />
                                <span className="slider"></span>
                            </label>
                        ) : (
                             <div className={`${formulaNewData.is_history_on !== formula.is_history_on
                                ? "param__status_unsaved" : formulaNewData.is_history_on === true
                                ? "param__status_on" : "param__status_off"
                             }`}> {formulaNewData.is_history_on === true ? "On" : "Off"}
                            </div>
                        )}
                    </div>
                    <div className={"strategy__info__item"}>
                        <div>Notifications</div>
                        {changeMod ? (
                        <label htmlFor={`notifications_slider_${formula.id}`} className={"checkbox_zipline"}>
                            <span className={"zipline"}></span>
                            <input id={`notifications_slider_${formula.id}`} type="checkbox" className={"switch"}
                                checked={formulaNewData.is_notified}
                                onChange={(e) =>
                                    setFormulaNewData((prev) => ({
                                        ...prev,
                                        is_notified: e.target.checked,
                                    }))
                                }
                            />
                            <span className="slider"></span>
                        </label>
                        ) : (
                            <div className={`${formulaNewData.is_notified !== formula.is_notified
                                ? "param__status_unsaved" : formulaNewData.is_notified === true
                                ? "param__status_on" : "param__status_off"
                             }`}> {formulaNewData.is_notified === true ? "On" : "Off"}
                            </div>
                        )}
                    </div>
                    <div className={"strategy__info__item"}>
                        <div>Active</div>
                        {changeMod ? (
                            <label htmlFor={`relevance_slider${formula.id}`} className={"checkbox_zipline"}>
                                <span className={"zipline"}></span>
                                <input id={`relevance_slider${formula.id}`} type="checkbox" className={"switch"}
                                    checked={formulaNewData.is_active}
                                    onChange={(e) => setFormulaNewData((prev) => ({
                                        ...prev,
                                        is_active: e.target.checked,
                                        }))
                                    }
                                />
                                <span className="slider"></span>
                            </label>
                        ) : (
                            <div className={`${formulaNewData.is_active !== formula.is_active
                                ? "param__status_unsaved" : formulaNewData.is_active === true
                                ? "param__status_on" : "param__status_off"
                             }`}> {formulaNewData.is_active === true ? "On" : "Off"}
                            </div>
                        )}
                    </div>
                    <div className={"strategy__info__item"}>
                        <div>Last Triggered</div>
                        <div>{formula.last_triggered || "Never"}</div>
                    </div>
                </div>
                <div className={"strategy__manipulations"}>
                    <input type="checkbox" id="strategy__checkbox" onChange={() => setChangeMod((prev) => !prev)}/>
                    <div className={"strategy__remove"} onClick={handleRemoveFormula}>remove</div>
                    <div className={"strategy__change__save"} onClick={handleSaveChanges}>save</div>
                    <label className={"strategy__change"} htmlFor="strategy__checkbox">change</label>
                    <label className={"strategy__change__cancle"} htmlFor="strategy__checkbox">cancle</label>
                </div>
                <span className={"line-1"}></span>
                <FormulaInput formula={formula.formula_raw}/>
                {historyData.length > 0 && (
                    <div className={"area__chart"}>
                        <div className="field__chart chart__strategy_history">
                            <Chart data={historyData} />
                            {!isFormulaHistoryLoading && hasNext && <svg onClick={loadPrevHistory} className="chart_additional_data" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 448 512">
                                <path
                                    d="M9.4 233.4c-12.5 12.5-12.5 32.8 0 45.3l160 160c12.5 12.5 32.8 12.5 45.3 0s12.5-32.8 0-45.3L109.2 288 416 288c17.7 0 32-14.3 32-32s-14.3-32-32-32l-306.7 0L214.6 118.6c12.5-12.5 12.5-32.8 0-45.3s-32.8-12.5-45.3 0l-160 160z">
                                </path>
                            </svg>}
                        </div>
                    </div>
                )}
            </div>
            )}
        </div>
    );
};

export default Strategy;