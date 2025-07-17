import React, {useEffect, useState} from 'react';
import "../styles/errors.css"
import {useFetching} from "../hooks/useFetching";
import MetricsService from "../API/MetricsService";
import AdaptiveLoading from "../components/UI/AdaptiveLoading";
import ErrorField from "../components/UI/ErrorField";
import {formatTimestamp} from "../utils/utils";

const Errors = () => {
    const [logsBasicInfo, setLogsBasicInfo] = useState([]);
    const [binanceLogs, setBinanceLogs] = useState([]);
    const [applicationLogs, setApplicationLogs] = useState([]);
    const [analyticsServiceLogs, setAnalyticsServiceLogs] = useState([]);
    const [openedLogType, setOpenedLogType] = useState(null);
    const [fetchDetailedLogs, , ] = useFetching(async (logsType) => {
        return await MetricsService.getDetailedLogs(logsType)
    }, 1000, 1000)
    const [fetchBasicLogsInfo, isBasicLogsInfo, basicLogsInfoError] = useFetching(async () => {
        return await MetricsService.getBasicLogs()
    }, 1000, 1000)

    const downloadErrors = (logsType) => {
        const jsonStr = JSON.stringify(getErrorsForType(logsType), null, 2);
        const blob = new Blob([jsonStr], {type: 'application/json'});
        const url = URL.createObjectURL(blob);

        const a = document.createElement('a');
        a.href = url;
        a.download = `${logsType}Logs.json`;
        a.click();

        URL.revokeObjectURL(url);
    };

    useEffect(() => {
        const loadData = async () => {
            if (!isBasicLogsInfo && !basicLogsInfoError) {
                const response = await fetchBasicLogsInfo()
                response && setLogsBasicInfo(response.data)
            }
        }
        void loadData()
    }, [isBasicLogsInfo])

    const handleToggleLogs = async (logType) => {
        setOpenedLogType(prev => (prev === logType ? null : logType));

        if (logType === 'binance' && binanceLogs.length === 0) {
            const res = await fetchDetailedLogs('binance');
            res && setBinanceLogs(res.data);
        } else if (logType === 'application' && applicationLogs.length === 0) {
            const res = await fetchDetailedLogs('application');
            res && setApplicationLogs(res.data);
        } else if (logType === 'analytics' && analyticsServiceLogs.length === 0) {
            const res = await fetchDetailedLogs('analytics');
            res && setAnalyticsServiceLogs(res.data);
        }
    };

    const getErrorsForType = (logType) => {
        switch (logType) {
            case 'binance':
                return binanceLogs;
            case 'application':
                return applicationLogs;
            case 'analytics':
                return analyticsServiceLogs;
            default:
                return [];
        }
    };

    return (
        <div className={"section__main"}>
            {isBasicLogsInfo ? (
                <div className={"loading__center"}>
                    <AdaptiveLoading/>
                </div>
            ) : basicLogsInfoError ? (
                <ErrorField/>
            ) : (
                <div className={"field__errors"}>
                    <div className={"errors__list"}>
                        {logsBasicInfo.map((lbi, index) => (
                        <div className={"errors__item"} key={lbi.type}>
                            <input id={`errors_reveal-${index}`} className={"checkbox__errors_reveal"} type={"checkbox"}
                                   onChange={() => handleToggleLogs(lbi.type)} checked={openedLogType === lbi.type} />
                            <div className={"errors__item__header"}>
                                {lbi.lines > 0 && (
                                <svg className={"svg_errors__download"} onClick={() => downloadErrors(lbi.type)}>
                                    <use xlinkHref={"#icon_download_file"}></use>
                                </svg>
                                )}
                                <div className={"errors__filename"}>{lbi.type} Errors</div>
                                {lbi.lines > 0 ? (
                                    <label htmlFor={`errors_reveal-${index}`} className={"label__errors_reveal"}>
                                        <svg className={"svg_errors__reveal"}>
                                            <use xlinkHref={"#icon_fullcontainer"}></use>
                                        </svg>
                                        <svg className={"svg_errors__reveal_exit"}>
                                            <use xlinkHref={"#icon_fullcontainer_exit"}></use>
                                        </svg>
                                    </label>
                                ) : (
                                    <div className={"label__errors_reveal"}>logs are empty</div>
                                )}
                            </div>
                            <div className={"errors__details"}>
                                <div className={"errors__details__header"}>
                                    <div className={"details__header__level"}>Level</div>
                                    <div className={"details__header__date"}>Date</div>
                                    <div className={"details__header__info"}>
                                        <div className={"details__header__message"}>Message</div>
                                        <div className={"details__header__error"}>Error</div>
                                    </div>
                                </div>
                                {getErrorsForType(lbi.type).map((applicationError, index2) => (
                                <React.Fragment key={index2}>
                                <input id={`details__info_${index2}`} className={"checkbox_details__info"} type={"checkbox"}/>
                                <div className={"errors__details__core"}>
                                    <div className={"details__level"}>{applicationError.level}</div>
                                    <div className={"details__date"}>{formatTimestamp(applicationError.timestamp)}</div>
                                    <label className={"details__info"} htmlFor={`details__info_${index2}`}>
                                        <div className={"details__message"}>{applicationError.event}</div>
                                        <div className={"details__error"}>{applicationError.event}</div>
                                    </label>
                                </div>
                                </React.Fragment>
                                ))}
                            </div>
                        </div>
                        ))}
                    </div>
                </div>
            )}
        </div>
    );
};

export default Errors;