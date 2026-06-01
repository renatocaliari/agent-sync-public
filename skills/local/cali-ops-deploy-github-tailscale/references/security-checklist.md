# Pre-Deploy Security Checklist

Run these checks before deploying to a new server.

## SSH Security

```bash
# Check PermitRootLogin
grep PermitRootLogin /etc/ssh/sshd_config
# Expected: PermitRootLogin no

# Check PasswordAuthentication
grep PasswordAuthentication /etc/ssh/sshd_config
# Expected: PasswordAuthentication no

# Check X11Forwarding
grep X11Forwarding /etc/ssh/sshd_config
# Expected: X11Forwarding no

# Check MaxAuthTries
grep MaxAuthTries /etc/ssh/sshd_config
# Expected: MaxAuthTries 3

# Check LoginGraceTime
grep LoginGraceTime /etc/ssh/sshd_config
# Expected: LoginGraceTime 30
```

## Firewall

```bash
# Check UFW status
sudo ufw status verbose
# Expected:
#   Status: active
#   Default: deny (incoming), allow (outgoing)
#   22/tcp on tailscale0 ALLOW
#   2222/tcp on tailscale0 ALLOW
#   80/tcp ALLOW
#   443/tcp ALLOW
```

## Tailscale

```bash
# Check Tailscale status
tailscale status
# Expected: server listed with correct IP

# Check Tailscale SSH
tailscale set --ssh=true  # already set?
tailscale status | grep ssh
```

## Deploy User

```bash
# Check deploy user exists and is non-root
id deploy
# Expected: uid=NNN(deploy) groups=... (NOT root)

# Check deploy user can't get shell
ssh deploy@<tailscale-ip> whoami
# Expected: fails (command= restricts)

# Check deploy user can run docker
ssh deploy@<tailscale-ip> docker ps
# Expected: lists containers
```

## Docker

```bash
# Check Docker socket permissions
ls -la /var/run/docker.sock
# Expected: srw-rw---- root docker

# Check no insecure registries
cat /etc/docker/daemon.json
# Expected: no "insecure-registries" or empty
```

## Ports

```bash
# Check what's listening
sudo ss -tlnp
# Expected: no processes on 0.0.0.0:3001, 3002, 4040, 9122, 9111
```
