import React from 'react';
import AdaptiveLoading from "../AdaptiveLoading";

const DefaultMetric = ({header, value}) => {
    return (
        <div className={"field__metric__default"}>
            <div className={"metric__header__default"}>{header}</div>
            {value != null ? (
                <div className={"metric__value__default"}>{value}</div>
            ) : (
                <AdaptiveLoading/>
            )}
        </div>
    );
};

export default DefaultMetric;