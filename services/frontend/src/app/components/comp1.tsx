import { gql } from "@apollo/client"
import { useQuery, useMutation } from "@apollo/client/react"
import OrderProductButton from "./order-product-button"
import SomeLinks from "./some-links"
import Products from "./products"
import Order from "./order"
import LastOrderCreated from './last-order-created';

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

export default function Comp1() {
  return (
    <div style={{ padding: 24 }}>
      <Products />
      <LastOrderCreated />
      {/* <Order /> */}
      <SomeLinks />
    </div>
  )
}
