#!/bin/bash

# Script Distribution Server Installer
# Version: v0.4.0
set -e

VERSION="v0.4.0"
SCRIPT_NAME="Script Distribution Server Installer"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Helper functions
info() { echo -e "${BLUE}[INFO]${NC} $1"; }
success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
warning() { echo -e "${YELLOW}[WARNING]${NC} $1"; }
error() { echo -e "${RED}[ERROR]${NC} $1"; }

show_version() {
    echo -e "${BLUE}$SCRIPT_NAME${NC}"
    echo -e "${BLUE}Version: $VERSION${NC}"
    echo
    echo "A lightweight script distribution server installer"
    echo "Repository: https://github.com/4ngel2769/simple-script-distribution"
    echo
    exit 0
}

show_help() {
    echo -e "${BLUE}$SCRIPT_NAME v$VERSION${NC}"
    echo
    echo "Usage: $0 [OPTIONS]"
    echo
    echo "Options:"
    echo "  -h, --help     Show this help message"
    echo "  -v, --version  Show version information"
    echo
    echo "Description:"
    echo "  Installs and configures a script distribution server with web admin dashboard."
    echo "  The server allows you to host and distribute shell scripts via clean URLs."
    echo
    echo "Examples:"
    echo "  $0              # Run interactive installation"
    echo "  $0 -v           # Show version"
    echo "  $0 --help       # Show this help"
    echo
    exit 0
}

parse_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            -h|--help)
                show_help
                ;;
            -v|--version)
                show_version
                ;;
            *)
                error "Unknown option: $1"
                echo "Use '$0 --help' for usage information"
                exit 1
                ;;
        esac
        shift
    done
}

# Check if running as root
if [[ $EUID -eq 0 ]]; then
   error "This script should not be run as root for security reasons"
   exit 1
fi

# Check dependencies
check_dependencies() {
    info "Checking dependencies..."
    
    # Check Docker
    if ! command -v docker &> /dev/null; then
        error "Docker is not installed. Please install Docker first."
        echo "Visit: https://docs.docker.com/get-docker/"
        exit 1
    fi
    
    # Check Docker Compose
    if ! command -v docker-compose &> /dev/null && ! sudo docker compose version &> /dev/null; then
        error "Docker Compose is not installed. Please install Docker Compose first."
        echo "Visit: https://docs.docker.com/compose/install/"
        exit 1
    fi
    
    # Check Git
    if ! command -v git &> /dev/null; then
        error "Git is not installed. Please install Git first."
        exit 1
    fi
    
    # Check Go (for password hashing)
    if ! command -v go &> /dev/null; then
        warning "Go is not installed. You'll need to generate password hash manually."
    fi
    
    success "All dependencies are installed"
}

# Get installation directory
get_install_dir() {
    echo
    info "Choose installation directory:"
    echo "1) /opt/script-distribution-server (recommended)"
    echo "2) ~/script-distribution-server"
    echo "3) Custom path"
    
    read -p "Enter choice [1-3]: " choice < /dev/tty
    
    case $choice in
        1)
            INSTALL_DIR="/opt/script-distribution-server"
            if [ ! -w "/opt" ]; then
                error "Cannot write to /opt. Please run: sudo chown $USER:$USER /opt"
                exit 1
            fi
            ;;
        2)
            INSTALL_DIR="$HOME/script-distribution-server"
            ;;
        3)
            read -p "Enter custom path: " INSTALL_DIR < /dev/tty
            ;;
        *)
            INSTALL_DIR="/opt/script-distribution-server"
            ;;
    esac
    
    info "Installation directory: $INSTALL_DIR"
}

# Clone repository
clone_repo() {
    info "Cloning repository..."
    
    if [ -d "$INSTALL_DIR" ]; then
        warning "Directory $INSTALL_DIR already exists"
        read -p "Remove and reinstall? (y/N): " confirm < /dev/tty
        if [[ $confirm =~ ^[Yy]$ ]]; then
            rm -rf "$INSTALL_DIR"
        else
            error "Installation cancelled"
            exit 1
        fi
    fi
    
    # Get GitHub repo URL
    read -p "Enter GitHub repository URL (https://github.com/user/repo): " REPO_URL < /dev/tty
    if [ -z "$REPO_URL" ]; then
        error "Repository URL is required"
        exit 1
    fi
    
    git clone "$REPO_URL" "$INSTALL_DIR"
    cd "$INSTALL_DIR"
    
    success "Repository cloned successfully"
}

# Configure environment
setup_environment() {
    info "Setting up environment configuration..."
    
    # Copy environment template
    cp .env.example .env
    cp admin/example.config.yaml admin/config.yaml
    
    # Get domain
    echo
    read -p "Enter your domain (e.g., get.yourdomain.com): " DOMAIN < /dev/tty
    if [ -n "$DOMAIN" ]; then
        sed -i "s/DOMAIN=.*/DOMAIN=$DOMAIN/" .env
    fi
    
    # Get admin credentials
    echo
    read -p "Enter admin username [admin]: " ADMIN_USER < /dev/tty
    ADMIN_USER=${ADMIN_USER:-admin}
    
    echo
    read -s -p "Enter admin password: " ADMIN_PASS < /dev/tty
    echo
    
    if [ -z "$ADMIN_PASS" ]; then
        warning "Using default password 'admin123' - CHANGE THIS IMMEDIATELY!"
        ADMIN_PASS="admin123"
    fi
    
    # Generate password hash
    if command -v go &> /dev/null; then
        info "Generating password hash..."
        cd admin
        
        info "Installing Go dependencies..."
        go mod download
        go mod tidy
        
        PASS_HASH=$(go run cmd/hash_password/hash_password.go "$ADMIN_PASS" 2>/dev/null | tail -n1)
        
        # Fallback to old path if new structure doesn't exist
        if [ -z "$PASS_HASH" ] && [ -f "hash_password.go" ]; then
            PASS_HASH=$(go run hash_password.go "$ADMIN_PASS" | tail -n1)
        fi
        
        # Update config
        sed -i "s/username: .*/username: $ADMIN_USER/" config.yaml
        sed -i "s|password_hash: .*|password_hash: \"$PASS_HASH\"|" config.yaml
        
        cd ..
        success "Password hash generated and configured"
    else
        warning "Cannot generate password hash without Go. Please do this manually."
        echo "Run: cd admin && go run cmd/hash_password/hash_password.go \"$ADMIN_PASS\""
        echo "Then update admin/config.yaml with the generated hash"
    fi
}

# Setup directories
setup_directories() {
    info "Setting up directories..."
    
    # Create scripts directory
    SCRIPTS_DIR="/var/www/scripts"
    if [ ! -d "$SCRIPTS_DIR" ]; then
        sudo mkdir -p "$SCRIPTS_DIR"
        sudo chown -R $USER:$USER "$SCRIPTS_DIR"
    fi
    
    # Copy initial index
    if [ -f "scripts/index.html" ]; then
        cp scripts/index.html "$SCRIPTS_DIR/"
    fi
    
    success "Directories created"
}

# Configure firewall (optional)
configure_firewall() {
    echo
    read -p "Configure firewall to allow ports 80 and 8080? (y/N): " configure_fw < /dev/tty
    
    if [[ $configure_fw =~ ^[Yy]$ ]]; then
        info "Configuring firewall..."
        
        if command -v ufw &> /dev/null; then
            sudo ufw allow 80/tcp
            sudo ufw allow 8080/tcp
            success "UFW rules added"
        elif command -v firewall-cmd &> /dev/null; then
            sudo firewall-cmd --permanent --add-port=80/tcp
            sudo firewall-cmd --permanent --add-port=8080/tcp
            sudo firewall-cmd --reload
            success "Firewall rules added"
        else
            warning "No supported firewall found. Please configure manually."
        fi
    fi
}

# Start services
start_services() {
    info "Building and starting services..."
    
    # Build and start
    sudo docker compose up --build -d
    
    # Wait a moment for services to start
    sleep 5
    
    # Check status
    if sudo docker compose ps | grep -q "Up"; then
        success "Services started successfully!"
        echo
        info "Access your script server at: http://localhost"
        info "Access admin dashboard at: http://localhost:8080/admin"
        echo
        info "Default credentials:"
        info "  Username: $ADMIN_USER"
        info "  Password: [as configured]"
        echo
        warning "Remember to:"
        warning "1. Change default passwords"
        warning "2. Configure your domain DNS"
        warning "3. Set up HTTPS (Cloudflare Tunnel recommended)"
    else
        error "Services failed to start. Check logs with: sudo docker compose logs"
        exit 1
    fi
}

main() {
    parse_args "$@"
    
    echo -e "${BLUE}================================${NC}"
    echo -e "${BLUE}$SCRIPT_NAME v$VERSION${NC}"
    echo -e "${BLUE}================================${NC}"
    echo
    
    check_dependencies
    get_install_dir
    clone_repo
    setup_environment
    setup_directories
    configure_firewall
    start_services
    
    echo
    success "Installation completed successfully!"
    echo
    info "Installed version: $VERSION"
    info "Next steps:"
    echo "1. Access admin dashboard and create your first script"
    echo "2. Update the index page through the admin panel"
    echo "3. Configure your domain DNS if using a custom domain"
    echo
    info "For more information, see: $INSTALL_DIR/docs/SETUP.md"
    echo
    info "Useful commands:"
    echo "  View logs: sudo docker compose logs -f"
    echo "  Stop services: sudo docker compose down"
    echo "  Restart services: sudo docker compose restart"
}

# # Run main function
# main "$@"
