export const defaultKeyboard = [
    {
        label: "Basic",
        type: "basic",
        rows: [
            [
                "7", "8", "9", "÷",
                { class: "sep-1", label: "" },
                { label: "", token: "frac", id: "frac"},
                { label: "", token: "matrix", id: "matrix"},
                { label: "", token: "abs", id: "mo"},
                { label: "", token: "sqrt", id: "sq"},
            ],
            [
                "4", "5", "6", "*",
                { class: "sep-1", label: ""},
                { label: "<", token: "<"},
                { label: "≤", token: "<="},
                { label: "", token: "^", id: "square2"},
                { label: "", token: "^2", id: "square"},
            ],
            [
                "1", "2", "3", "-",
                { class: "sep-1", label: ""},
                { label: ">", token: ">"},
                { label: "≥", token: ">="},
                { class: "backspace", label: "⌫", token: "backspace", id: "backspace"},
            ],
            [
                "0", ".", "=", "+",
                { class: "sep-1", label: "" },
                { label: "(", token: "brackets-l"},
                { label: ")", token: "brackets-r"},
                { class: "swap-left", label: "←", token: "left" },
                { class: "swap-right", label: "→", token: "right" },
            ],
        ]
    }
];

export function urlBase64ToUint8Array(base64String) {
    const padding = '='.repeat((4 - base64String.length % 4) % 4);
    const base64 = (base64String + padding).replace(/-/g, '+').replace(/_/g, '/');
    const rawData = atob(base64);
    return Uint8Array.from([...rawData].map(char => char.charCodeAt(0)));
}

export const formatNumber = (num) => { // маленьким учет не добавлять
    if (num >= 1e9) {
        return (num / 1e9).toFixed(1) + 'B';
    } else if (num >= 1e6) {
        return (num / 1e6).toFixed(1) + 'M';
    }
    return num.toString();
}

export const formatTimestamp = (timestamp) => {
    const date = typeof timestamp === 'string'
        ? new Date(timestamp)
        : new Date(timestamp * 1000);
    return date.toLocaleDateString([],
        {hour:'2-digit',  minute:'2-digit',  day:'2-digit',  month:'2-digit',  year:'numeric'});
}

export const formatUptime = (timestamp) => {
    const timestampMs = Number(timestamp) * 1000;
    const now = Date.now();

    const diffMs = now - timestampMs;
    const diffSec = diffMs / 1000;

    if (diffSec < 60) return `${diffSec.toFixed(1)}s`;
    if (diffSec < 3600) return `${(diffSec / 60).toFixed(1)}min`;
    if (diffSec < 86400) return `${(diffSec / 3600).toFixed(1)}hour`;
    return `${(diffSec / 86400).toFixed(2)}day`;
};

export const transformData = (data) => {
    return data.map(item => {
        const names = Object.keys(item).filter(key => key !== 'timestamp');

        const values = item[names[0]].split(', ').map(value => parseFloat(value.trim()));

        let result = { timestamp: item.timestamp };

        names[0].split(', ').forEach((name, index) => {
            result[name.trim()] = values[index];
        });
        return result;
    });
}

export const formatDuration = (stringSeconds) => {
    const days = Math.floor(+stringSeconds / 86400);
    const hours = Math.floor((+stringSeconds % 86400) / 3600);
    const minutes = Math.floor((+stringSeconds % 3600) / 60);
    const seconds = Math.floor(+stringSeconds % 60);

    const parts = [];
    if (days) parts.push(`${days} day${days !== 1 ? 's' : ''}`);
    if (hours) parts.push(`${hours} hour${hours !== 1 ? 's' : ''}`);
    if (minutes) parts.push(`${minutes} minute${minutes !== 1 ? 's' : ''}`);
    if (seconds || parts.length === 0) parts.push(`${seconds} second${seconds !== 1 ? 's' : ''}`);

    return parts.join(', ');
};

export const selectKlinesInterval = {"1m": "1m", "3m": "3m", "5m": "5m", "15m": "15m", "30m": "30m", "1h": "1h",
    "2h": "2h", "4h": "4h", "6h": "6h", "8h": "8h", "12h": "12h", "1d": "1d", "3d": "3d", "1w": "1w"};

export const calculateMA = (dayCount, data) => {
    let result = [];
    for (let i = 0; i < data.values.length; i++) {
        if (i < dayCount) {
            result.push('-');
            continue;
        }
        let sum = 0;
        for (let j = 0; j < dayCount; j++) {
            sum += data.values[i - j][1];
        }
        result.push(+(sum / dayCount).toFixed(3));
    }
    return result;
}
