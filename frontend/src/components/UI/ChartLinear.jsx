import {formatNumber, formatTimestamp} from "../../utils/utils";
import AdaptiveLoading from "./AdaptiveLoading";
import {XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, LineChart, Line, Brush} from 'recharts';

export const ChartLinear = ({data}) => {
    if (!data || data.length === 0) {
        return <AdaptiveLoading/>;
    }

    const allKeys = Array.from(new Set(
        data.flatMap(item => Object.keys(item).filter(key => key !== 'timestamp'))
    ));

    const getColor = index => {
        const hue = (index * 360) / allKeys.length;
        return `hsl(${hue}, 70%, 50%)`;
    };

    const colorMap = {};
    allKeys.forEach((key, index) => {
        colorMap[key] = getColor(index);
    });

    const CustomTooltip = ({active, payload, label, colors}) => {
        if (!active || !payload || payload.length === 0) return null;

        return (
            <div style={{
                backgroundColor: '#333',
                color: '#fff',
                padding: 10,
                borderRadius: 5
            }}>
                <p style={{marginBottom: 5}}>{formatTimestamp(label)}</p>
                {payload.map((entry) => (
                    <p
                        key={entry.name}
                        style={{
                            color: colors[entry.name] || '#fff',
                            backgroundColor: '#222',
                            padding: '2px 4px',
                            margin: 0,
                            borderRadius: 3
                        }}
                    >
                        {entry.name}: {formatNumber(entry.value)}
                    </p>
                ))}
            </div>
        );
    };

    return (
        <ResponsiveContainer width="100%" height={300}>
            <LineChart data={data} margin={{top: 20, right: 20, left: 0, bottom: 20}}>
                <CartesianGrid stroke="#444" strokeDasharray="3 3" vertical={false}/>
                <XAxis dataKey="timestamp" tick={{fill: '#aaa'}} tickFormatter={formatTimestamp}/>
                <YAxis tick={{fill: '#aaa'}}/>
                <Tooltip
                    content={<CustomTooltip colors={colorMap} />}
                    labelFormatter={formatTimestamp}
                    formatter={(value, key) => [formatNumber(value), key]}
                />

                {allKeys.map((key, index) => (
                    <Line
                        key={key}
                        type="monotone"
                        dataKey={key}
                        dot={false}
                        stroke={getColor(index)}
                        strokeWidth={2}
                        isAnimationActive={true}
                        activeDot={{r: 6}}
                    />
                ))}

                <Brush
                    dataKey="timestamp"
                    height={28}
                    stroke="#8884d8"
                    fill="#ffffff00"
                    travellerWidth={6}
                    tickFormatter={formatTimestamp}
                />
            </LineChart>
        </ResponsiveContainer>
    );
};

export default ChartLinear;