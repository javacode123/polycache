check:
	go mod tidy && go vet ./...

ut:
	go test -coverprofile=coverage.out ./... && go tool cover -func=coverage.out | grep total

generate:
	go generate ./...