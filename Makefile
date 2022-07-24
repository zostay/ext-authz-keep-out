.PHONY: build
build:
	docker build . -t zostay/ext-authz-keep-out

.PHONY: push
push: build
	docker push zostay/ext-authz-keep-out
