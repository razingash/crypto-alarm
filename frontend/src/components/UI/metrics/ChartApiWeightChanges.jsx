import React, {useEffect, useState} from 'react';
import {useFetching} from "../../../hooks/useFetching";
import MetricsService from "../../../API/MetricsService";
import {CartesianGrid, Line, LineChart, ResponsiveContainer, Tooltip, XAxis, YAxis} from "recharts";
import {formatTimestamp} from "../../../utils/utils";

const ChartApiWeightChanges = () => {
    const [chartData, setChartData] = useState([]);
    const [endpoints, setEndpoints] = useState([]);
    const [colorMap, setColorMap] = useState({});
    const [fetchApiWeights, isApiWeightsLoading, ApiWeightsError] = useFetching(async () => {
        return await MetricsService.getBinanceApiWeight()
    }, 1000, 1000)

    useEffect(() => {
        if (!isApiWeightsLoading && !ApiWeightsError) {
            const loadData = async () => {
                const res = await fetchApiWeights()
                if (res) {
                    const tempMap = new Map();
                    const allEndpoints = [];

                    res.forEach(({ endpoint, weights }) => {
                        allEndpoints.push(endpoint);
                        weights.forEach(({ created_at, weight }) => {
                            const ts = new Date(created_at).toISOString();
                            if (!tempMap.has(ts)) tempMap.set(ts, { created_at: ts });
                            tempMap.get(ts)[endpoint] = weight;
                        });
                    });

                    const sorted = Array.from(tempMap.values()).sort(
                        (a, b) => new Date(a.created_at) - new Date(b.created_at)
                    );

                    const colors = {};
                    allEndpoints.forEach((ep, i) => {
                        const hue = (i * 360) / allEndpoints.length;
                        colors[ep] = `hsl(${hue}, 70%, 50%)`;
                    });

                    setChartData(sorted);
                    setEndpoints(allEndpoints);
                    setColorMap(colors);
                }
            }
            void loadData()
        }
    }, [isApiWeightsLoading]);

    const CustomTooltip = ({ active, payload, label }) => {
        if (!active || !payload?.length) return null;
        return (
            <div className={"chart__tooltip"}>
                <p style={{ marginBottom: 5 }}>{formatTimestamp(label)}</p>
                {payload.map(entry => (
                    <p key={entry.name} className={"tooltip__item"} style={{color: colorMap[entry.name]}}>
                        {entry.name}: {entry.value}
                    </p>
                ))}
            </div>
        );
    };

    return (
        <>
        {chartData.length > 0 && (
            <div className="field__metric__default field__chart__api_weight">
                <div className={"metric__header__default"}>Binance endpoints weight</div>
                <ResponsiveContainer width="100%" height={180}>
                    <LineChart data={chartData} margin={{ top: 20, right: 20, left: 0, bottom: 20 }}>
                        <CartesianGrid stroke="#444" strokeDasharray="3 3" vertical={false} />
                        <XAxis dataKey="created_at" tick={{ fill: '#aaa' }} tickFormatter={formatTimestamp} />
                        <YAxis tick={{ fill: '#aaa' }} />
                        <Tooltip content={<CustomTooltip />} formatter={(value, key) => [value, key]} />
                        {endpoints.map((endpoint, i) => (
                            <Line key={endpoint} type="monotone" dataKey={endpoint} dot={false} stroke={colorMap[endpoint]} strokeWidth={2} connectNulls activeDot={{ r: 6 }} />
                        ))}
                    </LineChart>
                </ResponsiveContainer>
            </div>
        )}
        </>
    );
};
export default ChartApiWeightChanges;