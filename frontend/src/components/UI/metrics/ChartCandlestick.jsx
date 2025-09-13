import React, {useEffect, useMemo, useState} from 'react';
import ReactECharts from "echarts-for-react";
import {useFetching} from "../../../hooks/useFetching";
import StrategyService from "../../../API/modules/StrategyService";
import {calculateMA, selectKlinesInterval} from "../../../utils/utils";

const ChartCandlestick = () => {
    const [rawData, setRawData] = useState([]);
    const data = splitData(rawData);
    const [searchKey, setSearchKey] = useState("");
    const [searchResults, setSearchResults] = useState("");
    const [selectedPair, setSelectedPair] = useState(null);
    const [selectedInterval, setSelectedInterval] = useState('1m')
    const [availableCurrencies, setAvailableCurrencies] = useState([]);
    const [fetchCurrenciesPairs, isCurrenciesPairsLoading, currenciesPairsError] = useFetching(async () => {
        return await StrategyService.getKeyboard()
    }, 0, 1000)
    const [fetchKlines, isKlinesLoading, klinesError] = useFetching(async (selectedPair, selectedInterval) => {
        return await StrategyService.getBinanceKlines(selectedPair, selectedInterval)
    }, 0, 1000)

    useEffect(() => {
        const loadData = async () => {
            if (!isKlinesLoading && !klinesError && selectedPair != null && selectedInterval != null) {
                const data = await fetchKlines(selectedPair, selectedInterval);
                if (data) {
                    setRawData(data.data)
                }
            }
        }
        void loadData();
    }, [selectedPair, selectedInterval])

    useEffect(() => {
        const loadData = async () => {
            if (!isCurrenciesPairsLoading && availableCurrencies.length === 0 && !currenciesPairsError) {
                const data = await fetchCurrenciesPairs(selectedInterval);
                data?.currencies && setAvailableCurrencies(data.currencies)
            }
        }
        void loadData();
    }, [isCurrenciesPairsLoading])

    useEffect(() => {
        setSearchResults(searchKey);
    }, [searchKey]);

    const filteredFields = useMemo(() => {
        if (!searchResults) return availableCurrencies.slice(0, 5);
        return availableCurrencies
            .filter(item =>
                item.toLowerCase().includes(searchResults.toLowerCase())
            )
            .slice(0, 10);
    }, [searchResults, availableCurrencies]);

    function splitData(rawData) {
        let categoryData = [];
        let values = [];
        let volumes = [];
        for (let i = 0; i < rawData.length; i++) {
            const date = new Date(rawData[i][0]);
            const localDate = date.toLocaleString('ru-RU', {
                month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit'
            }).replace(',', '');
            categoryData.push(localDate);

            values.push([
                parseFloat(rawData[i][1]), parseFloat(rawData[i][4]),
                parseFloat(rawData[i][3]), parseFloat(rawData[i][2])
            ]);
            volumes.push([
                i, parseFloat(rawData[i][5]), rawData[i][1] > rawData[i][4] ? 1 : -1
            ]);
        }
        return {categoryData, values, volumes};
    }

    const option = {
        animation: false,
        legend: {
            bottom: 0,
            left: 'center',
            textStyle: {
                color: '#727272',
            },
            inactiveColor: '#555',
            selectedMode: true,
            data: ['Price', 'MA5', 'MA10', 'MA20', 'MA30']
        },
        tooltip: {
            trigger: 'axis',
            axisPointer: {
                type: 'cross'
            },
            borderWidth: 1,
            borderColor: '#fff', // ничего не делает
            padding: 10,
            textStyle: {color: '#000'},
            position: function (pos, params, el, elRect, size) {
                const obj = {top: 10};
                obj[['left', 'right'][+(pos[0] < size.viewSize[0] / 2)]] = 30;
                return obj;
            }
        },
        axisPointer: {
            link: [{xAxisIndex: 'all'}],
            label: {backgroundColor: '#777'} // a62a2a
        },
        toolbox: {
            feature: {
                dataZoom: {yAxisIndex: false},
                brush: {type: ['lineX', 'clear']}
            }
        },
        brush: {
            xAxisIndex: 'all',
            brushLink: 'all',
            outOfBrush: {colorAlpha: 0.1}
        },
        visualMap: {
            show: false,
            seriesIndex: 5,
            dimension: 2,
            pieces: [
                {value: 1, color: '#ec0000'},
                {value: -1, color: '#00da3c'}
            ]
        },
        grid: [
            {left: '10%', right: '8%', height: '50%'},
            {left: '10%', right: '8%', top: '63%', height: '16%'}
        ],
        xAxis: [
            {
                type: 'category',
                data: data.categoryData,
                boundaryGap: false,
                axisLine: {onZero: false},
                splitLine: {show: false},
                min: 'dataMin',
                max: 'dataMax',
                axisPointer: {z: 100}
            },
            {
                type: 'category',
                gridIndex: 1,
                data: data.categoryData,
                boundaryGap: false,
                axisLine: {onZero: false},
                axisTick: {show: false},
                splitLine: {show: false},
                axisLabel: {show: false},
                min: 'dataMin',
                max: 'dataMax'
            }
        ],
        yAxis: [
            {
                scale: true,
                splitArea: {
                    show: true,
                    areaStyle: {
                        color: ['rgba(52,52,52,0.5)', 'rgba(37,37,37,0.5)']
                    }
                },
                splitLine: {
                    show: true,
                    lineStyle: {
                        color: '#727272',
                        width: 1,
                        type: 'solid'
                    }
                }
            },
            {
                scale: true,
                gridIndex: 1,
                splitNumber: 2,
                axisLabel: {show: false},
                axisLine: {show: false},
                axisTick: {show: false},
                splitLine: {show: false},
            }
        ],
        dataZoom: [
            {
                type: 'inside',
                xAxisIndex: [0, 1],
                start: 70,
                end: 100
            },
            {
                show: true,
                xAxisIndex: [0, 1],
                type: 'slider',
                top: '85%',
                start: 70,
                end: 100,
                textStyle: {color: '#ffffff'},
            }
        ],
        series: [
            {
                name: 'Price',
                type: 'candlestick',
                data: data.values,
                itemStyle: {
                    color: '#00da3c',
                    color0: '#ec0000',
                    borderColor: undefined,
                    borderColor0: undefined
                }
            },
            {
                name: 'MA5',
                type: 'line',
                showSymbol: false,
                data: calculateMA(5, data),
                smooth: true,
                lineStyle: {
                    opacity: 0.6,
                    color: '#0095ff'
                },
            },
            {
                name: 'MA10',
                type: 'line',
                showSymbol: false,
                data: calculateMA(10, data),
                smooth: true,
                lineStyle: {opacity: 0.5}
            },
            {
                name: 'MA20',
                type: 'line',
                showSymbol: false,
                data: calculateMA(20, data),
                smooth: true,
                lineStyle: {opacity: 0.5}
            },
            {
                name: 'MA30',
                type: 'line',
                showSymbol: false,
                data: calculateMA(30, data),
                smooth: true,
                lineStyle: {opacity: 0.5}
            },
            {
                name: 'Volume',
                type: 'bar',
                xAxisIndex: 1,
                yAxisIndex: 1,
                data: data.volumes
            }
        ]
    };

    return (
        <div className={"field__candlestick"}>
            <input type={"checkbox"} id={"ckeckbox__candlestick__search"}/>
            <div className={"candlestick__search__field"}>
                <input id={"candlestick__search_field"} onChange={e => setSearchKey(e.target.value)}
                       value={searchKey} className={"input__default"} placeholder={"Search"}/>
                <svg className={"svg__clear_input"} onClick={() => setSearchKey('')}>
                    <use xlinkHref={"#icon_cross"}></use>
                </svg>
                <div className={"candlestick__search__results"}>
                    {filteredFields.map(item => (
                        <div key={item} className={"candlestick__search__result"} onClick={() => setSelectedPair(item)}>{item}</div>
                    ))}
                </div>
                <label htmlFor={"ckeckbox__candlestick__search"} className={"candlestick__search__close"}>
                    <svg className={"svg__switch_csf"}>
                        <use xlinkHref={"#icon_switch_on"}></use>
                    </svg>
                    <svg className={"svg__switch_csf"}>
                        <use xlinkHref={"#icon_switch_off"}></use>
                    </svg>
                </label>
            </div>
            <div className={"area__fullspace"}>
                <div className={"candlestick__header"}>
                    <div className={"candlestick__header__item"}>
                        <div>Interval:</div>
                        <select className={"select__default"} onChange={(e) => setSelectedInterval(e.target.value)}>
                        {Object.entries(selectKlinesInterval).map(([interval, type]) => (
                            <option className={"option__default"} key={interval} value={type}>{interval}</option>
                        ))}
                        </select>
                    </div>
                    <label htmlFor={"ckeckbox__candlestick__search"} className={"candlestick__header__item"}>Currency: {selectedPair || "unselected"}</label>
                </div>
                {rawData.length > 0 && (
                    <ReactECharts option={option} style={{ height: "450px", width: "100%" }} />
                )}
            </div>
        </div>
    );
};

export default ChartCandlestick;