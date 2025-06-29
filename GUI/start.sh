#!/bin/bash

# Define the binary name and output log file
BINARY_NAME="cf-manager"
LOG_FILE="output.log"
SOURCE_PATH="main.go" # Source path relative to the current directory (GUI)

echo "Building the $BINARY_NAME binary in the current directory (./GUI)..."
# Build the Go binary in the current directory (./GUI)
go build -o $BINARY_NAME $SOURCE_PATH

# Check if build was successful
if [ $? -ne 0 ]; then
    echo "Error: Build failed in ./GUI. Exiting."
    exit 1
fi

echo "Checking for existing $BINARY_NAME processes..."
# Find the PID of the running process, excluding the grep process itself
# Using pgrep is generally more reliable than parsing ps output
# Note: pgrep searches by process name, not directory, which is fine here.
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

echo "Starting $BINARY_NAME from the current directory (./GUI) with nohup, output redirected to $LOG_FILE..."
# Start the binary using nohup from the current directory (./GUI)
nohup ./$BINARY_NAME > $LOG_FILE 2>&1 &

echo "$BINARY_NAME started in the background from ./GUI."
echo "You can check the output in ./GUI/$LOG_FILE or use 'tail -f ./GUI/$LOG_FILE'."