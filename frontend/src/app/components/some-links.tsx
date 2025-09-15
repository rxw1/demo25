export default function Links() {
  return (
    <div style={{ padding: 0 }}>
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
