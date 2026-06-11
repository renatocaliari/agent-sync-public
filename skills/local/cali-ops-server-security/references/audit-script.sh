# Security Audit Script

Run this script to audit your server's security posture.

```bash
#!/bin/bash
# Server Security Audit Script
# Run as root or with sudo

set -euo pipefail

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

pass() { echo -e "${GREEN}✅ PASS${NC}: $1"; }
fail() { echo -e "${RED}❌ FAIL${NC}: $1"; }
warn() { echo -e "${YELLOW}⚠️  WARN${NC}: $1"; }

echo "🔒 Server Security Audit"
echo "========================"
echo ""

# ── Layer 2: UFW ──────────────────────────────────────────────
echo "── Layer 2: UFW Firewall ──"
if sudo ufw status | grep -q "Status: active"; then
  pass "UFW is active"
else
  fail "UFW is NOT active"
fi

if sudo ufw status verbose | grep -q "Default: deny (incoming)"; then
  pass "Default policy: deny incoming"
else
  warn "Default policy may not be deny incoming"
fi

# Check for overly permissive rules
if sudo ufw status | grep -q "Anywhere"; then
  warn "Found 'Anywhere' rules — check if intentional"
else
  pass "No overly permissive rules"
fi
echo ""

# ── Layer 3: Tailscale ────────────────────────────────────────
echo "── Layer 3: Tailscale ──"
if command -v tailscale &>/dev/null; then
  pass "Tailscale is installed"
  if tailscale status &>/dev/null; then
    pass "Tailscale is connected"
  else
    fail "Tailscale is NOT connected"
  fi
else
  fail "Tailscale is NOT installed"
fi
echo ""

# ── Layer 4: SSH ──────────────────────────────────────────────
echo "── Layer 4: SSH Security ──"
SSHD_CONFIG="/etc/ssh/sshd_config"

check_ssh_setting() {
  local setting=$1
  local expected=$2
  local actual=$(grep -E "^$setting" "$SSHD_CONFIG" 2>/dev/null | awk '{print $2}' | head -1)
  if [ "$actual" = "$expected" ]; then
    pass "$setting = $expected"
  elif [ -z "$actual" ]; then
    warn "$setting not set (default may be insecure)"
  else
    fail "$setting = $actual (expected $expected)"
  fi
}

check_ssh_setting "PermitRootLogin" "no"
check_ssh_setting "PasswordAuthentication" "no"
check_ssh_setting "X11Forwarding" "no"
check_ssh_setting "MaxAuthTries" "3"
check_ssh_setting "LoginGraceTime" "30"

# Check deploy user
if id "deploy" &>/dev/null; then
  pass "Deploy user exists"
  DEPLOY_GROUPS=$(id -nG deploy 2>/dev/null)
  if echo "$DEPLOY_GROUPS" | grep -q "docker"; then
    pass "Deploy user in docker group"
  else
    warn "Deploy user NOT in docker group"
  fi
else
  warn "Deploy user not found"
fi
echo ""

# ── Layer 5: Docker ───────────────────────────────────────────
echo "── Layer 5: Docker Security ──"
if command -v docker &>/dev/null; then
  pass "Docker is installed"

  # Socket permissions
  SOCKET_PERMS=$(ls -la /var/run/docker.sock 2>/dev/null | awk '{print $1}')
  if echo "$SOCKET_PERMS" | grep -q "srw-rw----"; then
    pass "Docker socket permissions: srw-rw----"
  else
    warn "Docker socket permissions: $SOCKET_PERMS (expected srw-rw----)"
  fi

  # Insecure registries
  if [ -f /etc/docker/daemon.json ]; then
    if grep -q "insecure-registries" /etc/docker/daemon.json; then
      fail "Insecure registries configured!"
    else
      pass "No insecure registries"
    fi
  else
    pass "No daemon.json (defaults are secure)"
  fi

  # Privileged containers
  PRIV_CONTAINERS=$(docker ps -q | xargs -I{} docker inspect {} --format '{{.Name}}: {{.HostConfig.Privileged}}' 2>/dev/null | grep "true" || true)
  if [ -z "$PRIV_CONTAINERS" ]; then
    pass "No privileged containers running"
  else
    fail "Privileged containers found: $PRIV_CONTAINERS"
  fi
else
  fail "Docker is NOT installed"
fi
echo ""

# ── Layer 6: Ports ────────────────────────────────────────────
echo "── Layer 6: Port Exposure ──"
EXPOSED=$(sudo ss -tlnp 2>/dev/null | grep "0\.0\.0\.0" | grep -v -E ":(22|80|443) " || true)
if [ -z "$EXPOSED" ]; then
  pass "No unexpected ports on 0.0.0.0"
else
  warn "Ports exposed on 0.0.0.0:"
  echo "$EXPOSED"
fi
echo ""

# ── Summary ───────────────────────────────────────────────────
echo "========================"
echo "Audit complete. Review any FAIL or WARN items above."
```

## Usage

```bash
# Save as audit.sh, then:
chmod +x audit.sh
sudo ./audit.sh
```
