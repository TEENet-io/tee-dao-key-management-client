.PHONY: proto certs build run test clean start

# Generate protobuf code
proto:
	@echo "Generating protobuf code..."
	@command -v protoc >/dev/null 2>&1 || { echo "Error: protoc not found. Install with: brew install protobuf"; exit 1; }
	@command -v protoc-gen-go >/dev/null 2>&1 || go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	@command -v protoc-gen-go-grpc >/dev/null 2>&1 || go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	@protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/user_task.proto proto/node_management.proto proto/appid_service.proto
	@echo "Proto generation completed!"

# Generate certificates
certs:
	@echo "Generating certificates..."
	@chmod +x generate-certs.sh
	@./generate-certs.sh

# Full build (used by start-test-env.sh)
build: build-dao build-config build-app build-example
	@echo "All components built successfully!"

# Build individual components
build-dao: proto
	@echo "Building mock DAO server..."
	@go build -o dao-server dao-server.go

build-config: proto
	@echo "Building config server..."
	@go build -o config-server mock-config-server.go

build-app: proto
	@echo "Building app node..."
	@go build -o app-node mock-app-node.go

build-example: proto
	@echo "Building example program..."
	@go build -o example-program example-user-program.go

# Run the server
run: build-dao certs
	@echo "Starting DAO server..."
	@./dao-server

# Run config server
run-config: build-config certs
	@echo "Starting config server..."
	@./config-server

# Run app node
run-app: build-app certs
	@echo "Starting app node..."
	@./app-node

# Run all servers
run-all: build certs
	@echo "Starting all servers..."
	@./config-server &
	@sleep 1
	@./app-node &
	@sleep 1
	@./dao-server

# Run example program
example: build-example
	@echo "Running example program..."
	@./example-program

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -f dao-server config-server app-node example-program
	@rm -rf certs/ logs/
	@rm -f proto/*.pb.go


# Quick start
start: build certs
	@echo "Starting test environment..."
	@./start-test-env.sh

# Full test
test: start
	@echo "Running example..."
	@sleep 3
	@./example-program