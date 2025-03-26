export const defaultKeyboard = [
    {
        label: 'Basic',
        rows: [
            [
                '[7]',
                '[8]',
                '[9]',
                '\\div',
                '[separator-5]',
                {class: 'small', latex: '\\frac{#@}{#0}'},
                {class: 'small', latex: '\\begin{pmatrix}#0\\\\#0\\end{pmatrix}'},
                '\\left\\lvert #0 \\right\\rvert',
                '\\sqrt{#@}',
            ],
            [
                '[4]',
                '[5]',
                '[6]',
                '\\times',
                '[separator-5]',
                '\\lt',
                '\\le',
                '#@^{#?}',
                '#@^2',
            ],
            [
                '1',
                '[2]',
                '[3]',
                '-',
                '[separator-5]',
                '\\gt',
                '\\ge',
                {label: '[backspace]', width: 2, shift: null},
            ],
            [
                '0',
                '.',
                '=',
                '+',
                '[separator-5]',
                '(',
                ')',
                {label: '[left]', shift: null},
                {label: '[right]', shift: null},
            ],
        ],
    },
    {
        label: '/v3/ticker/24hr',
        rows: [
            ["nodata"]
        ],
    },
    {
        label: '/v3/ticker/price',
        rows: [
            ["nodata"]
        ],
    },
]

export const defaultKeyboardV2 = [
    {
        label: "Basic",
        type: "basic",
        rows: [
            [
                '7',
                '8',
                '9',
                {label: '÷', latex: '\\div'},
                {class: 'sep-1', label: ''},
                {label: '', latex: '\\frac{#@}{#0}', id: 'frac'},
                {label: '', latex: '\\begin{pmatrix}#0\\\\#0\\end{pmatrix}', id: 'matrix'},
                {label: '', latex: '\\left\\lvert #0 \\right\\rvert', id: "mo"},
                {label: '', latex: '\\sqrt{#@}', id: "sq"},
            ],
            [
                '4',
                '5',
                '6',
                {label: '×', latex: '\\times'},
                {class: 'sep-1', label: ''},
                {label: '<', latex: '\\lt'},
                {label: '≤', latex: '\\le'},
                {label: 'x^y', latex: '#@^{#?}', id: "square2"},
                {label: '', latex: '#@^2', id: "square"},
            ],
            [
                '1',
                '2',
                '3',
                '−',
                {class: 'sep-1', label: ''},
                {label: '>', latex: '\\gt'},
                {label: '≥', latex: '\\ge'},
                {label: '', class: 'backspace', latex: 'backspace', id: "backspace"},
            ],
            [
                '0',
                '.',
                '=',
                '+',
                {class: 'sep-1', label: ''},
                '(',
                ')',
                {class: 'swap-left', label: '←', latex: 'left'},
                {class: 'swap-right', label: '→', latex: 'right'},
            ],
        ]
    },
    {
        label: "ticker 24hr",
        type: "flex",
        rows:
            [
                [
                    "priceChange",
                    "priceChangePercent",
                    "weightedAvgPrice",
                    "prevClosePrice",
                ],
            ],
    }
]