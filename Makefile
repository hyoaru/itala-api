.PHONY: build package

build:
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o bootstrap cmd/lambda/main.go

package: build
	zip -j function.zip bootstrap

