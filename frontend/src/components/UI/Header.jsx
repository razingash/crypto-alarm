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
                <div className={"header__items"}>
                    <Link to={"/"} className={"header__item"}>Main</Link>
                    <Link to={"/logs"} className={"header__item"}>Logs</Link>
                    <Link to={"/new-strategy"} className={"header__item"}>New strategy</Link>
                    <Link to={"/strategies"} className={"header__item"}>Strategies</Link>
                    <Link to={"/settings"} className={"header__item"}>Settings</Link>
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