.PHONY: build

build:
	go build -buildmode=plugin -o adapter-kokoro-io.so adapter.go
