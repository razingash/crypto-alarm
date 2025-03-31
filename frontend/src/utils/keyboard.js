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
                "(", ")",
                { class: "swap-left", label: "←", token: "left" },
                { class: "swap-right", label: "→", token: "right" },
            ],
        ]
    }
];
