"use client"
import { ApolloClient, HttpLink, InMemoryCache, gql } from "@apollo/client"
import { ApolloProvider } from "@apollo/client/react"
import { useQuery, useMutation } from "@apollo/client/react"
import Comp1 from "./components/comp1"

const client = new ApolloClient({
  link: new HttpLink({
    uri: process.env.NEXT_PUBLIC_GRAPHQL_URL || "http://localhost:8080/graphql",
  }),
  cache: new InMemoryCache(),
})

export default function Home() {
  return (
    <ApolloProvider client={client}>
      <Comp1 />
    </ApolloProvider>
  )
}
