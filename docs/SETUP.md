# Setup Guide

## Prerequisites

- Docker and Docker Compose
- A domain name (optional, but recommended)
- Basic Linux/terminal knowledge

## Installation

### 1. Clone and Setup

```bash
git clone https://github.com/4ngel2769/simple-script-distribution.git
cd simple-script-distribution

# Copy configuration templates
cp .env.example .env
cp admin/example.config.yaml admin/config.yaml
```

### 2. Configure Environment

Edit `.env` file:
```bash
# Update with your domain
DOMAIN=get.yourdomain.com
SCRIPTS_PATH=/var/www/scripts
```

### 3. Configure Admin Access

Generate a secure password hash:
```bash
cd admin
go run hash_password.go "your_secure_password_here"
```

Update `admin/config.yaml`:
```yaml
admin:
  username: admin
  password_hash: "YOUR_GENERATED_HASH_HERE"
```

### 4. Start Services

```bash
# Build and start
sudo docker compose up --build -d

# Check status
sudo docker compose ps

# View logs
sudo docker compose logs -f
```

### 5. Initial Setup

1. **Access admin panel**: `http://localhost:8080/admin`
2. **Login** with your credentials
3. **Create your first script** or import existing ones
4. **Update index page** to reflect your scripts

## Directory Structure

After setup, your file structure will look like:

```
script-distribution-server/
├── docker-compose.yml
├── .env                    # Your environment config
├── Caddyfile
├── admin/
│   ├── config.yaml         # Your admin config
│   └── ...
└── /var/www/scripts/       # Script storage (Docker volume)
    ├── index.html          # Auto-generated landing page
    ├── tor/                # Example script directory
    │   └── runme_tor.sh
    └── .backups/           # Automatic backups
        └── ...
```

## Production Deployment

### With Cloudflare Tunnel

1. **Setup Cloudflare Tunnel**:
   ```bash
   # Install cloudflared
   # Create tunnel: cloudflared tunnel create script-server
   # Configure tunnel to point to localhost:80
   ```

2. **Update Caddyfile** for tunnel:
   ```caddyfile
   {
       auto_https off  # Cloudflare handles SSL
   }
   
   :80 {
       # Your existing config
   }
   ```

3. **Start tunnel**:
   ```bash
   cloudflared tunnel run script-server
   ```

### With Traditional SSL

1. **Update Caddyfile** for HTTPS:
   ```caddyfile
   get.yourdomain.com {
       # Your existing config
   }
   ```

2. **Expose HTTPS port**:
   ```yaml
   # In docker-compose.yml
   ports:
     - "80:80"
     - "443:443"  # Add this
   ```

## Configuration Options

### Environment Variables (.env)

| Variable | Description | Default |
|----------|-------------|---------|
| `DOMAIN` | Your domain name | `localhost` |
| `HTTP_PORT` | HTTP port | `80` |
| `ADMIN_PORT` | Admin panel port | `8080` |
| `SCRIPTS_PATH` | Scripts storage path | `/var/www/scripts` |

### Admin Config (admin/config.yaml)

```yaml
admin:
  username: admin           # Admin username
  password_hash: "..."     # Bcrypt hash of password

scripts:
  - name: example          # URL path (/example)
    description: "..."     # Description shown on index
    icon: "📜"            # Emoji icon
    type: local           # 'local' or 'redirect'
    redirect_url: "..."   # Only for redirect type
```

## Troubleshooting

### Common Issues

**Admin panel not accessible:**
```bash
# Check if containers are running
docker compose ps

# Check admin container logs
docker compose logs admin-dashboard

# Verify port binding
netstat -tlnp | grep 8080
```

**Scripts not found:**
```bash
# Check script directory
ls -la /var/www/scripts/

# Check file permissions
sudo docker compose exec script-server ls -la /srv/scripts/

# Recreate containers
sudo docker compose down && sudo docker compose up -d
```

**Caddy config errors:**
```bash
# Validate Caddyfile syntax
sudo docker compose exec script-server caddy validate --config /etc/caddy/Caddyfile

# Reload Caddy config
sudo docker compose exec script-server caddy reload --config /etc/caddy/Caddyfile
```

### Performance Tuning

For high-traffic scenarios:

1. **Enable Caddy caching**:
   ```caddyfile
   # Add to your endpoints
   header Cache-Control "public, max-age=300"
   ```

2. **Use CDN** (Cloudflare, etc.) for static assets

3. **Monitor resources**:
   ```bash
   docker stats
   ```

## Backup and Recovery

### Backup Script Storage
```bash
# Backup scripts
tar -czf scripts-backup-$(date +%Y%m%d).tar.gz /var/www/scripts/

# Backup config
cp admin/config.yaml admin/config.yaml.backup
```

### Restore
```bash
# Stop services
sudo docker compose down

# Restore scripts
tar -xzf scripts-backup-YYYYMMDD.tar.gz -C /

# Restore config
cp admin/config.yaml.backup admin/config.yaml

# Restart
sudo docker compose up -d
```

## Security Considerations

1. **Change default passwords**
2. **Use strong password hashes**
3. **Enable HTTPS** (Cloudflare Tunnel recommended)
4. **Regular backups**
5. **Monitor access logs**
6. **Keep Docker images updated**

## Next Steps

- [API Documentation](/docs/API.md) - Learn about the admin API
- [Deployment Guide](/docs/DEPLOYMENT.md) - Advanced deployment scenarios
- Check the [GitHub Issues](https://github.com/4ngel2769/simple-script-distribution/issues) for known issues
