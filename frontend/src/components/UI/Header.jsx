import React, {useEffect, useState} from 'react';
import {Link} from "react-router-dom";
import {useStore} from "../../utils/store";

const Header = () => {
    const {isPwaMode} = useStore();
    const [deferredPrompt, setDeferredPrompt] = useState(null);

    useEffect(() => {
        const handleBeforeInstallPrompt = (e) => {
            e.preventDefault();
            setDeferredPrompt(e);
        };

        window.addEventListener("beforeinstallprompt", handleBeforeInstallPrompt);

        return () => {
            window.removeEventListener("beforeinstallprompt", handleBeforeInstallPrompt);
        };
    }, []);

    const handleInstall = () => {
        if (deferredPrompt) {
            deferredPrompt.prompt();
            setDeferredPrompt(null);
        }
    };

    return (
        <div className={"section__header"}>
            <div className={"header__field"}>
                <input id="menu__toggle" type="checkbox"/>
                <label htmlFor="menu__toggle" className="menu__button">
                    <span className="menu__bar"></span>
                    <span className="menu__bar"></span>
                    <span className="menu__bar"></span>
                </label>
                <div className={"header__items"}>
                    <label htmlFor="menu__toggle" className={"header__dropdown__close"}>
                        <svg className="svg__cross">
                            <use xlinkHref="#icon_cross"></use>
                        </svg>
                    </label>
                    <label htmlFor="menu__toggle" className={"header__item"}>
                        <Link to={"/"} className={"header__link"}>Main</Link>
                    </label>
                    <label htmlFor="menu__toggle" className={"header__item"}>
                        <Link to={"/errors"} className={"header__link"}>Errors</Link>
                    </label>
                    <label htmlFor="menu__toggle" className={"header__item"}>
                        <Link to={"/logs"} className={"header__link"}>Logs</Link>
                    </label>
                    <label htmlFor="menu__toggle" className={"header__item"}>
                        <Link to={"/strategies"} className={"header__link"}>Strategies</Link>
                    </label>
                    <label htmlFor="menu__toggle" className={"header__item"}>
                        <Link to={"/new-strategy"} className={"header__link"}>New strategy</Link>
                    </label>
                    <label htmlFor="menu__toggle" className={"header__item"}>
                        <Link to={"/variables"} className={"header__link"}>Variables</Link>
                    </label>
                    <label htmlFor="menu__toggle" className={"header__item"}>
                        <Link to={"/variables"} className={"header__link"}>New Variable</Link>
                    </label>
                    <label htmlFor="menu__toggle" className={"header__item"}>
                        <Link to={"/analytics"} className={"header__link"}>Analytics</Link>
                    </label>
                    <label htmlFor="menu__toggle" className={"header__item"}>
                        <Link to={"/settings"} className={"header__link"}>Settings</Link>
                    </label>
                </div>
            </div>
            {!isPwaMode &&
                <div className={"header__button__app"} onClick={handleInstall}>
                    <svg className={"svg__pwa"}>
                        <use xlinkHref={"#download_app"}></use>
                    </svg>
                </div>
            }
        </div>
    );
};

export default Header;