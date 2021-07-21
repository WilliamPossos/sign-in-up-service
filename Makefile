.PHONY: build clean deploy

deps:
	go get

build:
	env GOARCH=amd64 GOOS=linux go build -o bin/application application.go

clean:
	rm -rf ./bin ./vendor Gopkg.lock

deploy: clean build
	sls deploy --verbose
