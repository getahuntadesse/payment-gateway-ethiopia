.PHONY: help setup run-api run-worker run-all stop test clean migrate db-connect rabbitmq-ui

# Colors for output
GREEN = \033[0;32m
YELLOW = \033[1;33m
RED = \033[0;31m
NC = \033[0m # No Color

help:
	@echo "$(YELLOW)Ethiopian Payment Gateway - Make Commands$(NC)"
	@echo ""
	@echo "$(GREEN)setup$(NC)           - Setup database and dependencies"
	@echo "$(GREEN)run-api$(NC)         - Run the API server"
	@echo "$(GREEN)run-worker$(NC)      - Run the payment processor worker"
	@echo "$(GREEN)run-all$(NC)         - Run both API and worker"
	@echo "$(GREEN)stop$(NC)            - Stop all services"
	@echo "$(GREEN)test$(NC)            - Test the API with sample Ethiopian data"
	@echo "$(GREEN)clean$(NC)           - Clean up temporary files"
	@echo "$(GREEN)migrate$(NC)         - Run database migrations"
	@echo "$(GREEN)db-connect$(NC)      - Connect to PostgreSQL database"
	@echo "$(GREEN)rabbitmq-ui$(NC)     - Open RabbitMQ management UI"
	@echo ""

setup:
	@echo "$(YELLOW)Setting up Ethiopian Payment Gateway...$(NC)"
	@echo "1. Installing Go dependencies..."
	@go mod download
	@echo "2. Setting up database..."
	@powershell -Command "& {$$env:PGPASSWORD='postgres'; & 'C:\Program Files\PostgreSQL\15\bin\psql.exe' -U postgres -c \"CREATE DATABASE IF NOT EXISTS ethiopian_payments;\"}"
	@echo "$(GREEN)✅ Setup complete!$(NC)"

run-api:
	@echo "$(YELLOW)Starting Ethiopian Payment Gateway API...$(NC)"
	@echo "Server will run on: http://localhost:8080"
	@echo "API Documentation: http://localhost:8080/docs"
	@go run cmd/api/main.go

run-worker:
	@echo "$(YELLOW)Starting Ethiopian Payment Processor Worker...$(NC)"
	@echo "Worker is processing payments in ETB and USD..."
	@go run cmd/worker/main.go

run-all:
	@echo "$(YELLOW)Starting both API and Worker...$(NC)"
	@echo "Open two separate terminals and run:"
	@echo "1. make run-api"
	@echo "2. make run-worker"

stop:
	@echo "$(YELLOW)Stopping services...$(NC)"
	@taskkill /F /IM go.exe 2>nul || echo "No Go processes found"

test:
	@echo "$(YELLOW)Testing with Ethiopian sample data...$(NC)"
	@powershell -File .\scripts\test-ethiopian.ps1

clean:
	@echo "$(YELLOW)Cleaning up...$(NC)"
	@del /Q *.log 2>nul || echo "No log files found"
	@del /Q coverage.out 2>nul || echo "No coverage files found"

migrate:
	@echo "$(YELLOW)Running database migrations...$(NC)"
	@powershell -Command "& {$$env:PGPASSWORD='postgres'; & 'C:\Program Files\PostgreSQL\15\bin\psql.exe' -U postgres -d ethiopian_payments -f migrations\001_init_schema.up.sql}"
	@echo "$(GREEN)✅ Migrations applied!$(NC)"

db-connect:
	@echo "$(YELLOW)Connecting to PostgreSQL database...$(NC)"
	@powershell -Command "& {$$env:PGPASSWORD='postgres'; & 'C:\Program Files\PostgreSQL\15\bin\psql.exe' -U postgres -d ethiopian_payments}"

rabbitmq-ui:
	@echo "$(YELLOW)Opening RabbitMQ Management UI...$(NC)"
	@start http://localhost:15672
	@echo "Username: guest"
	@echo "Password: guest"