import { useState } from "react";
import {
  type Product
} from "../__generated__/graphql";

export default function Products({
  product,
  onOrder,
}: {
  product: Product
  onOrder: any
}) {
  const [qty, setQty] = useState<number>(1)

  const increment = () => setQty((q) => q + 1)
  const decrement = () => setQty((q) => (q > 1 ? q - 1 : 1))

  return (
    <div
      key={product.id}
      style={{
        width: "var(--width)",
        display: "grid",
        gridTemplateColumns: "auto 4em auto auto auto auto auto auto 4em",
        gap: 8,
      }}
    >
      <div>{product.id.slice(9, 14)}</div>
      <div>{product.name}</div>
      <div>${product.price}</div>
      <button
        onClick={(e) => {
          onOrder(product.id, product.price, qty)
        }}
      >
        ORDER
      </button>
      <button
        onClick={(e) => {
          const qty = Math.floor(Math.random() * 100) + 1
          setQty(qty)
        }}
      >
        RND
      </button>
      <button
        onClick={(e) => {
          setQty(1)
        }}
      >
        RST
      </button>
      <button
        onClick={(e) => {
          decrement()
        }}
      >
        -
      </button>
      <button
        onClick={(e) => {
          increment()
        }}
      >+</button>
      {qty}
    </div>
  )
}
