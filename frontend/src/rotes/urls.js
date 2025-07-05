import {lazy} from "react";

const Main = lazy(() => import("../pages/Main"));
const NewStrategy = lazy(() => import("../pages/NewStrategy"))
const Strategy = lazy(() => import("../pages/Strategy"))
const UserStrategies = lazy(() => import("../pages/Strategies"));
const Settigns = lazy(() => import("../pages/Settings"));

export const publicRotes = [
    {path: "/", component: <Main/>, key: "main"},
    {path: "/new-strategy/", component: <NewStrategy/>, key: "new-strategy"},
    {path: "/strategies/", component: <UserStrategies/>, key: "strategies"},
    {path: "/strategy/:id/", component: <Strategy/>, key: "strategy"},
    {path: "/settings/", component: <Settigns/>, key: "settings"}
]
