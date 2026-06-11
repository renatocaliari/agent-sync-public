---
name: cali-ops-deploy-github-tailscale
description: >
  [Cali] Deploy to a private server via Tailscale OIDC with ephemeral credentials.
  Use when: setting up CI/CD deploy pipeline, configuring Tailscale SSH + OpenSSH
  for GitHub Actions, creating deploy users with restricted access, writing deploy.yml
  workflows, or troubleshooting deploy authentication failures.
  Triggers: "deploy to server", "tailscale deploy", "deploy pipeline", "CI/CD deploy",
  "deploy workflow", "deploy setup", "deploy key", "deploy user".

metadata:
  frequency: rare
  category: infra
  context-cost: low
disable-model-invocation: true
---

# Deploy to Private Server via Tailscale OIDC

Zero-trust deployment: ephemeral OIDC credentials, restricted SSH keys, WireGuard tunnel.

## Architecture

```
GITHUB RUNNER                    SERVER (100.120.175.47)
   │                                 │
   │ Tailscale OIDC (ephemeral)      │
   │ OpenSSH porta 2222              │
   │ chave ed25519 restrita          │
   │ command= limita comandos        │
   └─────────────────────────────────┘
         │
   UFW: só tailscale0
   Internet: bloqueada
```

## Prerequisites

- [ ] Tailscale account + admin access
- [ ] Server with Ubuntu/Debian + root SSH access
- [ ] Tailscale installed on server (`curl -fsSL https://tailscale.com/install.sh | sh`)
- [ ] Docker + Docker Compose on server
- [ ] GitHub repo with Actions enabled

## Setup Steps

### Phase 1: Tailscale ACLs

1. Go to https://login.tailscale.com/admin/acls
2. Add tag owner:
   ```json
   "tagOwners": {
     "tag:continuous-integration": ["autogroup:admin"]
   }
   ```
3. Add grants:
   ```json
   {
     "src": ["autogroup:member", "tag:continuous-integration"],
     "dst": ["<SERVER_IP>"],
     "ip": ["tcp:22", "tcp:2222"]
   }
   ```
4. Create OIDC credential at https://login.tailscale.com/admin/settings/oauth
   - Issuer: GitHub
   - Subject: `repo:{owner}/{repo}:ref:refs/heads/{branch}`
   - Scope: Auth Keys → Write
   - Tag: `tag:continuous-integration`
   - Copy Client ID + Audience

### Phase 2: Server Setup

Run as root on the server:

```bash
# Enable Tailscale SSH
tailscale set --ssh=true

# Discover Tailscale IP
TAILSCALE_IP=$(ip -4 addr show tailscale0 | grep -oP 'inet \K[\d.]+')

# Add port 2222 to SSH socket (Ubuntu systemd)
mkdir -p /etc/systemd/system/ssh.socket.d
cat > /etc/systemd/system/ssh.socket.d/override.conf << EOF
[Socket]
ListenStream=
ListenStream=0.0.0.0:22
ListenStream=[::]:22
ListenStream=$TAILSCALE_IP:2222
FreeBind=yes
EOF
systemctl daemon-reload
systemctl restart ssh.socket
systemctl restart ssh

# Firewall
ufw default deny incoming
ufw default allow outgoing
ufw allow in on tailscale0 to any port 22
ufw allow in on tailscale0 to any port 2222
ufw allow 80/tcp
ufw allow 443/tcp
ufw --force enable

# Create deploy user
adduser --disabled-password --gecos "" deploy
usermod -aG docker deploy

# Restricted SSH key
cat > /home/deploy/.ssh/authorized_keys << 'KEYEOF'
command="cd /opt/{project} && docker compose pull && docker rm -f {container} 2>/dev/null; docker compose up -d && docker image prune -f",no-agent-forwarding,no-port-forwarding,no-user-rc,no-X11-forwarding ssh-ed25519 AAA... github-actions-deploy-{project}
KEYEOF
chmod 600 /home/deploy/.ssh/authorized_keys
chown -R deploy:deploy /home/deploy/.ssh

# Copy Docker registry auth to deploy user
cp /root/.docker/config.json /home/deploy/.docker/config.json
chown -R deploy:deploy /home/deploy/.docker
```

### Phase 3: GitHub Secrets

| Secret | Value |
|---|---|
| `TS_OAUTH_CLIENT_ID` | OIDC Client ID |
| `TS_AUDIENCE` | OIDC Audience |
| `DEPLOY_SSH_KEY` | Private ed25519 key |

### Phase 4: Workflow

See `references/workflow-template.yml` for the full template.

Key elements:
- `id-token: write` permission for OIDC
- `tailscale/github-action@v4` for ephemeral auth
- SSH with restricted key + `command=`

### Workflow trigger options

The template in `references/workflow-template.yml` uses `on: push: [main]`.
For release-driven deploys, change the trigger:

```yaml
on:
  release:
    types: [published, prereleased]
```

This deploys on both stable and alpha/beta releases. Use `[published]` only
if you want to skip prereleases.

## Security Measures

| Measure | What it protects |
|---|---|
| OIDC (ephemeral JWT) | No permanent credentials on runner |
| `command=` in authorized_keys | Even with key, only docker compose allowed |
| Port 2222 on Tailscale IP only | Invisible on internet |
| UFW + tailscale0 interface | Only WireGuard traffic |
| `deploy` user (non-root) | Least privilege |
| Tailscale SSH (personal) | You access without keys, by identity |
| Ephemeral runner node | Disappears after workflow |

## Examples

### Example 1: First-time setup

```bash
# On server (as root):
curl -fsSL https://tailscale.com/install.sh | sh
tailscale set --ssh=true
tailscale status  # verify connection

# Create deploy user + restricted key
adduser --disabled-password --gecos "" deploy
usermod -aG docker deploy
```

### Example 2: Deploy from GitHub Actions

```yaml
- name: Authenticate with Tailscale
  uses: tailscale/github-action@v4
  with:
    oauth-client-id: ${{ secrets.TS_OAUTH_CLIENT_ID }}
    audience: ${{ secrets.TS_AUDIENCE }}
    tags: tag:continuous-integration

- name: Deploy
  shell: bash
  run: |
    echo "$SSH_KEY" > /tmp/deploy_key && chmod 600 /tmp/deploy_key
    ssh -i /tmp/deploy_key -p 2222 deploy@100.120.175.47 \
      'cd /opt/app && docker compose pull && docker compose up -d'
    shred -u /tmp/deploy_key
```

### Example 3: Verify deployment

```bash
# Check deploy user
ssh deploy@100.120.175.47  # should NOT get shell (command= restricts)
ssh deploy@100.120.175.47 whoami  # should fail (command= only allows docker)

# Check UFW
sudo ufw status verbose  # should show only tailscale0 ports
```

## Edge Cases

### OIDC token expired
- Runner must complete within token TTL (usually 1 hour)
- If expired, re-run the workflow — new OIDC token is minted automatically

### Server unreachable from runner
- Verify Tailscale status: `tailscale status` on both ends
- Check ACL grants include the server IP
- Verify UFW allows port 2222 on tailscale0

### Deploy user can't pull images
- Copy Docker registry auth: `cp /root/.docker/config.json /home/deploy/.docker/config.json`
- Verify ghcr.io token is valid

### Port 2222 not listening
- Check systemd socket override: `systemctl status ssh.socket`
- Verify `ListenStream` includes Tailscale IP
- Restart: `systemctl daemon-reload && systemctl restart ssh.socket`

## Test Cases

### Should

- [ ] Server only accepts SSH on port 22 (Tailscale) and 2222 (Tailscale IP)
- [ ] Internet cannot reach SSH ports (UFW blocks)
- [ ] Deploy user cannot get a shell (command= restricts)
- [ ] Deploy user can run docker compose pull/up
- [ ] OIDC token is ephemeral (no permanent credentials)
- [ ] Runner node disappears after workflow

### Should NOT

- [ ] Deploy user should NOT have root access
- [ ] Deploy user should NOT be able to run arbitrary commands
- [ ] SSH key should NOT have agent forwarding
- [ ] Port 2222 should NOT be accessible from internet
- [ ] Runner should NOT store permanent credentials

## References

- `references/workflow-template.yml` — Full GitHub Actions workflow
- `references/server-setup.sh` — Server setup script
- `references/security-checklist.md` — Pre-deploy security verification
