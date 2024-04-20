#!/bin/bash

# Start up the Redpanda 
docker-compose up -d redpanda-0

# Wait for Redpanda to be ready
echo "Waiting for Redpanda to be ready..."
sleep 10  

# Create a Kafka topic with 3 partitions
docker-compose exec redpanda-0 rpk topic create test --partitions 3 --replicas 1

# Start the kafka-producer-consumer-tester
echo "Starting tester..."
docker-compose up kafka-producer-consumer-tester  -d
docker attach kafka-producer-consumer-tester-1
app_exit_code=$?


if [ $app_exit_code -ne 0 ]; then
    echo "Error: The tester has exited with code $app_exit_code"

    # Optional: Fetch and print the latest logs from the application for debugging
    # Assuming the service name in docker-compose for your app is kafka-producer-consumer-tester
    echo "Fetching logs for error details..."
    docker-compose logs kafka-producer-consumer-tester
fi


# Regardless of the app's exit code, shut down all services
echo "Shutting down..."
docker-compose down

# Exit with the code from the application
exit $app_exit_code
