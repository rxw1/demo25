"use client";
import {
    ApolloClient,
    ApolloLink,
    HttpLink,
    InMemoryCache,
} from "@apollo/client";
import { GraphQLWsLink } from "@apollo/client/link/subscriptions";
import { ApolloProvider } from "@apollo/client/react";
import { OperationTypeNode } from "graphql";
import { createClient } from "graphql-ws";
import Comp1 from "./components/comp1";

export const graphQLEndpoint =
    process.env.NEXT_PUBLIC_GRAPHQL_URL || "http://localhost:8080/graphql";
const wsUrl = graphQLEndpoint.replace(/^http/, "ws");

const httpLink = new HttpLink({
    uri: graphQLEndpoint,
});

const wsLink = new GraphQLWsLink(
    createClient({
        url: wsUrl,
    })
);

const splitLink = ApolloLink.split(
    ({ operationType }) => {
        return operationType === OperationTypeNode.SUBSCRIPTION;
    },
    wsLink,
    httpLink
);

const client = new ApolloClient({
    link: splitLink,

    cache: new InMemoryCache(),
});

export default function Home() {
    return (
        <ApolloProvider client={client}>
            <Comp1 />
        </ApolloProvider>
    );
}
