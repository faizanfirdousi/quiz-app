.PHONY: dev-backend dev-frontend build-lambda tidy

dev-backend:
	cd backend && go run ./cmd/local

dev-frontend:
	cd frontend && npm run dev

build-lambda:
	cd backend && \
	for dir in cmd/lambda/*/; do \
		name=$$(basename $$dir); \
		echo "Building $$name..."; \
		GOOS=linux GOARCH=amd64 go build -o dist/$$name/bootstrap ./$$dir; \
		cd dist/$$name && zip bootstrap.zip bootstrap && cd ../..; \
	done
	@echo "All Lambda functions built successfully."

tidy:
	cd backend && go mod tidy
	cd frontend && npm install
