import React, {useEffect, useRef, useState} from 'react';
import "../styles/strategy.css"
import StrategyItem from "../components/UI/StrategyItem";
import {useFetching} from "../hooks/useFetching";
import TriggersService from "../API/TriggersService";
import {useObserver} from "../hooks/useObserver";

const UserStrategies = () => {
    const [page, setPage] = useState(1);
    const [hasNext, setNext] = useState(false);
    const lastElement = useRef();
    const [formulas, setFormulast] = useState([]);
    const [fetchFormulas, isFormulasLoading, ] = useFetching(async () => {
        const data = await TriggersService.getUserFormulas(page)
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
        console.log(page)
    }, [page])

    return (
        <div className={"section__main"}>
            <div className={"strategies__list"}>
                {formulas.length > 0 && (
                    formulas.map((formula, index) => (
                        <div className={"strategy__item"} key={formula.id} ref={index === formulas.length - 1 ? lastElement : null}>
                            <StrategyItem formula={formula}/>
                        </div>
                    ))
                )}
            </div>
        </div>
    );
};

export default UserStrategies;