import React from 'react';
import { Suspense } from "react";
import {publicRotes} from "../rotes/urls";
import {Navigate, Route, Routes} from "react-router-dom";

const AppRouter = () => {
    return (
        <Suspense fallback={<></>}>
            <Routes>
                {publicRotes.map(route =>
                    <Route path={route.path} element={route.component} key={route.key}></Route>
                )}
                <Route path="*" element={<Navigate to="" replace />} key={"redirect"}/>
            </Routes>
        </Suspense>
    );
};

export default AppRouter;