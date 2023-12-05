import {
    ApolloClient,
    ApolloLink,
    createHttpLink,
    InMemoryCache,
} from "@apollo/client";
import { onError } from "@apollo/client/link/error";
import { from } from "apollo-link";

import {UserState} from "@/app/components/AuthContext";

export function getApolloClient(auth: UserState | null) {
    const httpLink = createHttpLink({
        uri: 'https://api.archireport.dev/graphql',
    });

    const authLink = new ApolloLink((operation, forward) => {
        operation.setContext(({ headers = {} }) => ({
            headers: {
                ...headers,
                authorization: `Bearer ${auth?.token}`,
            },
        }));

        return forward(operation);
    });

    const errorLink = onError(({ graphQLErrors, networkError }) => {
        if (graphQLErrors) {
            // eslint-disable-next-line no-console
            graphQLErrors.map(({ message, locations, path }) => console.error(
                `[GraphQL error]: Message: ${message}, Location: ${locations}, Path: ${path}`,
            ));
        }

        if (networkError) {
            // eslint-disable-next-line no-console
            console.error(`[Network error]: ${networkError}`);
        }
    });

    return new ApolloClient({
        // @ts-ignore
        link: from([errorLink, authLink, httpLink]),
        cache: new InMemoryCache(),
    });
}
