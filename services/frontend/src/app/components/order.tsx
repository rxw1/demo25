import { gql } from "@apollo/client";
import { useQuery } from "@apollo/client/react";
import { OrdersDocument, Order } from "../__generated__/graphql";

const Q = gql`
  query Orders {
    orders {
      id
      productId
      qty
      createdAt
    }
  }
`

type Data = {
  orders: Order[]
}

export default function Orders({}) {
  const { data } = useQuery<Data>(OrdersDocument, {})

  return (
    <div style={{ width: "var(--width)" }}>
      <h3>Orders ({data?.orders.length})</h3>
      {data?.orders.map((p) => (
        <div
          key={p.id}
          style={{
            display: "grid",
            gridTemplateColumns: "1fr 1fr 1fr auto",
          }}
        >
          <div>{p.id.slice(9, 14)}</div>
          <div>{p.productId.slice(9, 14)}</div>
          <div>{p.qty}</div>
          <div>{new Date(p.createdAt).toLocaleString()}</div>
        </div>
      ))}
    </div>
  )
}
