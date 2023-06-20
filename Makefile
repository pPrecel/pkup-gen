.PHONY: build
build:
	go build -o .out/pkup main.go

.PHONY: verify
verify:
	./hack/verify.sh
