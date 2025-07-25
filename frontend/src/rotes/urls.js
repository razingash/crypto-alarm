import {lazy} from "react";

const Main = lazy(() => import("../pages/Main"));
const Errors = lazy(() => import("../pages/Errors"));
const Logs = lazy(() => import("../pages/Logs"));
const NewStrategy = lazy(() => import("../pages/NewStrategy"))
const Strategy = lazy(() => import("../pages/Strategy"))
const UserStrategies = lazy(() => import("../pages/Strategies"));
const Settigns = lazy(() => import("../pages/Settings"));
const Analytics = lazy(() => import("../pages/Analytics"));
const Variables = lazy(() => import("../pages/Variables"));
const Variable = lazy(() => import("../pages/Variable"));
const NewVariable = lazy(() => import("../pages/NewVariable"));

export const publicRotes = [
    {path: "/", component: <Main/>, key: "main"},
    {path: "/errors", component: <Errors/>, key: "logs"},
    {path: "/logs", component: <Logs/>, key: "logs"},
    {path: "/new-strategy/", component: <NewStrategy/>, key: "new-strategy"},
    {path: "/strategies/", component: <UserStrategies/>, key: "strategies"},
    {path: "/strategies/:id/", component: <Strategy/>, key: "strategy"},
    {path: "/settings/", component: <Settigns/>, key: "settings"},
    {path: "/analytics/", component: <Analytics/>, key: "analytics"},
    {path: "/new-variable/", component: <NewVariable/>, key: "new-variable"},
    {path: "/variables/", component: <Variables/>, key: "variables"},
    {path: "/variables/:id/", component: <Variable/>, key: "variable"},
]
