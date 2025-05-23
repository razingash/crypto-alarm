import React, {useState} from 'react';
import {formatDuration, formatTimestamp} from "../utils/utils";

const EndpointItem = ({endpoint, handleSaveChanges}) => {
    const [changeMod, setChangeMod] = useState(false);
    const [oldCooldown, setOldCooldown] = useState(endpoint.Cooldown);
    const [cooldown, setCooldown] = useState(endpoint.Cooldown);

    return (
        <>
            {endpoint.IsActual ? (
                <div className={"api__status api-actual"}></div>
            ) : (
                <div className={"api__status api-unactual"}></div>
            )}
            <div className={"api__endpoint"}>{endpoint.Api}</div>
            {changeMod ? (
                 <input className={"input__api__cooldown"} value={cooldown} placeholder={"cooldown in seconds..."}
                    type={"number"} min={1} max={999999} onChange={(e) => setCooldown(() => e.target.value)}
                />
            ) : (
               <div className={`api__cooldown ${cooldown !== oldCooldown && "param__status_unsaved"}`}>{formatDuration(cooldown)}</div>
            )}
            <div className={"api__lastUpdate"}>{formatTimestamp(Math.floor(Date.parse(endpoint.LastUpdate) / 1000))}</div>
            <div className={"api__change"} onClick={() => setChangeMod(!changeMod)}>change</div>
            {changeMod && (
                <div onClick={() => handleSaveChanges(endpoint.Id, +cooldown, setCooldown, +oldCooldown, setOldCooldown, setChangeMod)}>save</div>
            )}
        </>
    );
};

export default EndpointItem;