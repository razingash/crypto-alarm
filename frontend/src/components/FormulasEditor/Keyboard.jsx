import React, {useEffect, useMemo, useRef, useState} from 'react';
import {defaultKeyboard} from "../../utils/utils";
import AdaptiveLoading from "../UI/AdaptiveLoading";
import {useFetching} from "../../hooks/useFetching";
import TriggersService from "../../API/TriggersService";

const Keyboard = ({onKeyPress}) => {
    const [availableApi, setAvailableApi] = useState([]);
    const [availableCurrencies, setAvailableCurrencies] = useState([]);
    const [selectedIndex, setSelectedIndex] = useState(0);
    const [searchKey, setSearchKey] = useState("");
    const [selectedCurrency, setSelectedCurrency] = useState(null);

    const [isSearch, setIsSearch] = useState(false);
    const [delayedSearchKey, setDelayedSearchKey] = useState(""); // отсроченный поиск
    const listRef = useRef(null);
    const [canScrollLeft, setCanScrollLeft] = useState(false);
    const [canScrollRight, setCanScrollRight] = useState(false);

    const [fetchKeyboard, isKeyboardLoading, ] = useFetching(async () => {
        return await TriggersService.getKeyboard()
    }, 0, 1000)

    const selectedKeyboard = defaultKeyboard[selectedIndex];

    useEffect(() => {
        const loadData = async () => {
            if (!isKeyboardLoading && availableApi.length === 0) {
                const data = await fetchKeyboard();
                if (data) {
                    data?.api && setAvailableApi(data.api)
                    data?.currencies && setAvailableCurrencies(data.currencies)
                }
            }
        }
        void loadData();
    }, [isKeyboardLoading])

    useEffect(() => {
        //поиск начинается спустя 500мс после того как пользователь закончит вводить инфу
        const timeout = setTimeout(() => {
            setDelayedSearchKey(searchKey);
            setIsSearch(true);
        }, 500);

        return () => clearTimeout(timeout);
    }, [searchKey]);

    const filteredFields = useMemo(() => {
        if (!delayedSearchKey) return availableCurrencies;
        return availableCurrencies.filter(item =>
            item.toLowerCase().includes(delayedSearchKey.toLowerCase())
        );
    }, [delayedSearchKey, availableCurrencies]);

    useEffect(() => {
        if (delayedSearchKey.length > 0) {
            setIsSearch(false);
        }
    }, [delayedSearchKey]);

    const getNearestHiddenElement = (direction) => {
        if (!listRef.current) return null;

        const container = listRef.current;
        const items = Array.from(container.children);

        for (let item of items) {
            const itemRect = item.getBoundingClientRect();
            const containerRect = container.getBoundingClientRect();

            if (direction === "right" && itemRect.right > containerRect.right) {
                return item;
            }
            if (direction === "left" && itemRect.left < containerRect.left) {
                return item;
            }
        }

        return null;
    };

    const scrollToNearestElement = (direction) => {
        const element = getNearestHiddenElement(direction);
        if (element) {
            element.scrollIntoView({behavior: "smooth", block: "nearest", inline: "start"});
        }
    };

    useEffect(() => {
        const updateScrollState = () => {
            if (!listRef.current) return;
            const container = listRef.current;
            setCanScrollLeft(container.scrollLeft > 0);
            setCanScrollRight(container.scrollLeft + container.clientWidth < container.scrollWidth);
        };

        if (listRef.current) {
            updateScrollState();
            listRef.current.addEventListener("scroll", updateScrollState);
            window.addEventListener("resize", updateScrollState);
        }

        return () => {
            if (listRef.current) {
                listRef.current.removeEventListener("scroll", updateScrollState);
            }
            window.removeEventListener("resize", updateScrollState);
        };
    }, []);

    const clearSearchImmediately = () => {
        setSearchKey('');
        setDelayedSearchKey('');
        setIsSearch(false);
    };

    const handleKeyClick = (key) => {
        onKeyPress(key);
    };

    const handleVariable = (variable) => {
        if (selectedCurrency) {
            const v = `\\text{\\textcolor{orange}{${selectedCurrency}}\\_${variable}}`
            handleKeyClick(v)
        }
    }

    return (
        <div className={"formula__keyboard"}>
            <div className={"keyboard__labels"}>
                {canScrollLeft && (
                    <div className={"labels__before"} onClick={() => scrollToNearestElement("left")}>&#171;</div>
                )}
                <div className={"labels__list"} ref={listRef}>
                    <div className={`label__item ${selectedIndex === 0 ? "choosed_label" : ""}`}
                         onClick={() => setSelectedIndex(0)}> Basic
                    </div>
                    <div className={`label__item ${selectedIndex === -1 ? "choosed_label" : ""}`}
                         onClick={() => setSelectedIndex(-1)}>Currency
                    </div>
                    {availableApi &&
                        Object.keys(availableApi).map((key, index) => (
                            <div key={key} className={`label__item ${selectedIndex === index+1 ? "choosed_label" : ""}`}
                                onClick={() => setSelectedIndex(index+1)}> {key}
                            </div>
                        ))}
                </div>
                {canScrollRight && (
                    <div className={"labels__right"} onClick={() => scrollToNearestElement("right")}>&#187;</div>
                )}
            </div>
            {selectedKeyboard?.type && selectedKeyboard.type === "basic" ? (
            <div className="basic_keyboard__list">
                {selectedKeyboard.rows.map((row, rowIndex) => (
                    <div key={rowIndex} className="basic_keyboard__row">
                        {row.map((item) => (
                            <div key={item.token || item.toString()}
                                 onClick={() => handleKeyClick(item.token ? item : item.toString())}
                                 className={`basic_keyboard__item ${typeof item === "object" && item.class ? item.class : ""}`}
                            >
                                {item.id === "frac" ? (
                                    <div className={"fraction"}>
                                        <div className={"selected_element"}>▢</div>
                                        <div className={"span"}></div>
                                        <div className={"square"}>▢</div>
                                    </div>
                                ) : item.id === "matrix" ? (
                                    <div className={"row_button"}>
                                        <div>(</div>
                                        <div className={"fraction"}>
                                            <div className={"square"}>▢</div>
                                            <div className={"span"}></div>
                                            <div className={"square"}>▢</div>
                                        </div>
                                        <div>)</div>
                                    </div>
                                ) : item.id === "mo" ? (
                                    <div className={"row_button"}>
                                        <span className={"module-left"}/>
                                        <div>▢</div>
                                        <span className={"module-right"}/>
                                    </div>
                                ) : item.id === "sq" ? (
                                    <div className={"row_button"}>
                                        <div className={"button__sqrt"}>√</div>
                                        <div className={"selected_element_v2"}></div>
                                    </div>
                                ) : item.id === "square" ? (
                                    <div className={"row_button"}>
                                        <div className={"selected_element_v3"}></div>
                                        <div className={"button_rank"}>2</div>
                                    </div>
                                ) : item.id === "square2" ? (
                                    <div className={"row_button"}>
                                        <div className={"selected_element_v3"}></div>
                                        <div className={"button_rank"}>▢</div>
                                    </div>
                                ) : item.id === "backspace" ? (
                                    <div className={"backspace__item"}>⌫</div>
                                ) : (
                                    typeof item === "string" ? item : item.label
                                )}
                            </div>
                        ))}
                    </div>
                ))}
            </div>
            ) : (
            <div className={"keyboard__choices"}>
                <div className="dynamic_keyboard__list">
                    {Object.values(availableApi)[selectedIndex-1]?.map(value => (
                        <div key={value} className="dynamic_keyboard__item" onClick={() => handleVariable(value)}>{value}</div>
                    ))}
                </div>
            </div>
            )}
            {selectedIndex === -1 && (
            <div className={"keyboard__choices"}>
                <div className={"keaboard__input__area"}>
                    <div className={"keayboard__input__field"}>
                        {searchKey.length > 0 && (
                            <div className={"cross_field"} onClick={clearSearchImmediately}>
                                <div className={"keyboard__input__cross"}></div>
                            </div>
                        )}
                        {searchKey.length === 0 ? (
                            <svg className={"keyboard__input__lupa"} viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg">
                                <circle cx="10" cy="10" r="6"/>
                                <line x1="14" y1="14" x2="20" y2="20"/>
                            </svg>
                        ) : isSearch ? (
                            <div className={"keyboard__input__loading"}><AdaptiveLoading/></div>
                        ) : (
                            <svg onClick={clearSearchImmediately} className={"keyboard__input__rollback"} xmlns="http://www.w3.org/2000/svg" viewBox="0 0 448 512">
                                <path d="M9.4 233.4c-12.5 12.5-12.5 32.8 0 45.3l160 160c12.5 12.5 32.8 12.5 45.3 0s12.5-32.8 0-45.3L109.2 288 416 288c17.7 0 32-14.3 32-32s-14.3-32-32-32l-306.7 0L214.6 118.6c12.5-12.5 12.5-32.8 0-45.3s-32.8-12.5-45.3 0l-160 160z"/>
                            </svg>
                        )}
                        <input id={"keyboard__search_field"} onChange={e => setSearchKey(e.target.value)}
                               value={searchKey} className={"keyboard__search_field"} placeholder={"Search"}></input>
                    </div>
                    <div className={"current_currency_3"}>Current currency: {selectedCurrency || "None"}</div>
                </div>
                <div className="dynamic_keyboard__list">
                    {filteredFields.map(item => (
                        <div key={item} className="dynamic_keyboard__item" onClick={() => setSelectedCurrency(item)}>{item}</div>
                    ))}
                    {filteredFields.length === 0 && <div className="dynamic_keyboard__no_results">Nothing was found</div>}
                </div>
            </div>
            )}
        </div>
    );
};

export default Keyboard;