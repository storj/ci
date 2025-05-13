REGISTRY_HOST ?= storjlabs

.PHONY: build-slim
build-slim:
	docker buildx build \
	    --load \
		--tag $(REGISTRY_HOST)/ci:slim \
		--platform linux/amd64 \
		-f images/Dockerfile .

.PHONY: build-and-push-images
build-and-push-images:
	docker buildx build \
		--target=ci \
		--push \
		--pull \
		--tag $(REGISTRY_HOST)/ci:latest \
		--platform linux/amd64 \
		-f images/Dockerfile .

	docker buildx build \
		--target=ci-slim \
		--push \
		--tag $(REGISTRY_HOST)/ci:slim \
		--platform linux/amd64,linux/arm64 \
		-f images/Dockerfile .

.PHONY: clean
clean:
	docker rmi $(REGISTRY_HOST)/ci:latest $(REGISTRY_HOST)/ci:slim
