import React, {useEffect, useState} from 'react';
import {Link, NavLink} from "react-router-dom";
import {useStore} from "../../utils/store";

const Header = () => {
    const {isPwaMode} = useStore();
    const [deferredPrompt, setDeferredPrompt] = useState(null);
    const [menuOpen, setMenuOpen] = useState(false);

    const handleLinkClick = () => {
        setTimeout(() => {
            const toggle = document.getElementById('menu__toggle');
            document.body.style.overflow = 'none';
            if (toggle) toggle.checked = false;
        }, 0);
    };

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
                <label htmlFor="menu__toggle" className="menu__button" onClick={() => setMenuOpen(prev => !prev)}>
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
                    <NavLink to={"/"} className={({ isActive }) => `header__item ${isActive ? "link_active" : ""}`} onClick={handleLinkClick}>
                        <svg className={"svg__header"}>
                            <use xlinkHref={"#icon_town_council"}></use>
                        </svg>
                        <div>Main</div>
                    </NavLink>
                    <NavLink to={"/errors"} className={({ isActive }) => `header__item ${isActive ? "link_active" : ""}`} onClick={handleLinkClick}>
                        <svg className={"svg__header"}>
                            <use xlinkHref={"#icon_judjments_hammer"}></use>
                        </svg>
                        <div>Errors</div>
                    </NavLink>
                    <NavLink to={"/logs"} className={({ isActive }) => `header__item ${isActive ? "link_active" : ""}`} onClick={handleLinkClick}>
                        <svg className={"svg__header"}>
                            <use xlinkHref={"#icon_newspaper"}></use>
                        </svg>
                        <div>Logs</div>
                    </NavLink>
                    <NavLink to={"/strategies"} className={({ isActive }) => `header__item ${isActive ? "link_active" : ""}`} onClick={handleLinkClick}>
                        <svg className={"svg__header"}>
                            <use xlinkHref={"#icon_rook"}></use>
                        </svg>
                        <div>Strategies</div>
                    </NavLink>
                    <NavLink to={"/new-strategy"} className={({ isActive }) => `header__item ${isActive ? "link_active" : ""}`} onClick={handleLinkClick}>
                        <svg className={"svg__header"}>
                            <use xlinkHref={"#icon_rook"}></use>
                        </svg>
                        <div>New strategy</div>
                    </NavLink>
                    <NavLink to={"/variables"} className={({ isActive }) => `header__item ${isActive ? "link_active" : ""}`} onClick={handleLinkClick}>
                        <svg className={"svg__header"}>
                            <use xlinkHref={"#icon_oprosnik"}></use>
                        </svg>
                        <div>Variables</div>
                    </NavLink>
                    <NavLink to={"/new-variable"} className={({ isActive }) => `header__item ${isActive ? "link_active" : ""}`} onClick={handleLinkClick}>
                        <svg className={"svg__header"}>
                            <use xlinkHref={"#icon_oprosnik"}></use>
                        </svg>
                        <div>New Variable</div>
                    </NavLink>
                    <NavLink to={"/analytics"} className={({ isActive }) => `header__item ${isActive ? "link_active" : ""}`} onClick={handleLinkClick}>
                        <svg className={"svg__header"}>
                            <use xlinkHref={"#icon_stonks"}></use>
                        </svg>
                        <div>Analytics</div>
                    </NavLink>
                    <NavLink to={"/settings"} className={({ isActive }) => `header__item ${isActive ? "link_active" : ""}`} onClick={handleLinkClick}>
                        <svg className={"svg__header"}>
                            <use xlinkHref={"#icon_gear"}></use>
                        </svg>
                        <div>Settings</div>
                    </NavLink>
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