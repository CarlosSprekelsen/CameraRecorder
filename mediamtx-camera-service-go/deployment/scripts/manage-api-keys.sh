#!/bin/bash

# API Key Management Script for MediaMTX Camera Service
# Properly organized deployment script for API key management

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
SERVER_CONFIG="/opt/camera-service/config/default.yaml"
API_KEYS_FILE="/opt/camera-service/api-keys.json"
TEST_KEYS_DIR="$PROJECT_ROOT/config/test/api-keys"
DEPLOYMENT_KEYS_DIR="$PROJECT_ROOT/../deployment/keys"

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Show usage
show_usage() {
    echo "API Key Management Script for MediaMTX Camera Service"
    echo ""
    echo "Usage: $0 <command> [options]"
    echo ""
    echo "Commands:"
    echo "  generate <environment>  Generate API keys for specified environment"
    echo "  install <environment>   Install API keys to server"
    echo "  backup                  Backup existing API keys"
    echo "  restore <backup_file>   Restore API keys from backup"
    echo "  list                    List existing API keys"
    echo "  test                    Test API key authentication"
    echo "  clean                   Clean temporary files"
    echo ""
    echo "Environments:"
    echo "  test        Test environment (for development/testing)"
    echo "  development Development environment"
    echo "  staging     Staging environment"
    echo "  production  Production environment"
    echo ""
    echo "Examples:"
    echo "  $0 generate test"
    echo "  $0 install test"
    echo "  $0 backup"
    echo "  $0 test"
}

# Generate API keys for specified environment
generate_keys() {
    local environment=$1
    
    if [ -z "$environment" ]; then
        log_error "Environment not specified"
        show_usage
        exit 1
    fi
    
    log_info "Generating API keys for environment: $environment"
    
    case $environment in
        test)
            generate_test_keys
            ;;
        development|staging|production)
            generate_production_keys "$environment"
            ;;
        *)
            log_error "Invalid environment: $environment"
            show_usage
            exit 1
            ;;
    esac
}

# Generate test keys (for development/testing)
generate_test_keys() {
    local output_file="$TEST_KEYS_DIR/test-keys.json"
    local env_file="$TEST_KEYS_DIR/test-keys.env"
    
    log_info "Generating test API keys..."
    
    # Ensure directory exists
    mkdir -p "$TEST_KEYS_DIR"
    
    # Generate keys
    generate_api_keys_json "$output_file" "test"
    generate_env_file "$output_file" "$env_file"
    
    log_success "Test API keys generated:"
    log_info "  JSON: $output_file"
    log_info "  ENV:  $env_file"
    
    # Copy to client for testing
    local client_keys="$PROJECT_ROOT/../MediaMTX-Camera-Service-Client/client/tests/fixtures/test_api_keys.json"
    cp "$output_file" "$client_keys"
    log_info "  Client: $client_keys"
}

# Generate production keys
generate_production_keys() {
    local environment=$1
    local output_dir="$DEPLOYMENT_KEYS_DIR/$environment"
    local output_file="$output_dir/api-keys.json"
    
    log_info "Generating production API keys for: $environment"
    
    # Ensure directory exists
    mkdir -p "$output_dir"
    
    # Generate keys with longer expiry for production
    generate_api_keys_json "$output_file" "$environment" "365d"  # 1 year expiry
    
    log_success "Production API keys generated: $output_file"
    log_warning "Keep these keys secure and do not commit to version control!"
}

# Generate API keys JSON file
generate_api_keys_json() {
    local output_file=$1
    local environment=$2
    local expiry=${3:-"90d"}
    
    # Initialize the API keys JSON structure
    cat > "$output_file" << 'EOF'
{
  "keys": {
EOF

    local first_key=true
    
    # Generate keys for each role
    roles=("viewer" "operator" "admin")
    for role in "${roles[@]}"; do
        log_info "Generating API key for role: $role"
        
        # Generate secure random key (32 bytes)
        key_bytes=$(openssl rand -hex 32)
        
        # Create key ID with environment prefix
        key_id="${environment}_${role}_$(date +%s)_$(echo $key_bytes | cut -c1-8)"
        
        # Create the full key with csk_ prefix (base64url format)
        full_key="csk_$(echo $key_bytes | base64 | tr -d '=' | tr '+/' '-_' | tr -d '\n')"
        
        # Calculate expiry
        if [[ "$expiry" == *"d" ]]; then
            days=$(echo $expiry | sed 's/d//')
            expires_at=$(date -d "+${days} days" -Iseconds)
        else
            expires_at=$(date -d "+90 days" -Iseconds)
        fi
        created_at=$(date -Iseconds)
        
        # Add to JSON (handle comma placement)
        if [ "$first_key" = true ]; then
            first_key=false
        else
            echo "," >> "$output_file"
        fi
        
        cat >> "$output_file" << EOF
    "$key_id": {
      "id": "$key_id",
      "key": "$full_key",
      "role": "$role",
      "created_at": "$created_at",
      "expires_at": "$expires_at",
      "description": "Generated API key for $role role in $environment environment",
      "last_used": "1970-01-01T00:00:00Z",
      "usage_count": 0,
      "status": "active"
    }
EOF
    done
    
    # Close the JSON structure
    cat >> "$output_file" << 'EOF'
  }
}
EOF
}

# Generate environment file
generate_env_file() {
    local json_file=$1
    local env_file=$2
    
    cat > "$env_file" << 'EOF'
# API Keys Environment Variables
# Generated by manage-api-keys.sh

EOF

    # Extract keys from JSON and create environment variables
    if command -v jq >/dev/null 2>&1; then
        viewer_key=$(jq -r '.keys | to_entries[] | select(.value.role == "viewer") | .value.key' "$json_file")
        operator_key=$(jq -r '.keys | to_entries[] | select(.value.role == "operator") | .value.key' "$json_file")
        admin_key=$(jq -r '.keys | to_entries[] | select(.value.role == "admin") | .value.key' "$json_file")
        
        echo "export TEST_VIEWER_KEY=\"$viewer_key\"" >> "$env_file"
        echo "export TEST_OPERATOR_KEY=\"$operator_key\"" >> "$env_file"
        echo "export TEST_ADMIN_KEY=\"$admin_key\"" >> "$env_file"
        
        # Add legacy JWT token variables (using API keys)
        echo "" >> "$env_file"
        echo "# Legacy JWT Tokens (use API keys instead)" >> "$env_file"
        echo "export TEST_VIEWER_TOKEN=\"$viewer_key\"" >> "$env_file"
        echo "export TEST_OPERATOR_TOKEN=\"$operator_key\"" >> "$env_file"
        echo "export TEST_ADMIN_TOKEN=\"$admin_key\"" >> "$env_file"
        
        # Add server configuration
        echo "" >> "$env_file"
        echo "# Server Configuration" >> "$env_file"
        echo "export CAMERA_SERVICE_HOST=localhost" >> "$env_file"
        echo "export CAMERA_SERVICE_PORT=8002" >> "$env_file"
        echo "export CAMERA_SERVICE_WS_PATH=/ws" >> "$env_file"
        echo "export CAMERA_SERVICE_HEALTH_PORT=8003" >> "$env_file"
        echo "export CAMERA_SERVICE_HEALTH_PATH=/health" >> "$env_file"
    else
        log_warning "jq not available, environment variables not created"
    fi
}

# Install API keys to server
install_keys() {
    local environment=$1
    
    if [ -z "$environment" ]; then
        log_error "Environment not specified"
        show_usage
        exit 1
    fi
    
    local keys_file
    case $environment in
        test)
            keys_file="$TEST_KEYS_DIR/test-keys.json"
            ;;
        development|staging|production)
            keys_file="$DEPLOYMENT_KEYS_DIR/$environment/api-keys.json"
            ;;
        *)
            log_error "Invalid environment: $environment"
            exit 1
            ;;
    esac
    
    if [ ! -f "$keys_file" ]; then
        log_error "API keys file not found: $keys_file"
        log_info "Generate keys first with: $0 generate $environment"
        exit 1
    fi
    
    log_info "Installing API keys to server..."
    
    # Backup existing keys
    if [ -f "$API_KEYS_FILE" ]; then
        backup_keys
    fi
    
    # Install new keys
    sudo cp "$keys_file" "$API_KEYS_FILE"
    sudo chown camera-service:camera-service "$API_KEYS_FILE"
    sudo chmod 600 "$API_KEYS_FILE"
    
    log_success "API keys installed to server"
    log_info "Restart the camera service to apply changes:"
    log_info "  sudo systemctl restart camera-service"
}

# Backup existing API keys
backup_keys() {
    if [ -f "$API_KEYS_FILE" ]; then
        local backup_file="$API_KEYS_FILE.backup.$(date +%s)"
        sudo cp "$API_KEYS_FILE" "$backup_file"
        sudo chown "$(whoami):$(whoami)" "$backup_file"
        log_success "API keys backed up to: $backup_file"
    else
        log_warning "No existing API keys file found to backup"
    fi
}

# Test API key authentication
test_authentication() {
    log_info "Testing API key authentication..."
    
    # Check if test keys exist
    local test_keys="$TEST_KEYS_DIR/test-keys.json"
    if [ ! -f "$test_keys" ]; then
        log_error "Test API keys not found. Generate them first:"
        log_info "  $0 generate test"
        exit 1
    fi
    
    # Extract a test key
    if command -v jq >/dev/null 2>&1; then
        local test_key=$(jq -r '.keys | to_entries[0].value.key' "$test_keys")
        log_info "Testing with key: ${test_key:0:20}..."
        
        # Test ping (should work without auth)
        log_info "Testing ping (no auth required)..."
        curl -s "http://localhost:8003/health" | jq . || log_warning "Health endpoint test failed"
        
        log_info "Authentication test completed"
        log_info "Run client tests to verify full authentication:"
        log_info "  cd MediaMTX-Camera-Service-Client/client"
        log_info "  source .test_env"
        log_info "  npm run test:integration"
    else
        log_warning "jq not available, cannot test authentication"
    fi
}

# Clean temporary files
clean_files() {
    log_info "Cleaning temporary files..."
    
    # Remove old files from root directory
    rm -f /home/carlossprekelsen/CameraRecorder/server_api_keys.json
    rm -f /home/carlossprekelsen/CameraRecorder/server_api_keys.env
    rm -f /home/carlossprekelsen/CameraRecorder/generated_api_keys.json
    rm -f /home/carlossprekelsen/CameraRecorder/temp_*_key.json
    
    log_success "Temporary files cleaned"
}

# Main execution
main() {
    local command=$1
    local environment=$2
    
    case $command in
        generate)
            generate_keys "$environment"
            ;;
        install)
            install_keys "$environment"
            ;;
        backup)
            backup_keys
            ;;
        test)
            test_authentication
            ;;
        clean)
            clean_files
            ;;
        *)
            show_usage
            exit 1
            ;;
    esac
}

# Run main function
main "$@"
