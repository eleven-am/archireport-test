'use client';

import { gql, useMutation } from '@apollo/client';
import {useEffect} from "react";
import {useAuthContext, UserState} from "@/app/components/AuthContext";

const LOGIN = gql`
    mutation ($email: EmailAddress!, $password: String!) {
        login(email: $email, password: $password) {
            token
            account {
                _id
                properties {
                    email
                    firstname
                    lastname
                }
            }
        }
    }
`;

export const useLogin = () => {
    const { setAuthError, setAuth } = useAuthContext();
    const [login, { data, loading, error }] = useMutation(LOGIN);

    useEffect(() => {
        if (data && data.login) {
            const userData = data.login as UserState;
            setAuth(userData);
        }
    }, [data, setAuth]);

    useEffect(() => {
        if (error) {
            setAuthError(error.message);
        }
    }, [error, setAuthError]);

    return {
        login,
        data,
        loading,
        error,
    };
}

