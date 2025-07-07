import React from 'react';
import {Area, AreaChart, CartesianGrid, ReferenceArea, ResponsiveContainer, Tooltip, XAxis, YAxis} from "recharts";

const ChartAvailability = ({ data }) => {
    const chartData = data.map((entry) => ({
        timestamp: new Date(entry.timestamp).toISOString(),
        webserver: entry.type === 1 ? entry.isAvailable : null,
        binance: entry.type === 2 ? entry.isAvailable : null,
    }));

    const reducedData = chartData.reduce((acc, item) => {
        const last = acc[acc.length - 1];
        if (last && last.timestamp === item.timestamp) {
            acc[acc.length - 1] = {
                timestamp: item.timestamp,
                webserver: item.webserver ?? last.webserver,
                binance: item.binance ?? last.binance,
            };
        } else {
            acc.push(item);
        }
        return acc;
    }, []);

    const binanceIntervals = [];
    let currentStart = null;
    let currentState = null;

    for (let i = 0; i < reducedData.length; i++) {
        const point = reducedData[i];
        if (point.binance !== null && point.binance !== currentState) {
            if (currentStart !== null) {
                binanceIntervals.push({
                    start: currentStart,
                    end: point.timestamp,
                    available: currentState === 1,
                });
            }
            currentStart = point.timestamp;
            currentState = point.binance;
        }
    }

    if (currentStart !== null && currentState !== null) {
        binanceIntervals.push({
            start: currentStart,
            end: reducedData[reducedData.length - 1]?.timestamp,
            available: currentState === 1,
        });
    }

    return (
        <ResponsiveContainer width="100%" height={400}>
            <AreaChart data={reducedData}>
                <CartesianGrid strokeDasharray="3 3" />
                <XAxis
                    dataKey="timestamp"
                    tickFormatter={(tick) => new Date(tick).toLocaleTimeString()}
                />
                <YAxis domain={[0, 1]} ticks={[0, 1]} />
                <Tooltip />

                {binanceIntervals.map(({ start, end, available }, index) => (
                    <ReferenceArea
                        key={index}
                        x1={start}
                        x2={end}
                        y1={0}
                        y2={1}
                        fill={available ? "rgba(0,255,0,0.1)" : "rgba(255,0,0,0.1)"}
                        stroke="none"
                    />
                ))}

                <Area
                    type="monotone"
                    dataKey="webserver"
                    stroke="#007bff"
                    fill="#007bff55"
                    name="Webserver"
                    connectNulls
                />
                <Area
                    type="monotone"
                    dataKey="binance"
                    stroke="#28a745"
                    fill="#28a74555"
                    name="Binance"
                    connectNulls
                />
            </AreaChart>
        </ResponsiveContainer>
    );
};

export default ChartAvailability;