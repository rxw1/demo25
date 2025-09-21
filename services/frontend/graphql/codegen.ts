import type { CodegenConfig } from '@graphql-codegen/cli'

// Allow overriding schema URL via env var at codegen time
const schemaUrl = process.env.NEXT_PUBLIC_GRAPHQL_URL || 'http://localhost:8080/graphql'

const config: CodegenConfig = {
  schema: schemaUrl,
  documents: ['./src/app/**/*.{ts,tsx}'],
  generates: {
    './src/app/__generated__/': {
      preset: 'client',
    },
  },
}
export default config
