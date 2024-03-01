## How to get latest Airstack GraphQL schema

1. `npm install -g get-graphql-schema`
2. `get-graphql-schema https://api.airstack.xyz/gql > new-schema.graphql`

## How to generate code

1. Add your query under /queries
2. Modify genqlient.yaml operations to include all your query files
3. Run `go generate ./...`

