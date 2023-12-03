import {useCallback, useEffect, useMemo} from "react";
import {useLocalStorage} from "@/app/hooks/useLocalStorage";

type Theme = 'light' | 'dark';

export const useTheme = (): [boolean, () => void] => {
    const [theme, setTheme] = useLocalStorage<Theme>('theme', 'light');

    useEffect(() => {
        if (theme === 'dark') {
            document.documentElement.classList.add('dark');
        } else {
            document.documentElement.classList.remove('dark');
        }
    }, [theme]);

    const toggleTheme = useCallback(() => {
        setTheme(prev => prev === 'dark' ? 'light' : 'dark');
    }, [setTheme]);

    const isDark = useMemo(() => theme === 'dark', [theme]);

    return [isDark, toggleTheme];
}
