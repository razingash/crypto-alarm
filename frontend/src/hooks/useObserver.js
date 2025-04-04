import {useEffect, useRef} from "react";

export const useObserver = (ref, callback, isLoading, hasNextPage, page, setPage) => {
    const observer = useRef();
    useEffect(() => {
        if (isLoading || !hasNextPage) return;
        const cb = (entries) => {
            if (entries[0].isIntersecting) {
                console.log(`loading items for page ${page}`);
                setPage((prevPage) => prevPage + 1)
                callback()
            }
        }

        observer.current = new IntersectionObserver(cb);
        if (ref.current) {
            observer.current.observe(ref.current);
        }

        return () => {
            if (observer.current && ref.current) {
                observer.current.unobserve(ref.current);
            }
        }
    }, [isLoading]);
}