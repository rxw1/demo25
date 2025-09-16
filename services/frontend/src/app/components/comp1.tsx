import { gql } from "@apollo/client"
import { useQuery, useMutation } from "@apollo/client/react"
import OrderProductButton from "./order-product-button"
import SomeLinks from "./some-links"
import Products from "./products"
import Order from "./order"

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

// type ProductByIdData = {
//   productById?: {
//     id: string
//     name: string
//     price: number
//   }
// }

export default function Comp1() {
  // const { data } = useQuery<ProductByIdData>(FetchProductByIdQuery, {
  //   variables: { id: "p1" },
  // })
  // const [createOrder] = useMutation(CreateOrderMutation)
  return (
    <div style={{ padding: 24 }}>
      {/* <OrderProductButton
          productId={data.productById.id}
          onOrderCreated={(order) => console.log("Order created:", order)}
        /> */}

      <Products />
      <Order />

      {/* <pre>{JSON.stringify(data?.productById, null, 2)}</pre> */}

      {/* {data?.productById?.id && (
        <OrderProductButton
          productId={data.productById.id}
          onOrderCreated={(order) => console.log("Order created:", order)}
        />
      )} */}

      <SomeLinks />
    </div>
  )
}
