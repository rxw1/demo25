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

export default function Products() {
  const { data } = useQuery<Products>(Q, {})
  const [createOrder] = useMutation<OrderResult>(M)

  function handleOrder(productId: string, price: number, qty: number) {
    createOrder({ variables: { productId, qty } })
      .then((response) => {
        console.log("Order created:", response.data?.createOrder)
      })
      .catch((error) => {
        console.error("Error creating order:", error)
      })
  }

  return (
    <div>
      <h3>Products</h3>
      <div
        style={{
          display: "flex",
          flexDirection: "row",
          gap: 16,
        }}
      >
        {data?.products.map((p) => (
          <div key={p.id} style={{}}>
            {p.id} {p.name} ${p.price}{" "}
            <button
              onClick={(e) => {
                const input = e.currentTarget
                  .nextElementSibling as HTMLInputElement | null
                const qty = input ? parseInt(input.value || "0", 10) : 1
                handleOrder(p.id, p.price, qty)
              }}
            >
              ORDER
            </button>
            <input type="number" defaultValue={1} style={{ width: 40 }} />
          </div>
        ))}
      </div>
    </div>
  )
}
