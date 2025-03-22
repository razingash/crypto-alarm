import React from 'react';
import {Link} from "react-router-dom";
import {useAuth} from "../../hooks/context/useAuth";

const Header = () => {
    const { isAuth, logout } = useAuth();

    return (
        <div className={"section__header"}>
            <div className={"header__field"}>
                <div className={"header__items"}>
                    {isAuth && (
                        <>
                        <Link to={"/profile"} className={"header__item"}>Profile</Link>
                        <Link to={"/profile/strategies"} className={"header__item"}>Strategies</Link>
                            <Link to={"/new-strategy"} className={"header__item"}>New strategy</Link>
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
        </div>
    );
};

export default Header;