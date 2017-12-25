cat list | head -n 2 | perl assemble_ghapi_url.pl | xargs -I{} sh -c "curl -s {} | jq -c -M ." > contents
