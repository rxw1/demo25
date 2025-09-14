import { gql } from "@apollo/client"
import { useMutation } from "@apollo/client/react"

const CREATE_ORDER = gql`
  mutation CreateOrderMutation($productId: ID!, $qty: Int!) {
    createOrder(productId: $productId, qty: $qty) {
      id
      productId
      qty
      createdAt
    }
  }
`

type OrderProductButtonProps = {
  productId: string
  qty?: number
  onOrderCreated?: (order: {
    id: string
    productId: string
    qty: number
    createdAt: string
  }) => void
}

export default function OrderProductButton({
  productId,
  qty = 1,
  onOrderCreated,
}: OrderProductButtonProps) {
  type CreateOrderResult = {
    createOrder: {
      id: string
      productId: string
      qty: number
      createdAt: string
    }
  }

  const [createOrder, { loading, error }] =
    useMutation<CreateOrderResult>(CREATE_ORDER)

  const handleClick = async () => {
    try {
      const res = await createOrder({
        variables: { productId, qty },
      })
      if (onOrderCreated && res.data?.createOrder) {
        onOrderCreated(res.data.createOrder)
      }
    } catch (e) {
      // handle error if needed
    }
  }

  return (
    <button onClick={handleClick} disabled={loading}>
      {loading ? "Ordering..." : `Order ${productId}`}
      {error && <span style={{ color: "red", marginLeft: 8 }}>Error</span>}
    </button>
  )
}
