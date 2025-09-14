import { gql } from "@apollo/client"
import { useQuery, useMutation } from "@apollo/client/react"
import OrderProductButton from "./order-product-button"

const FetchProductsQuery = gql`
  query FetchProducts {
    products {
      id
      name
      price
    }
  }
`

const FetchProductByIdQuery = gql`
  query FetchProductById($id: ID!) {
    productById(productId: $id) {
      id
      name
      price
    }
  }
`
const CreateOrderMutation = gql`
  mutation CreateOrderMutation($productId: ID!, $qty: Int!) {
    createOrder(productId: $productId, qty: $qty) {
      id
      productId
      qty
      createdAt
    }
  }
`

type ProductByIdData = {
  productById?: {
    id: string
    name: string
    price: number
  }
}

export default function Comp1() {
  const { data } = useQuery<ProductByIdData>(FetchProductByIdQuery, {
    variables: { id: "p1" },
  })
  const [createOrder] = useMutation(CreateOrderMutation)
  return (
    <div style={{ padding: 24 }}>
      <pre>{JSON.stringify(data?.productById, null, 2)}</pre>

      <h3>Order Product</h3>

      {data?.productById?.id && (
        <OrderProductButton
          productId={data.productById.id}
          onOrderCreated={(order) => console.log("Order created:", order)}
        />
      )}

      <h3>Links</h3>
      <ul>
        <li>
          <strong>GraphQL Playground:</strong>
          <a href="http://localhost:8080/">http://localhost:8080/</a>
        </li>
        <li>
          <strong>Postgres URL:</strong>
          <a href="postgres://app:app@postgres:5432/app?sslmode=disable">
            postgres://app:app@postgres:5432/app?sslmode=disable
          </a>
        </li>
        <li>
          <strong>Product Service:</strong>
          <a href="http://localhost:8081">http://localhost:8081</a>
        </li>
      </ul>
    </div>
  )
}
