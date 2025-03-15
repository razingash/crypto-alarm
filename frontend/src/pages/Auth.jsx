import React, {useCallback, useEffect, useState} from 'react';
import "../styles/auth.css"
import {useAuth} from "../hooks/context/useAuth";

const Auth = () => {
    const [username, SetUsername] = useState('')
    const [password, SetPassword] = useState('')
    const [isNewbie, setIsNewbie] = useState(false);
    const [canSubmit, setCanSubmit] = useState(true);
    const [activeElement, setActiveElement] = useState(null);
    const { login, register, loginError, registerError } = useAuth();

    const registerUser = async (e) => {
        e.preventDefault();
        if (!canSubmit) return;

        setCanSubmit(false);
        await register(username.value, password.value)

        setTimeout(() => {
            setCanSubmit(true);
        }, 1000)
    }

    const loginUser = async (e) => {
        e.preventDefault();
        if (!canSubmit) return;

        setCanSubmit(false);
        await login(username.value, password.value);

        setTimeout(() => {
            setCanSubmit(true);
        }, 1000)
    }

    const handleSignUpClick = useCallback((event) => {
        setIsNewbie(true);
        setActiveElement(event.target);
    }, [])

    const handleSignInClick = useCallback((event) => {
        setIsNewbie(false);
        setActiveElement(event.target);
    }, [])

    useEffect(() => {
        setActiveElement(document.querySelector('.auth__item:first-child'));
    }, [])

    return (
        <div className={"section__main section__authentication"}>
            <div className={"field__authentication"}>
                <div className={"auth__header"}>
                    <div className={`auth__item ${activeElement?.innerText === 'Sign in' ? 'active' : ''}`} onClick={handleSignInClick}>
                        Sign in
                    </div>
                    <div className={`auth__item ${activeElement?.innerText === 'Sign up' ? 'active' : ''}`} onClick={handleSignUpClick}>
                        Sign up
                    </div>
                </div>
                <div className={"auth_seperator"}></div>
                {isNewbie ? (<form onSubmit={loginUser}><div className={"field__auth"}>
                        <div className={"field__input"}>
                            <input onChange={e => SetUsername(e.target.value)} value={username} className={"input__login"} type={"text"} placeholder={"username..."}/>
                            <svg className="svg__auth-help">
                                <use xlinkHref="#icon_user"></use>
                            </svg>
                        </div>
                        <div className={"field__input"}>
                            <input onChange={e => SetPassword(e.target.value)} value={password} className={"input__password"} type={"password"} placeholder={"password..."}/>
                            <svg className={"svg__auth-help"} viewBox="0 0 20 20">
                                <path d="M10.07 0a6.1 6.1 0 0 0-6.1 6.1v2.035H2.348c-.705 0-1.276.571-1.276 1.276v9.313c0 .704.571 1.276 1.276 1.276h15.3c.704 0 1.276-.571 1.276-1.276V9.411c0-.705-.571-1.276-1.276-1.276h-1.622V6.1a6.03 6.03 0 0 0-5.96-6.1zm-.014 2.634a3.47 3.47 0 0 1 3.412 3.525v1.977H6.531V6.159c0-1.947 1.578-3.525 3.525-3.525z"/>
                            </svg>
                        </div>
                </div><div className={"auth__footer"}>
                        <div className={"reset__password"}>forgot password?</div>
                        <button className={"button__submit"}>log in</button>
                </div>
                </form>
                ) : (<form onSubmit={registerUser}><div className={"field__auth"}>
                        <div className={"field__input"}>
                            <input onChange={e => SetUsername(e.target.value)} value={username} className={"input__login"} type={"text"} placeholder={"username..."}/>
                            <svg className="svg__auth-help">
                                <use xlinkHref="#icon_user"></use>
                            </svg>
                        </div>
                        <div className={"field__input"}>
                            <input onChange={e => SetPassword(e.target.value)} value={password} className={"input__password"} type={"password"} placeholder={"password..."}/>
                            <svg className={"svg__auth-help"} viewBox="0 0 20 20">
                                <path d="M10.07 0a6.1 6.1 0 0 0-6.1 6.1v2.035H2.348c-.705 0-1.276.571-1.276 1.276v9.313c0 .704.571 1.276 1.276 1.276h15.3c.704 0 1.276-.571 1.276-1.276V9.411c0-.705-.571-1.276-1.276-1.276h-1.622V6.1a6.03 6.03 0 0 0-5.96-6.1zm-.014 2.634a3.47 3.47 0 0 1 3.412 3.525v1.977H6.531V6.159c0-1.947 1.578-3.525 3.525-3.525z"/>
                            </svg>
                        </div>
                </div><div className={"auth__footer"}>
                        <div className={"reset__password"}>forgot password?</div>
                        <button className={"button__submit"}>register</button>
                </div>
                </form>
                )}
            </div>
        </div>
    );
};

export default Auth;