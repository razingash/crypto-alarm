import React, {useEffect, useRef, useState} from 'react';
import "../styles/workspace.css"
import {useFetching} from "../hooks/useFetching";
import WorkspaceService from "../API/WorkspaceService";
import {useObserver} from "../hooks/useObserver";
import AdaptiveLoading from "../components/UI/AdaptiveLoading";
import ErrorField from "../components/UI/ErrorField";
import {Link, useNavigate} from "react-router-dom";

const Workspaces = () => {
    const [page, setPage] = useState(1);
    const [hasNext, setNext] = useState(false);
    const lastElement = useRef();
    const navigate = useNavigate();
    const [newDiagramName, setNewDiagramName] = useState(null)
    const [localError, setLocalError] = useState(null);
    const [diagrams, setDiagrams] = useState([])
    const [fetchDiagrams, isDiagramsLoading, diagramsError] = useFetching(async () => {
        const data = await WorkspaceService.getDiagrams({page: page})
        setDiagrams((prevDiagrams) => {
            const newDiagrams = data.data.filter(
                (diagram) => !prevDiagrams.some((obj) => obj.id === diagram.id)
            )
            return [...prevDiagrams, ...newDiagrams]
        })
        setNext(data.has_next)
    }, 0, 1000)

    const [fetchNewDiagram, , newDiagramError] = useFetching(async (name) => {
        return await WorkspaceService.createDiagram(name)
    }, 0, 1000)
    const [removeDiagram, ,] = useFetching(async (id) => {
        return await WorkspaceService.deleteDiagram(id)
    }, 0, 1000)

    useObserver(lastElement, fetchDiagrams, isDiagramsLoading, hasNext, page, setPage);

    useEffect(() => {
        const loadData = async () => {
            await fetchDiagrams();
        }
        void loadData();
        console.log(diagrams)
    }, [page])

    const sendNewDiagram = async () => {
        if (newDiagramName && newDiagramName.trim() === '') {
            setLocalError("Name is required");
            setTimeout(() => setLocalError(null), 4000);
            return;
        }

        const response = await fetchNewDiagram(newDiagramName);
        if (response && response.status === 200) {
            navigate(`/workspaces/${response.data.id}`);
        }
    };

    const handleRemoveDiagram = async (id) => {
        const isConfirmed = window.confirm('Are you sure you want delete this diagram?');
        if (!isConfirmed) return;

        const response = await removeDiagram(id);
        if (response && response.status === 200) {
            setDiagrams(prev => prev.filter(diagram => diagram.id !== id));
        }
    };

    return (
        <div className={"section__main"}>
            {isDiagramsLoading ? (
                <div className={"loading__center"}>
                    <AdaptiveLoading/>
                </div>
            ) : diagramsError ? (
                <ErrorField/>
            ) : (
            <div className={"field__workspace_list"}>
                <div className={"workspace__list field__scrollable"}>
                    <input type={"checkbox"} id={"new_workspace"}/>
                    <label className={"workspace__new__workspace"} htmlFor={"new_workspace"}>New Workspace</label>
                    <div className={"workspace__new"}>
                        <input className={"input__default input__workspace"} placeholder={"input workspace name..."}
                               type="text" maxLength={150} onChange={(e) => setNewDiagramName(e.target.value)}/>
                        {localError && <div className={"cell__error"}>{localError}</div>}
                        {newDiagramError?.error && <div className={"cell__error"}>{newDiagramError?.error}</div>}
                        <div className={"workspace__new__footer"}>
                            <label className={"button__cancle"} htmlFor="new_workspace">Cancle</label>
                            <div className={"button__save"} onClick={() => sendNewDiagram()}>Create Workspace</div>
                        </div>
                    </div>
                    <div className={"workspace__core"}>
                    {diagrams.length > 0 ? (
                        <div className={"workspace__items"}>
                            {diagrams.map((diagram) => (
                            <div className={"workspace__item"} key={diagram.id}>
                                <div className={"diagram__item"}>{diagram.name}</div>
                                <Link to={`/workspaces/${diagram.id}`} className={"link__default"}>{diagram.name}</Link>
                                <div className={"diagram__remove"} onClick={() => handleRemoveDiagram(diagram.id)}>
                                    <svg className={"svg__widget_extension"}>
                                        <use xlinkHref={"#icon_trash_can"}></use>
                                    </svg>
                                </div>
                            </div>
                            ))}
                        </div>
                    ) : (
                        <ErrorField message={"You don't possess any workspaces yet"}/>
                    )}
                    </div>
                </div>
            </div>
            )}
        </div>
    );
};

export default Workspaces;