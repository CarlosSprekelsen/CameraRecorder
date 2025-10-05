# MediaMTX Camera Service Makefile
# Provides convenient commands for development and documentation

.PHONY: help build test clean docs-validate docs-audit docs-generate docs-serve

# Default target
help:
	@echo "MediaMTX Camera Service - Available Commands"
	@echo "============================================="
	@echo ""
	@echo "Development:"
	@echo "  build          - Build the camera service"
	@echo "  test           - Run all tests"
	@echo "  test-coverage  - Run tests with coverage report"
	@echo "  clean          - Clean build artifacts"
	@echo ""
	@echo "Documentation:"
	@echo "  docs-validate  - Validate documentation against implementation"
	@echo "  docs-audit     - Run comprehensive documentation audit"
	@echo "  docs-generate  - Generate API documentation from code"
	@echo "  docs-serve     - Serve documentation locally"
	@echo ""
	@echo "Quality Assurance:"
	@echo "  lint           - Run code linting"
	@echo "  format         - Format code"
	@echo "  security       - Run security checks"
	@echo ""
	@echo "Deployment:"
	@echo "  install        - Install the service"
	@echo "  uninstall      - Uninstall the service"
	@echo "  restart        - Restart the service"

# Build the camera service
build:
	@echo "🔨 Building camera service..."
	go build -o camera-service ./cmd/camera-service
	@echo "✅ Build completed"

# Run all tests
test:
	@echo "🧪 Running tests..."
	go test -v ./...
	@echo "✅ Tests completed"

# Run tests with coverage
test-coverage:
	@echo "🧪 Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "✅ Coverage report generated: coverage.html"

# Integration Tests (cross-package coverage)
integration-cover:
	@echo "🧪 Running integration tests with coverage..."
	@mkdir -p coverage/integration
	@go test ./tests/integration/... -coverpkg=./internal/... -coverprofile=coverage/integration/integration.out -v
	@go tool cover -func=coverage/integration/integration.out | tail -1

integration-cover-html:
	@mkdir -p coverage/integration
	@go test ./tests/integration/... -coverpkg=./internal/... -coverprofile=coverage/integration/integration.out -v
	@go tool cover -html=coverage/integration/integration.out -o coverage/integration/integration.html
	@echo "✅ HTML: coverage/integration/integration.html"

# Clean build artifacts
clean:
	@echo "🧹 Cleaning build artifacts..."
	rm -f camera-service
	rm -f coverage.out coverage.html
	rm -rf /tmp/implemented_methods.txt
	rm -rf /tmp/documented_methods.txt
	rm -rf /tmp/missing_docs.txt
	rm -rf /tmp/extra_docs.txt
	@echo "✅ Clean completed"

# Validate documentation
docs-validate:
	@echo "📚 Validating documentation..."
	@chmod +x scripts/validate-documentation.sh
	@./scripts/validate-documentation.sh
	@echo "✅ Documentation validation completed"

# Run comprehensive documentation audit
docs-audit:
	@echo "📊 Running documentation audit..."
	@echo "📋 Implementation vs Documentation Audit"
	@echo "======================================"
	@chmod +x scripts/validate-documentation.sh
	@./scripts/validate-documentation.sh
	@echo ""
	@echo "📈 Documentation Coverage Report"
	@echo "==============================="
	@if [ -f "/tmp/implemented_methods.txt" ]; then \
		echo "Implemented methods: $$(wc -l < /tmp/implemented_methods.txt)"; \
	fi
	@if [ -f "/tmp/documented_methods.txt" ]; then \
		echo "Documented methods: $$(wc -l < /tmp/documented_methods.txt)"; \
	fi
	@if [ -f "/tmp/missing_docs.txt" ]; then \
		echo "Missing documentation: $$(wc -l < /tmp/missing_docs.txt)"; \
		if [ -s "/tmp/missing_docs.txt" ]; then \
			echo ""; \
			echo "Missing methods:"; \
			cat /tmp/missing_docs.txt | sed 's/^/  - /'; \
		fi; \
	fi
	@if [ -f "/tmp/extra_docs.txt" ]; then \
		echo "Extra documentation: $$(wc -l < /tmp/extra_docs.txt)"; \
		if [ -s "/tmp/extra_docs.txt" ]; then \
			echo ""; \
			echo "Extra methods:"; \
			cat /tmp/extra_docs.txt | sed 's/^/  - /'; \
		fi; \
	fi

# Generate API documentation from code
docs-generate:
	@echo "📝 Generating API documentation from code..."
	@echo "Extracting methods from implementation..."
	@grep -r "Method.*func" ./internal/websocket/ | \
		grep -v test | \
		grep -o '"[^"]*"' | \
		sed 's/"//g' | \
		sort | uniq > /tmp/generated_methods.txt
	@echo "Generated methods list: /tmp/generated_methods.txt"
	@echo "✅ API documentation generation completed"

# Serve documentation locally
docs-serve:
	@echo "🌐 Serving documentation locally..."
	@if command -v python3 >/dev/null 2>&1; then \
		python3 -m http.server 8080 --directory docs/; \
	elif command -v python >/dev/null 2>&1; then \
		python -m SimpleHTTPServer 8080; \
	else \
		echo "❌ Python not found. Please install Python to serve documentation."; \
		exit 1; \
	fi

# Run code linting
lint:
	@echo "🔍 Running code linting..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "⚠️  golangci-lint not found. Installing..."
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
		golangci-lint run; \
	fi
	@echo "✅ Linting completed"

# Format code
format:
	@echo "🎨 Formatting code..."
	go fmt ./...
	@echo "✅ Code formatting completed"

# Run security checks
security:
	@echo "🔒 Running security checks..."
	@if command -v gosec >/dev/null 2>&1; then \
		gosec ./...; \
	else \
		echo "⚠️  gosec not found. Installing..."
		go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
		gosec ./...; \
	fi
	@echo "✅ Security checks completed"

# Install the service
install:
	@echo "📦 Installing camera service..."
	@sudo cp camera-service /opt/camera-service/
	@sudo cp -r config /opt/camera-service/
	@sudo cp -r docs /opt/camera-service/
	@sudo chown -R camera-service:camera-service /opt/camera-service
	@sudo systemctl enable camera-service
	@sudo systemctl start camera-service
	@echo "✅ Service installed and started"

# Uninstall the service
uninstall:
	@echo "🗑️  Uninstalling camera service..."
	@sudo systemctl stop camera-service
	@sudo systemctl disable camera-service
	@sudo rm -rf /opt/camera-service
	@echo "✅ Service uninstalled"

# Restart the service
restart:
	@echo "🔄 Restarting camera service..."
	@sudo systemctl restart camera-service
	@echo "✅ Service restarted"

# Development setup
dev-setup:
	@echo "🛠️  Setting up development environment..."
	@echo "Installing development tools..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
	@echo "Setting up git hooks..."
	@cp scripts/pre-commit-doc-validation.sh .git/hooks/pre-commit
	@chmod +x .git/hooks/pre-commit
	@echo "✅ Development environment setup completed"

# Full quality check
quality-check: lint format security test docs-validate
	@echo "✅ Full quality check completed"

# Documentation workflow
docs-workflow: docs-validate docs-audit docs-generate
	@echo "✅ Documentation workflow completed"

# Complete development workflow
dev-workflow: clean build test docs-validate lint security
	@echo "✅ Complete development workflow completed"