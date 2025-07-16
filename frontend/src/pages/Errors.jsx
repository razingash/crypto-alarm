import React, {useEffect, useState} from 'react';
import "../styles/errors.css"
import {useFetching} from "../hooks/useFetching";
import MetricsService from "../API/MetricsService";
import AdaptiveLoading from "../components/UI/AdaptiveLoading";
import ErrorField from "../components/UI/ErrorField";
import {formatTimestamp} from "../utils/utils";

const Errors = () => {
    const [errors, setErrors] = useState([]);
    const [fetchErrors, isErrorsLoading, errorsError] = useFetching(async () => {
        return await MetricsService.getCriticalErrorLogs()
    }, 1000, 1000)

    const downloadErrors = () => {
        const jsonStr = JSON.stringify(errors, null, 2);
        const blob = new Blob([jsonStr], {type: 'application/json'});
        const url = URL.createObjectURL(blob);

        const a = document.createElement('a');
        a.href = url;
        a.download = 'errors.json';
        a.click();

        URL.revokeObjectURL(url);
    };

    useEffect(() => {
        const loadData = async () => {
            if (!isErrorsLoading && !errorsError) {
                const response = await fetchErrors()
                response && setErrors(response.data)
            }
        }
        void loadData()
    }, [isErrorsLoading])

    return (
        <div className={"section__main"}>
            {isErrorsLoading ? (
                <div className={"loading__center"}>
                    <AdaptiveLoading/>
                </div>
            ) : errorsError ? (
                <ErrorField/>
            ) : errors.length > 0 ? (
                <div className={"field__errors"}>
                    <div className={"errors__list"}>
                        <div className={"errors__item"}>
                            <input id={"errors_reveal"} type={"checkbox"}/>
                            <div className={"errors__item__header"}>
                                <svg className={"svg_errors__download"} onClick={downloadErrors}>
                                    <use xlinkHref={"#icon_download_file"}></use>
                                </svg>
                                <div className={"errors__filename"}>Critical Errors</div>
                                <label htmlFor={"errors_reveal"} className={"label__errors_reveal"}>
                                    <svg className={"svg_errors__reveal"}>
                                        <use xlinkHref={"#icon_fullcontainer"}></use>
                                    </svg>
                                    <svg className={"svg_errors__reveal_exit"}>
                                        <use xlinkHref={"#icon_fullcontainer_exit"}></use>
                                    </svg>
                                </label>
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
                                {errors.map((applicationError, index) => (
                                <React.Fragment key={index}>
                                <input id={`details__info_${index}`} className={"checkbox_details__info"} type={"checkbox"}/>
                                <div className={"errors__details__core"}>
                                    <div className={"details__level"}>{applicationError.level}</div>
                                    <div className={"details__date"}>{formatTimestamp(applicationError.timestamp)}</div>
                                    <label className={"details__info"} htmlFor={`details__info_${index}`}>
                                        <div className={"details__message"}>{applicationError.event}</div>
                                        <div className={"details__error"}>{applicationError.event}</div>
                                    </label>
                                </div>
                                </React.Fragment>
                                ))}
                            </div>
                        </div>
                    </div>
                </div>
            ) : (isErrorsLoading === false && !errorsError) && (
                <ErrorField message={"Application hasn't recorded any unexpected errors yet. If a problem occurs and it isn't visible here, then the error isn't processed by the application. You should notify developer(me) as soon as possible"}/>
            )}
        </div>
    );
};

export default Errors;