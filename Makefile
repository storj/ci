.PHONY: build-image
build-image:
	docker build --pull -t storjlabs/ci:latest .

.PHONY: push-image
push-image:
	docker push storjlabs/ci:latest

.PHONY: clean-image
clean-image:
	docker rmi storjlabs/ci:latest
