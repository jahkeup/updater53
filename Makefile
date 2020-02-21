build:
	go build -mod=readonly .

test:
	go test -v -race -mod=readonly ./...
