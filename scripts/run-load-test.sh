#!/bin/bash

# k6 Load Test Runner Script
# This script manages the Go application and runs k6 load tests with optional profiling

set -e  # Exit on error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
APP_PORT="${APP_PORT:-8080}"
APP_URL="http://localhost:${APP_PORT}"
PROFILE_DIR="${PROFILE_DIR:-./profiles}"
COLLECT_PROFILES="${COLLECT_PROFILES:-false}"
K6_SCRIPT="${K6_SCRIPT:-./k6/load-test.js}"

# Print colored message
print_info() {
    echo -e "${BLUE}ℹ${NC} $1"
}

print_success() {
    echo -e "${GREEN}✓${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

print_error() {
    echo -e "${RED}✗${NC} $1"
}

# Check if k6 is installed
check_k6() {
    if ! command -v k6 &> /dev/null; then
        print_error "k6 is not installed"
        echo ""
        echo "Install k6 from: https://k6.io/docs/get-started/installation/"
        echo ""
        echo "Quick install options:"
        echo "  macOS:   brew install k6"
        echo "  Linux:   See k6/README.md for apt instructions"
        echo "  Windows: choco install k6"
        exit 1
    fi
    print_success "k6 is installed ($(k6 version))"
}

# Check if Go is installed
check_go() {
    if ! command -v go &> /dev/null; then
        print_error "Go is not installed"
        exit 1
    fi
    print_success "Go is installed ($(go version))"
}

# Build the Go application
build_app() {
    print_info "Building Go application..."
    GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "local")
    BUILT_AT=$(date)
    
    go build -v -ldflags="-X 'main.gitCommit=${GIT_COMMIT}' -X 'main.builtAt=${BUILT_AT}'" -o ./timelapse-app
    
    if [ $? -eq 0 ]; then
        print_success "Application built successfully"
    else
        print_error "Failed to build application"
        exit 1
    fi
}

# Start the Go application with profiling
start_app() {
    print_info "Starting Go application with profiling on port ${APP_PORT}..."
    
    ./timelapse-app -port ":${APP_PORT}" -pprof &
    APP_PID=$!
    
    # Wait for app to start
    sleep 3
    
    # Check if app is running
    if ! kill -0 $APP_PID 2>/dev/null; then
        print_error "Failed to start application"
        exit 1
    fi
    
    # Test if app is responding
    if curl -s "${APP_URL}/version" > /dev/null; then
        print_success "Application is running (PID: ${APP_PID})"
    else
        print_error "Application is not responding"
        kill $APP_PID 2>/dev/null || true
        exit 1
    fi
}

# Stop the Go application
stop_app() {
    if [ -n "$APP_PID" ]; then
        print_info "Stopping application (PID: ${APP_PID})..."
        kill $APP_PID 2>/dev/null || true
        wait $APP_PID 2>/dev/null || true
        print_success "Application stopped"
    fi
}

# Collect CPU profile
collect_cpu_profile() {
    print_info "Collecting CPU profile (30 seconds)..."
    mkdir -p "$PROFILE_DIR"
    
    curl -s "${APP_URL}/debug/pprof/profile?seconds=30" -o "${PROFILE_DIR}/cpu.prof"
    
    if [ $? -eq 0 ] && [ -s "${PROFILE_DIR}/cpu.prof" ]; then
        print_success "CPU profile saved to ${PROFILE_DIR}/cpu.prof"
    else
        print_warning "Failed to collect CPU profile"
    fi
}

# Collect memory profile
collect_mem_profile() {
    print_info "Collecting memory heap profile..."
    mkdir -p "$PROFILE_DIR"
    
    curl -s "${APP_URL}/debug/pprof/heap" -o "${PROFILE_DIR}/mem.prof"
    
    if [ $? -eq 0 ] && [ -s "${PROFILE_DIR}/mem.prof" ]; then
        print_success "Memory profile saved to ${PROFILE_DIR}/mem.prof"
    else
        print_warning "Failed to collect memory profile"
    fi
}

# Collect goroutine profile
collect_goroutine_profile() {
    print_info "Collecting goroutine profile..."
    mkdir -p "$PROFILE_DIR"
    
    curl -s "${APP_URL}/debug/pprof/goroutine" -o "${PROFILE_DIR}/goroutine.prof"
    
    if [ $? -eq 0 ] && [ -s "${PROFILE_DIR}/goroutine.prof" ]; then
        print_success "Goroutine profile saved to ${PROFILE_DIR}/goroutine.prof"
    else
        print_warning "Failed to collect goroutine profile"
    fi
}

# Run k6 load test
run_load_test() {
    print_info "Running k6 load test..."
    echo ""
    
    # Run k6 with custom environment variable
    K6_BASE_URL="${APP_URL}" k6 run "$K6_SCRIPT"
    K6_EXIT_CODE=$?
    
    echo ""
    if [ $K6_EXIT_CODE -eq 0 ]; then
        print_success "Load test completed successfully - all thresholds passed!"
    else
        print_error "Load test failed - some thresholds were not met"
    fi
    
    return $K6_EXIT_CODE
}

# Collect profiles in background during test
collect_profiles_async() {
    if [ "$COLLECT_PROFILES" = "true" ]; then
        print_info "Profiles will be collected after load test completes"
        
        # Collect memory and goroutine profiles immediately
        collect_mem_profile
        collect_goroutine_profile
        
        # CPU profile note
        print_info "To collect CPU profile during load test, run in another terminal:"
        echo "  curl ${APP_URL}/debug/pprof/profile?seconds=30 -o ${PROFILE_DIR}/cpu.prof"
    fi
}

# Analyze profiles
analyze_profiles() {
    if [ "$COLLECT_PROFILES" = "true" ] && [ -d "$PROFILE_DIR" ]; then
        echo ""
        print_info "Profile analysis commands:"
        echo ""
        
        if [ -f "${PROFILE_DIR}/cpu.prof" ]; then
            echo "  CPU Profile:"
            echo "    go tool pprof -top ${PROFILE_DIR}/cpu.prof"
            echo "    go tool pprof -http=:8081 ${PROFILE_DIR}/cpu.prof"
        fi
        
        if [ -f "${PROFILE_DIR}/mem.prof" ]; then
            echo ""
            echo "  Memory Profile:"
            echo "    go tool pprof -top ${PROFILE_DIR}/mem.prof"
            echo "    go tool pprof -http=:8082 ${PROFILE_DIR}/mem.prof"
        fi
        
        if [ -f "${PROFILE_DIR}/goroutine.prof" ]; then
            echo ""
            echo "  Goroutine Profile:"
            echo "    go tool pprof -top ${PROFILE_DIR}/goroutine.prof"
        fi
        
        echo ""
    fi
}

# Cleanup on exit
cleanup() {
    stop_app
}

trap cleanup EXIT

# Main execution
main() {
    echo ""
    echo "========================================="
    echo "   k6 Load Test Runner"
    echo "========================================="
    echo ""
    
    # Check prerequisites
    check_go
    check_k6
    echo ""
    
    # Build and start app
    build_app
    start_app
    echo ""
    
    # Collect profiles in background if requested
    collect_profiles_async
    
    # Run the load test
    run_load_test
    TEST_EXIT_CODE=$?
    
    # Show profile analysis commands
    analyze_profiles
    
    echo ""
    print_info "Test complete!"
    echo ""
    
    exit $TEST_EXIT_CODE
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --collect-profiles)
            COLLECT_PROFILES=true
            shift
            ;;
        --port)
            APP_PORT="$2"
            APP_URL="http://localhost:${APP_PORT}"
            shift 2
            ;;
        --profile-dir)
            PROFILE_DIR="$2"
            shift 2
            ;;
        --k6-script)
            K6_SCRIPT="$2"
            shift 2
            ;;
        --help)
            echo "Usage: $0 [OPTIONS]"
            echo ""
            echo "Options:"
            echo "  --collect-profiles    Collect CPU and memory profiles during test"
            echo "  --port PORT          Application port (default: 8080)"
            echo "  --profile-dir DIR    Directory for profile output (default: ./profiles)"
            echo "  --k6-script PATH     Path to k6 test script (default: ./k6/load-test.js)"
            echo "  --help               Show this help message"
            echo ""
            echo "Environment variables:"
            echo "  APP_PORT             Same as --port"
            echo "  PROFILE_DIR          Same as --profile-dir"
            echo "  COLLECT_PROFILES     Set to 'true' to collect profiles"
            echo "  K6_SCRIPT            Same as --k6-script"
            echo ""
            exit 0
            ;;
        *)
            print_error "Unknown option: $1"
            echo "Use --help for usage information"
            exit 1
            ;;
    esac
done

# Run main
main
