import {lazy} from "react";

const Main = lazy(() => import("../pages/Main"));
const Auth = lazy(() => import("../pages/Auth"));
const NewStrategy = lazy(() => import("../pages/NewStrategy"))
const UserStrategies = lazy(() => import("../pages/profile/UserStrategies"));

export const publicRotes = [
    {path: "/", component: <Main/>, key: "main"},
]

export const unprivateRotes = [
    {path: "/authentication/", component: <Auth/>, key: "login"}
]

// вместо profile сделать settings, чтобы универсальнее было
export const privateRotes = [
    {path: "/new-strategy/", component: <NewStrategy/>, key: "new-strategy"},
    {path: "/strategies/", component: <UserStrategies/>, key: "user-strategies"},
]
