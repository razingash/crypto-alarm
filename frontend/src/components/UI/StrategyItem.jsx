import React, {useState} from 'react';
import FormulaInput from "../FormulasEditor/FormulaInput";

const StrategyItem = ({formula}) => {
    const [isHistoryAvailable, setIsHistoryAvailable] = useState(formula.is_history_on);
    const [isNoisy, setIsNoisy] = useState(formula.is_notified);
    const [isActive, setIsActive] = useState(formula.is_active);

    return (
        <div className={"strategy__item"} key={formula.id}>
            <div className={"strategy__item__header"}>
                <div className={"strategy__weight"}>Weight: 80</div>
                <div className={"strategy__name"}>{formula.name}</div>
            </div>
            <div className={"strategy__description"}>{formula.description}</div>
            <div className={"strategy__info"}>
                <div className={"strategy__info__item"}>
                    <div>History</div>
                    <label htmlFor={`history_slider${formula.id}`} className={"checkbox_zipline"}>
                        <span className={"zipline"}></span>
                        <input id={`history_slider${formula.id}`} type="checkbox" className={"switch"}
                               onChange={() => setIsHistoryAvailable(!isHistoryAvailable)} checked={isHistoryAvailable}
                        />
                        <span className="slider"></span>
                    </label>
                </div>
                <div className={"strategy__info__item"}>
                    <div>Notifications</div>
                    <label htmlFor={`notifications_slider_${formula.id}`} className={"checkbox_zipline"}>
                        <span className={"zipline"}></span>
                        <input id={`notifications_slider_${formula.id}`} type="checkbox" className={"switch"}
                               onChange={() => setIsNoisy(!isNoisy)} checked={isNoisy}
                        />
                        <span className="slider"></span>
                    </label>
                </div>
                <div className={"strategy__info__item"}>
                    <div>Active</div>
                    <label htmlFor={`relevance_slider${formula.id}`} className={"checkbox_zipline"}>
                        <span className={"zipline"}></span>
                        <input id={`relevance_slider${formula.id}`} type="checkbox" className={"switch"}
                               onChange={() => setIsActive(!isActive)} checked={isActive}
                        />
                        <span className="slider"></span>
                    </label>
                </div>
                <div className={"strategy__info__item"}>
                    <div>Last Triggered</div>
                    <div>{formula.last_triggered}</div>
                </div>
            </div>
            <FormulaInput formula={formula.formula}/>
            <div className={"button__show_more"}></div>
        </div>
    );
};

export default StrategyItem;