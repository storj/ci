REGISTRY_HOST ?= storjlabs

IMAGE_NAME = ci
IMAGE_TAG = latest

IMAGE_FULL = $(REGISTRY_HOST)/$(IMAGE_NAME):$(IMAGE_TAG)

.build:
	docker build -t $(IMAGE_FULL) -f images/$(IMAGE_NAME)/Dockerfile .

.push:
	docker push $(IMAGE_FULL)

.clean:
	docker rmi $(IMAGE_FULL)


.PHONY: build-image
build-image:
	$(MAKE) .build

.PHONY: push-image
push-image:
	$(MAKE) .push

.PHONY: clean-image
clean-image:
	$(MAKE) .clean
