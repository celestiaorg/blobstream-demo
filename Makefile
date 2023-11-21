verify:
	go run main.go

install:
	go mod tidy
	go mod download