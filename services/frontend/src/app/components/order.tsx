import { gql } from "@apollo/client"
import { useQuery } from "@apollo/client/react"

const Q = gql`
  query FetchOrders {
    orders {
      id
      productId
      qty
      createdAt
      # price
    }
  }
`

type Order = {
  id: string
  productId: string
  qty: number
  createdAt: string
  // price: number
}

type Orders = {
  orders: Order[]
}

export default function Products() {
  const { data } = useQuery<Orders>(Q, {})

  return (
    <div style={{ width: 480 }}>
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
          {/* <div>Price: {p.price}</div> */}
        </div>
      ))}
    </div>
  )
}
