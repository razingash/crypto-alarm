import {lazy} from "react";

const Main = lazy(() => import("../pages/Main"));
const Errors = lazy(() => import("../pages/Errors"));
const Logs = lazy(() => import("../pages/Logs"));
const NewStrategy = lazy(() => import("../pages/NewStrategy"))
const Strategy = lazy(() => import("../pages/Strategy"))
const UserStrategies = lazy(() => import("../pages/Strategies"));
const Settigns = lazy(() => import("../pages/Settings"));
const Analytics = lazy(() => import("../pages/Analytics"));

export const publicRotes = [
    {path: "/", component: <Main/>, key: "main"},
    {path: "/errors", component: <Errors/>, key: "logs"},
    {path: "/logs", component: <Logs/>, key: "logs"},
    {path: "/new-strategy/", component: <NewStrategy/>, key: "new-strategy"},
    {path: "/strategies/", component: <UserStrategies/>, key: "strategies"},
    {path: "/strategy/:id/", component: <Strategy/>, key: "strategy"},
    {path: "/settings/", component: <Settigns/>, key: "settings"},
    {path: "/analytics/", component: <Analytics/>, key: "analytics"}
]
