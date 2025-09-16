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
    <div>
      <h3>Orders ({data?.orders.length})</h3>
      {data?.orders.map((p) => (
        <div key={p.id} style={{ marginBottom: 12 }}>
          <div>Order ID: {p.id}</div>
          <div>Product ID: {p.productId}</div>
          <div>Quantity: {p.qty}</div>
          <div>Created At: {new Date(p.createdAt).toLocaleString()}</div>
          {/* <div>Price: {p.price}</div> */}
        </div>
      ))}
    </div>
  )
}
