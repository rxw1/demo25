/* eslint-disable */
import * as types from './graphql';
import { TypedDocumentNode as DocumentNode } from '@graphql-typed-document-node/core';

/**
 * Map of all GraphQL operations in the project.
 *
 * This map has several performance disadvantages:
 * 1. It is not tree-shakeable, so it will include all operations in the project.
 * 2. It is not minifiable, so the string of a GraphQL query will be multiple times inside the bundle.
 * 3. It does not support dead code elimination, so it will add unused operations.
 *
 * Therefore it is highly recommended to use the babel or swc plugin for production.
 * Learn more about it here: https://the-guild.dev/graphql/codegen/plugins/presets/preset-client#reducing-bundle-size
 */
type Documents = {
    "\n  query Orders {\n    orders {\n      id\n      productId\n      qty\n      createdAt\n    }\n  }\n": typeof types.OrdersDocument,
    "\n  subscription LastOrderCreated {\n    lastOrderCreated {\n      id\n      productId\n      qty\n      createdAt\n    }\n  }\n": typeof types.LastOrderCreatedDocument,
    "\n  query FetchProducts {\n    products {\n      id\n      name\n      price\n    }\n  }\n": typeof types.FetchProductsDocument,
    "\n  mutation CreateOrder($productId: ID!, $qty: Int!) {\n    createOrder(productId: $productId, qty: $qty) {\n      id\n      productId\n      qty\n      createdAt\n    }\n  }\n": typeof types.CreateOrderDocument,
};
const documents: Documents = {
    "\n  query Orders {\n    orders {\n      id\n      productId\n      qty\n      createdAt\n    }\n  }\n": types.OrdersDocument,
    "\n  subscription LastOrderCreated {\n    lastOrderCreated {\n      id\n      productId\n      qty\n      createdAt\n    }\n  }\n": types.LastOrderCreatedDocument,
    "\n  query FetchProducts {\n    products {\n      id\n      name\n      price\n    }\n  }\n": types.FetchProductsDocument,
    "\n  mutation CreateOrder($productId: ID!, $qty: Int!) {\n    createOrder(productId: $productId, qty: $qty) {\n      id\n      productId\n      qty\n      createdAt\n    }\n  }\n": types.CreateOrderDocument,
};

/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 *
 *
 * @example
 * ```ts
 * const query = graphql(`query GetUser($id: ID!) { user(id: $id) { name } }`);
 * ```
 *
 * The query argument is unknown!
 * Please regenerate the types.
 */
export function graphql(source: string): unknown;

/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "\n  query Orders {\n    orders {\n      id\n      productId\n      qty\n      createdAt\n    }\n  }\n"): (typeof documents)["\n  query Orders {\n    orders {\n      id\n      productId\n      qty\n      createdAt\n    }\n  }\n"];
/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "\n  subscription LastOrderCreated {\n    lastOrderCreated {\n      id\n      productId\n      qty\n      createdAt\n    }\n  }\n"): (typeof documents)["\n  subscription LastOrderCreated {\n    lastOrderCreated {\n      id\n      productId\n      qty\n      createdAt\n    }\n  }\n"];
/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "\n  query FetchProducts {\n    products {\n      id\n      name\n      price\n    }\n  }\n"): (typeof documents)["\n  query FetchProducts {\n    products {\n      id\n      name\n      price\n    }\n  }\n"];
/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "\n  mutation CreateOrder($productId: ID!, $qty: Int!) {\n    createOrder(productId: $productId, qty: $qty) {\n      id\n      productId\n      qty\n      createdAt\n    }\n  }\n"): (typeof documents)["\n  mutation CreateOrder($productId: ID!, $qty: Int!) {\n    createOrder(productId: $productId, qty: $qty) {\n      id\n      productId\n      qty\n      createdAt\n    }\n  }\n"];

export function graphql(source: string) {
  return (documents as any)[source] ?? {};
}

export type DocumentType<TDocumentNode extends DocumentNode<any, any>> = TDocumentNode extends DocumentNode<  infer TType,  any>  ? TType  : never;