.PHONY: build run dev clean test deploy

# Build the application
build:
	go build -o bin/app cmd/server/main.go

# Run the application
run: build
	./bin/app

# Development with hot reload
dev:
	air

# Clean build artifacts
clean:
	rm -rf bin/
	rm -f internal/db/app.db

# Run tests
test:
	go test ./...

# Production build
build-prod:
	GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o bin/app-linux cmd/server/main.go
	GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o bin/app-darwin cmd/server/main.go
	GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o bin/app-windows.exe cmd/server/main.go

# Deploy (example for Docker)
deploy:
	docker build -t  .
	docker run -p 8080:8080 

# Install dependencies
install:
	go mod download
	go mod tidy
