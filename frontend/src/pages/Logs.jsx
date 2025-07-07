import React, {useEffect, useState} from 'react';
import {useFetching} from "../hooks/useFetching";
import SettingsService from "../API/SettingsService";
import AdaptiveLoading from "../components/UI/AdaptiveLoading";
import ErrorField from "../components/UI/ErrorField";
import ChartAvailability from "../components/UI/ChartAvailability";

const Logs = () => {
    const [logs, setLogs] = useState(null);
    const [fetchLogs, isLogsLoading, LogsError] = useFetching(async () => {
        return await SettingsService.getLogs()
    }, 1000, 1000)


    useEffect(() => {
        const loadData = async () => {
            if (!isLogsLoading && logs === null){
                const data = await fetchLogs();
                if (data) {
                    setLogs(data.data);
                    console.log(data.data)
                }
            }
        }
        void loadData();
    }, [isLogsLoading])

    return (
        <div className={"section__main"}>
            {isLogsLoading ? (
                <div className={"loading__center"}>
                    <AdaptiveLoading/>
                </div>
            ) : LogsError ? (
                <ErrorField/>
            ) : logs?.length > 0 ? (
                <div className={"field__settings__api"}>
                   <ChartAvailability data={logs}/>
                </div>
            ) : (isLogsLoading === false && logs.length === 0) && (
                <ErrorField message={"logs are empty, this is non-standard situation"}/>
            )}
        </div>
    );
};

export default Logs;