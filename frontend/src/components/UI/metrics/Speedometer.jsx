import React from 'react';

const Speedometer = ({header, percentage}) => {
    const radius = 45;
    const circumference = Math.PI * radius;
    const offset = circumference - (percentage / 100) * circumference;

    const getColor = () => {
        if (percentage <= 55) return '#39c559';
        if (percentage <= 80) return '#f9cd1d';
        return '#c53939';
    };

    return (
        <div className="field__metric__default metric__2x2">
            <div className="metric__header__default">{header}</div>
            <div className="speedometer__value" style={{ color: getColor(percentage) }}>{percentage}%</div>

            <svg className="speedometer" viewBox="0 0 120 60">
                <circle className={"speedometer__circle"} cx="60" cy="60" r={radius} stroke="#ffffff21"
                        strokeDasharray={circumference} strokeDashoffset={0} transform="rotate(-180 60 60)"
                />
                <circle className={"speedometer__circle"} cx="60" cy="60" r={radius} stroke={getColor()}
                        strokeDasharray={circumference} strokeDashoffset={offset} transform="rotate(-180 60 60)"
                />
            </svg>
        </div>
    );
};
export default Speedometer;