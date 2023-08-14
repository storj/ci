REGISTRY_HOST ?= storjlabs

.PHONY: build-slim
build-slim:
	docker build \
		--tag $(REGISTRY_HOST)/ci:slim \
		-f images/ci-slim/Dockerfile .

.PHONY: build-and-push-images
build-and-push-images:
	docker buildx build \
		--push \
		--pull \
		--tag $(REGISTRY_HOST)/ci:latest \
		--platform linux/amd64 \
		-f images/ci/Dockerfile .

	docker buildx build \
		--push \
		--tag $(REGISTRY_HOST)/ci:slim \
		--platform linux/amd64,linux/arm64 \
		-f images/ci-slim/Dockerfile .

.PHONY: clean
clean:
	docker rmi $(REGISTRY_HOST)/ci:latest $(REGISTRY_HOST)/ci:slim
