import type { CodegenConfig } from '@graphql-codegen/cli'

const config: CodegenConfig = {
  schema: 'http://localhost:8080/graphql',
  documents: ['./src/app/**/*.{ts,tsx}'],
  generates: {
    './src/app/__generated__/': {
      preset: 'client',
    },
  },
}
export default config
