import React from 'react';
import {CartesianGrid, Legend, Line, LineChart, ResponsiveContainer, Tooltip, XAxis, YAxis} from "recharts";
import AdaptiveLoading from "../AdaptiveLoading";
import {formatNumber} from "../../../utils/utils";

const ChartSystemLoad = ({data}) => {
    const CustomTooltip = ({active, payload, label}) => {
        if (!active || !payload || payload.length === 0) return null;

        return (
            <div className={"chart__tooltip"}>
                <p style={{marginBottom: 5}}>{label}</p>
                {payload.map((entry) => (
                    <p className={"tooltip__item"} key={entry.name} style={{color: entry.color}}>
                        {entry.name}: {formatNumber(entry.value)}
                    </p>
                ))}
            </div>
        );
    };

    return (
        <div className={"field__metric__default metric__chart__full"}>
            <div className={"metric__header__default"}>System load</div>
            {data ? (
            <ResponsiveContainer width="100%" height={180}>
                <LineChart data={data} margin={{top: 20, right: 30, left: 20, bottom: 5}}>
                    <CartesianGrid stroke="#333" strokeDasharray="3 3"/>
                    <XAxis dataKey="time"/>
                    <YAxis/>
                    <Tooltip
                        content={<CustomTooltip/>}
                        formatter={(value, key) => [formatNumber(value), key]}
                    />
                    <Legend/>
                    <Line type="monotone" dataKey="1m" stroke="#00ff00" dot={false}/>
                    <Line type="monotone" dataKey="5m" stroke="#ffcc00" dot={false}/>
                    <Line type="monotone" dataKey="15m" stroke="#66ccff" dot={false}/>
                </LineChart>
            </ResponsiveContainer>
            ) : (
              <AdaptiveLoading/>
            )}
        </div>
    );
};

export default ChartSystemLoad;