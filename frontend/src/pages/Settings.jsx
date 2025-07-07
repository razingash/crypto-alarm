import React, {useEffect, useState} from 'react';
import {useFetching} from "../hooks/useFetching";
import SettingsService from "../API/SettingsService";
import "../styles/settings.css"
import EndpointItem from "../components/EndpointItem";
import AdaptiveLoading from "../components/UI/AdaptiveLoading";
import ErrorField from "../components/UI/ErrorField";

const Settings = () => {
    const [endpointsSettings, setEndpointsSettings] = useState(null);
    const [fetchSettings, isSettingsLoading, SettingsError] = useFetching(async () => {
        return await SettingsService.getSettings()
    }, 1000, 1000)
    const [fetchUpdateCooldown, , ] = useFetching(async (id, cooldown) => {
        return await SettingsService.updateApiCooldown(id, cooldown)
    }, 1000, 1000)

    useEffect(() => {
        const loadData = async () => {
            if (!isSettingsLoading && endpointsSettings === null){
                const data = await fetchSettings();
                if (data) {
                    setEndpointsSettings(data.data);
                }
            }
        }
        void loadData();
    }, [isSettingsLoading])

    const handleSaveChanges = async (id, cooldown, setCooldown, oldCooldown, setOldCooldown, setChangeMod) => {
        if (oldCooldown !== cooldown) {
            const response = await fetchUpdateCooldown(id, cooldown);
            if (response && response === 'OK') {
                setCooldown(cooldown);
                setOldCooldown(cooldown)
                setChangeMod(false);
            } else {
                alert("Failed to save changes.");
            }
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
                        {endpointsSettings && endpointsSettings.map((item) => (
                            <div className={"api__item"} key={item.Id}>
                                <EndpointItem endpoint={item} handleSaveChanges={handleSaveChanges}/>
                            </div>
                        ))}
                    </div>
                </div>
            )}
        </div>
    );
};

export default Settings;