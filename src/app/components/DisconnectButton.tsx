'use client';

import {memo, useCallback} from 'react';
import {RoundButton} from "@/app/components/Button";
import {LogOutIcon} from "@/app/components/Icons";
import {useAuthContext} from "@/app/components/AuthContext";

export const DisconnectButton = memo(function DisconnectButton() {
    const {setAuth} = useAuthContext();

    const disconnect = useCallback(() => {
        setAuth(null);
    }, [setAuth]);

    return (
        <RoundButton
            tooltip={'log out'}
            className={'text-red-500 dark:text-red-400'}
            Icon={<LogOutIcon/>}
            handleClick={disconnect}
        />
    );
});
