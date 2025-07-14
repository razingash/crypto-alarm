import React, {useEffect, useRef, useState} from 'react';
import "../styles/strategy.css"
import {useFetching} from "../hooks/useFetching";
import StrategyService from "../API/StrategyService";
import {useObserver} from "../hooks/useObserver";
import {Link} from "react-router-dom";
import FormulaInput from "../components/FormulasEditor/FormulaInput";
import AdaptiveLoading from "../components/UI/AdaptiveLoading";
import ErrorField from "../components/UI/ErrorField";

const Strategies = () => {
    const [page, setPage] = useState(1);
    const [hasNext, setNext] = useState(false);
    const lastElement = useRef();
    const [strategies, setStrategies] = useState([]);
    const [fetchStrategies, isStrategiesLoading, strategiesError] = useFetching(async () => {
        const data = await StrategyService.getStrategies({page: page})
        setStrategies((prevStrategies) => {
            const newStrategies = data.data.filter(
                (strategy) => !prevStrategies.some((obj) => obj.id === strategy.id)
            )
            return [...prevStrategies, ...newStrategies]
        })
        setNext(data.has_next)
    }, 0, 1000)

    useObserver(lastElement, fetchStrategies, isStrategiesLoading, hasNext, page, setPage);

    useEffect(() => {
        const loadData = async () => {
            await fetchStrategies();
        }
        void loadData();
    }, [page])

    return (
        <div className={"section__main"}>
            {isStrategiesLoading ? (
                <div className={"loading__center"}>
                    <AdaptiveLoading/>
                </div>
            ) : strategiesError ? (
                <ErrorField/>
            ) : strategies.length > 0 ? (
                <div className={"strategies__list"}>
                    {strategies.map((strategy, index) => (
                    <div className={"strategy__item"} key={strategy.id} ref={index === strategies.length - 1 ? lastElement : null}>
                        <div className={"strategy__item__header"}>
                            <div className={"strategy__weight"}>Cooldown: {strategy.cooldown}</div>
                            <Link to={`/strategy/${strategy.id}`} className={"strategy__name"}>
                                {strategy.name || `Nameless formula with id ${strategy.id}`}
                            </Link>
                        </div>
                        {strategy.description && (
                        <div className={"strategy__description"}>{strategy.description}</div>
                        )}
                        <div className={"strategy__info"}>
                            <div className={"strategy__info__item"}>
                                <div>History</div>
                                {strategy.is_history_on === true ? (
                                    <div className={"param__status_on"}>On</div>
                                ) : (
                                    <div className={"param__status_off"}>Off</div>
                                )}
                            </div>
                            <div className={"strategy__info__item"}>
                                <div>Notifications</div>
                                {strategy.is_notified === true ? (
                                    <div className={"param__status_on"}>On</div>
                                ) : (
                                    <div className={"param__status_off"}>Off</div>
                                )}
                            </div>
                            <div className={"strategy__info__item"}>
                                <div>Active</div>
                                {strategy.is_active === true ? (
                                    <div className={"param__status_on"}>On</div>
                                ) : (
                                    <div className={"param__status_off"}>Off</div>
                                )}
                            </div>
                            <div className={"strategy__info__item"}>
                                <div>Last Triggered</div>
                                <div>{strategy.last_triggered || "Never"}</div>
                            </div>
                        </div>
                        {strategy.conditions.map((condition) => (
                             <FormulaInput formula={condition.formula_raw}/>
                        ))}
                        <div className={"button__show_more"}></div>
                    </div>
                    ))}
                </div>
            ) : (isStrategiesLoading === false && !strategiesError) && (
                <ErrorField message={"You don't possess any strategies yet"}/>
            )}
        </div>
    );
};

export default Strategies;