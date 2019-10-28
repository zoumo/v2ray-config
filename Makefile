.PHONY: all
all: build


.PHONY: build
build:
	go run main.go


.PHONY: container
container: build
	docker build -t zoumo/v2ray:latest .

.PHONY: push
push:
	docker push zoumo/v2ray:latest 
