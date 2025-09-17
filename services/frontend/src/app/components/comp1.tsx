import { useState } from "react";
import { Order } from "../__generated__/graphql";
import { graphQLEndpoint } from "../page";
import Orders from "./orders";
import Products from "./products";

export default function Comp1() {
  const [orders, setOrders] = useState<Order[]>([])
  const playgroundUrl = graphQLEndpoint.replace("/graphql", "/")

  return (
    <div
      style={{
        padding: "var(--padding)",
      }}
    >
      <header>demo 2025/9/11</header>
      <main>
        <Products />
        <Orders />
      </main>
      <footer
        style={{
          display: "flex",
          flexDirection: "column",
          fontSize: "var(--font-size-sm)",
          color: "var(--ink-40)",
          gap: "var(--gap)",
          marginTop: "var(--margin)",
          paddingTop: "var(--padding)",
          borderTop: "1px solid var(--ink-20)",
        }}
      >
        <div>
          <strong>GraphQL Playground: </strong>
          <a href={playgroundUrl}>{playgroundUrl}</a>
        </div>
      </footer>
    </div>
  )
}
