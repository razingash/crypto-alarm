import React, {useEffect, useState} from 'react';
import "../../styles/modules.css"
import {useNavigate, useSearchParams} from "react-router-dom";
import {useFetching} from "../../hooks/useFetching";
import NotificationTelegramService from "../../API/modules/NotificationTelegramService";
import WorkspaceService from "../../API/WorkspaceService";


const NotificationTelegram = () => {
    const navigate = useNavigate();
    const [searchParams] = useSearchParams();
    const id = searchParams.get("id");
    const workflowId = searchParams.get("workflowId");
    const elementId = searchParams.get("elementId");
    const nodeId = searchParams.get("nodeId");
    const [error, setError] = useState(null);

    const [inputsState, setInputsState] = useState({
        bots: [],
        selectedBot: "new",
        name: "",
        token: "",
        chat_id: "",
        signal: "",
        message: "",
    });

    const [fetchNotification] = useFetching(async (id) => {
        return await NotificationTelegramService.get(id);
    }, 0, 1000);

    const [fetchNewNotification] = useFetching(async (data) => {
        return await NotificationTelegramService.create(searchParams, data);
    }, 0, 1000);

    const [fetchUpdatedNotification] = useFetching(async (data) => {
        return await NotificationTelegramService.update(searchParams, data);
    }, 0, 1000);

    const [fetchAttachedNotification] = useFetching(async (data) => {
        return await WorkspaceService.updateDiagramNodes(workflowId, data);
    }, 0, 1000);

    useEffect(() => {
        const loadData = async () => {
            const data = await fetchNotification(id);
            if (data?.data) {
                const res = data.data;

                const selectedBot = res.name && res.bots.includes(res.name) ? res.name : "new";
                setInputsState({
                    bots: res.bots || [],
                    selectedBot: selectedBot,
                    name: res.name || "",
                    token: res.token || "",
                    chat_id: res.chat_id || "",
                    signal: res.signal?.toString() || "",
                    message: res.message || "",
                });
            }
        };

        if (id) {
            void loadData();
        }
    }, [id]);

    const handleChange = (field, value) => {
        setInputsState((prev) => ({ ...prev, [field]: value }));
    };

    const handleBotChange = (value) => {
        if (value === "new" && id) {
            return;
        }

        if (value === "new") {
            setInputsState((prev) => ({
                ...prev,
                selectedBot: "new",
                name: "",
                token: "",
                chat_id: "",
            }));
        } else {
            const bot = inputsState.bots.find((b) => b.name === value);
            setInputsState((prev) => ({
                ...prev,
                selectedBot: value,
                name: bot?.name || "",
                token: bot?.token || "",
                chat_id: bot?.chat_id || "",
            }));
        }
    };

    const loadData = async () => {
        const data = await fetchNotification();
        if (data?.data) {
            const res = data.data;
            const selectedBot = res.name && res.bots?.includes(res.name) ? res.name : "new";

            setInputsState({
                bots: res.bots || [],
                selectedBot,
                name: res.name || "",
                token: res.token || "",
                chat_id: res.chat_id || "",
                signal: res.signal?.toString() || "",
                message: res.message || "",
            });
        }
    };

    const save = async () => {
        if (id && inputsState.selectedBot === "new") {
            return;
        }

        const payload = {
            bot:
                inputsState.selectedBot === "new"
                    ? {
                        name: inputsState.name,
                        token: inputsState.token,
                        chat_id: inputsState.chat_id,
                    }
                    : {
                        name: inputsState.selectedBot,
                    },
            message: {
                element_id: elementId,
                message: inputsState.message,
                signal:
                    inputsState.signal === "true" ||
                    inputsState.signal === true,
            },
        };

        if (inputsState.selectedBot === "new") {
            const response = await fetchNewNotification(payload);
            setError(response?.error ?? null);

            if (response?.pk) {
                const resp = await fetchAttachedNotification({
                    attachNotificationTelegram: {
                        nodeId,
                        itemID: response.pk.toString(),
                    },
                });
                if (resp?.status === 200) {
                    await loadData();
                }
            }
            return;
        }

        if (id) {
            await fetchUpdatedNotification(payload);
            return;
        }

        const response = await fetchNewNotification(payload);
        setError(response?.error ?? null);
        if (response?.pk) {
            const resp = await fetchAttachedNotification({
                attachNotificationTelegram: {
                    nodeId,
                    itemID: response.pk.toString(),
                },
            });
            if (resp?.status === 200) {
                navigate(`/notification-telegram/?id=${nodeId}&${searchParams}`);
            }
        }
    };

    return (
        <div className={"section__main"}>
            <div className={"field__default field__scrollable field__full"}>
                <div className={"cell__center"}>Telegram Notification</div>
                <span className={"span__default"}/>
                <div className={"module__list"}>
                    <div className={"module__item"}>
                        {inputsState?.bots.length > 0 && (
                            <div className="cell__row">
                                <div className="module__tip">Bot:</div>
                                <select className="select__default" value={inputsState.selectedBot}
                                     onChange={(e) => handleBotChange(e.target.value)}>
                                    {!id && <option value="new">New</option>}
                                    {inputsState.bots.map((bot, idx) => (
                                        <option key={idx} value={bot}>{bot}</option>
                                    ))}
                                </select>
                            </div>
                        )}
                        {inputsState.selectedBot  === "new" && (
                        <>
                        <div className={"cell__row"}>
                            <div className={"module__tip"}>Bot name:</div>
                            <input className="input__module" type="text" maxLength={150} placeholder="input bot name..."
                                value={inputsState.name} onChange={(e) => handleChange("name", e.target.value)}
                            />
                        </div>
                        <div className={"cell__row"}>
                            <div className={"module__tip"}>Bot token:</div>
                            <input className={"input__module"} type="text" maxLength={150} placeholder={"input bot token..."}
                                value={inputsState.token} onChange={(e) => handleChange("token", e.target.value)}/>
                        </div>
                        <div className={"cell__row"}>
                            <div className={"module__tip"}>Chat id:</div>
                            <input className={"input__module"} type="text" maxLength={150} placeholder={"input chat id..."}
                                value={inputsState.chat_id} onChange={(e) => handleChange("chat_id", e.target.value)}/>
                        </div>
                        </>
                        )}
                    </div>
                    <div className={"module__item"}>
                        <div className={"cell__row"}>
                            <div className={"module__tip"}>Signal:</div>
                            <input className={"input__module"} type="text" placeholder={"input signal..."}
                                value={inputsState.signal} onChange={(e) => handleChange("signal", e.target.value)}/>
                        </div>
                        <div className={"module__tip"}>Message:</div>
                        <textarea className="textarea__default" maxLength={1000} placeholder="input message..."
                            value={inputsState.message} onChange={(e) => handleChange("message", e.target.value)}
                        />
                    </div>
                    <div className={"default__footer"}>
                        {id ? (
                            <div className={"button__save"} onClick={() => save()}>update</div>
                        ) : (
                            <div className={"button__save"} onClick={() => save()}>save</div>
                        )}
                        {error && (
                            <div className={"cell__error"}>{error}</div>
                        )}
                    </div>
                </div>
            </div>
        </div>
    );
};

export default NotificationTelegram;