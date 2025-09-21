#!/usr/bin/env zsh

port=8080
url=http://localhost:${port}/graphql
query=${1:-'query {products{id}}'}

set +x
curl --request POST \
  --header 'content-type: application/json' \
  --url $url \
  --data '{"query":'${(qqq)query}'}'
set -x
