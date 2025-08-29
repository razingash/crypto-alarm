import React, {useEffect, useRef, useState} from 'react';
import {Graph} from "@antv/x6";
import "../styles/workspace.css"
import {useFetching} from "../hooks/useFetching";
import WorkspaceService from "../API/WorkspaceService";
import {useNavigate, useParams} from "react-router-dom";

const Workspace = () => {
    const {id} = useParams();
    const navigate = useNavigate();
    const containerRef = useRef(null);
    const [graph, setGraph] = useState(null);
    const [showMenu, setShowMenu] = useState(false);
    const [hasStrategy, setHasStrategy] = useState(false);
    const [diagram, setDiagram] = useState(null);
    const [isModified, setIsModified] = useState(false);
    const [nodeMenu, setNodeMenu] = useState({visible: false, x: 0, y: 0, node: null,});
    // добавить обработку ошибок, чтобы их было видно в правом углу
    const [fetchSavedDiagram, , ] = useFetching(async (data) => {
        return await WorkspaceService.updateDiagram(id, data)
    }, 0, 1000)
    const [fetchDiagram, , ] = useFetching(async () => {
        return await WorkspaceService.getDiagrams({id: id})
    }, 0, 1000)

    useEffect(() => {
        const loadData = async () => {
            const res = await fetchDiagram();
            const raw = res?.data?.data ?? null;

            const g = new Graph({
                container: containerRef.current,
                grid: false,
                panning: true,
                connecting: {
                    router: "orth",
                    connector: "rounded",
                    allowBlank: false,
                    validateConnection({sourceCell, sourcePort, targetCell, targetPort}) {
                        if (sourceCell === targetCell) return false;

                        const sourceNode = g.getCellById(sourceCell);
                        const targetNode = g.getCellById(targetCell);
                        if (!sourceNode || !targetNode) return false;

                        const sourceType = sourceNode.getData()?.type;
                        const targetType = targetNode.getData()?.type;
                        if (targetType === "strategy") return false;

                        if (sourceType === "strategy" && targetType === "trigger") {
                            return sourcePort === "out" && targetPort === "in";
                        }

                        if (sourceType === "trigger" && targetType === "orchestrator") return true;
                        if (sourceType === "trigger") return false;

                        return true;
                    },
                    createEdge() {
                        return g.createEdge({
                            attrs: {
                                line: {
                                    stroke: "var(--darkened-text)",
                                    strokeWidth: 2,
                                },
                            },
                        });
                    },
                },
            });
            setGraph(g);
            let parsedData = null;
            if (raw) {
                try {
                    parsedData = JSON.parse(raw);
                } catch (err) {
                    console.error('Failed to parse diagram JSON', err, raw);
                }
            }

            if (parsedData && parsedData.cells && parsedData.cells.length > 0) {
                g.fromJSON(parsedData);

                const hasStrNode = g.getNodes().some(node => {
                    const d = node.getData?.() || node.getData?.call?.(node) || node.get('data');
                    return node.id === 'strategy' || d?.type === 'strategy';
                });

                if (!hasStrNode) {
                    createWidget("strategy", g);
                    setHasStrategy(true);
                } else {
                    setHasStrategy(true);
                }

                setDiagram(parsedData);
            } else {
                createWidget("strategy", g);
                setHasStrategy(true);
                const snapshot = g.toJSON();
                setDiagram(snapshot);
            }

            const checkChanges = () => {
                if (!g || !diagram) return;
                const current = g.toJSON();
                setIsModified(JSON.stringify(current) !== JSON.stringify(diagram));
            };

            const events = [
                'node:added',
                'node:removed',
                'edge:added',
                'edge:removed',
                'node:change:position',
                'edge:change:source',
                'edge:change:target',
            ];
            events.forEach(evt => g.on(evt, checkChanges));

            g.on("node:click", ({node}) => {
                const bbox = node.getBBox();
                const point = g.localToClient(bbox.x, bbox.y + bbox.height);
                const containerRect = containerRef.current.getBoundingClientRect();

                setNodeMenu({
                    visible: true,
                    x: point.x - containerRect.left,
                    y: point.y - containerRect.top + 8,
                    node,
                });
            });
            g.on("blank:click", () => {
                setNodeMenu({visible: false, x: 0, y: 0, node: null});
            });
            g.on("node:change:position", () => {
                setNodeMenu({visible: false, x: 0, y: 0, node: null});
            });

            const cleanup = () => {
                events.forEach(evt => g.off(evt, checkChanges));
                g.off("node:click");
                g.off("blank:click");
                g.off("node:change:position");
                try {
                    g.dispose();
                } catch (err) {
                }
            };
            (containerRef.current || {}).__x6_cleanup = cleanup;
        };

        void loadData();

        return () => {
            const cl = containerRef.current && containerRef.current.__x6_cleanup;
            if (typeof cl === 'function') cl();
        };
    }, []);

    useEffect(() => {
        if (!graph || !diagram) return;

        const checkChanges = () => {
            const currentDiagram = graph.toJSON();
            setIsModified(JSON.stringify(currentDiagram) !== JSON.stringify(diagram));
        };

        const events = [
            'node:added',
             'node:removed',
             'edge:added',
             'edge:removed',
             'node:change:position',
             'edge:change:source',
             'edge:change:target',
        ];

        events.forEach(evt => graph.on(evt, checkChanges));

        return () => {
            events.forEach(evt => graph.off(evt, checkChanges));
        };
    }, [graph, diagram]);

    const updateDiagram = async () => {
        if (!graph) return;

        const currentDiagram = graph.toJSON();
        const response = await fetchSavedDiagram({diagram: JSON.stringify(currentDiagram)});

        if (response && response.status === 200) {
            setDiagram(currentDiagram);
            setIsModified(false);
        }
    };

    const createWidget = (type, graphInstance) => {
        if (!graphInstance) return;

        const baseConfig = {
            x: 100 + Math.random() * 200,
            y: 100 + Math.random() * 200,
            width: 120,
            height: 50,
        };

        if (type === "strategy") {
            if (hasStrategy) {
                return;
            }
            graphInstance.addNode({
                id: "strategy",
                ...baseConfig,
                label: "strategy",
                data: {type: "strategy"},
                shape: "rect",
                attrs: {
                    body: {fill: "var(--container-background)", stroke: "#13C2C2", rx: 6, ry: 6},
                    label: {fill: "#fff"}
                },
                ports: {
                    groups: {
                        out: {
                            position: "right",
                            attrs: {
                                circle: {
                                    r: 8,
                                    magnet: true,
                                    fill: 'var(--container-background)',
                                    stroke: '#13C2C2',
                                    strokeWidth: 2,
                                },
                            },
                        },
                    },
                    items: [{id: "out", group: "out"}],
                },
            });
            setHasStrategy(true);
        }

        if (type === "trigger") { // скорее всего будет удален
            graphInstance.addNode({
                ...baseConfig,
                label: "Trigger",
                data: {type:"trigger"},
                shape: "rect",
                attrs: {
                    body: {fill: "var(--container-background)", stroke: "#FAAD14", rx: 4, ry: 4},
                    label: {fill: "#fff"}
                },
                ports: {
                    groups: {
                        out: {
                            position: "right",
                            attrs: {
                                circle: {
                                    r: 8,
                                    magnet: true,
                                    fill: 'var(--container-background)',
                                    stroke: '#FAAD14',
                                    strokeWidth: 2,
                                },
                            },
                        }, in: {
                            position: "left",
                            attrs: {
                                circle: {
                                    r: 8,
                                    magnet: true,
                                    fill: 'var(--container-background)',
                                    stroke: '#FAAD14',
                                    strokeWidth: 2,
                                },
                            },
                        }
                    },
                    items: [{id: "out", group: "out"}, {id: "in", group: "in"}],
                },
            });
        }

        if (type === "orchestrator") {
            graphInstance.addNode({
                ...baseConfig,
                label: "Orchestrator",
                data: {type:"orchestrator"},
                shape: "polygon",
                attrs: {
                    body: {
                        fill: "var(--container-background)",
                        stroke: "#2F54EB",
                    },
                    label: {
                        fill: "#fff",
                    },
                },
                ports: {
                    groups: {
                        in: {
                            position: "top",
                            attrs: {
                                circle: {
                                    r: 8,
                                    magnet: true,
                                    fill: 'var(--container-background)',
                                    stroke: '#2F54EB',
                                    strokeWidth: 2,
                                },
                            },
                        }, out: {
                            position: "bottom",
                            attrs: {
                                circle: {
                                    r: 8,
                                    magnet: true,
                                    fill: 'var(--container-background)',
                                    stroke: '#2F54EB',
                                    strokeWidth: 2,
                                },
                            },
                        }
                    },
                    items: [{id: "in", group: "in"}, {id: "out", group: "out"}],
                },
                points: [
                    [0.5, 0],
                    [1, 0.25],
                    [1, 0.75],
                    [0.5, 1],
                    [0, 0.75],
                    [0, 0.25],
                ],
            });
        }

        if (type === "notification") {
            graphInstance.addNode({
                ...baseConfig,
                label: "Notification",
                data: {type: "notification"},
                shape: "circle",
                attrs: {
                    body: {fill: "var(--container-background)", stroke: "#EB2F96"},
                    label: {fill: "#fff"}
                },
                ports: {
                    groups: {
                        in: {
                            position: "top",
                            attrs: {
                                circle: {
                                    r: 8,
                                    magnet: true,
                                    fill: 'var(--container-background)',
                                    stroke: '#EB2F96',
                                    strokeWidth: 2,
                                },
                            },
                        }
                    },
                    items: [{id: "in", group: "in"}],
                },
            });
        }

        setShowMenu(false);
    };

    const modifyStrategy = () => {
        const node = nodeMenu.node;
        if (!node) return;

        const strategyId = node.getData()?.strategyId;
        if (strategyId) {
            navigate(`/strategies/${strategyId}`);
        } else {
            navigate(`/new-strategy?workflowId=${id}&nodeId=${node.id}`);
        }
    }

    const modifyOrchestrator = () => {
        const node = nodeMenu.node;
        if (!node) return;

        const orchestratorId = node.getData()?.orchestratorId;
        if (orchestratorId) {
            navigate(`/orchestrator/${orchestratorId}?workflowId=${id}&nodeId=${node.id}`);
        } else {
            navigate(`/orchestrator/new?workflowId=${id}&nodeId=${node.id}`);
        }
    }

    const modifyWidget = () => {
        const node = nodeMenu.node;
        if (!node) return;

        const type = node.getData()?.type;

        switch (type) {
            case "strategy":
                modifyStrategy(node);
                break;
            case "orchestrator":
                modifyOrchestrator(node);
                break;
            case "trigger":
                //modifyTrigger(node);
                break;
            case "notification":
                //modifyNotification(node);
                break;
            default:
                console.warn("Неизвестный тип виджета:", type);
        }
    };

    return (
        <div className={"section__main"}>
            <div className={"field__workspace field__scrollable"}>
                {isModified && (
                <div className={"widget_load_data"} onClick={() => updateDiagram()}>
                    <svg className={"svg__save_to_db"}>
                        <use xlinkHref={"#icon_load_to_cloud"}></use>
                    </svg>
                </div>
                )}
                <div className={"widget__add_modules"} onClick={() => setShowMenu(!showMenu)}>+</div>
                <div ref={containerRef} className={"workspace"}/>
                {nodeMenu.visible && (
                    <div className={"widget__extensions"} style={{top: nodeMenu.y, left: nodeMenu.x}}>
                        <div onClick={() => {
                            graph.removeNode(nodeMenu.node);
                            setNodeMenu({visible: false, x: 0, y: 0, node: null})
                        }}>
                            <div className={"widget__extension"}>
                                <svg className={"svg__widget_extension"}>
                                    <use xlinkHref={"#icon_trash_can"}></use>
                                </svg>
                            </div>
                        </div>
                        <div className={"widget__extension"} onClick={() => modifyWidget()}>
                            <svg className={"svg__widget_extension"}>
                                <use xlinkHref={"#icon_gear"}></use>
                            </svg>
                        </div>
                    </div>
                )}
                <div className={`workspace__modules ${showMenu ? "show" : ""}`}>
                    <div className={"workspace__module"} onClick={() => createWidget("strategy", graph)}>
                        <svg className={"svg__workspace_modules"}>
                            <use xlinkHref={"#icon_rook"}></use>
                        </svg>
                        <div>Strategy</div>
                    </div>
                    <div className={"workspace__module"} onClick={() => createWidget("trigger", graph)}>
                        <svg className={"svg__workspace_modules"} xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20">
                            <path d="M12.03 3.313L7.936 9.224l-.052.075-1.023 1.477h2.666l-1.563 5.911 4.094-5.911.052-.075 1.023-1.477h-2.666z"/>
                            <path d="M2.007 12.5h-.633c1.224 3.012 4.177 5.137 7.629 5.137 4.232-.007 7.77-3.221 8.181-7.434l.881.881c.144.144.358.161.48.038s.105-.336-.039-.48L17.2 9.336a.43.43 0 0 0-.023-.022c-.071-.061-.155-.094-.236-.098-.081.004-.165.037-.236.098a.56.56 0 0 0-.024.022l-1.307 1.307c-.144.144-.161.357-.039.48s.336.105.48-.038l.775-.775a7.65 7.65 0 0 1-7.586 6.745c-3.123 0-5.807-1.872-6.996-4.554zM16.62 7.12h.633c-1.224-3.012-4.177-5.137-7.629-5.137-4.232.007-7.77 3.221-8.181 7.434l-.881-.881c-.144-.144-.358-.161-.48-.038s-.105.336.039.48l1.307 1.307a.43.43 0 0 0 .023.022c.071.061.155.094.236.098.081-.004.165-.037.236-.098a.56.56 0 0 0 .024-.022l1.307-1.307c.144-.144.161-.357.039-.48s-.336-.105-.48.038l-.775.775a7.65 7.65 0 0 1 7.586-6.745c3.123 0 5.807 1.872 6.996 4.554z" transform="matrix(-.7713 -.9192 .9192 -.7713 8.167 26.12)"/>
                            <path d="M56.47 31.63v.921c4.383-1.781 7.474-6.078 7.474-11.1-.01-6.158-4.688-11.31-10.82-11.9l1.282-1.282c.209-.209.234-.521.056-.699s-.489-.153-.698.056l-1.902 1.902a.59.59 0 0 0-.031.034.57.57 0 0 0-.142.343c.006.118.053.24.142.343a.78.78 0 0 0 .032.034l1.902 1.902c.209.209.52.235.698.057s.153-.489-.056-.699l-1.128-1.128c5.593.667 9.807 5.406 9.815 11.04 0 4.544-2.724 8.45-6.627 10.18z"/>
                        </svg>
                        <div>Condition</div>
                    </div>
                    <div className={"workspace__module"} onClick={() => createWidget("orchestrator", graph)}>
                        <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" className={"svg__workspace_modules"}>
                            <path d="M18.94 6.293c-.47.002-.882.314-1.011.766h-2.632l-.984.98h-1.06a3.31 3.31 0 0 0-2.655-1.353 3.31 3.31 0 1 0 0 6.62 3.31 3.31 0 0 0 2.636-1.326h1.095l.952.955h2.63c.129.453.541.766 1.011.768.584 0 1.058-.473 1.057-1.057s-.474-1.057-1.057-1.057a1.06 1.06 0 0 0-1.011.764h-2.389l-.952-.955h-.988a3.31 3.31 0 0 0 .297-1.08h2.404c.129.452.541.764 1.011.766.584 0 1.058-.473 1.057-1.057s-.474-1.057-1.057-1.057c-.47.002-.882.315-1.011.767h-2.398a3.3 3.3 0 0 0-.29-1.113h.959l.002-.002.981-.977h2.392c.129.452.541.764 1.011.766.584 0 1.058-.473 1.057-1.057s-.474-1.057-1.057-1.057zm0 .449a.61.61 0 0 1 .609.608.61.61 0 0 1-.609.609.61.61 0 0 1-.607-.6c.005-.349.275-.616.607-.617zm-8.342.553a2.7 2.7 0 1 1 0 5.402 2.7 2.7 0 1 1 0-5.402zm6.696 2.121a.61.61 0 0 1 .609.608.61.61 0 0 1-.609.609.61.61 0 0 1-.608-.609.61.61 0 0 1 .608-.608zm1.629 2.62a.61.61 0 0 1 .609.608.61.61 0 0 1-.609.609.61.61 0 0 1-.608-.609.61.61 0 0 1 .608-.608z"/>
                            <path d="M10 .001a9.99 9.99 0 0 0-2.338.301l.386 2.632-.295.092-.082.026-.3.11-.168.065-.199.088-.271.126-.069.036-.28.146-1.588-2.134c-.677.405-1.303.889-1.866 1.441-.552.563-1.036 1.189-1.441 1.866l2.134 1.588-.151.288-.029.056-.13.281-.085.193-.065.169-.11.3-.024.078-.094.299-2.632-.386A10 10 0 0 0 0 10a10 10 0 0 0 .301 2.338l2.632-.386.094.299.024.078.11.3.065.169.085.193.13.281.029.056.151.288-2.134 1.588a9.98 9.98 0 0 0 1.441 1.866c.563.552 1.189 1.036 1.866 1.441l1.573-2.114a7.35 7.35 0 0 0 1.676.691l-.383 2.61a10 10 0 0 0 2.338.301H10v-5.026a5.028 5.028 0 0 1-4.481-4.974A5.028 5.028 0 0 1 10 5.025V.001z"/>
                        </svg>
                        <div>Orchestrator</div>
                    </div>
                    <div className={"workspace__module"} onClick={() => createWidget("notification", graph)}>
                        <svg viewBox="0 0 20 20" xmlns="http://www.w3.org/2000/svg" className={"svg__workspace_modules"}>
                            <path d="m15.42 18.65c-1.634-1.103 1.2e-5 7e-6 -4.701-3.593l-0.9686 0.9075c-2.283 2.14-2.573 2.296-2.424 1.308 0.04915-0.3267 0.1771-1.439 0.2843-2.47 0.1307-1.258 0.2775-1.97 0.4456-2.16 0.3667-0.4147 8.692-7.824 8.441-7.675l-11.28 6.699-2.382-0.7962c-1.31-0.4377-2.487-0.9009-2.616-1.03-0.3133-0.3133-0.2953-0.7423 0.04334-1.035 0.3014-0.2606 18.33-7.363 18.82-7.417 0.3737-0.1225 0.8902-0.2106 0.9129 0.2399 0.0573 0.331-3.032 16.24-3.304 16.65-0.2432 0.3746-0.8213 0.5406-1.282 0.3682z"/>
                        </svg>
                        <div>Notification</div>
                    </div>
                </div>
            </div>
        </div>
    );
};

export default Workspace;