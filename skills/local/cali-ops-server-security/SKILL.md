---
name: cali-ops-server-security
description: >
  [Cali] Server security audit and hardening for private servers behind Tailscale.
  Use when: auditing server security, hardening SSH/firewall/Docker, checking for
  vulnerabilities, setting up fail2ban, reviewing port exposure, or responding to
  security alerts. Covers 6 layers: CloudFlare, UFW, Tailscale, SSH, Docker, Application.
  Triggers: "server security", "security audit", "harden server", "SSH hardening",
  "firewall rules", "UFW config", "fail2ban", "port security", "Docker security",
  "vulnerability check", "security review".

metadata:
  frequency: rare
  category: infra
  context-cost: low
disable-model-invocation: true
---

# Server Security Audit & Hardening

6-layer security model for private servers behind Tailscale.

## Security Layers

```
Internet → CloudFlare → UFW → Tailscale → SSH → Docker → App
```

| Layer | Purpose | Tool |
|---|---|---|
| 1. CloudFlare | CDN, DDoS, SSL, WAF, hide real IP | CloudFlare DNS + Tunnel |
| 2. UFW | Port filtering, default deny | UFW firewall |
| 3. Tailscale | WireGuard tunnel, private network | Tailscale |
| 4. SSH | Authentication, access control | OpenSSH / Tailscale SSH |
| 5. Docker | Container isolation, socket security | Docker Engine |
| 6. Application | Process isolation, env vars, volumes | Docker Compose |

## Required Stack

```bash
# Server packages
tailscale          # WireGuard tunnel
ufw                # Firewall
docker             # Container runtime
docker-compose     # Multi-container apps
fail2ban           # (optional) Brute-force protection

# SSH config
openssh-server     # SSH daemon
```

## Audit Commands

### Layer 2: UFW

```bash
sudo ufw status verbose
# Check:
#   Status: active
#   Default: deny (incoming), allow (outgoing)
#   Only tailscale0 ports allowed (22, 2222)
#   80/tcp and 443/tcp allowed (for web)
```

### Layer 3: Tailscale

```bash
tailscale status
# Check:
#   Server listed with correct IP
#   SSH enabled
#   Correct tags assigned

tailscale set --ssh=true  # ensure SSH enabled
```

### Layer 4: SSH

```bash
# Check critical settings
grep -E "PermitRootLogin|PasswordAuthentication|X11Forwarding|MaxAuthTries|LoginGraceTime" /etc/ssh/sshd_config

# Expected:
#   PermitRootLogin no
#   PasswordAuthentication no
#   X11Forwarding no
#   MaxAuthTries 3
#   LoginGraceTime 30

# Check deploy user
id deploy
cat /home/deploy/.ssh/authorized_keys
# Verify: command= restricts to docker compose only
# Verify: no-agent-forwarding, no-port-forwarding, no-X11-forwarding
```

### Layer 5: Docker

```bash
# Socket permissions
ls -la /var/run/docker.sock
# Expected: srw-rw---- root docker

# No insecure registries
cat /etc/docker/daemon.json | grep insecure
# Expected: nothing or empty

# Containers not exposing ports on 0.0.0.0
sudo ss -tlnp | grep -E "0\.0\.0\.0:(3001|3002|4040|9122|9111)"
# Expected: nothing
```

### Layer 6: Application

```bash
# Check container isolation
docker ps --format "{{.Names}}: {{.Ports}}"
# Expected: no 0.0.0.0 bindings for internal services

# Check volumes
docker inspect <container> | grep -A5 "Mounts"
# Expected: named volumes or specific host paths, not /
```

## Vulnerability Categories

### Critical (fix immediately)

| Issue | Risk | Fix |
|---|---|---|
| `PermitRootLogin yes` | Root SSH access | Set to `no` |
| `PasswordAuthentication yes` | Brute force risk | Set to `no` |

### Medium (fix soon)

| Issue | Risk | Fix |
|---|---|---|
| `X11Forwarding yes` | Display hijacking | Set to `no` |
| Processes on `0.0.0.0` | Port exposure | Bind to `127.0.0.1` or `tailscale0` |
| No fail2ban | No brute-force protection | Install + configure |

### Low (consider)

| Issue | Risk | Fix |
|---|---|---|
| No `AllowUsers` | Any user can SSH | Add `AllowUsers deploy` |
| No log monitoring | Missed intrusion attempts | Set up log alerts |

## Hardening Checklist

### SSH Hardening

```bash
# Add to /etc/ssh/sshd_config
PermitRootLogin no
PasswordAuthentication no
X11Forwarding no
MaxAuthTries 3
LoginGraceTime 30
AllowUsers deploy

# Restart SSH
sudo systemctl restart ssh
```

### UFW Hardening

```bash
# Verify rules
sudo ufw status numbered

# Remove any overly permissive rules
sudo ufw delete <rule-number>

# Rate limiting (optional)
sudo ufw limit 22/tcp
```

### Docker Hardening

```bash
# Restrict socket access
sudo chmod 660 /var/run/docker.sock
sudo chown root:docker /var/run/docker.sock

# No privileged containers
docker inspect <container> | grep Privileged
# Expected: false

# Read-only rootfs (optional)
docker run --read-only ...
```

### fail2ban (optional)

```bash
sudo apt install fail2ban
sudo systemctl enable fail2ban

# Configure
cat > /etc/fail2ban/jail.local << 'EOF'
[sshd]
enabled = true
port = 22,2222
filter = sshd
logpath = /var/log/auth.log
maxretry = 3
bantime = 3600
EOF

sudo systemctl restart fail2ban
```

## Examples

### Example 1: Full audit

```bash
# Run all checks
echo "=== UFW ===" && sudo ufw status verbose
echo "=== SSH ===" && grep -E "PermitRootLogin|PasswordAuthentication|X11Forwarding" /etc/ssh/sshd_config
echo "=== Tailscale ===" && tailscale status
echo "=== Docker ===" && docker ps --format "{{.Names}}: {{.Ports}}"
echo "=== Ports ===" && sudo ss -tlnp | grep -E "0\.0\.0\.0"
```

### Example 2: Fix critical vulnerabilities

```bash
# Fix SSH (run as root)
sudo sed -i 's/PermitRootLogin yes/PermitRootLogin no/' /etc/ssh/sshd_config
sudo sed -i 's/PasswordAuthentication yes/PasswordAuthentication no/' /etc/ssh/sshd_config
sudo sed -i 's/X11Forwarding yes/X11Forwarding no/' /etc/ssh/sshd_config
sudo systemctl restart ssh
```

### Example 3: Check for exposed ports

```bash
# Find processes listening on 0.0.0.0
sudo ss -tlnp | grep "0\.0\.0\.0" | awk '{print $NF, $4}'
# Should only show: sshd (22), caddy (80,443)
```

## Edge Cases

### Server is public (no Tailscale)
- UFW must allow SSH on standard port (22) from specific IPs
- Consider fail2ban for brute-force protection
- Use key-based auth only (no passwords)

### Multiple deploy users
- Each user needs separate `command=` in authorized_keys
- Use `AllowUsers` to whitelist all deploy users
- Verify each user's Docker group membership

### Root access needed temporarily
- Use `sudo` from deploy user instead of PermitRootLogin
- Add deploy user to sudoers: `deploy ALL=(ALL) NOPASSWD: /usr/bin/docker`
- Never leave PermitRootLogin yes permanently

### Container exposes ports
- Check `docker-compose.yml` for `ports:` section
- Bind to `127.0.0.1:host_port` instead of `host_port:host_port`
- Use UFW to block if needed

## Test Cases

### Should

- [ ] UFW active with default deny incoming
- [ ] SSH only accessible via Tailscale (port 22, 2222)
- [ ] PermitRootLogin set to no
- [ ] PasswordAuthentication set to no
- [ ] Deploy user is non-root with docker group
- [ ] Deploy SSH key has command= restriction
- [ ] Docker socket owned by root:docker
- [ ] No processes listening on 0.0.0.0 for internal services

### Should NOT

- [ ] Root should NOT be able to SSH directly
- [ ] Password auth should NOT be enabled
- [ ] Internal services should NOT bind to 0.0.0.0
- [ ] Deploy user should NOT have shell access
- [ ] Docker socket should NOT be world-readable
