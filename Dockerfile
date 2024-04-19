FROM golang:1.22.2-alpine3.19 as builder

RUN apk add --no-cache make

# Set the Current Working Directory inside the container
WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

# Build the application using the Makefile.
RUN make build

FROM golang:1.22.2-alpine3.19

WORKDIR /root/

COPY --from=builder /app/build/kafka-producer-consumer-tester .

# Command to run the executable
CMD ["./kafka-producer-consumer-tester"]
