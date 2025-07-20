import React, {useEffect, useState} from 'react';
import {useFetching} from "../hooks/useFetching";
import SettingsService from "../API/SettingsService";
import "../styles/settings.css"
import AdaptiveLoading from "../components/UI/AdaptiveLoading";
import ErrorField from "../components/UI/ErrorField";
import {formatDuration, formatTimestamp} from "../utils/utils";

const Settings = () => {
    const [changeMod, setChangeMod] = useState(false);
    const [initialSettings, setInitialSettings] = useState(null);
    const [editedSettings, setEditedSettings] = useState(null);
    const [fetchSettings, isSettingsLoading, SettingsError] = useFetching(async () => {
        return await SettingsService.getSettings()
    }, 1000, 1000)

    const [fetchUpdateData, ,] = useFetching(async (updates) => {
        return await SettingsService.updateApiSettings(updates);
    });

    useEffect(() => {
        const loadData = async () => {
            if (!isSettingsLoading && initialSettings === null) {
                const response = await fetchSettings();
                if (response?.data) {
                    setInitialSettings(JSON.parse(JSON.stringify(response.data)));
                    setEditedSettings(JSON.parse(JSON.stringify(response.data)));
                }
            }
        };
        void loadData();
    }, [isSettingsLoading]);

    const handleCooldownChange = (index, value) => {
        const updated = { ...editedSettings, api: [...editedSettings.api] };
        updated.api[index].cooldown = +value;
        setEditedSettings(updated);
    };

    const handleHistoryToggle = (index, checked) => {
        const updated = {...editedSettings, api: [...editedSettings.api]};
        updated.api[index].is_history_on = checked;
        setEditedSettings(updated);
    };

    const handleConfigToggle  = (index, checked) => {
        const updated = {...editedSettings, config: [...editedSettings.config]};
        updated.config[index].is_active = checked;
        setEditedSettings(updated);
    }

    const handleSaveAll = async () => {
        if (!initialSettings || !editedSettings) return;

        const apiChanged = editedSettings.api.reduce((acc, item) => {
            const original = initialSettings.api.find(o => o.id === item.id);
            if (!original) return acc;

            const update = {endpoint: item.api};
            let hasChanges = false;

            if (item.cooldown !== original.cooldown) {
                update.cooldown = item.cooldown;
                hasChanges = true;
            }
            if (item.is_history_on !== original.is_history_on) {
                update.history = item.is_history_on;
                hasChanges = true;
            }

            if (hasChanges) {
                acc.push(update);
            }
            return acc;
        }, []);

        const configChanged = editedSettings.config.reduce((acc, item) => {
            const original = initialSettings.config.find(o => o.id === item.id);
            if (!original) return acc;

            if (item.is_active !== original.is_active) {
                acc.push({
                    id: item.id,
                    is_active: item.is_active
                });
            }
            return acc;
        }, []);

        const payload = {};
        if (apiChanged.length > 0) payload.api = apiChanged;
        if (configChanged.length > 0) payload.config = configChanged;

        if (Object.keys(payload).length > 0) {
            const response = await fetchUpdateData(payload);
            if (response === "OK") {
                setInitialSettings(JSON.parse(JSON.stringify(editedSettings)));
                setChangeMod(false);
                document.getElementById("api__checkbox").checked = false;
                alert("Changes saved successfully.");
            } else {
                alert("Failed to save changes.");
            }
        } else {
            alert("No changes to save.");
        }
    };

    return (
        <div className={"section__main"}>
            {isSettingsLoading ? (
                <div className={"loading__center"}>
                    <AdaptiveLoading/>
                </div>
            ) : SettingsError ? (
                <ErrorField/>
            ) : (
            <div className={"field__settings__api"}>
                <div className={"area__settings"}>
                    <div className={"settings__config"}>
                        {editedSettings?.config && editedSettings.config.map((item, index) => (
                        <div className={"config__item"} key={item.id}>
                            {changeMod ? (
                            <input id={"config__asl"} className={"checkbox__item"} type={"checkbox"}
                                   checked={item.is_active} onChange={(e) => handleConfigToggle(index, e.target.checked)}/>
                            ) : (
                            <div className={`api__history ${item.is_active !== initialSettings.config[index].is_active ? "param__status_unsaved" : ""} ${item.is_active ? "api-actual" : "api-unactual"}`}>
                                {item.is_active ? "On" : "Off"}
                            </div>
                            )}
                            <div>Average System Load</div>
                        </div>
                        ))}
                    </div>
                    <div className={"settings__api__list"}>
                        <div className={"api__item"}>
                            <div className={"api__availability"}>Available</div>
                            <div className={"api__history"}>History</div>
                            <div className={"api__endpoint"}>Endpoint</div>
                            <div className={"api__cooldown"}>Cooldown</div>
                            <div className={"api__last_update"}>Last Update</div>
                        </div>
                        {editedSettings?.api && editedSettings.api.map((item, index) => (
                            <div className={"api__item"} key={item.id}>
                                {item.is_actual ? (
                                    <div className={"api__availability api-actual"}>Yes</div>
                                ) : (
                                    <div className={"api__availability api-unactual"}>No</div>
                                )}
                                {changeMod ? (
                                    <input className={"api__history"} type="checkbox" checked={item.is_history_on}
                                        onChange={(e) => handleHistoryToggle(index, e.target.checked)}
                                    />
                                ) : (
                                    <div className={`api__history ${item.is_history_on !== initialSettings.api[index].is_history_on ? "param__status_unsaved" : ""} ${item.is_history_on ? "api-actual" : "api-unactual"}`}>
                                        {item.is_history_on ? "Yes" : "No"}
                                    </div>
                                )}
                                <div className={"api__endpoint endpoint-actual"}>{item.Api}</div>
                                {changeMod ? (
                                     <input className={"input__api__cooldown"} value={item.cooldown} placeholder={"cooldown in seconds..."}
                                        type={"number"} min={1} max={999999} onChange={(e) => handleCooldownChange(index, e.target.value)}
                                    />
                                ) : (
                                    <div className={`api__cooldown ${item.cooldown !== initialSettings.api[index].cooldown ? "param__status_unsaved" : ""}`}>{formatDuration(item.cooldown)}</div>
                                )}
                                <div className={"api__last_update"}>{formatTimestamp(Math.floor(Date.parse(item.LastUpdate) / 1000))}</div>
                            </div>
                        ))}
                        <div className={"api__settings"}>
                            <input type="checkbox" id="api__checkbox" onChange={() => setChangeMod((prev) => !prev)}/>
                            <label id={"api__change"} className={"button__change"} htmlFor="api__checkbox">change</label>
                            <label id={"api__cancle"} className={"button__cancle"} htmlFor="api__checkbox">cancle</label>
                            <div id={"api__save"} className={"button__save"} onClick={handleSaveAll}>save</div>
                        </div>
                    </div>
                </div>
            </div>
            )}
        </div>
    );
};

export default Settings;