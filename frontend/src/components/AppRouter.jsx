import React from 'react';
import { Suspense } from "react";
import {privateRotes, publicRotes, unprivateRotes} from "../rotes/urls";
import {Navigate, Route, Routes} from "react-router-dom";
import {useAuth} from "../hooks/context/useAuth";
import {useApiInterceptors} from "../hooks/useApiInterceptor";

const AppRouter = () => {
    useApiInterceptors();
    const {isAuth, loading} = useAuth();

    if (loading) {
        return <></>
    }

    return (
        <Suspense fallback={<></>}>
            <Routes>
                {isAuth ? (
                    privateRotes.map(route =>
                        <Route path={route.path} element={route.component} key={route.key}>
                            {route.children && route.children.map(child => (
                                <Route path={child.path} element={child.component} key={child.key} />
                            ))}
                        </Route>
                    )
                ) : (
                    unprivateRotes.map(route =>
                        <Route path={route.path} element={route.component} key={route.key}></Route>
                    )
                )}
                {publicRotes.map(route =>
                    <Route path={route.path} element={route.component} key={route.key}></Route>
                )}
                <Route path="*" element={<Navigate to="" replace />} key={"redirect"}/>
            </Routes>
        </Suspense>
    );
};

export default AppRouter;