#!/bin/bash
set -e

echo "======================================="
echo "   TEE DAO Local Test Environment"
echo "======================================="

# Check dependencies
check_dependencies() {
    echo "Checking system dependencies..."
    
    if ! command -v go &> /dev/null; then
        echo "Error: Go language environment not found"
        echo "Please install Go 1.22 or higher: https://golang.org/dl/"
        exit 1
    fi
    
    if ! command -v protoc &> /dev/null; then
        echo "Error: protoc compiler not found"
        echo "On macOS run: brew install protobuf"
        echo "On Ubuntu run: sudo apt-get install protobuf-compiler"
        exit 1
    fi
    
    if ! command -v openssl &> /dev/null; then
        echo "Error: openssl not found"
        echo "On macOS run: brew install openssl"
        exit 1
    fi
    
    echo "✓ All dependencies check passed"
}

# Setup environment
setup_environment() {
    echo "Setting up test environment..."
    
    # Generate protobuf code
    echo "Generating protobuf code..."
    make proto
    
    # Generate certificates
    echo "Generating TLS certificates..."
    make certs
    
    # Build all components
    echo "Building project components..."
    make build
    
    echo "✓ Environment setup completed"
}

# Start services
start_services() {
    echo "Starting services..."
    echo "  Config server port: 50052"
    echo "  DAO server port: 50051"
    echo "  App node port: 50053"
    
    # Create logs directory
    mkdir -p logs
    
    # Start config server
    echo "Starting config server..."
    ./config-server > logs/config-server.log 2>&1 &
    CONFIG_PID=$!
    echo "  Config server PID: $CONFIG_PID"
    
    # Wait for config server to start
    sleep 2
    
    # Start app node
    echo "Starting app node..."
    ./app-node > logs/app-node.log 2>&1 &
    APP_PID=$!
    echo "  App node PID: $APP_PID"
    
    # Wait for app node to start
    sleep 2
    
    # Start DAO server
    echo "Starting DAO server..."
    ./dao-server > logs/dao-server.log 2>&1 &
    DAO_PID=$!
    echo "  DAO server PID: $DAO_PID"
    
    # Wait for DAO server to start
    sleep 2
    
    # Save PIDs to files
    echo $CONFIG_PID > logs/config-server.pid
    echo $APP_PID > logs/app-node.pid
    echo $DAO_PID > logs/dao-server.pid
    
    echo "✓ All services started successfully"
}

# Run tests
run_tests() {
    echo "Running built-in tests..."
    ./example-program
    echo "✓ Tests completed"
}

# Show status
show_status() {
    echo ""
    echo "======================================="
    echo "   Test Environment Ready!"
    echo "======================================="
    echo ""
    echo "Service Status:"
    echo "  Config Server: localhost:50052 (PID: $(cat logs/config-server.pid 2>/dev/null || echo 'N/A'))"
    echo "  DAO Server:    localhost:50051 (PID: $(cat logs/dao-server.pid 2>/dev/null || echo 'N/A'))"
    echo "  App Node:      localhost:50053 (PID: $(cat logs/app-node.pid 2>/dev/null || echo 'N/A'))"
    echo ""
    
    # Display available App IDs from log file
    echo "Available App IDs for testing:"
    if [ -f logs/app-node.log ]; then
        grep -A 20 "Available App IDs for testing:" logs/app-node.log | grep "  -" | head -10
        echo ""
        echo "💡 Usage Tips:"
        echo "   Copy any of the above App IDs to use in your client programs"
        echo "   Each App ID corresponds to different signature protocol and curve combinations"
    else
        echo "  (App IDs will be available once App node starts successfully)"
    fi
    echo ""
    
    echo "Log Files:"
    echo "  Config server logs: logs/config-server.log"
    echo "  DAO server logs:    logs/dao-server.log"
    echo "  App node logs:      logs/app-node.log"
    echo ""
    echo "Usage:"
    echo "  1. Your program should connect to config server: localhost:50052"
    echo "  2. View logs: tail -f logs/*.log"
    echo "  3. Run example program: see example-user-program.go"
    echo "  4. Stop services: ./stop-test-env.sh"
    echo ""
}

# Create stop script
create_stop_script() {
    cat > stop-test-env.sh << 'EOF'
#!/bin/bash
echo "Stopping test environment..."

if [ -f logs/config-server.pid ]; then
    CONFIG_PID=$(cat logs/config-server.pid)
    if kill -0 $CONFIG_PID 2>/dev/null; then
        kill $CONFIG_PID
        echo "✓ Config server stopped (PID: $CONFIG_PID)"
    fi
    rm -f logs/config-server.pid
fi

if [ -f logs/app-node.pid ]; then
    APP_PID=$(cat logs/app-node.pid)
    if kill -0 $APP_PID 2>/dev/null; then
        kill $APP_PID
        echo "✓ App node stopped (PID: $APP_PID)"
    fi
    rm -f logs/app-node.pid
fi

if [ -f logs/dao-server.pid ]; then
    DAO_PID=$(cat logs/dao-server.pid)
    if kill -0 $DAO_PID 2>/dev/null; then
        kill $DAO_PID
        echo "✓ DAO server stopped (PID: $DAO_PID)"
    fi
    rm -f logs/dao-server.pid
fi

echo "Test environment stopped"
EOF
    chmod +x stop-test-env.sh
}

# Main function
main() {
    # Check if running in correct directory
    if [ ! -f "dao-server.go" ] || [ ! -f "mock-config-server.go" ]; then
        echo "Error: Please run this script in the tee-dao-mock-server directory"
        exit 1
    fi
    
    check_dependencies
    setup_environment
    create_stop_script
    start_services
    
    # Run tests (optional)
    if [ "$1" = "--with-test" ]; then
        run_tests
    fi
    
    show_status
}

# Handle interrupt signals
cleanup() {
    echo ""
    echo "Received interrupt signal, cleaning up environment..."
    if [ -f stop-test-env.sh ]; then
        ./stop-test-env.sh
    fi
    exit 0
}

trap cleanup INT TERM

main "$@"