import React, {useState} from 'react';
import FormulaInput from "../FormulasEditor/FormulaInput";

const StrategyItem = ({formula}) => {
    const [isChecked, setIsChecked] = useState(false);

    const handlePushNotificationToggle = (e) => {
        setIsChecked(e.target.checked);
    };

    return (
        <div className={"strategy__item"} key={formula.formula}>
            <div className={"strategy__item__header"}>
                <div className={"strategy__weight"}>Weight: 80</div>
                <div className={"strategy__name"}>{formula.name}</div>
            </div>
            <div className={"strategy__description"}>{formula.description}</div>
            <div className={"strategy__info"}>
                <div className={"strategy__info__item"}>
                    <div>History</div>
                    <label htmlFor="history_slider" className={"checkbox_zipline"}>
                        <span className={"zipline"}></span>
                        <input id="history_slider" type="checkbox" className={"switch"}
                               onChange={handlePushNotificationToggle} checked={formula.is_history_on}
                        />
                        <span className="slider"></span>
                    </label>
                </div>
                <div className={"strategy__info__item"}>
                    <div>Notifications</div>
                    <label htmlFor="notifications_slider" className={"checkbox_zipline"}>
                        <span className={"zipline"}></span>
                        <input id="notifications_slider" type="checkbox" className={"switch"}
                               onChange={handlePushNotificationToggle} checked={formula.is_notified}
                        />
                        <span className="slider"></span>
                    </label>
                </div>
                <div className={"strategy__info__item"}>
                    <div>Active</div>
                    <label htmlFor="relevance_slider" className={"checkbox_zipline"}>
                        <span className={"zipline"}></span>
                        <input id="relevance_slider" type="checkbox" className={"switch"}
                               onChange={handlePushNotificationToggle} checked={formula.is_active}
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