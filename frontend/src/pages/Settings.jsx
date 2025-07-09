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
        const updated = [...editedSettings];
        updated[index].Cooldown = +value;
        setEditedSettings(updated);
    };

    const handleHistoryToggle = (index, checked) => {
        const updated = [...editedSettings];
        updated[index].IsHistoryOn = checked;
        setEditedSettings(updated);
    };

    const handleSaveAll = async () => {
        if (!initialSettings || !editedSettings) return;

        const changed = editedSettings.reduce((acc, item, index) => {
            const original = initialSettings[index];
            const update = { endpoint: item.Api };

            let hasChanges = false;
            if (item.Cooldown !== original.Cooldown) {
                update.cooldown = item.Cooldown;
                hasChanges = true;
            }
            if (item.IsHistoryOn !== original.IsHistoryOn) {
                update.history = item.IsHistoryOn;
                hasChanges = true;
            }

            if (hasChanges) {
                acc.push(update);
            }
            return acc;
        }, []);

        if (changed.length > 0) {
            const response = await fetchUpdateData(changed);
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
                    <div className={"api__list"}>
                        <div className={"api__item"}>
                            <div className={"api__availability"}>Available</div>
                            <div className={"api__history"}>History</div>
                            <div className={"api__endpoint"}>Endpoint</div>
                            <div className={"api__cooldown"}>Cooldown</div>
                            <div className={"api__last_update"}>Last Update</div>
                        </div>
                        {editedSettings && editedSettings.map((item, index) => (
                            <div className={"api__item"} key={item.Id}>
                                {item.IsActual ? (
                                    <div className={"api__availability api-actual"}>Yes</div>
                                ) : (
                                    <div className={"api__availability api-unactual"}>No</div>
                                )}
                                {changeMod ? (
                                    <input className={"api__history"} type="checkbox" checked={item.IsHistoryOn}
                                        onChange={(e) => handleHistoryToggle(index, e.target.checked)}
                                    />
                                ) : (
                                    <div className={`api__history ${item.IsHistoryOn !== initialSettings[index].IsHistoryOn ? "param__status_unsaved" : ""} ${item.IsHistoryOn ? "api-actual" : "api-unactual"}`}>
                                        {item.IsHistoryOn ? "Yes" : "No"}
                                    </div>
                                )}
                                <div className={"api__endpoint endpoint-actual"}>{item.Api}</div>
                                {changeMod ? (
                                     <input className={"input__api__cooldown"} value={item.Cooldown} placeholder={"cooldown in seconds..."}
                                        type={"number"} min={1} max={999999} onChange={(e) => handleCooldownChange(index, e.target.value)}
                                    />
                                ) : (
                                    <div className={`api__cooldown ${item.Cooldown !== initialSettings[index].Cooldown ? "param__status_unsaved" : ""}`}>{formatDuration(item.Cooldown)}</div>
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
            )}
        </div>
    );
};

export default Settings;