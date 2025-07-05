import React, {useEffect, useRef, useState} from 'react';
import "../styles/strategy.css"
import {useFetching} from "../hooks/useFetching";
import TriggersService from "../API/TriggersService";
import {useObserver} from "../hooks/useObserver";
import {Link} from "react-router-dom";
import FormulaInput from "../components/FormulasEditor/FormulaInput";

const Strategies = () => {
    const [page, setPage] = useState(1);
    const [hasNext, setNext] = useState(false);
    const lastElement = useRef();
    const [formulas, setFormulast] = useState([]);
    const [fetchFormulas, isFormulasLoading, ] = useFetching(async () => {
        const data = await TriggersService.getUserFormulas({page: page})
        setFormulast((prevFormula) => {
            const newFormulas = data.data.filter(
                (formula) => !prevFormula.some((obj) => obj.id === formula.id)
            )
            return [...prevFormula, ...newFormulas]
        })
        setNext(data.has_next)
    }, 0, 1000)

    useObserver(lastElement, fetchFormulas, isFormulasLoading, hasNext, page, setPage);

    useEffect(() => {
        const loadData = async () => {
            await fetchFormulas();
        }
        void loadData();
    }, [page])

    return (
        <div className={"section__main"}>
            <div className={"strategies__list"}>
                {formulas.length > 0 ? (
                    formulas.map((formula, index) => (
                        <div className={"strategy__item"} key={formula.id} ref={index === formulas.length - 1 ? lastElement : null}>
                            <div className={"strategy__item__header"}>
                                <div className={"strategy__weight"}>Cooldown: {formula.cooldown}</div>
                                <Link to={`/strategy/${formula.id}`} className={"strategy__name"}>
                                    {formula.name || `Nameless formula with id ${formula.id}`}
                                </Link>
                            </div>
                            {formula.description && (
                            <div className={"strategy__description"}>{formula.description}</div>
                            )}
                            <div className={"strategy__info"}>
                                <div className={"strategy__info__item"}>
                                    <div>History</div>
                                    {formula.is_history_on === true ? (
                                        <div className={"param__status_on"}>On</div>
                                    ) : (
                                        <div className={"param__status_off"}>Off</div>
                                    )}
                                </div>
                                <div className={"strategy__info__item"}>
                                    <div>Notifications</div>
                                    {formula.is_notified === true ? (
                                        <div className={"param__status_on"}>On</div>
                                    ) : (
                                        <div className={"param__status_off"}>Off</div>
                                    )}
                                </div>
                                <div className={"strategy__info__item"}>
                                    <div>Active</div>
                                    {formula.is_active === true ? (
                                        <div className={"param__status_on"}>On</div>
                                    ) : (
                                        <div className={"param__status_off"}>Off</div>
                                    )}
                                </div>
                                <div className={"strategy__info__item"}>
                                    <div>Last Triggered</div>
                                    <div>{formula.last_triggered || "Never"}</div>
                                </div>
                            </div>
                            <FormulaInput formula={formula.formula_raw}/>
                            <div className={"button__show_more"}></div>
                        </div>
                    ))
                ) : (
                    <div className={"strategy__none"}>You don't possess any strategies yet</div>
                )}
            </div>
        </div>
    );
};

export default Strategies;