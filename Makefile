PROJECT?=hub.ahiho.com/ahiho/squirrel-srv
APP?=squirrel
PORT?=8080
SRC_PATH=$(shell pwd)

RELEASE?=0.2.1
COMMIT?=$(shell git rev-parse --short HEAD)
BUILD_TIME?=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
CONTAINER_IMAGE?=reg.ahiho.com/ahiho/${APP}

GOOS?=linux
GOARCH?=amd64
CGO_ENABLED?=1

clean:
	rm -f ${APP}

build: clean
	docker run --rm -it -v "${SRC_PATH}:/app" golang:1.11.4 bash -c "cd /app && CGO_ENABLED=${CGO_ENABLED} GOOS=${GOOS} GOARCH=${GOARCH} go build -a -installsuffix cgo -ldflags \"-s -w -X ${PROJECT}/pkg/version.Release=${RELEASE} -X ${PROJECT}/pkg/version.Commit=${COMMIT} -X ${PROJECT}/pkg/version.BuildTime=${BUILD_TIME}\" -o ${APP}"

container:
	docker build -t $(CONTAINER_IMAGE):$(RELEASE) .

run: container
	docker stop $(APP) || true && docker rm $(APP) || true
	docker run --name ${APP} -p ${PORT}:${PORT} --rm \
		-e "PORT=${PORT}" \
		$(CONTAINER_IMAGE):$(RELEASE)

test:
	go test -v -race ./...

push: container
	docker push $(CONTAINER_IMAGE):$(RELEASE)

kubernetes: push
	for t in $(shell find ./deployments/kubernetes -type f -name "*.yaml"); do \
        cat $$t | \
        	gsed -E "s/\{\{(\s*)\.Release(\s*)\}\}/$(RELEASE)/g" | \
        	gsed -E "s/\{\{(\s*)\.ServiceName(\s*)\}\}/$(APP)/g"; \
        echo ---; \
    done > tmp.yaml
	kubectl apply -f tmp.yaml

