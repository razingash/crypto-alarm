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
