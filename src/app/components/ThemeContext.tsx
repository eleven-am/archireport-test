'use client';

import {createContext, memo, ReactNode, useContext} from "react";
import {useTheme} from "@/app/hooks/useTheme";

interface ThemeState {
    darkMode: boolean;
    toggleDarkMode: () => void;
}

const ThemeContext = createContext<ThemeState>({
    darkMode: false,
    toggleDarkMode: () => {},
});

export const ThemeProvider = memo(function ThemeProvider({children}: {children: ReactNode}) {
    const [darkMode, toggleDarkMode] = useTheme();

    return (
        <ThemeContext.Provider value={{darkMode, toggleDarkMode}}>
            {children}
        </ThemeContext.Provider>
    )
})

export const useThemeContext = () => {
    const context = useContext(ThemeContext);

    if (context === undefined) {
        throw new Error('useThemeContext must be used within a ThemeProvider');
    }

    return context;
}
