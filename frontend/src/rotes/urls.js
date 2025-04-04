import {lazy} from "react";
import Strategy from "../pages/Strategy";

const Main = lazy(() => import("../pages/Main"));
const Auth = lazy(() => import("../pages/Auth"));
const NewStrategy = lazy(() => import("../pages/NewStrategy"))
const UserStrategies = lazy(() => import("../pages/UserStrategies"));

export const publicRotes = [
    {path: "/", component: <Main/>, key: "main"},
]

export const unprivateRotes = [
    {path: "/authentication/", component: <Auth/>, key: "login"}
]

// вместо profile сделать settings, чтобы универсальнее было
export const privateRotes = [
    {path: "/new-strategy/", component: <NewStrategy/>, key: "new-strategy"},
    {path: "/strategies/", component: <UserStrategies/>, key: "strategies"},
    {path: "/strategy/:id/", component: <Strategy/>, key: "strategy"}
]
