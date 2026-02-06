.PHONY: build docker-build docker-run run clean

# Build native binary
build:
	go build -o scoundrel-server

# Build Docker image
docker-build:
	docker build -t scoundrel .

# Run with Docker Compose
docker-run:
	docker-compose up -d

# Run Docker container manually on port 22
docker-run-manual:
	docker run -d -p 22:22 --name scoundrel-game scoundrel

# Run Docker container on alternate port (alongside system SSH)
docker-run-alt:
	docker run -d -p 2222:22 --name scoundrel-game scoundrel

# Stop Docker container
docker-stop:
	docker-compose down
	# or: docker stop scoundrel-game && docker rm scoundrel-game

# Run native server (requires root for port 22)
run:
	./scoundrel-server

# Clean build artifacts
clean:
	rm -f scoundrel-server
	docker-compose down 2>/dev/null || true
	docker rm -f scoundrel-game 2>/dev/null || true

# View logs
logs:
	docker-compose logs -f
