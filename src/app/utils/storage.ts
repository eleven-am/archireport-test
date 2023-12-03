
export const createStorage = <T> (key: string, initialValue: T) => {
    const get = () => {
        if (typeof window === 'undefined') {
            return initialValue;
        }

        try {
            const item = window.localStorage.getItem(key);

            return item ? JSON.parse(item) : initialValue;
        } catch (error) {
            console.error(error);

            return initialValue;
        }
    }

    const set = (value: T) => {
        try {
            window.localStorage.setItem(key, JSON.stringify(value));
        } catch (error) {
            console.error(error);
        }
    }

    const remove = () => {
        localStorage.removeItem(key);
    }

    return {
        get,
        set,
        remove,
    }
}
