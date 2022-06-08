REGISTRY_HOST ?= storjlabs

IMAGE_NAME = ci
IMAGE_TAG = latest

IMAGE_FULL = $(REGISTRY_HOST)/$(IMAGE_NAME):$(IMAGE_TAG)

## build-image will not produce a usable image reference as buildx is unable to load into docker with it's current set
## of drivers (docker, docker-container, kubernetes).
.PHONY: build-image
build-image:
	docker buildx build . \
		--pull \
		-f images/$(IMAGE_NAME)/Dockerfile \
		--platform linux/amd64,linux/arm64 \
		--tag $(IMAGE_FULL)

## push-image's invocation of buildx will reuse parts from the docker build cache where possible. So this may seem like
## we're rebuilding the image, but we're likely just pulling from the layer cache.
.PHONY: push-image
push-image:
	docker buildx build . \
		--pull \
		-f images/$(IMAGE_NAME)/Dockerfile \
		--platform linux/amd64,linux/arm64 \
		--tag $(IMAGE_FULL) \
		--push

.PHONY: clean
clean:
	docker buildx prune
