param (
    $command
)

if (-not $command)  {
    $command = "start"
}

$ProjectRoot = "${PSScriptRoot}/.."

$env:API_ENVIRONMENT="Development"
$env:API_PORT="8080"
$env:API_MONGODB_USERNAME="root"
$env:API_MONGODB_PASSWORD="neUhaDnes"

function mongo {
    docker compose --file ${ProjectRoot}/deployments/docker-compose/compose.yaml $args
}

switch ($command) {
    "start" {
        try {
            mongo up --detach
            go run ${ProjectRoot}/cmd/sprava-krvi-api-service
            mongo down
        }
    }
    "openapi" {
        docker run --rm -ti -v ${ProjectRoot}:/local openapitools/openapi-generator-cli generate -c /local/scripts/generator-cfg.yaml
    }
    "mongo" {
        mongo up
    }
    default {
        throw "Unknown command: $command"
    }
}
