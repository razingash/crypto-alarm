import React, {useCallback, useEffect, useRef, useState} from 'react';
import FormulaInput from "../../components/Keyboard/FormulaInput";
import "../../styles/modules.css"
import {useNavigate, useParams, useSearchParams} from "react-router-dom";
import {useFetching} from "../../hooks/useFetching";
import OrchestratorService from "../../API/Widgets/OrchestratorService";
import WorkspaceService from "../../API/WorkspaceService";

const Orchestrator = () => {
    const {id} = useParams(); // /orchestrator/:id
    const [searchParams] = useSearchParams();
    const workflowId = searchParams.get("workflowId");
    const nodeId = searchParams.get("nodeId");
    const nextId = useRef(2);
    const navigate = useNavigate();
    const [inputsState, setInputsState] = useState([]);
    const [orchestratorParts, setOrchestratorParts] = useState([]);
    const [orchestrator, setOrchestrator] = useState(null);

    const [fetchNewOrchestrator, , ] = useFetching(async (data) => {
        return await OrchestratorService.create(searchParams, data)
    }, 0, 1000)
    const [fetchOrchestrator, , ] = useFetching(async () => {
        return await OrchestratorService.get(id)
    }, 0, 1000)
    const [fetchOrchestratorParts, , ] = useFetching(async () => {
        return await OrchestratorService.getParts(searchParams)
    }, 0, 1000)
    const [fetchAttachedStrategy, ,] = useFetching(async (data) => {
        return await WorkspaceService.updateDiagramNodes(workflowId, data)
    }, 0, 1000)
    const [fetchUpdatedOrchestrator, , ] = useFetching(async (data) => {
        return await OrchestratorService.update(id, data)
    }, 0, 1000)
    const [fetchRemovedOrchestrator, , ] = useFetching(async () => {
        return await OrchestratorService.delete(id)
    }, 0, 1000)
    
    useEffect(() => {
        const loadData = async () => {
            if (id) {
                const data = await fetchOrchestrator();
                data && setOrchestrator(data)
            }
            const data = await fetchOrchestratorParts();
            data && setOrchestratorParts(data)
        }
        void loadData();
    }, [id])

    useEffect(() => {
        if (id && orchestrator?.inputs) {
            setInputsState(orchestrator.inputs.map(input => ({
                id: input.id ?? nextId.current++,
                formula: input.formula,
                tag: input.tag,
                sources: input.sources
            })));
        }
    }, [id, orchestrator]);

    const addInput = () => setInputsState(prev => [...prev, {id: nextId.current++, formula: "", tag: "", sources: []}]);
    const removeInput = (id) => setInputsState(prev => prev.filter(i => i.id !== id));

    const updateInput = useCallback((id, field, value) => {
        setInputsState(prev => prev.map(i => (i.id === id ? {...i, [field]: value} : i)));
    }, []);

    const saveOrchestrator = async () => {
        const inputs = inputsState.map(s => {
            const regex = /formula(\d+)/g;
            let match;
            const usedIds = new Set();

            while ((match = regex.exec(s.formula)) !== null) {
                usedIds.add(Number(match[1]));
            }

            const sources = Array.from(usedIds).map(csfId => {
                const part = orchestratorParts.find(p => p.csf_id === csfId);
                return part ? {source_type: "binance", source_id: part.csf_id} : null;
            }).filter(Boolean);

            return {
                formula: s.formula,
                tag: s.tag,
                sources
            };
        });
        const response = await fetchNewOrchestrator(inputs);
        if (response?.pk) {
            const resp = await fetchAttachedStrategy({
                attachOrchestrator: { nodeId, itemID: response.pk.toString() }
            });
            if (resp?.status === 200) {
                console.error(resp)
                navigate(`/orchestrator/${response.pk}?${searchParams}`);
            }
        }
    };

    return (
        <div className={"section__main"}>
            <div className={"field__default field__scrollable field__full"}>
                <div className={"cell__center"}>Attached Formulas</div>
                <div className={"associacions"}>
                    {orchestratorParts.length > 0 && orchestratorParts.map(part => (
                        <React.Fragment key={part.csf_id}>
                        <input className={"default__checkbox"} type={"checkbox"} id={`associacion_${part.csf_id}`}/>
                        <label className={"associacion__formula"} htmlFor={`associacion_${part.csf_id}`}>
                            <div>formula{part.csf_id}</div>
                            <FormulaInput formula={part.formula_raw}/>
                        </label>
                        </React.Fragment>
                    ))}
                </div>
                <span className={"span__default"}></span>
                <div className={"signals__list"}>
                    {inputsState.map(signal => (
                        <div key={signal.id} className={'signal__item'}>
                            <div className={"cell__row"}>
                                <div className={"signal__tip"}>Signal:</div>
                                <input className={"input__signal"} type="text" maxLength={100} placeholder={"input signal input..."}
                                    value={signal.formula} onChange={(e) => updateInput(signal.id, "formula", e.target.value)}
                                />
                            </div>
                            <div className={"cell__row"}>
                                <div className={"signal__tip"}>Signal Tag:</div>
                                <input className={"input__signal_outp"} type="text" maxLength={100} placeholder={"input signal output tag..."}
                                    value={signal.tag} onChange={(e) => updateInput(signal.id, "tag", e.target.value)}
                                />
                                <svg className={"svg__trash_can"} onClick={() => removeInput(signal.id)}
                                     onKeyDown={(e) => (e.key === "Enter" || e.key === " ") && removeInput(signal.id)}
                                >
                                    <use xlinkHref={"#icon_trash_can"}></use>
                                </svg>
                            </div>
                        </div>
                    ))}
                </div>
                <div className={"default__footer"}>
                    <div className={"button__cancle"} onClick={addInput}>add signal</div>
                    {id ? (
                        <div className={"button__save"} onClick={() => saveOrchestrator()}>update</div>
                    ) : (
                        <div className={"button__save"} onClick={() => saveOrchestrator()}>save</div>
                    )}
                </div>
            </div>
        </div>
    );
};

export default Orchestrator;