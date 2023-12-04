import {ChangeEvent, useCallback, memo} from "react";

interface AuthInputProps {
    type: string;
    name: string;
    value: string;
    onChange?: (name: string, value: string) => void;
    placeholder: string;
    autoComplete?: string;
    required?: boolean;
    disabled?: boolean;
}

export const AuthInput = memo(function AuthInput({ type, name, value, onChange, placeholder, autoComplete, required, disabled }: AuthInputProps) {
    const onChangeCallback = useCallback((event: ChangeEvent<HTMLInputElement>) => {
        if (onChange) {
            onChange(event.target.name, event.target.value);
        }
    }, [onChange]);

    return (
        <input
            placeholder={placeholder}
            autoComplete={autoComplete}
            required={required}
            disabled={disabled}
            type={type}
            name={name}
            value={value}
            onChange={onChangeCallback}
            className="w-full text-zinc-500 text-base font-medium leading-6 whitespace-nowrap border border-[color:var(--border,#EBEBEB)] dark:border-[color:var(--border,#383944)] bg-neutral-50 dark:bg-zinc-800 justify-center px-4 py-5 rounded-md border-solid items-start max-md:pr-5"
        />
    )
});
