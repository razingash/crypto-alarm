import {createContext, useContext, useEffect, useState} from "react";
import NotificationService from "../API/NotificationService";
import {urlBase64ToUint8Array} from "./utils";
import {useAuth} from "../hooks/context/useAuth";

export const StoreContext = createContext(null);

export const useStore = () => {
    return useContext(StoreContext);
}

export const NotificationProvider = ({children}) => {
    const {isAuth} = useAuth();
    const [pushNotification, setPushNotification] = useState(Notification.permission === "granted");

    const getInitialPwaMode = () => {
        const fromInstall = localStorage.getItem('wasJustInstalled') === 'true';

        return (
            window.matchMedia('(display-mode: standalone)').matches ||
            window.matchMedia('(display-mode: window-controls-overlay)').matches ||
            window.matchMedia('(display-mode: minimal-ui)').matches ||
            window.matchMedia('(display-mode: fullscreen)').matches ||
            window.navigator.standalone === true ||
            document.referrer.startsWith('android-app://') ||
            fromInstall
        );
    };

    const [isPwaMode, setIsPwaMode] = useState(getInitialPwaMode());

    useEffect(() => {
        if (localStorage.getItem('wasJustInstalled') === 'true') {
            localStorage.removeItem('wasJustInstalled');
        }
    }, []);

    useEffect(() => {
        const onAppInstalled = () => {
            localStorage.setItem('wasJustInstalled', 'true');
            setIsPwaMode(true);
        };
        window.addEventListener('appinstalled', onAppInstalled);
        return () => window.removeEventListener('appinstalled', onAppInstalled);
    }, []);

    useEffect(() => {
        if (isPwaMode && Notification.permission === 'default') {
            setTimeout(async () => {
                const wantsPush = window.confirm(
                    "This PWA application could send reports about strategies with custom cooldowns." +
                    "It will be difficult to turn notifications on again after blocking them, but if it happens by chance just reinstalling this application"
                );

                if (wantsPush) {
                    const permission = await Notification.requestPermission();
                    if (permission === 'granted') {
                        setPushNotification(true);
                    } else {
                        alert("You have blocked notifications. Now the application won't be able to send you reports about triggers");
                    }
                }
            }, 1000);
        }
    }, [isPwaMode]);

    useEffect(() => {
        const setupPushSubscription = async () => {
            if (!('serviceWorker' in navigator)) return;
            if (Notification.permission !== 'granted') return;
            try {
                const registration = await navigator.serviceWorker.ready;
                let subscription = await registration.pushManager.getSubscription();
                if (!subscription) {
                    const vapidKey = await NotificationService.getVapidKey();

                    subscription = await registration.pushManager.subscribe({
                        userVisibleOnly: true,
                        applicationServerKey: urlBase64ToUint8Array(vapidKey)
                    });
                }
                const subscriptionJSON = subscription.toJSON();
                console.log(subscriptionJSON)
                const { endpoint, keys: { p256dh, auth } } = subscription.toJSON();
                await NotificationService.subscribeToPushNotifications(endpoint, p256dh, auth);
            } catch (err) {
                console.error('Push subscription failed:', err);
            }
        };
        if (isAuth) {
            setupPushSubscription();
        }
    }, [isAuth]);

    const triggerPushNotification = (message) => {
        if ('serviceWorker' in navigator) {
            navigator.serviceWorker.ready.then(registration => {
                registration.active.postMessage({
                    action: 'triggerPush',
                    body: message
                });
            });
        }
    };

    return (
        <StoreContext.Provider
            value={{isPwaMode, triggerPushNotification, pushNotification, setPushNotification}}>
            {children}
        </StoreContext.Provider>
    )
}
