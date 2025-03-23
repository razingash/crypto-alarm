import {createContext, useContext, useEffect, useRef, useState} from "react";
import {useFetching} from "../useFetching";
import AuthService from "../../API/AuthService";

export const AuthContext = createContext(null);

export const useAuth = () => {
    return useContext(AuthContext);
}

export const AuthProvider = ({ children }) => {
    const [loading, setLoading] = useState(true);
    const [isAuth, setIsAuth] = useState(false);
    const tokenRef = useRef({access: null});

    const [fetchRegisteredUser, , registerError] = useFetching(async (username, password) => {
        return await AuthService.register(username, password)
    }, 0, 1000)
    const [fetchLoginUser, , loginError] = useFetching(async (username, password) => {
        return await AuthService.login(username, password);
    }, 0, 1000)
    const [fetchLogoutUser] = useFetching(async (refreshToken) => {
        return await AuthService.logout(refreshToken);
    }, 0, 10)
    const [fetchVerifiedToken] = useFetching(async (token) => {
        return await AuthService.verifyToken(token);
    }, 0, 1000)
    const [fetchRefreshedToken] = useFetching(async (refreshToken) => {
        return await AuthService.refreshAccessToken(refreshToken)
    }, 0, 1000)

    const login = async (username, password) => { // good
        const data = await fetchLoginUser(username, password);
        if (data != null) {
            localStorage.setItem('token', data.refresh);
            tokenRef.current.access = data.access;
            setIsAuth(true);
            console.log(data)
        } else {
            console.log("error during login")
        }
    }

    const register = async (username, password) => {
        const data = await fetchRegisteredUser(username, password);
        if (data != null) {
            localStorage.setItem('token', data.refresh)
            tokenRef.current.access = data.access;
            setIsAuth(true);
            console.log(data)
        } else {
            console.log("error during register")
        }
    }

    const logout = async () => { // good
        const refreshToken = localStorage.getItem('token');
        if (!refreshToken) { //
            setIsAuth(false);
        }
        else if (refreshToken) {
             await fetchLogoutUser(refreshToken);
        }
        localStorage.removeItem('token')
        tokenRef.current.access = null;
        setIsAuth(false);
    }

    const validateRefreshToken = async () => { // good
        const refreshToken = localStorage.getItem('token')
        if (refreshToken) {
            const result = await fetchVerifiedToken(refreshToken);
            if (result != null) { // all good
                return result?.isValid === true;
            } else { // error, so exit
                return false
            }
        } else {
            setIsAuth(false);
            await logout(); // there will be no request bcs refreshToken undefinded
        }
    }

    const refreshAccessToken = async () => { // must be good?
        const isRefreshTokenValid = await validateRefreshToken();
        const refreshToken = localStorage.getItem('token');
        if (isRefreshTokenValid === true) {
            const data = await fetchRefreshedToken(refreshToken);
            if (data) {
                tokenRef.current.access = data.access;
                setIsAuth(true);
                return data.access
            }
        } else { // refresh token has expired
            localStorage.removeItem('token')
            tokenRef.current.access = null;
            setIsAuth(false);
        }
    }

    useEffect(() => {
        setLoading(true);
        const initializeAuth = async () => {
            const isRefreshTokenValid = await validateRefreshToken();
            if (isRefreshTokenValid) { // if true, then will be an authorization attempt
                await refreshAccessToken();
            } else {
                await logout();
            }
            setLoading(false);
        }
        void initializeAuth();
    }, [isAuth]);

    return (
        <AuthContext.Provider
            value={{
                loading, isAuth, tokenRef, register, login, loginError, registerError,
                logout, refreshAccessToken, validateRefreshToken
            }}>
            {children}
        </AuthContext.Provider>
    )
}
