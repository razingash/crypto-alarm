import {useEffect, useRef, useState} from 'react';
import {websocketBaseURL} from "./useApiInterceptor";

const UseWebSocket = (link) => {
    const [prevValue, setPrevValue] = useState(null);
    const [value, setValue] = useState(null);
    const [connected, setConnected] = useState(false);
    const socketRef = useRef();

    useEffect(() => {
        socketRef.current = new WebSocket(`${websocketBaseURL}${link}`);

        socketRef.current.onopen = () => {
            console.log('connected')
            setConnected(true);

        }

        socketRef.current.onmessage = (event) => {
            const message = JSON.parse(event.data);
            console.log(message)
            if (JSON.stringify(message) !== JSON.stringify(prevValue)) {
                setPrevValue(message);
                setValue(message);
            }
        }

        socketRef.current.onclose = () => {
            console.log('closing')
        }

        socketRef.current.onerror = () => {
            console.log('error')
        }
        return () => {
            if (socketRef.current) {
                socketRef.current.close();
                console.log('WebSocket connection closed on cleanup');
            }
        };
    }, [link])

    return [value, setValue, prevValue, connected];
};

export default UseWebSocket;