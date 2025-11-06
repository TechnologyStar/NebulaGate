FRONTEND_DIR = ./web
BACKEND_DIR = .

.PHONY: all build-frontend start-backend build-heimdall start-heimdall

all: build-frontend start-backend

build-frontend:
    @echo "Building frontend..."
    @cd $(FRONTEND_DIR) && bun install && DISABLE_ESLINT_PLUGIN='true' VITE_REACT_APP_VERSION=$(cat VERSION) bun run build

start-backend:
    @echo "Starting backend dev server..."
    @cd $(BACKEND_DIR) && go run main.go &

build-heimdall:
    @echo "Building Heimdall gateway..."
    @cd $(BACKEND_DIR) && go build -o bin/heimdall ./cmd/heimdall

start-heimdall:
    @echo "Starting Heimdall gateway..."
    @cd $(BACKEND_DIR) && ./bin/heimdall &
