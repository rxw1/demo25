const { ApolloServer } = require('apollo-server');
const { ApolloGateway, IntrospectAndCompose } = require("@apollo/gateway");

const gateway = new ApolloGateway({
    supergraphSdl: new IntrospectAndCompose({
        subgraphs: [
            { name: 'products', url: 'http://localhost:8080/graphql' },
            { name: 'orders', url: 'http://localhost:8081/graphql' },
        ]
    })
});

const server = new ApolloServer({
    gateway,
    // subscriptions: true,
});

server.listen().then(({ url }) => {
    console.log(`🚀 Server ready at ${url}`);
});
