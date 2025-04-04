import React from 'react';

const Main = () => { // FAQ
    return (
        <div className={"section__main"}>
            <div className={"faq__list"}>
                <div className={"description__item"}>
                    <div className={"description__header"}>How the trigger system works</div>
                    <div>
                        Users can create their own expressions with custom notifications, the standard keyboard includes fractions, powers up to 9, roots and modules. Trigonometric functions are not supported, and are unlikely to ever be. It is also possible to plot a graph of the triggering of custom expressions
                    </div>
                </div>
                <div className={"description__item"}>
                    <div className={"description__header"}>Available cryptocurrencies</div>
                    <div>
                        All cryptocurrencies are available that can be worked with in all available endpoints - more than 3000, about 99.8% of all cryptocurrencies on Binance, excluding only the newest ones that are not yet supported by all endpoints
                    </div>
                </div>
                <div className={"description__item"}>
                    <div className={"description__header"}>Available endpoints</div>
                    <div>
                        All variables from all significant endpoints are available, such as /v3/ticker/24hr, /v3/ticker/price, etc. There is currently no access to websockets
                    </div>
                </div>
                <div className={"description__item"}>
                    <div className={"description__header"}>Change response system</div>
                    <div>
                        If for some reason Binance stops supporting cryptocurrencies or variables received from specific APIs, users will receive a notification and formulas associated with unsupported data will no longer be supported
                    </div>
                </div>
                <div className={"description__item"}>
                    <div className={"description__header"}>Notification system</div>
                    <div>
                        Notifications can be sent to an application available on Android and a PWA application. In case there are problems with them, it is possible to use a telegram bot as an alternative
                    </div>
                </div>
            </div>
        </div>
    );
};

export default Main;