import React from 'react';
import AdaptiveLoading from "../AdaptiveLoading";

const Speedometer = ({header, percentage}) => {
    const circumference = Math.PI * 45;
    const offset = circumference - (percentage / 100) * circumference;

    const getColor = () => {
        if (percentage <= 55) return '#39c559';
        if (percentage <= 80) return '#f9cd1d';
        return '#c53939';
    };

    return (
        <div className="field__metric__default metric__2x2">
            <div className="metric__header__default">{header}</div>
            {percentage ? (
            <>
            <div className="speedometer__value" style={{ color: getColor(percentage) }}>{percentage}%</div>
            <svg className="speedometer" viewBox="0 0 120 60">
                <circle className={"speedometer__circle"} cx="60" cy="60" r={45} stroke="#ffffff21"
                        strokeDasharray={circumference} strokeDashoffset={0} transform="rotate(-180 60 60)"
                />
                <circle className={"speedometer__circle"} cx="60" cy="60" r={45} stroke={getColor()}
                        strokeDasharray={circumference} strokeDashoffset={offset} transform="rotate(-180 60 60)"
                />
            </svg>
            </>
            ) : (
                <AdaptiveLoading/>
            )}
        </div>
    );
};
export default Speedometer;