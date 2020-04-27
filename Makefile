NAME = scheduling-webhook
IMAGE_PREFIX = gcr.io/quatreapp
IMAGE_NAME = k8s-scheduling-webhook
IMAGE_VERSION = $$(git log --abbrev-commit --format=%h -s | head -n 1)

export GO111MODULE=on

app: deps
	go build -v -o $(NAME) cmd/main.go

deps:
	go get -v ./...

test: deps
	go test -v ./... -cover
	
docker:
	docker build --no-cache -t $(IMAGE_PREFIX)/$(IMAGE_NAME):$(IMAGE_VERSION) .
	docker tag $(IMAGE_PREFIX)/$(IMAGE_NAME):$(IMAGE_VERSION) $(IMAGE_PREFIX)/$(IMAGE_NAME):latest

push:
	docker push $(IMAGE_PREFIX)/$(IMAGE_NAME):$(IMAGE_VERSION)
	docker push $(IMAGE_PREFIX)/$(IMAGE_NAME):latest
