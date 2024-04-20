#!/bin/bash

# Start up Redpanda in Docker
docker-compose up -d redpanda-0

# Wait for Redpanda to be ready
echo "Waiting for Redpanda to be ready..."
sleep 10

# Create a Kafka topic with 3 partitions
docker-compose exec redpanda-0 rpk topic create test --partitions 3 --replicas 1

# Start the kafka-producer-consumer-tester locally using Make
echo "Starting tester locally..."
make run
app_exit_code=$?

# Check if the local run was successful
if [ $app_exit_code -ne 0 ]; then
    echo "Error: The tester has exited with code $app_exit_code"
fi

# Regardless of the app's exit code, shut down all Docker services
echo "Shutting down Docker services..."
docker-compose down

# Exit with the code from the local application
exit $app_exit_code
