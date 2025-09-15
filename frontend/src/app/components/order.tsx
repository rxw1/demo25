import { gql } from "@apollo/client"
import { useQuery } from "@apollo/client/react"

const Q = gql`
  query FetchOrders {
    orders {
      id
      name
      price
    }
  }
`

type Order = {
  id: string
  name: string
  price: number
}

type Orders = {
  products: Order[]
}

export default function Produts() {
  const { data } = useQuery<Orders>(Q, {})

  return (
    <div>
      {data?.products.map((p) => (
        <div key={p.id} style={{ marginBottom: 12 }}>
          <button onClick={() => alert(`Order ${p.name} for ${p.price}?`)}>
            {p.name} {p.price}
          </button>
        </div>
      ))}
    </div>
  )
}
