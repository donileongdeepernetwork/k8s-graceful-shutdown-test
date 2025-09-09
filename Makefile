.PHONY: build run-server run-client clean k8s-apply k8s-delete k8s-logs-server k8s-logs-client k8s-restart k8s-delete-pod k8s-list-pods

IMAGE_NAME := ghcr.io/donileongdeepernetwork/k8s-graceful-shutdown-test
NAMESPACE := k8s-graceful-shutdown-test

# Build the Docker image locally
build:
	docker build -t $(IMAGE_NAME) .

# Run the server in Docker (pulls from registry if not built locally)
run-server:
	docker run --rm -p 8080:8080 $(IMAGE_NAME)

# Run the client in Docker (assuming server is running)
run-client:
	docker run --rm -e SERVER_URL=ws://host.docker.internal:8080/ws $(IMAGE_NAME) /app/client

# Apply Kubernetes resources using Kustomize
k8s-apply:
	kubectl apply -k k8s/

# Delete Kubernetes resources using Kustomize
k8s-delete:
	kubectl delete -k k8s/

# Restart deployments to pull latest images
k8s-restart:
	kubectl rollout restart deployment/server-deployment deployment/client-deployment -n $(NAMESPACE)

# List all pods in the namespace
k8s-list-pods:
	kubectl get pods -n $(NAMESPACE) --no-headers -o custom-columns="NAME:.metadata.name,STATUS:.status.phase,AGE:.metadata.creationTimestamp"

# Delete a specific pod
k8s-delete-pod:
	@if [ -z "$(POD_NAME)" ]; then \
		echo "Available pods:"; \
		$(MAKE) k8s-list-pods; \
		echo ""; \
		echo "Usage: make k8s-delete-pod POD_NAME=<pod-name>"; \
		exit 1; \
	fi
	kubectl delete pod $(POD_NAME) -n $(NAMESPACE)

# View server logs
k8s-logs-server:
	kubectl logs -n $(NAMESPACE) -l app=server --tail=100 -f --prefix

# View client logs
k8s-logs-client:
	kubectl logs -n $(NAMESPACE) -l app=client --tail=100 -f --prefix

# Clean up Docker images
clean:
	docker rmi $(IMAGE_NAME) || true
