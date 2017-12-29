cat list | perl assemble_ghapi_url.pl | xargs -I{} sh -c 'curl -H "Authorization: token $TOKEN" -s '{}' | jq -c -M .' > contents

# port
web: 8080
indexer: 8081
renderer: 8082
gaelv: 9090

# run
dev_appserver.py app.yaml --port=8080 --logs_path=/tmp/log.db --env_var GITHUB_API_TOKEN=xxx --env_var RENDERER_SCHEME=http --env_var RENDERER_HOST=localhost --env_var RENDERER_PORT=8082
