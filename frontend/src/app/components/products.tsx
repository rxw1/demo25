import { gql } from "@apollo/client"
import { useMutation, useQuery } from "@apollo/client/react"
import { Hahmlet } from "next/font/google"

const Q = gql`
  query FetchProducts {
    products {
      id
      name
      price
    }
  }
`

const M = gql`
  mutation CreateOrderMutation($productId: ID!, $qty: Int!) {
    createOrder(productId: $productId, qty: $qty) {
      id
      productId
      qty
      createdAt
    }
  }
`

type Product = {
  id: string
  name: string
  price: number
}

type Products = {
  products: Product[]
}

type OrderResult = {
  createOrder: {
    id: string
    productId: string
    qty: number
    createdAt: string
  }
}

export default function Produts() {
  const { data } = useQuery<Products>(Q, {})
  const [createOrder] = useMutation<OrderResult>(M)

  function handleOrder(productId: string, price: number) {
    createOrder({ variables: { productId, qty: 1 } })
      .then((response) => {
        console.log("Order created:", response.data?.createOrder)
      })
      .catch((error) => {
        console.error("Error creating order:", error)
      })
  }

  return (
    <div
      style={{
        display: "flex",
        flexDirection: "row",
        gap: 16,
        padding: 24,
      }}
    >
      {data?.products.map((p) => (
        <div key={p.id} style={{}}>
          {p.id} {p.name} ${p.price}{" "}
          <button onClick={() => handleOrder(p.id, p.price)}>ORDER</button>
        </div>
      ))}
    </div>
  )
}
