.PHONY: lint fmt test bench

lint:
	golangci-lint run --fix ./...

fmt:
	gofumpt -w .
	gci write .

test:
	go test ./...

bench:
	go test -bench=. -benchmem ./parser/
