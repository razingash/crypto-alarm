import {lazy} from "react";

const Main = lazy(() => import("../pages/Main"));
const Auth = lazy(() => import("../pages/Auth"));
const NewStrategy = lazy(() => import("../pages/NewStrategy"))
const Profile = lazy(() => import("../pages/profile/Profile"));
const UserStrategies = lazy(() => import("../pages/profile/UserStrategies"));
const UserStrategy = lazy(() => import("../pages/profile/UserStrategy"))

export const publicRotes = [
    {path: "/", component: <Main/>, key: "main"},
    {path: "/new-strategy/", component: <NewStrategy/>, key: "new-strategy"},
]

export const unprivateRotes = [
    {path: "/authentication/", component: <Auth/>, key: "login"}
]

export const privateRotes = [
    {path: "/profile/", component: <Profile/>, key: "profile", children: [
        {path: "/profile/strategies/", component: <UserStrategies/>, key: "user-strategies"},
        {path: "/profile/strategy/", component: <UserStrategy/>, key: "strategy"}
        ]
    },

]
