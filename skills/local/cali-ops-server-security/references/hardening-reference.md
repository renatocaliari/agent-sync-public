# Server Hardening Reference

Quick-reference for hardening a private server behind Tailscale.

## SSH Hardening

```bash
# /etc/ssh/sshd_config
PermitRootLogin no
PasswordAuthentication no
X11Forwarding no
MaxAuthTries 3
LoginGraceTime 30
AllowUsers deploy

# Restart
sudo systemctl restart ssh
```

## UFW Rules

```bash
# Default
sudo ufw default deny incoming
sudo ufw default allow outgoing

# Tailscale-only SSH
sudo ufw allow in on tailscale0 to any port 22
sudo ufw allow in on tailscale0 to any port 2222

# Web (if using CloudFlare tunnel)
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp

# Enable
sudo ufw --force enable
```

## Docker Hardening

```bash
# Restrict socket
sudo chmod 660 /var/run/docker.sock
sudo chown root:docker /var/run/docker.sock

# No privileged containers
# In docker-compose.yml:
#   privileged: false  (default)

# Read-only rootfs (optional)
# In docker-compose.yml:
#   read_only: true

# No new capabilities (optional)
# In docker-compose.yml:
#   cap_drop:
#     - ALL
```

## fail2ban

```bash
sudo apt install fail2ban

# /etc/fail2ban/jail.local
[sshd]
enabled = true
port = 22,2222
filter = sshd
logpath = /var/log/auth.log
maxretry = 3
bantime = 3600
findtime = 600

sudo systemctl enable fail2ban
sudo systemctl restart fail2ban
```

## Deploy User

```bash
# Create
adduser --disabled-password --gecos "" deploy
usermod -aG docker deploy

# Restricted key
cat > /home/deploy/.ssh/authorized_keys << 'KEYEOF'
command="cd /opt/{project} && docker compose pull && docker compose up -d && docker image prune -f",no-agent-forwarding,no-port-forwarding,no-user-rc,no-X11-forwarding ssh-ed25519 AAA... github-actions-deploy-{project}
KEYEOF
chmod 600 /home/deploy/.ssh/authorized_keys
chown -R deploy:deploy /home/deploy/.ssh

# Docker registry auth
cp /root/.docker/config.json /home/deploy/.docker/config.json
chown -R deploy:deploy /home/deploy/.docker
```

## Monitoring

```bash
# Check auth logs
sudo tail -f /var/log/auth.log | grep -i "failed\|invalid\|refused"

# Check fail2ban
sudo fail2ban-client status sshd

# Check UFW logs
sudo tail -f /var/log/ufw.log

# Check Docker logs
docker logs <container> --tail 100 -f
```

## Emergency Lockdown

If you suspect intrusion:

```bash
# Lock down SSH immediately
sudo ufw deny 22/tcp
sudo ufw deny 2222/tcp

# Only allow Tailscale
sudo ufw allow in on tailscale0 to any port 22

# Check active sessions
who
w
last -10

# Check for suspicious processes
ps aux | grep -v "\[" | sort -nrk 3 | head -20

# Check cron for persistence
crontab -l
ls -la /etc/cron.*
```
