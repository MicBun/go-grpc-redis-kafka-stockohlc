# Run this command from root project directory
# sh scripts/mockery.sh

# Create mock for all interfaces
docker container run --rm -v "$PWD":/src -v $(go env GOCACHE):/cache/go -e GOCACHE=/cache/go -v ${GOPATH}/pkg:/go/pkg -w /src vektra/mockery:v2.35.0
