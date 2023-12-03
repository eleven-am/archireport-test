'use client';

import {Dispatch, SetStateAction, useCallback, useRef, useState} from "react";
import {createStorage} from "@/app/utils/storage";

export const useLocalStorage = <T extends unknown> (key: string, initialValue: T): [T, Dispatch<SetStateAction<T>>] => {
    const storageRef = useRef(createStorage(key, initialValue));
    const [storedValue, setStoredValue] = useState<T>(storageRef.current.get());

    const setValue = useCallback((value: T | ((storedValue: T) => T)) => {
        setStoredValue(prev => {
            if (value instanceof Function) {
                const newValue = value(prev);

                storageRef.current.set(newValue);

                return newValue;
            } else {
                storageRef.current.set(value);

                return value;
            }
        });
    }, []);

    return [storedValue, setValue];
}
