<div align="center">
  <h1>Kafka Producer Consumer Tester</h1>

</div>

## Table of Contents
- [Assessment Description](#assessment-description)
- [Solution Overview](#solution-overview)
- [Getting Started](#getting-started)
  - [Prerequisites](#prerequisites)
  - [Installation](#installation)
- [Running the Tests](#running-the-tests)
- [Contributing](#contributing)
- [License](#license)

## Assessment Description

This assessment is designed as an example to evaluate the skills relevant to Kafka, event-driven architecture, and data persistence. The task is to create a system that involves:

1. **Kafka Producer:** Generates and publishes 1000 messages in batches to a Kafka topic. Each message carries a state: either *failed*, *completed*, or *in-progress*.

2. **Kafka Consumer:** Subscribes to the topic, reads the messages, and categorizes them based on their state. The messages are then stored in a persistent storage system according to their categorized state.

3. **Testing Suite:** Ensures comprehensive testing to verify that no messages are dropped during the process and all messages are correctly classified into their respective states.

The goal is to demonstrate proficiency in handling streaming data, ensuring data integrity, and implementing a testable, event-driven system.

## Solution Overview

// TODO

*This section need to detail the approach taken to solve the provided assessment, highlighting the architecture, design patterns, and technologies used. It will include explanations of the Kafka setup, the logic for message state categorization, and the persistent storage solution.*

- **Architecture Diagram:** *Include a high-level architecture diagram showcasing the producer-consumer workflow and data storage.*

- **Key Components and Technologies:** *List and briefly describe the main components and technologies used in the solution, such as Kafka, the programming language (e.g., Go), and the database for persistent storage.*

- **Design Considerations:** *Discuss any significant design decisions and how they impact the system's efficiency, reliability, and scalability.*

## Getting Started

### Prerequisites

- Docker
- Kafka (with Zookeeper)
- Go (version >=1.22)

### Installation

1. Clone the repository: `git clone (https://github.com/raw-leak/kafka-producer-consumer-tester.git`
2. Navigate to the project directory: `cd kafka-producer-consumer-tester`
3. Start Kafka and Zookeeper services using Docker: `docker-compose up -d`
4. Run the producer and consumer services: `go run producer/main.go` and `go run consumer/main.go`

## Running the Tests
// TODO

*Need to explain how to run the automated tests for this system.*

## Contributing

Please read [CONTRIBUTING.md](https://github.com/raw-leak/kafka-producer-consumer-tester/CONTRIBUTING.md) for details on our code of conduct, and the process for submitting pull requests to us.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
