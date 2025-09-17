import { gql } from "@apollo/client";
import { useSubscription } from "@apollo/client/react";
import { useEffect, useState } from "react";
import { LastOrderCreatedDocument, Order } from "../__generated__/graphql";
import AllKnownOrders from "./all-known-orders";

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

const useLocalStorage = (key: string, fallback: string) => {
  const [value, setValue] = useState(
    JSON.parse(localStorage.getItem(key) ?? fallback)
  )

  useEffect(() => {
    localStorage.setItem(key, JSON.stringify(value))
  }, [value, key])

  return [value, setValue]
}

const shortId = (id: string) => id.slice(6, 9 + 2)

export type Data = {
  lastOrderCreated: Order
}

export default function Orders() {
  const { data, loading } = useSubscription<Data>(LastOrderCreatedDocument, {
    // variables: { orderID }
  })

  const [orders, setOrders] = useState<Order[]>([])

  useEffect(() => {
    if (typeof window !== "undefined") {
      const storedOrders = localStorage.getItem("orders")
      if (storedOrders) {
        try {
          setOrders(JSON.parse(storedOrders))
        } catch {
          setOrders([])
        }
      }
    }
  }, [])

  // persist orders whenever they change
  useEffect(() => {
    if (typeof window !== "undefined") {
      localStorage.setItem("orders", JSON.stringify(orders))
    }
  }, [orders])

  useEffect(() => {
    if (!data?.lastOrderCreated) return
    setOrders((prev) => {
      const exists = prev.find((o) => o.id === data.lastOrderCreated.id)
      if (exists) return prev
      const newOrders = [data.lastOrderCreated, ...prev].slice(0, 7)
      return newOrders
    })
  }, [data])

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
              width: "var(--width)",
              marginTop: 10,
              display: "flex",
              flexDirection: "column",
              gap: 4,
            }}
          >
            <div>
              <strong>orderID:</strong>
              <div title={data.lastOrderCreated.id}>
                {shortId(data.lastOrderCreated.id)}
              </div>
            </div>

            <div>
              <strong>productID:</strong>
              <div title={data.lastOrderCreated.productId}>
                {shortId(data.lastOrderCreated.productId)}
              </div>
            </div>

            <div>
              <strong>qty:</strong>
              {data.lastOrderCreated.qty}
            </div>

            <div>
              <strong>createdAt:</strong>
              <div title={data.lastOrderCreated.createdAt}>
                {new Date(data.lastOrderCreated.createdAt).toLocaleString()}
              </div>
            </div>
          </div>
        )) || <div>No data</div>}
      </div>
      <AllKnownOrders orders={orders} />
    </div>
  )
}
