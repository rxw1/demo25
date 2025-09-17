import { gql } from "@apollo/client"
import { useMutation, useQuery } from "@apollo/client/react"
import {
  CreateOrderDocument,
  CreateOrderMutation,
  CreateOrderMutationVariables,
  FetchProductsDocument,
  FetchProductsQuery,
} from "../__generated__/graphql"
import ProductComponent from "./product"

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
  mutation CreateOrder($productId: ID!, $qty: Int!) {
    createOrder(productId: $productId, qty: $qty) {
      id
      productId
      qty
      createdAt
    }
  }
`

export default function Products() {
  const { data, loading } = useQuery<FetchProductsQuery>(FetchProductsDocument)
  const [createOrder] = useMutation<
    CreateOrderMutation,
    CreateOrderMutationVariables
  >(CreateOrderDocument)

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
      <h4>Products</h4>
      <div
        style={{
          width: "var(--width)",
        }}
      >
        {!loading &&
          data?.products.map((p) => (
            <ProductComponent key={p.id} product={p} onOrder={handleOrder} />
          ))}
      </div>
    </div>
  )
}
