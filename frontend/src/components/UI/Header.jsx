import React, {useEffect, useState} from 'react';
import {Link} from "react-router-dom";
import {useAuth} from "../../hooks/context/useAuth";
import {useStore} from "../../utils/store";

const Header = () => {
    const {isPwaMode} = useStore();
    const { isAuth, logout } = useAuth();
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
                    {isAuth && (
                        <>
                        <Link to={"/new-strategy"} className={"header__item"}>New strategy</Link>
                        <Link to={"/strategies"} className={"header__item"}>Strategies</Link>
                        </>
                    )}
                    <Link to={"#"} className={"header__item"}>Settings</Link>
                    {isAuth ? (
                        <div onClick={async () => await logout()} className={"header__item"}>log out</div>
                    ) : (
                        <Link to={"/authentication"} className={"header__item"}>log in</Link>
                    )}
                </div>
            </div>
            {!isPwaMode &&
                <div className={"header__button__app"} onClick={handleInstall}>
                    <svg className={"svg__translator"}>
                        <use xlinkHref={"#download_app"}></use>
                    </svg>
                </div>
            }
        </div>
    );
};

export default Header;