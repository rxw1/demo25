#!/usr/bin/env zsh

q=${1:-'query {products{id}}'}

set +x
curl -s --request POST \
  --header 'content-type: application/json' \
  --url http://localhost:4000/ \
  --data '{"query":'${(qqq)q}'}'
set -x
