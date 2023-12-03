'use client';

import {createContext, ReactNode, useCallback, useContext, useEffect, useState} from "react";
import {useLocalStorage} from "@/app/hooks/useLocalStorage";
import {ApolloProvider} from "@apollo/client";
import {getApolloClient} from "@/app/gql/apolloClient";
import {usePathname, useRouter} from "next/navigation";

export interface UserState {
    token: string;
    code: string;
    _id: string;
    properties: {
        email: string;
        firstname: string;
        lastname: string;
    };
}

interface AuthContext {
    auth: UserState | null;
    error: string | null;
    setAuth: (auth: UserState | null) => void;
    setAuthError: (error: string | null) => void;
    getNewToken: () => Promise<string>;
}

interface AuthProviderProps {
    children: ReactNode;
}

const AuthContext = createContext<AuthContext>({
    auth: null,
    error: null,
    setAuth: () => {},
    setAuthError: () => {},
    getNewToken: () => Promise.resolve(''),
});

const defaultAuth: UserState = {
    token: '',
    code: '',
    _id: '',
    properties: {
        email: 'email@example.com',
        firstname: 'John',
        lastname: 'Doe',
    },
};

export const AuthProvider = ({children}: AuthProviderProps) => {
    const [auth, setAuth] = useLocalStorage<UserState | null>('auth', defaultAuth);
    const [error, setAuthError] = useState<string | null>(null);
    const router = useRouter();
    const pathname = usePathname();

    const getNewToken = useCallback(() => {
        return Promise.resolve('');
    }, []);

    useEffect(() => {
        if (auth === null && pathname !== '/auth') {
            router.push('/auth');
        } else if (auth !== null && pathname === '/auth') {
            router.push('/');
        }
    }, [auth, pathname, router]);

    return (
        <ApolloProvider client={getApolloClient(auth)}>
            <AuthContext.Provider value={{auth, error, setAuth, setAuthError, getNewToken}}>
                {children}
            </AuthContext.Provider>
        </ApolloProvider>
    );
};

export const useAuthContext = () => {
    const context = useContext(AuthContext);

    if (context === undefined) {
        throw new Error('useAuthContext must be used within a AuthProvider');
    }

    return context;
}
