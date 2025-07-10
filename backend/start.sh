#!/bin/bash

# Define the binary name and output log file
BINARY_NAME="cf-manager-backend"
LOG_FILE="backend.log"
SOURCE_PATH="main.go"

echo "Building the $BINARY_NAME binary in the backend directory..."
# Build the Go binary in the backend directory
go build -o $BINARY_NAME $SOURCE_PATH

# Check if build was successful
if [ $? -ne 0 ]; then
    echo "Error: Build failed in backend directory. Exiting."
    exit 1
fi

echo "Checking for existing $BINARY_NAME processes..."
# Find the PID of the running process, excluding the grep process itself
PID=$(pgrep $BINARY_NAME)

if [ ! -z "$PID" ]; then
    echo "Found running process with PID: $PID. Killing it..."
    # Kill the process
    kill $PID
    # Wait a bit for the process to terminate gracefully
    sleep 2
    # Optional: check if the process is still running and force kill if necessary
    if ps -p $PID > /dev/null; then
        echo "Process $PID did not stop gracefully, forcing kill..."
        kill -9 $PID
        sleep 1
    fi
else
    echo "No existing $BINARY_NAME process found."
fi

echo "Starting $BINARY_NAME from the backend directory with nohup, output redirected to $LOG_FILE..."
# Start the binary using nohup from the backend directory
nohup ./$BINARY_NAME > $LOG_FILE 2>&1 &

echo "$BINARY_NAME started in the background from backend directory."
echo "You can check the output in backend/$LOG_FILE or use 'tail -f backend/$LOG_FILE'."
echo "Frontend is served at http://localhost:3000" 