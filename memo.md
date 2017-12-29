cat list | perl assemble_ghapi_url.pl | xargs -I{} sh -c 'curl -H "Authorization: token $TOKEN" -s '{}' | jq -c -M .' > contents

# port
web: 8080
indexer: 8081
renderer: 8082
gaelv: 9090
