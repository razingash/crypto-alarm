import React, {useEffect, useState} from 'react';
import "../../styles/strategy.css"
import StrategyItem from "../../components/UI/StrategyItem";
import {useFetching} from "../../hooks/useFetching";
import TriggersService from "../../API/TriggersService";

const UserStrategies = () => {
    const [formulas, setFormulast] = useState([]);
    const [fetchFormulas, isFormulasLoading, ] = useFetching(async (formula) => {
        return await TriggersService.getUserFormulas(formula)
    }, 0, 1000)

    useEffect(() => {
        const loadData = async () => {
            if (!isFormulasLoading && formulas.length === 0) {
                const response = await fetchFormulas();
                console.log(response.data)
                response && setFormulast(response.data)
            }
        }

        void loadData()
    }, [])

    return (
        <div className={"section__main"}>
            <div className={"strategies__list"}>
                {formulas.length > 0 && (
                    formulas.map((formula) => (
                        <StrategyItem formula={formula}/>
                    ))
                )}
            </div>
        </div>
    );
};

export default UserStrategies;