# Stage 1: Build
FROM golang:1.23 AS builder

# Set the Current Working Directory inside the container
WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

# Copy the source code into the container
COPY . .

# Copy the Firebase service account key
COPY ./handler/cloudassignment2-test-firebase-adminsdk-fbsvc-3a8f40042b.json ./serviceAccountKey.json

# Build the Go app
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o myapp .

# Stage 2: Create a small image
FROM alpine:latest

# Set the Current Working Directory inside the container
WORKDIR /assignment_02/

# Copy the Firebase service account key
# COPY 3a8f40042bb83a8621624b98776db8f7a19b56ff /app/serviceAccountKey.json/

# Expose port 8080
EXPOSE 8080

# Command to run the executable
CMD ["./assignment_02"]