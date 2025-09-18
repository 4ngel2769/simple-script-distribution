# Script Distribution Server

A lightweight, self-hosted script distribution server with web-based admin dashboard. Inspired by popular one-liner installations like Pi-hole and CasaOS.

## Features

🚀 **Clean One-Liner Installation URLs**
- `curl -fsSL https://get.yourdomain.com/script | sudo bash`
- Pi-hole style simplicity

⚙️ **Web Admin Dashboard**
- Create, edit, and delete scripts through a web interface
- Real-time script content editor
- Manage redirects to external scripts (GitHub, etc.)

🐳 **Docker-First Design**
- Easy deployment with Docker Compose
- Lightweight containers (Caddy + Go)
- Volume-based script storage

🔐 **Secure by Default**
- Session-based authentication
- Bcrypt password hashing
- Cloudflare Tunnel ready

## Quick Start

1. **Clone the repository:**
   ```bash
   git clone https://github.com/yourusername/script-distribution-server.git
   cd script-distribution-server
   ```

2. **Configure environment:**
   ```bash
   cp .env.example .env
   cp admin/config.yaml.example admin/config.yaml
   # Edit both files with your settings
   ```

3. **Generate admin password:**
   ```bash
   cd admin
   go run hash_password.go "your_secure_password"
   # Copy the hash to config.yaml
   ```

4. **Start the server:**
   ```bash
   docker compose up -d
   ```

5. **Access admin dashboard:**
   - Open `http://localhost:8080/admin` (or your domain)
   - Login with your configured credentials

## Usage

### For Script Users
```bash
# Download and execute a script
curl -fsSL https://get.yourdomain.com/tor | sudo bash

# Just download
curl -O https://get.yourdomain.com/docker

# View available scripts
curl https://get.yourdomain.com/
```

### For Administrators
- **Admin Dashboard**: `https://yourdomain.com/admin`
- **Create Scripts**: Add local scripts or redirects to external URLs
- **Edit Content**: Built-in code editor for script files
- **Manage Index**: Auto-generate the main landing page

## Architecture

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│                 │    │                  │    │                 │
│   Users         │───▶│   Caddy Server   │───▶│   Admin Panel   │
│                 │    │   (Port 80)      │    │   (Port 8080)   │
│                 │    │                  │    │                 │
└─────────────────┘    └──────────────────┘    └─────────────────┘
                                │
                                ▼
                       ┌──────────────────┐
                       │                  │
                       │  Script Storage  │
                       │  (/var/www/...)  │
                       │                  │
                       └──────────────────┘
```

## Documentation

- [Setup Guide](docs/SETUP.md) - Detailed installation and configuration
- [API Reference](docs/API.md) - Admin API endpoints
- [Deployment](docs/DEPLOYMENT.md) - Production deployment guide

## Example Scripts

The server supports two types of scripts:

### Local Scripts
Stored and served directly from your server:
```bash
curl -fsSL https://get.yourdomain.com/tor | sudo bash
```

### Redirect Scripts  
Redirects to external URLs (like GitHub raw files):
```bash
curl -fsSL https://get.yourdomain.com/external-script | sudo bash
# ↓ Redirects to ↓
# https://raw.githubusercontent.com/user/repo/main/script.sh
```

## Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feature-name`
3. Commit changes: `git commit -am 'Add feature'`
4. Push to branch: `git push origin feature-name`
5. Submit a pull request

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Acknowledgments

- Inspired by [Pi-hole](https://pi-hole.net/)'s simple installation approach
- Built with [Caddy](https://caddyserver.com/) and [Go Fiber](https://gofiber.io/)
- Thanks to the self-hosted community for the inspiration!
