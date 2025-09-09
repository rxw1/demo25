'use client';
import { ApolloClient, HttpLink, InMemoryCache, gql } from "@apollo/client";
import { ApolloProvider } from "@apollo/client/react";
import { useQuery, useMutation } from "@apollo/client/react";

const client = new ApolloClient({
  link: new HttpLink({ uri: process.env.NEXT_PUBLIC_GRAPHQL_URL || 'http://localhost:8080/graphql' }),
  cache: new InMemoryCache(),
});

//client
//  .query({
//    query: gql`
//      query GetLocations {
//        locations {
//          id
//          name
//          description
//          photo
//        }
//      }
//    `,
//  })
//  .then((result) => console.log(result));

const Q = gql`query($id: ID!){ productById(id:$id){ id name price } }`;
const M = gql`mutation($productId:ID!,$qty:Int!){ createOrder(productId:$productId, qty:$qty){ id productId qty createdAt } }`;

function PageInner() {
  const { data } = useQuery(Q, { variables: { id: 'p1' } });
  const [createOrder] = useMutation(M);
  return (
    <div style={{ padding: 24 }}>
      <pre>{JSON.stringify(data?.productById, null, 2)}</pre>

      <button onClick={
        () => createOrder({
          variables: {
            productId: 'p1',
            qty: 1
          }
        })
      }>Create Order</button>
    </div>
  );
}

export default function Home() {
  return <ApolloProvider client={client}>
    <PageInner />
  </ApolloProvider>;
}

//export default function Page() {
//  return (
//    <ApolloProvider client={client}>
//      <div style={{ padding: 24 }}>
//        <h1>jfpoc</h1>
//        <p>Open the console to see the fetched locations data.</p>
//      </div>
//    </ApolloProvider>
//  );
//}

// export default Page;


//export default function Page() {
//  return <h1>Hello Next.js!</h1>
//}
