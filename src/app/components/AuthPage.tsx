'use client';

import {AuthInput} from "@/app/components/Input";
import {memo, useCallback, useState} from "react";
import {ThemeButton} from "@/app/components/ThemeButton";
import {useLogin} from "@/app/gql/login";
import {ArchiReportLogo} from "@/app/components/Icons";

export const AuthPage = memo(function AuthPage() {
    const [{email, password}, setValues] = useState({email: '', password: ''});
    const {login, loading} = useLogin();

    const setValue = useCallback((name: string, value: string) => {
        setValues((state) => ({...state, [name]: value}));
    }, []);

    const handleSubmit = useCallback(() => {
        void login({
            variables: {
                email,
                password,
            }
        });
    }, [email, login, password]);

    return (
        <div className="flex relative flex-col justify-center items-center h-screen w-full">
            <ThemeButton className="absolute top-10 right-10"/>
            <ArchiReportLogo className="aspect-[4.27] object-contain object-center w-[222px] overflow-hidden self-center max-w-full mt-48" />
            <div
                className="text-zinc-500 dark:text-neutral-300 text-center text-md md:text-lg font-medium leading-7 max-w-[476px] mt-6 max-md:max-w-full">
                L’application pour vos suivis de chantier et de projets
            </div>
            <div
                className="justify-between items-stretch shadow-sm bg-white dark:bg-zinc-700 flex w-[400px] max-w-full flex-col mt-16 mb-40 px-7 py-10 rounded-2xl gap-y-5">
                <div
                    className="justify-between items-stretch border border-[color:var(--border,#EBEBEB)] dark:border-[color:var(--border,#383944)] bg-neutral-50 dark:bg-zinc-800 flex gap-5 pl-1 pr-14 py-1 rounded-full border-solid max-md:pr-5">
                    <div
                        className="text-black dark:text-white text-center text-sm font-semibold whitespace-nowrap justify-center items-stretch border border-[color:var(--border,#EBEBEB)] dark:border-[color:var(--border,#383944)] dark:bg-zinc-700 shadow-sm bg-white grow px-10 py-2 rounded-full border-solid max-md:px-5">
                        Se connecter
                    </div>
                    <div
                        className="text-zinc-500 text-center text-sm font-medium self-center grow whitespace-nowrap my-auto">
                        S’inscrire
                    </div>
                </div>
                <div className="flex flex-col w-full gap-y-1">
                    <AuthInput
                        placeholder="Email"
                        autoComplete="email"
                        type="email"
                        name="email"
                        value={email}
                        onChange={setValue}
                    />
                    <AuthInput
                        placeholder="Mot de passe"
                        autoComplete="password"
                        type="password"
                        name="password"
                        value={password}
                        onChange={setValue}
                    />
                </div>
                <button
                    className="text-white text-center text-sm font-semibold whitespace-nowrap justify-center items-center shadow-sm bg-indigo-500 px-16 py-2.5 rounded-full max-md:px-5"
                    onClick={handleSubmit}
                    disabled={loading}
                >
                    Se connecter
                </button>
            </div>
        </div>
    );
})
