import React from 'react';
import {Area, AreaChart, CartesianGrid, ReferenceArea, ResponsiveContainer, Tooltip, XAxis, YAxis} from "recharts";

const ChartAvailability = ({ data }) => {
    let chartData = data.map((entry) => ({
        timestamp: new Date(entry.timestamp).toISOString(),
        webserver: entry.type === 1 ? entry.isAvailable : null,
        binance: entry.type === 2 ? entry.isAvailable : null,
    }));

    const firstTimestamp = chartData[0]?.timestamp;
    if (firstTimestamp) {
        chartData.unshift({
            timestamp: firstTimestamp,
            webserver: 1,
            binance: null,
        });
    }

    let lastWebserverValue = null;
    let lastBinanceValue = null;
    let lastWebserverTimestamp = null;

    for (const item of chartData) {
        if (item.webserver != null) {
            lastWebserverValue = item.webserver;
            lastWebserverTimestamp = item.timestamp;
        }
        if (item.binance != null) {
            lastBinanceValue = item.binance;
        }
    }

    if (lastWebserverTimestamp != null) {
        chartData.push({
            timestamp: lastWebserverTimestamp,
            webserver: lastWebserverValue,
            binance: lastBinanceValue,
        });
    }

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
        <div className={"field__metric__default metric__chart__full"}>
            <div className={"metric__header__default"}>Webserver and Binance availability</div>
            <ResponsiveContainer width="100%" height={180}>
                <AreaChart data={reducedData} margin={{ top: 10, right: 30, left: -30, bottom: 0 }}>
                    <CartesianGrid strokeDasharray="3 3" />
                    <XAxis dataKey="timestamp" tickFormatter={(tick) => new Date(tick).toLocaleTimeString()}/>
                    <YAxis tick={false} yAxisId="web" domain={[1, 2]} ticks={[1, 2]} />
                    <YAxis tick={false} yAxisId="bin" domain={[-1, 0]} ticks={[-1, 0]} hide />
                    <Tooltip />

                    {binanceIntervals.map(({ start, end, available }, index) => (
                        <ReferenceArea
                            key={index}
                            x1={start}
                            x2={end}
                            y1={-1}
                            y2={0}
                            fill={available ? "rgba(0,255,0,0.1)" : "rgba(255,0,0,0.1)"}
                            stroke="none"
                        />
                    ))}

                    <Area type="stepAfter" yAxisId="web" dataKey="webserver" stroke="#007bff" fill="#007bff55" name="Webserver" connectNulls/>
                    <Area type="stepAfter" yAxisId="bin" dataKey="binance" stroke="#28a745" fill="#28a74555" name="Binance" connectNulls/>
                </AreaChart>
            </ResponsiveContainer>
        </div>
    );
};

export default ChartAvailability;