import React, {useCallback, useEffect, useRef, useState} from 'react';
import FormulaInput from "../../components/Keyboard/FormulaInput";
import "../../styles/modules.css"
import {useNavigate, useParams, useSearchParams} from "react-router-dom";
import {useFetching} from "../../hooks/useFetching";
import OrchestratorService from "../../API/Widgets/OrchestratorService";

const Orchestrator = () => {
    const {id} = useParams(); // /orchestrator/:id
    const [searchParams] = useSearchParams();
    const [signals, setSignals] = useState([{id: 1, formula: "", tag: ""}]);
    const nextId = useRef(2);
    const navigate = useNavigate();
    const [orchestratorParts, setOrchestratorParts] = useState([]);
    const [orchestrator, setOrchestrator] = useState(null);

    const [fetchNewOrchestrator, , ] = useFetching(async () => {
        return await OrchestratorService.create()
    }, 0, 1000)
    const [fetchUpdatedOrchestrator, , ] = useFetching(async (data) => {
        return await OrchestratorService.update(id, data)
    }, 0, 1000)
    const [fetchOrchestrator, , ] = useFetching(async () => {
        return await OrchestratorService.get(id)
    }, 0, 1000)
    const [fetchOrchestratorParts, , ] = useFetching(async () => {
        return await OrchestratorService.getParts(searchParams)
    }, 0, 1000)
    const [fetchRemovedOrchestrator, , ] = useFetching(async () => {
        return await OrchestratorService.delete(id)
    }, 0, 1000)

    useEffect(() => {
        const loadData = async () => {
            const data = await fetchOrchestratorParts();
            console.log(data)
            data && setOrchestratorParts(data)
        }
        void loadData();
        console.log(orchestratorParts)
    }, [])

    useEffect(() => {
        const loadData = async () => {
            if (id) {
                const data = await fetchOrchestrator();
                data && setOrchestrator(data)
            }
        }
        void loadData();
    }, [])

    const addSignal = useCallback(() => {
        setSignals(prev => [...prev, {id: nextId.current++, formula: "", tag: ""}]);
    }, []);

    const removeSignal = useCallback((id) => {
        setSignals(prev => prev.filter(s => s.id !== id));
    }, []);

    const updateSignal = useCallback((id, field, value) => {
        setSignals(prev => prev.map(s => (s.id === id ? {...s, [field]: value} : s)));
    }, []);

    const saveOrchestrator = async () => {
        const inputs = signals.map(s => ({
            source_type: "binance",
            formula: s.formula
        }));

        const data = await fetchNewOrchestrator(inputs);
        if (data?.id) {
            navigate(`/orchestrator/${data.id}`);
        }
    };

    return (
        <div className={"section__main"}>
            <div className={"field__default field__scrollable field__full"}>
                <div className={"cell__center"}>Attached Formulas</div>
                <div className={"associacions"}>
                    {orchestratorParts.length > 0 && orchestratorParts.map(part => (
                        <React.Fragment key={part.formula_id}>
                        <input className={"default__checkbox"} type={"checkbox"} id={`associacion_${part.formula_id}`}/>
                        <label className={"associacion__formula"} htmlFor={`associacion_${part.formula_id}`}>
                            <div>formula{part.formula_id}</div>
                            <FormulaInput formula={part.formula_raw}/>
                        </label>
                        </React.Fragment>
                    ))}
                </div>
                <span className={"span__default"}></span>
                <div className={"signals__list"}>
                    {signals.map(signal => (
                        <div key={signal.id} className={'signal__item'}>
                            <div className={"cell__row"}>
                                <div className={"signal__tip"}>Signal:</div>
                                <input className={"input__signal"} type="text" maxLength={100} placeholder={"input signal input..."}
                                    value={signal.formula} onChange={(e) => updateSignal(signal.id, "formula", e.target.value)}
                                />
                            </div>
                            <div className={"cell__row"}>
                                <div className={"signal__tip"}>Signal Tag:</div>
                                <input className={"input__signal_outp"} type="text" maxLength={100} placeholder={"input signal output tag..."}
                                    value={signal.tag} onChange={(e) => updateSignal(signal.id, "tag", e.target.value)}
                                />
                                <svg className={"svg__trash_can"} onClick={() => removeSignal(signal.id)}
                                     onKeyDown={(e) => (e.key === "Enter" || e.key === " ") && removeSignal(signal.id)}
                                >
                                    <use xlinkHref={"#icon_trash_can"}></use>
                                </svg>
                            </div>
                        </div>
                    ))}
                </div>
                <div className={"default__footer"}>
                    <div className={"button__cancle"} onClick={addSignal}>add signal</div>
                    <div className={"button__save"} onClick={() => saveOrchestrator()}>save</div>
                </div>
            </div>
        </div>
    );
};

export default Orchestrator;