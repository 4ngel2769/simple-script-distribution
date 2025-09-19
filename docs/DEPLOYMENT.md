# Production Deployment Guide

## Overview

This guide covers deploying the Script Distribution Server in production environments with proper security, monitoring, and scaling considerations.

## Server Requirements

### Minimum Requirements
- **CPU**: 1 core
- **RAM**: 512MB
- **Storage**: 2GB free space
- **OS**: Linux (Ubuntu 20.04+ recommended) *`(Standard support ended for Ubuntu 20.04)`*

### Recommended for Production
- **CPU**: 2+ cores
- **RAM**: 2GB+
- **Storage**: 10GB+ free space
- **OS**: Linux (Ubuntu 22.04+ recommended)

## Production Deployment Options

### Option 1: Cloudflare Tunnel (Recommended)

**Advantages:**
- ✅ Automatic HTTPS
- ✅ DDoS protection
- ✅ No port forwarding needed
- ✅ Works behind NAT/firewall

**Setup:**

1. **Install Cloudflared:**
   ```bash
   # Ubuntu/Debian
   wget -q https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-amd64.deb
   sudo dpkg -i cloudflared-linux-amd64.deb
   ```

2. **Create Tunnel:**
   ```bash
   cloudflared tunnel login
   cloudflared tunnel create script-server
   ```

3. **Configure Tunnel:**
   ```yaml
   # ~/.cloudflared/config.yml
   tunnel: your-tunnel-id
   credentials-file: /home/user/.cloudflared/your-tunnel-id.json
   
   ingress:
     - hostname: get.yourdomain.com
       service: http://localhost:2020
     - service: http_status:404
   ```

4. **Update DNS:**
   ```bash
   cloudflared tunnel route dns script-server get.yourdomain.com
   ```

5. **Run Tunnel:**
   ```bash
   # Test
   cloudflared tunnel run script-server
   
   # Install as service
   sudo cloudflared service install
   sudo systemctl enable --now cloudflared
   ```

### Option 2: Traditional HTTPS with Let's Encrypt

**Setup:**

1. **Update Caddyfile:**
   ```caddyfile
   get.yourdomain.com {
       # Remove auto_https off
       
       # Your existing handlers
       handle /admin* {
           reverse_proxy admin-dashboard:8080
       }
       # ... rest of config
   }
   ```

2. **Expose HTTPS Port:**
   ```yaml
   # docker-compose.yml
   services:
     script-server:
       ports:
         - "80:80"
         - "443:443"  # Add this
   ```

3. **Configure Firewall:**
   ```bash
   sudo ufw allow 80/tcp
   sudo ufw allow 443/tcp
   ```

### Option 3: Reverse Proxy (Nginx/Apache)

Use if you already have a web server:

**Nginx Config:**
```nginx
server {
    listen 80;
    server_name get.yourdomain.com;
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name get.yourdomain.com;
    
    # SSL configuration
    ssl_certificate /path/to/cert.pem;
    ssl_certificate_key /path/to/key.pem;
    
    location / {
        proxy_pass http://localhost:2020;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
    
    location /admin {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

## Security Hardening

### 1. **System Security**
```bash
# Update system
sudo apt update && sudo apt upgrade -y

# Configure firewall
sudo ufw enable
sudo ufw default deny incoming
sudo ufw default allow outgoing
sudo ufw allow ssh
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp

# Disable root login
sudo sed -i 's/PermitRootLogin yes/PermitRootLogin no/' /etc/ssh/sshd_config
sudo systemctl restart ssh
```

### 2. **Application Security**
- ✅ Change default admin password immediately
- ✅ Use strong, unique passwords (20+ characters)
- ✅ Enable automatic security updates
- ✅ Regular backup your configurations
- ✅ Monitor access logs

### 3. **Docker Security**
```bash
# Run containers as non-root user
# Add to docker-compose.yml:
services:
  script-server:
    user: "1000:1000"  # Replace with your user ID
    
  admin-dashboard:
    user: "1000:1000"
```

## Monitoring and Logging

### 1. **Container Monitoring**
```bash
# Monitor containers
sudo docker stats

# View logs
sudo docker compose logs -f

# Set up log rotation
# Add to docker-compose.yml:
services:
  script-server:
    logging:
      driver: "json-file"
      options:
        max-size: "10m"
        max-file: "3"
```

### 2. **System Monitoring**
```bash
# Install monitoring tools
sudo apt install htop iotop nethogs

# Check disk usage
df -h
du -sh /var/www/scripts/*

# Monitor network
netstat -tulnp | grep :80
```

### 3. **Log Monitoring**
```bash
# View Caddy logs
sudo docker compose logs script-server

# View admin logs
sudo docker compose logs admin-dashboard

# System logs
journalctl -fu docker
```

## Backup and Disaster Recovery

### 1. **Automated Backup Script**
```bash
#!/bin/bash
# backup.sh

BACKUP_DIR="/backup/script-server"
DATE=$(date +%Y%m%d_%H%M%S)

# Create backup directory
mkdir -p "$BACKUP_DIR"

# Backup configurations
tar -czf "$BACKUP_DIR/config_$DATE.tar.gz" \
    /opt/script-distribution-server/admin/config.yaml \
    /opt/script-distribution-server/.env \
    /opt/script-distribution-server/docker-compose.yml

# Backup scripts
tar -czf "$BACKUP_DIR/scripts_$DATE.tar.gz" \
    /var/www/scripts/

# Cleanup old backups (keep 30 days)
find "$BACKUP_DIR" -name "*.tar.gz" -mtime +30 -delete

echo "Backup completed: $DATE"
```

### 2. **Automated Backup with Cron**
```bash
# Add to crontab
crontab -e

# Daily backup at 2 AM
0 2 * * * /opt/script-distribution-server/backup.sh >> /var/log/backup.log 2>&1
```

## Performance Optimization

### 1. **Caching**
```caddyfile
# Add to Caddyfile
handle @script_request {
    header Cache-Control "public, max-age=300"
    # ... existing config
}
```

### 2. **Resource Limits**
```yaml
# docker-compose.yml
services:
  script-server:
    deploy:
      resources:
        limits:
          memory: 256M
          cpus: '0.5'
        reservations:
          memory: 128M
          cpus: '0.25'
          
  admin-dashboard:
    deploy:
      resources:
        limits:
          memory: 512M
          cpus: '0.5'
```

## Scaling Considerations

### Multiple Instances
For high availability, consider:

1. **Load Balancer** (nginx/HAProxy)
2. **Shared Storage** for scripts directory
3. **Database** for script configurations (instead of YAML)
4. **Container Orchestration** (Docker Swarm/Kubernetes)

### CDN Integration
- Use Cloudflare or AWS CloudFront
- Cache script files for faster global delivery
- Reduce server load

## Troubleshooting

### Common Issues

**Scripts not accessible:**
```bash
# Check container status
sudo docker compose ps

# Check file permissions
ls -la /var/www/scripts/

# Check Caddy config
sudo docker compose exec script-server caddy validate --config /etc/caddy/Caddyfile
```

**Admin panel not working:**
```bash
# Check admin container
sudo docker compose logs admin-dashboard

# Verify config file
cat admin/config.yaml

# Test admin port
curl -I http://localhost:8080/admin
```

**Performance issues:**
```bash
# Check resource usage
sudo docker stats

# Monitor disk space
df -h

# Check network connectivity
curl -I https://yourdomain.com/health
```

## Maintenance Tasks

### Weekly
- [ ] Check system updates
- [ ] Review access logs
- [ ] Verify backups
- [ ] Monitor disk usage

### Monthly  
- [ ] Update Docker images
- [ ] Review security logs
- [ ] Test disaster recovery
- [ ] Performance optimization review

### Quarterly
- [ ] Security audit
- [ ] Capacity planning
- [ ] Documentation updates
- [ ] Dependency updates
