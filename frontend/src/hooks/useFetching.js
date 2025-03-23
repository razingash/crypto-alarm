import {useCallback, useState} from "react";

export const useFetching = (callback, delay=0, maxRetries=1) => {
    const [retryCount, setRetryCount] = useState(0);
    const [isLoading, setIsLoading] = useState(null);
    const [error, setError] = useState(null);
    const [isSpammed, setIsSpammed] = useState(null);

    const fetching = useCallback(async (...args) => {
        if (isSpammed || retryCount >= maxRetries) return;
        //console.log('usefetching')
        try {
            setIsSpammed(true);
            setIsLoading(true);
            return await callback(...args);
        } catch (e) {
            console.log(e)
            console.log(e?.response?.data)
            setError(e?.response?.data || e?.message);
            setRetryCount(prev => prev + 1);
        } finally {
            setIsLoading(false);
            setTimeout(() => {
                setIsSpammed(false);
            }, delay);
        }
    }, [callback, delay, isSpammed, retryCount, maxRetries]);

    return [fetching, isLoading, error];
};