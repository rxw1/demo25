import { Order } from "../__generated__/graphql";

const shortId = (id: string) => id.slice(6, 9 + 2)

export default function AllKnownOrders({ orders }: { orders: Order[] }) {
  return (
    <div>
      <h4>Last Orders</h4>
<div>
            <button onClick={() => localStorage.clear()}>Clear LocalStorage</button></div>
      <div style={{}}>
        {orders.map((o) => (
          <div
            key={o.id}
            style={{
              display: "grid",
              gridTemplateColumns: "auto auto auto 1fr",
              gap: 10,
            }}
          >
            <div title={o.id}>{shortId(o.id)}</div>
            <div title={o.productId}>{shortId(o.productId)}</div>
            <div>{o.qty}</div>
            <div>{new Date(o.createdAt).toLocaleString()}</div>
          </div>
        ))}
      </div>
    </div>
  )
}
