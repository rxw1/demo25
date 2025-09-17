"use client"
import {
  ApolloClient,
  ApolloLink,
  HttpLink,
  InMemoryCache,
  gql,
} from "@apollo/client"
import { GraphQLWsLink } from "@apollo/client/link/subscriptions"
import { ApolloProvider, useMutation, useQuery } from "@apollo/client/react"
import { getMainDefinition } from "@apollo/client/utilities"
import { OperationTypeNode } from "graphql"
import { createClient } from "graphql-ws"
import Comp1 from "./components/comp1"

const httpLink = new HttpLink({
  uri: process.env.NEXT_PUBLIC_GRAPHQL_URL || "http://localhost:8080/graphql",
})

const wsLink = new GraphQLWsLink(
  createClient({
    url: "ws://localhost:8080/graphql",
  })
)

const splitLink = ApolloLink.split(
  ({ operationType }) => {
    return operationType === OperationTypeNode.SUBSCRIPTION
  },
  wsLink,
  httpLink
)

const client = new ApolloClient({
  link: splitLink,

  cache: new InMemoryCache(),
})

export default function Home() {
  return (
    <ApolloProvider client={client}>
      <Comp1 />
    </ApolloProvider>
  )
}
