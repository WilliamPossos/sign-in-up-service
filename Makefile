.PHONY: build clean deploy

build:
    go get -u ./...
	env GOARCH=amd64 GOOS=linux -o bin/application application.go -ldflags="-s -w"
	env GOARCH=amd64 GOOS=linux -o bin/application application.go -ldflags="-s -w"

clean:
	rm -rf ./bin ./vendor Gopkg.lock

deploy: clean build
	sls deploy --verbose
