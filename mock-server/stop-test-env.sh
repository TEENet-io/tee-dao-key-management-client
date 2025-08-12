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
