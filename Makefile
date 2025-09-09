.PHONY: build run-server run-client clean

IMAGE_NAME := ghcr.io/donileongdeepernetwork/k8s-graceful-shutdown-test

# Build the Docker image locally
build:
	docker build -t $(IMAGE_NAME) .

# Run the server in Docker (pulls from registry if not built locally)
run-server:
	docker run --rm -p 8080:8080 $(IMAGE_NAME)

# Run the client in Docker (assuming server is running)
run-client:
	docker run --rm -e SERVER_URL=ws://host.docker.internal:8080/ws $(IMAGE_NAME) /app/client

# Clean up Docker images
clean:
	docker rmi $(IMAGE_NAME) || true
