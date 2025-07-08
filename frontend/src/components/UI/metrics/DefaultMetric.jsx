import React from 'react';

const DefaultMetric = ({header, value}) => {
    return (
        <div className={"field__metric__default"}>
            <div className={"metric__header__default"}>{header}</div>
            <div className={"metric__value__default"}>{value}</div>
        </div>
    );
};

export default DefaultMetric;