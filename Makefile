REGISTRY_HOST ?= storjlabs

.PHONY: build-slim
build-slim:
	docker build \
		--tag $(REGISTRY_HOST)/ci:slim \
		-f images/ci-slim/Dockerfile .

.PHONY: build-images
build-images:
	docker buildx build \
		--load \
		--pull \
		--tag $(REGISTRY_HOST)/ci:latest \
		--platform linux/amd64 \
		-f images/ci/Dockerfile .

	docker buildx build \
		--load \
		--tag $(REGISTRY_HOST)/ci:slim \
		--platform linux/amd64,linux/arm64 \
		-f images/ci-slim/Dockerfile .

.PHONY: push-images
push-images:
	docker push $(REGISTRY_HOST)/ci:latest
	docker push $(REGISTRY_HOST)/ci:slim

.PHONY: clean
clean:
	docker rmi $(REGISTRY_HOST)/ci:latest $(REGISTRY_HOST)/ci:slim
