import React, {useEffect, useMemo, useRef, useState} from 'react';
import "../styles/keyboard.css"
import AdaptiveLoading from "./UI/AdaptiveLoading";

/*
1) прокрутка верхней херни
2) разделить на модули базовой клавиатуры и динамической

*/

const KeyboardV2 = () => {
    const [selectedIndex, setSelectedIndex] = useState(0);
    const [searchKey, setSearchKey] = useState("");
    const [isSearch, setIsSearch] = useState(false);
    const [delayedSearchKey, setDelayedSearchKey] = useState(""); // отсроченный поиск
    const listRef = useRef(null);
    const [canScrollLeft, setCanScrollLeft] = useState(false);
    const [canScrollRight, setCanScrollRight] = useState(false);

    const availableLabels = [
        "Basic", "complain", "pastoral", "funny", "company", "sedate", "legal", "selective", "daily", "cemetery", "bat",
        "current", "untidy", "groan", "reproduce", "squash", "woman", "adhesive", "earthy", "abnormal", "flashy",
        "tumble", "dogs", "dazzling", "meddle", "driving", "scintillating", "powerful", "famous", "feeling", "jump",
        "observe", "home", "craven", "selfish", "null", "naughty", "wiry", "soggy", "damage", "right", "six", "cannon"
        , "tomatoes", "impulse", "roll", "ripe", "rainstorm", "bizarre", "combative", "exchange", "town", "ghost",
        "hose", "striped", "slave", "saw", "texture", "courageous", "vulgar", "boy"
    ]

    const availableFields = useMemo(() => [
        "1 test item", "2 test item", "3 test item", "4 test item", "5 test item", "6 test item",
        "7 test item", "8 test item", "9 test item", "10 test item", "11 test item", "12 test item",
        "13 test item", "14 test item",
    ], []);

    useEffect(() => {
        //поиск начинается спустя 500мс после того как пользователь закончит вводить инфу
        const timeout = setTimeout(() => {
            setDelayedSearchKey(searchKey);
            setIsSearch(true);
        }, 500);

        return () => clearTimeout(timeout);
    }, [searchKey]);

    const filteredFields = useMemo(() => {
        if (!delayedSearchKey) return availableFields;
        return availableFields.filter(item =>
            item.toLowerCase().includes(delayedSearchKey.toLowerCase())
        );
    }, [delayedSearchKey, availableFields]);

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

    return (
        <div className={"section__main"}>
            <div className={"formula__input"} id={"formula__input"}>latex?</div>
            <div className={"formula__keyboard"}>
                <div className={"keyboard__labels"}>
                    <div className={"labels__before"} onClick={() => scrollToNearestElement("left")}>&#171;</div>
                    <div className={"labels__list"} ref={listRef}>
                        {availableLabels.map((label, index) => (
                            <div key={index} className={`label__item ${selectedIndex === index ? "choosed_label" : ""}`}
                                onClick={() => setSelectedIndex(index)}>{label}
                            </div>
                        ))}
                    </div>
                    <div className={"labels__right"} onClick={() => scrollToNearestElement("right")}>&#187;</div>
                </div>
                <div className={"keyboard__choices"}>
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
                    <div className={"keyboard__list"}>
                        {filteredFields.map(item => (
                            <div key={item} className="keyboard__item">{item}</div>
                        ))}
                        {filteredFields.length === 0 && <div className="keyboard__no_results">Nothing was found</div>}
                    </div>
                </div>
            </div>
        </div>
    );
};

export default KeyboardV2;