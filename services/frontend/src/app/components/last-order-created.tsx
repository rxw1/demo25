import { gql } from "@apollo/client"
import { useSubscription } from "@apollo/client/react"
import { useEffect, useState } from "react"
import { last } from "rxjs"

const S = gql`
  subscription LastOrderCreated {
    lastOrderCreated {
      id
      productId
      qty
      createdAt
    }
  }
`

type Order = {
  id: string
  productId: string
  qty: number
  createdAt: string
}

type Data = {
  lastOrderCreated: Order
}

const useLocalStorage = (key: string, fallback: string) => {
  const [value, setValue] = useState(
    JSON.parse(localStorage.getItem(key) ?? fallback)
  )

  useEffect(() => {
    localStorage.setItem(key, JSON.stringify(value))
  }, [value, key])

  return [value, setValue]
}

export default function LastOrderCreated() {
  const { data, loading } = useSubscription<Data>(S, {
    // variables: { orderID }
  })

  const [orders, setOrders] = useState<Order[]>([])

  useEffect(() => {
    if (typeof window !== "undefined") {
      const storedOrders = localStorage.getItem("orders")
      if (storedOrders) {
        setOrders(JSON.parse(storedOrders))
      }
    }
  }, [])

  useEffect(() => {
    localStorage.setItem("orders", JSON.stringify(orders))
  }, [orders])

  if (data?.lastOrderCreated) {
    const exists = orders.find((o) => o.id === data.lastOrderCreated.id)
    if (!exists) {
      setOrders((prev) => [data.lastOrderCreated, ...prev].slice(0, 10))
    }

    localStorage.setItem("orders", JSON.stringify(orders))
  }

  return (
    <div>
      <h4>Last Order Created</h4>
      <div>
        {loading && <div>Loading...</div>}
        {(!loading && data && (
          <div
            style={{
              fontFamily: "monospace",
              fontSize: 12,
              width: 200,
              marginTop: 10,
            }}
          >
            <div>
              <strong>orderID:</strong>
              {data.lastOrderCreated.id}
            </div>

            <div>
              <strong>productID:</strong>
              {data.lastOrderCreated.productId}
            </div>

            <div>
              <strong>qty:</strong>
              {data.lastOrderCreated.qty}
            </div>

            <div>
              <strong>createdAt</strong>
              {data.lastOrderCreated.createdAt}
            </div>
          </div>
        )) || <div>No data</div>}
      </div>
      <div>
        <h4>All received orders</h4>
        <div
          style={{
            fontFamily: "monospace",
            fontSize: 12,
            width: 480,
            marginTop: 10,
          }}
        >
          {orders.map((o) => (
            <div
              key={o.id}
              style={{
                display: "grid",
                gridTemplateColumns: "1fr 1fr 1fr auto",
              }}
            >
              <div>{o.id.slice(9, 14)}</div>
              <div>{o.productId.slice(9, 14)}</div>
              <div>{o.qty}</div>
              <div>{new Date(o.createdAt).toLocaleString()}</div>
              {/* <div>Price: {o.price}</div> */}
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}
