cat list | perl assemble_ghapi_url.pl | xargs -I{} sh -c 'curl -H "Authorization: token $TOKEN" -s '{}' | jq -c -M .' > contents
