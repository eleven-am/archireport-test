'use client';

import {memo} from "react";
import {MoonIcon, SunIcon} from "@/app/components/Icons";
import {RoundButton} from "@/app/components/Button";
import {useThemeContext} from "@/app/components/ThemeContext";

interface ThemeButtonProps {
    className?: string;
}

export const ThemeButton = memo(function ThemeButton({className}: ThemeButtonProps) {
    const {darkMode, toggleDarkMode} = useThemeContext();

    return (
        <RoundButton
            className={className}
            Icon={darkMode ? <SunIcon /> : <MoonIcon />}
            tooltip={darkMode ? 'light mode' : 'dark mode'}
            handleClick={toggleDarkMode}
        />
    )
});
