#!/bin/bash
#
# docker-server-dashboard.sh - Real-time ASCII dashboard for Docker server monitoring
#
# Optimized for single SSH connection - all data collected in one pass
#

SERVER="${DEPLOY_SERVER:-}"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
MAGENTA='\033[0;35m'
CYAN='\033[0;36m'
WHITE='\033[1;37m'
DIM='\033[2m'
NC='\033[0m'

STATUS_HEALTHY="🟢"
STATUS_STARTING="🟡"
STATUS_UNHEALTHY="🔴"

log_error() { echo -e "${RED}❌ $1${NC}" >&2; }
log_info() { echo -e "${BLUE}ℹ️  $1${NC}"; }
log_success() { echo -e "${GREEN}✅ $1${NC}"; }

if [ -z "$SERVER" ]; then
    echo ""
    echo -e "${CYAN}╔══════════════════════════════════════════════════════════╗${NC}"
    echo -e "${CYAN}║${WHITE}  🐳 Docker Server Dashboard${NC}${CYAN}                                    ║${NC}"
    echo -e "${CYAN}╚══════════════════════════════════════════════════════════╝${NC}"
    echo ""
    read -p "📌 Server SSH (e.g., root@server.example.com): " SERVER
    echo ""
fi

# SSH options to suppress known_hosts warnings
SSH_OPTS="-o ConnectTimeout=10 -o BatchMode=yes -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null"

log_info "Connecting to ${SERVER}..."
if ! ssh $SSH_OPTS "$SERVER" 2>/dev/null "echo 'OK'" >/dev/null 2>&1; then
    log_error "Failed to connect to ${SERVER}."
    exit 1
fi
log_success "Connected to ${SERVER}"

log_info "Fetching all server data in one connection..."

# ============================================================================
# COLLECT ALL DATA IN A SINGLE SSH SESSION
# ============================================================================
# Using marker lines to separate sections:
#   <<<SECTION>>> - marks start of a section
#   --- - internal divider within ORPHANED (containers | images | volumes)
#   ||| - internal divider within ORPHANED (all_volumes | used_volumes)

ALL_DATA=$(ssh $SSH_OPTS "$SERVER" 2>/dev/null /bin/bash << 'REMOTE_SCRIPT'
set -e

echo "<<<INFO>>>"
echo "$(hostname)"
echo "$(uptime -p)"
echo "$(uname -sr)"
docker --version 2>/dev/null | sed 's/Docker version //'

echo "<<<DISK>>>"
df -BG / 2>/dev/null | tail -1 | awk '{print $2","$3","$4","$5}' | sed 's/G//g'

echo "<<<MEMORY>>>"
free -m 2>/dev/null | grep '^Mem:' | awk '{print $2","$3","$4}'

echo "<<<CONTAINERS>>>"
docker ps -a --format '{{.Names}}|{{.Image}}|{{.Status}}|{{.Ports}}' 2>/dev/null || true

echo "<<<IMAGES>>>"
docker images --format '{{.Repository}}|{{.Tag}}|{{.Size}}' 2>/dev/null || true

echo "<<<VOLUMES>>>"
docker ps -a --format '{{.Names}}' 2>/dev/null | while read c; do
    docker inspect "$c" --format '{{range .Mounts}}{{.Name}}|{{.Destination}}{{end}}' 2>/dev/null | while read line; do
        [ -n "$line" ] && echo "$c|$line"
    done
done || true

echo "<<<CONTAINER_IPS>>>"
docker ps --format '{{.Names}}' 2>/dev/null | while read c; do
    ip=$(docker inspect "$c" --format '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' 2>/dev/null)
    net=$(docker inspect "$c" --format '{{range $k,$v := .NetworkSettings.Networks}}{{$k}}{{end}}' 2>/dev/null)
    [ -n "$ip" ] && echo "$c|$ip|$net"
done || true

echo "<<<ORPHAN_CONTAINERS>>>"
docker ps -a --filter 'status=exited' --filter 'status=dead' --filter 'status=created' --format '{{.Names}}|{{.Image}}' 2>/dev/null || true

echo "<<<ORPHAN_IMAGES>>>"
docker images --filter 'dangling=true' --format '{{.Repository}}|{{.Tag}}|{{.Size}}' 2>/dev/null || true

echo "<<<ALL_VOLUMES>>>"
docker volume ls -q 2>/dev/null || true

echo "<<<USED_VOLUMES>>>"
docker ps -a --format '{{.Names}}' 2>/dev/null | while read c; do
    docker inspect "$c" --format '{{range .Mounts}}{{.Name}}{{end}}' 2>/dev/null || true
done | sort -u || true

echo "<<<CRON>>>"
crontab -l 2>/dev/null | grep -v '^#' | grep -v '^\$' || true

echo "<<<TAILSCALE>>>"
tailscale status --json 2>/dev/null | python3 -c 'import sys,json;print(json.load(sys.stdin).get("Self",{}).get("DNSName","").rstrip("."))' 2>/dev/null || echo ""

echo "<<<ROUTES>>>"
python3 << 'PYEOF'
import re, subprocess, json

routes = []

# Parse Caddyfile
try:
    with open('/etc/caddy/Caddyfile') as f:
        content = f.read()
    lines = content.split('\n')
    i = 0
    while i < len(lines):
        line = lines[i].strip()
        m = re.match(r'handle_path\s+/([^/*]+)', line)
        if m:
            path = m.group(1)
            j = i + 1
            while j < len(lines) and not lines[j].strip().startswith('}'):
                pm = re.search(r'reverse_proxy\s+(\S+)', lines[j])
                if pm:
                    routes.append(('path', '/' + path + '/', pm.group(1), ''))
                    break
                j += 1
            i = j; continue
        if re.match(r':(\d+)\s*{', line):
            port = re.match(r':(\d+)', line).group(1)
            depth = 1; j = i + 1; backend = None
            while j < len(lines) and depth > 0:
                cline = lines[j].strip()
                depth += cline.count('{') - cline.count('}')
                pm = re.search(r'reverse_proxy\s+(\S+)', cline)
                if pm: backend = pm.group(1)
                j += 1
            if backend:
                routes.append(('port', ':' + port, backend, ''))
            i = j; continue
        i += 1
except: pass

# Build container → port mapping
try:
    out = subprocess.check_output(['docker','ps','--format','{{.Names}}|{{.Image}}|{{.ID}}'], text=True)
    for cline in out.strip().split('\n'):
        if not cline: continue
        parts = cline.split('|')
        cname, cimage, cid = parts[0], parts[1], parts[2] if len(parts) > 2 else ''
        netmode = ''
        try:
            netmode = subprocess.check_output(['docker','inspect',cid,'--format','{{.HostConfig.NetworkMode}}'], text=True).strip()
        except: pass

        # Get published port mappings
        try:
            insp = subprocess.check_output(['docker','inspect',cid,'--format','{{json .NetworkSettings.Ports}}'], text=True)
            pmap = json.loads(insp)
            for container_port, host_bindings in pmap.items():
                if host_bindings:
                    for hb in host_bindings:
                        host_port = hb.get('HostPort','')
                        if host_port:
                            for ri, (rt, tg, bk, _) in enumerate(routes):
                                if bk and (':' + host_port in bk or bk.endswith(':' + host_port)):
                                    routes[ri] = (rt, tg, bk, cname)
        except: pass

        # For host-network containers
        if netmode == 'host':
            try:
                top_out = subprocess.check_output(['docker','top',cid,'-eo','pid'], text=True)
                pids = set()
                for tp in top_out.strip().split('\n'):
                    tp = tp.strip()
                    if tp and tp.isdigit():
                        pids.add(tp)
                
                ss_out = subprocess.check_output(['ss','-tlnp'], text=True)
                for ss_line in ss_out.split('\n'):
                    pid_m = re.search(r'pid=(\d+)', ss_line)
                    port_m = re.search(r':(\d+)\s', ss_line)
                    if pid_m and port_m and pid_m.group(1) in pids:
                        port = port_m.group(1)
                        for ri, (rt, tg, bk, _) in enumerate(routes):
                            if bk and (':' + port in bk or bk.endswith(':' + port)):
                                routes[ri] = (rt, tg, bk, cname)
            except: pass
        
        # For internal networks
        if netmode != 'host':
            try:
                net_json = subprocess.check_output(['docker','inspect',cid,'--format','{{json .NetworkSettings.Networks}}'], text=True)
                nets = json.loads(net_json)
                for net_name, net_info in nets.items():
                    ip = net_info.get('IPAddress', '')
                    if ip:
                        for ri, (rt, tg, bk, _) in enumerate(routes):
                            if bk and bk.startswith(ip):
                                routes[ri] = (rt, tg, bk, cname)
            except: pass
except: pass

for rtype, target, backend, cname in routes:
    label = cname if cname else (target.replace('/','') if rtype=='path' else target.replace(':',''))
    print(f'{rtype}|{label}|{target}|{backend}')
PYEOF

REMOTE_SCRIPT
)

# ============================================================================
# PARSE ALL DATA
# ============================================================================

# Initialize
SERVER_HOSTNAME=""
SERVER_UPTIME=""
SERVER_OS=""
DOCKER_VER=""
DISK_TOTAL=0; DISK_USED=0; DISK_AVAIL=0; DISK_PCT=0
MEM_TOTAL=0; MEM_USED=0; MEM_AVAIL=0; MEM_PCT=0
CONTAINERS=""; IMAGES=""; VOLUME_MOUNTS=""; CONTAINER_IPS=""
ORPHAN_CONTAINERS=""; ORPHAN_IMAGES=""; ALL_VOLUMES=""; USED_VOLUMES=""
CRON_JOBS=""; TAILSCALE_DOMAIN=""; ROUTES=""

# Parse using state machine
STATE="WAIT"
while IFS= read -r line; do
    case "$line" in
        "<<<INFO>>>") STATE="INFO"; continue ;;
        "<<<DISK>>>") STATE="DISK"; continue ;;
        "<<<MEMORY>>>") STATE="MEMORY"; continue ;;
        "<<<CONTAINERS>>>") STATE="CONTAINERS"; continue ;;
        "<<<IMAGES>>>") STATE="IMAGES"; continue ;;
        "<<<VOLUMES>>>") STATE="VOLUMES"; continue ;;
        "<<<CONTAINER_IPS>>>") STATE="CONTAINER_IPS"; continue ;;
        "<<<ORPHAN_CONTAINERS>>>") STATE="ORPHAN_CONTAINERS"; continue ;;
        "<<<ORPHAN_IMAGES>>>") STATE="ORPHAN_IMAGES"; continue ;;
        "<<<ALL_VOLUMES>>>") STATE="ALL_VOLUMES"; continue ;;
        "<<<USED_VOLUMES>>>") STATE="USED_VOLUMES"; continue ;;
        "<<<CRON>>>") STATE="CRON"; continue ;;
        "<<<TAILSCALE>>>") STATE="TAILSCALE"; continue ;;
        "<<<ROUTES>>>") STATE="ROUTES"; continue ;;
    esac
    
    case "$STATE" in
        INFO)
            [ -z "$SERVER_HOSTNAME" ] && SERVER_HOSTNAME="$line" || \
            [ -z "$SERVER_UPTIME" ] && SERVER_UPTIME="$line" || \
            [ -z "$SERVER_OS" ] && SERVER_OS="$line" || \
            [ -z "$DOCKER_VER" ] && DOCKER_VER="$line"
            ;;
        DISK)
            IFS=',' read -r DISK_TOTAL DISK_USED DISK_AVAIL DISK_PCT <<< "$line"
            DISK_PCT=${DISK_PCT//%/}
            ;;
        MEMORY)
            IFS=',' read -r MEM_TOTAL MEM_USED MEM_AVAIL <<< "$line"
            ;;
        CONTAINERS) CONTAINERS="${CONTAINERS}${line}"$'\n' ;;
        IMAGES) IMAGES="${IMAGES}${line}"$'\n' ;;
        VOLUMES) VOLUME_MOUNTS="${VOLUME_MOUNTS}${line}"$'\n' ;;
        CONTAINER_IPS) CONTAINER_IPS="${CONTAINER_IPS}${line}"$'\n' ;;
        ORPHAN_CONTAINERS) ORPHAN_CONTAINERS="${ORPHAN_CONTAINERS}${line}"$'\n' ;;
        ORPHAN_IMAGES) ORPHAN_IMAGES="${ORPHAN_IMAGES}${line}"$'\n' ;;
        ALL_VOLUMES) ALL_VOLUMES="${ALL_VOLUMES}${line}"$'\n' ;;
        USED_VOLUMES) USED_VOLUMES="${USED_VOLUMES}${line}"$'\n' ;;
        CRON) CRON_JOBS="${CRON_JOBS}${line}"$'\n' ;;
        TAILSCALE) TAILSCALE_DOMAIN="$line" ;;
        ROUTES) ROUTES="${ROUTES}${line}"$'\n' ;;
    esac
done <<< "$ALL_DATA"

# Calculate orphaned volumes
ORPHAN_VOLUMES=""
ORPHAN_VOLUME_COUNT=0
while IFS= read -r vol; do
    [ -z "$vol" ] && continue
    if ! echo "$USED_VOLUMES" | grep -q "^${vol}$"; then
        ORPHAN_VOLUMES="${ORPHAN_VOLUMES}${vol}"$'\n'
        ORPHAN_VOLUME_COUNT=$((ORPHAN_VOLUME_COUNT + 1))
    fi
done <<< "$ALL_VOLUMES"

count_lines() { local var="$1"; [ -z "$var" ] && echo "0" || echo -n "$var" | wc -l | tr -d ' '; }

# Calculate counts
CONTAINER_COUNT=$(echo "$CONTAINERS" | grep -c . 2>/dev/null || echo "0")
RUNNING_COUNT=$(echo "$CONTAINERS" | grep -c "Up" 2>/dev/null || echo "0")
STOPPED_COUNT=$((CONTAINER_COUNT - RUNNING_COUNT))
IMAGE_COUNT=$(echo "$IMAGES" | grep -c . 2>/dev/null || echo "0")
VOLUME_COUNT=$(echo "$VOLUME_MOUNTS" | grep -c . 2>/dev/null || echo "0")
ORPHAN_CONTAINER_COUNT=$(count_lines "$ORPHAN_CONTAINERS")
ORPHAN_IMAGE_COUNT=$(count_lines "$ORPHAN_IMAGES")
CRON_COUNT=$(echo "$CRON_JOBS" | grep -c . 2>/dev/null || echo "0")

[ -n "$MEM_TOTAL" ] && [ "$MEM_TOTAL" -gt 0 ] 2>/dev/null && MEM_PCT=$((MEM_USED * 100 / MEM_TOTAL))
[ -z "$DISK_PCT" ] && DISK_PCT=0

log_success "Data collected"

# ============================================================================
# RENDER DASHBOARD
# ============================================================================

DISK_BAR_WIDTH=$((DISK_PCT * 30 / 100))
MEM_BAR_WIDTH=$((MEM_PCT * 30 / 100))

disk_bar=""; for ((i=0; i<DISK_BAR_WIDTH; i++)); do disk_bar+="█"; done
for ((i=DISK_BAR_WIDTH; i<30; i++)); do disk_bar+="░"; done

mem_bar=""; for ((i=0; i<MEM_BAR_WIDTH; i++)); do mem_bar+="█"; done
for ((i=MEM_BAR_WIDTH; i<30; i++)); do mem_bar+="░"; done

MEM_TOTAL_GB=$(awk "BEGIN {printf \"%.1f\", ${MEM_TOTAL:-0}/1024}")
MEM_USED_GB=$(awk "BEGIN {printf \"%.1f\", ${MEM_USED:-0}/1024}")
MEM_AVAIL_GB=$(awk "BEGIN {printf \"%.1f\", ${MEM_AVAIL:-0}/1024}")

TIMESTAMP=$(date -u +"%Y-%m-%d %H:%M:%S UTC")

echo ""
echo -e "${CYAN}╔══════════════════════════════════════════════════════════════════╗${NC}"
echo -e "${CYAN}║${WHITE}  🐳 DOCKER SERVER DASHBOARD${NC}${CYAN}                                      ║${NC}"
echo -e "${CYAN}║${NC}  Server: ${YELLOW}${SERVER}${NC}${CYAN}                                          ║${NC}"
echo -e "${CYAN}║${NC}  Updated: ${DIM}${TIMESTAMP}${NC}${CYAN}                                         ║${NC}"
echo -e "${CYAN}╠══════════════════════════════════════════════════════════════════╣${NC}"
echo -e "${CYAN}║${NC}  🖥️  Server: ${WHITE}${SERVER_HOSTNAME}${NC} | ${DIM}${SERVER_OS}${NC} | Docker ${DOCKER_VER}${NC}"
echo -e "${CYAN}║${NC}  ⏱️  Uptime: ${DIM}${SERVER_UPTIME}${NC}${CYAN}                                                    ║${NC}"
echo -e "${CYAN}╠══════════════════════════════════════════════════════════════════╣${NC}"
echo -e "${CYAN}║${NC}  ${GREEN}💾 DISK USAGE${NC}                    ${MAGENTA}🧠 MEMORY USAGE${NC}${CYAN}                ║${NC}"
echo -e "${CYAN}║${NC}  ${GREEN}┌──────────────────────────┐${NC}    ${MAGENTA}┌──────────────────────────┐${NC}${CYAN}   ║${NC}"
echo -e "${CYAN}║${NC}  ${GREEN}│${NC}${disk_bar}${GREEN}│${NC}    ${MAGENTA}│${NC}${mem_bar}${MAGENTA}│${NC}${CYAN}   ║${NC}"
echo -e "${CYAN}║${NC}  ${GREEN}│${NC} ${DIM}${DISK_USED}GB / ${DISK_TOTAL}GB${NC}            ${GREEN}│${NC}    ${MAGENTA}│${NC} ${DIM}${MEM_USED_GB}GB / ${MEM_TOTAL_GB}GB${NC}            ${MAGENTA}│${NC}${CYAN}   ║${NC}"
echo -e "${CYAN}║${NC}  ${GREEN}│${NC} ${DIM}${DISK_AVAIL}GB available${NC}          ${GREEN}│${NC}    ${MAGENTA}│${NC} ${DIM}${MEM_AVAIL_GB}GB available${NC}          ${MAGENTA}│${NC}${CYAN}   ║${NC}"
echo -e "${CYAN}║${NC}  ${GREEN}└──────────────────────────┘${NC}    ${MAGENTA}└──────────────────────────┘${NC}${CYAN}   ║${NC}"

[ "$DISK_PCT" -gt 80 ] 2>/dev/null && echo -e "${CYAN}║${NC}  ${RED}⚠️  WARNING: Disk usage above 80%!${NC}${CYAN}                                     ║${NC}"
[ "$MEM_PCT" -gt 80 ] 2>/dev/null && echo -e "${CYAN}║${NC}  ${RED}⚠️  WARNING: Memory usage above 80%!${NC}${CYAN}                                    ║${NC}"

echo -e "${CYAN}╠══════════════════════════════════════════════════════════════════╣${NC}"
echo -e "${CYAN}║${NC}  ${WHITE}📦 CONTAINERS${NC} (${GREEN}${RUNNING_COUNT} running${NC}, ${YELLOW}${STOPPED_COUNT} stopped${NC})${CYAN}                            ║${NC}"

if [ "$CONTAINER_COUNT" -gt 0 ] && [ -n "$CONTAINERS" ]; then
    echo -e "${CYAN}║${NC}  ${GREEN}┌────────────┬─────────────────────┬──────────┬─────────────────┐${NC}${CYAN} ║${NC}"
    echo -e "${CYAN}║${NC}  ${GREEN}│${NC} ${DIM}NAME      ${NC}│${NC} ${DIM}IMAGE              ${NC}│${NC} ${DIM}STATUS   ${NC}│${NC} ${DIM}PORTS            ${NC}│${NC} ${GREEN}║${NC}"
    echo -e "${CYAN}║${NC}  ${GREEN}├────────────┼─────────────────────┼──────────┼─────────────────┤${NC}${CYAN} ║${NC}"
    
    echo "$CONTAINERS" | while IFS='|' read -r name image status ports; do
        [ -z "$name" ] && continue
        status_icon="$STATUS_HEALTHY"
        echo "$status" | grep -q "unhealthy" && status_icon="$STATUS_UNHEALTHY"
        echo "$status" | grep -q "starting" && status_icon="$STATUS_STARTING"
        name_short=$(echo "$name" | cut -c1-10)
        image_short=$(echo "$image" | sed 's|.*/||' | cut -c1-19)
        status_short=$(echo "$status" | cut -c1-8)
        ports_short=$(echo "$ports" | cut -c1-15)
        printf "${CYAN}║${NC}  ${GREEN}│${NC} ${status_icon} %-9s${NC}│${NC} %-19s${NC}│${NC} %-8s${NC}│${NC} %-15s${NC}│${NC} ${GREEN}║${NC}\n" "$name_short" "$image_short" "$status_short" "$ports_short"
    done
    echo -e "${CYAN}║${NC}  ${GREEN}└────────────┴─────────────────────┴──────────┴─────────────────┘${NC}${CYAN} ║${NC}"
else
    echo -e "${CYAN}║${NC}  ${DIM}No containers found${NC}${CYAN}                                              ║${NC}"
fi

echo -e "${CYAN}╠══════════════════════════════════════════════════════════════════╣${NC}"
echo -e "${CYAN}║${NC}  ${WHITE}📡 ROTEAMENTO${NC}${CYAN}                                                       ║${NC}"

if [ -n "$ROUTES" ] && [ "$ROUTES" != " " ]; then
    echo -e "${CYAN}║${NC}  ${WHITE}🌐${NC} ${DIM}${TAILSCALE_DOMAIN:-server.tailnet.ts.net}${NC}${CYAN}                                    ║${NC}"
    echo -e "${CYAN}║${NC}  ${CYAN}┌──────────────────────────────┬────────┬──────────────────────────────────────┐${NC}${CYAN} ║${NC}"
    echo -e "${CYAN}║${NC}  ${CYAN}│${NC} ${DIM}SERVIÇO${NC}                       ${CYAN}│${NC} ${DIM}TIPO${NC}  ${CYAN}│${NC} ${DIM}URL                                               ${NC}│${NC} ${CYAN}║${NC}"
    echo -e "${CYAN}║${NC}  ${CYAN}├──────────────────────────────┼────────┼──────────────────────────────────────┤${NC}${CYAN} ║${NC}"
    echo "$ROUTES" | while IFS='|' read -r rtype label target backend; do
        [ -z "$rtype" ] && continue
        if [ "$rtype" = "path" ]; then
            url="https://${TAILSCALE_DOMAIN}${target}"
        else
            url="https://${TAILSCALE_DOMAIN}${target}/"
        fi
        service_short=$(echo "$label" | cut -c1-30)
        type_short="path "
        [ "$rtype" = "port" ] && type_short="porta"
        url_short=$(echo "$url" | cut -c1-54)
        printf "${CYAN}║${NC}  ${CYAN}│${NC} %-30s${NC}${CYAN}│${NC} %-6s${NC}${CYAN}│${NC} ${WHITE}%-54s${NC}${CYAN}│${NC} ${CYAN}║${NC}\n" "$service_short" "$type_short" "$url_short"
    done
    echo -e "${CYAN}║${NC}  ${CYAN}└──────────────────────────────┴────────┴─────────────────────────────────────────────────────┘${NC}${CYAN} ║${NC}"
else
    echo -e "${CYAN}║${NC}  ${DIM}Nenhuma rota configurada${NC}${CYAN}                                          ║${NC}"
fi

echo -e "${CYAN}╠══════════════════════════════════════════════════════════════════╣${NC}"
echo -e "${CYAN}║${NC}  ${WHITE}🖼️  IMAGES${NC} (${IMAGE_COUNT} total)${CYAN}                                        ║${NC}"

if [ "$IMAGE_COUNT" -gt 0 ] && [ -n "$IMAGES" ]; then
    echo -e "${CYAN}║${NC}  ${BLUE}┌─────────────────────────────┬──────────┬─────────────┐${NC}${CYAN}       ║${NC}"
    echo -e "${CYAN}║${NC}  ${BLUE}│${NC} ${DIM}REPOSITORY                   ${NC}│${NC} ${DIM}TAG     ${NC}│${NC} ${DIM}SIZE       ${NC}│${NC} ${BLUE}║${NC}"
    echo -e "${CYAN}║${NC}  ${BLUE}├─────────────────────────────┼──────────┼─────────────┤${NC}${CYAN}       ║${NC}"
    echo "$IMAGES" | while IFS='|' read -r repo tag size; do
        [ -z "$repo" ] && continue
        repo_short=$(echo "$repo" | sed 's|^ghcr.io/||' | cut -c1-27)
        tag_short=$(echo "$tag" | cut -c1-8)
        printf "${CYAN}║${NC}  ${BLUE}│${NC} %-27s${NC}│${NC} %-8s${NC}│${NC} %-11s${NC}│${NC} ${BLUE}║${NC}\n" "$repo_short" "$tag_short" "$size"
    done
    echo -e "${CYAN}║${NC}  ${BLUE}└─────────────────────────────┴──────────┴─────────────┘${NC}${CYAN}       ║${NC}"
fi

echo -e "${CYAN}╠══════════════════════════════════════════════════════════════════╣${NC}"
echo -e "${CYAN}║${NC}  ${WHITE}📁 VOLUMES${NC} (${VOLUME_COUNT} in use)${CYAN}                                        ║${NC}"

if [ "$VOLUME_COUNT" -gt 0 ] && [ -n "$VOLUME_MOUNTS" ]; then
    echo -e "${CYAN}║${NC}  ${YELLOW}┌─────────────────┬──────────────┬──────────────────────────┐${NC}${CYAN}  ║${NC}"
    echo -e "${CYAN}║${NC}  ${YELLOW}│${NC} ${DIM}VOLUME           ${NC}│${NC} ${DIM}CONTAINER   ${NC}│${NC} ${DIM}MOUNT PATH              ${NC}│${NC} ${YELLOW}║${NC}"
    echo -e "${CYAN}║${NC}  ${YELLOW}├─────────────────┼──────────────┼──────────────────────────┤${NC}${CYAN}  ║${NC}"
    echo "$VOLUME_MOUNTS" | while IFS='|' read -r container vol path; do
        [ -z "$vol" ] && continue
        vol_short=$(echo "$vol" | cut -c1-15)
        container_short=$(echo "$container" | cut -c1-12)
        printf "${CYAN}║${NC}  ${YELLOW}│${NC} %-15s${NC}│${NC} %-12s${NC}│${NC} %-24s${NC}│${NC} ${YELLOW}║${NC}\n" "$vol_short" "$container_short" "$path"
    done
    echo -e "${CYAN}║${NC}  ${YELLOW}└─────────────────┴──────────────┴──────────────────────────┘${NC}${CYAN}  ║${NC}"
fi

echo -e "${CYAN}╠══════════════════════════════════════════════════════════════════╣${NC}"
echo -e "${CYAN}║${NC}  ${RED}⚠️  ORPHANED RESOURCES${NC}${CYAN}                                          ║${NC}"

if [ "${ORPHAN_CONTAINER_COUNT:-0}" -gt 0 ] && [ -n "$ORPHAN_CONTAINERS" ]; then
    echo -e "${CYAN}║${NC}  ${YELLOW}Stopped containers:${NC}${CYAN}                                        ║${NC}"
    echo "$ORPHAN_CONTAINERS" | while IFS='|' read -r name image; do
        [ -z "$name" ] && continue
        echo -e "${CYAN}║${NC}    - ${RED}${name}${NC} (${DIM}${image}${NC})${CYAN}                                ║${NC}"
    done
else
    echo -e "${CYAN}║${NC}  ${GREEN}✅ No orphaned containers${NC}${CYAN}                                        ║${NC}"
fi

if [ "${ORPHAN_IMAGE_COUNT:-0}" -gt 0 ] && [ -n "$ORPHAN_IMAGES" ]; then
    echo -e "${CYAN}║${NC}  ${YELLOW}Unused images:${NC}${CYAN}                                             ║${NC}"
    echo "$ORPHAN_IMAGES" | while IFS='|' read -r repo tag size; do
        [ -z "$repo" ] && continue
        echo -e "${CYAN}║${NC}    - ${RED}${repo}:${tag}${NC} (${DIM}${size}${NC})${CYAN}                            ║${NC}"
    done
else
    echo -e "${CYAN}║${NC}  ${GREEN}✅ No orphaned images${NC}${CYAN}                                           ║${NC}"
fi

if [ "${ORPHAN_VOLUME_COUNT:-0}" -gt 0 ] && [ -n "$ORPHAN_VOLUMES" ]; then
    echo -e "${CYAN}║${NC}  ${YELLOW}Dangling volumes:${NC}${CYAN}                                          ║${NC}"
    echo "$ORPHAN_VOLUMES" | while IFS= read -r vol; do
        [ -z "$vol" ] && continue
        echo -e "${CYAN}║${NC}    - ${RED}${vol}${NC}${CYAN}                                        ║${NC}"
    done
else
    echo -e "${CYAN}║${NC}  ${GREEN}✅ No orphaned volumes${NC}${CYAN}                                           ║${NC}"
fi

echo -e "${CYAN}╠══════════════════════════════════════════════════════════════════╣${NC}"
echo -e "${CYAN}║${NC}  ${WHITE}🌐 NETWORK${NC} (container IPs)${CYAN}                                    ║${NC}"

if [ -n "$CONTAINER_IPS" ]; then
    echo -e "${CYAN}║${NC}  ${CYAN}┌────────────┬────────────────┬─────────────────────────────┐${NC}${CYAN}  ║${NC}"
    echo -e "${CYAN}║${NC}  ${CYAN}│${NC} ${DIM}CONTAINER  ${NC}│${NC} ${DIM}IP ADDRESS    ${NC}│${NC} ${DIM}NETWORK                      ${NC}│${NC} ${CYAN}║${NC}"
    echo -e "${CYAN}║${NC}  ${CYAN}├────────────┼────────────────┼─────────────────────────────┤${NC}${CYAN}  ║${NC}"
    echo "$CONTAINER_IPS" | while IFS='|' read -r container ip network; do
        [ -z "$container" ] && continue
        printf "${CYAN}║${NC}  ${CYAN}│${NC} %-10s${NC}│${NC} %-14s${NC}│${NC} %-27s${NC}│${NC} ${CYAN}║${NC}\n" "$container" "$ip" "$network"
    done
    echo -e "${CYAN}║${NC}  ${CYAN}└────────────┴────────────────┴─────────────────────────────┘${NC}${CYAN}  ║${NC}"
else
    echo -e "${CYAN}║${NC}  ${DIM}No container network info available${NC}${CYAN}                          ║${NC}"
fi

# Cron translation function
translate_cron() {
    local min="$1" hour="$2" dom="$3" mon="$4" dow="$5"
    local result=""
    
    if [[ "$hour" == */* ]]; then
        result="Every ${hour#*/}h"
        echo "$result"
        return
    fi
    
    if [[ "$min" == */* ]]; then
        result="Every ${min#*/}m"
        echo "$result"
        return
    fi
    
    local time_str=$(printf "%02d:%02d" "$hour" "$min" 2>/dev/null || echo "${hour}:${min}")
    
    local dow_text=""
    case "$dow" in
        "*") dow_text="" ;;
        "1") dow_text="Mon" ;;
        "2") dow_text="Tue" ;;
        "3") dow_text="Wed" ;;
        "4") dow_text="Thu" ;;
        "5") dow_text="Fri" ;;
        "6") dow_text="Sat" ;;
        "0") dow_text="Sun" ;;
        "1,3,5") dow_text="Mon,Wed,Fri" ;;
        "0,6"|"6,0") dow_text="Sat,Sun" ;;
        "1-5") dow_text="Mon-Fri" ;;
        "1-6") dow_text="Mon-Sat" ;;
        "0-6") dow_text="Daily" ;;
        *) dow_text="$dow" ;;
    esac
    
    local mon_text=""
    case "$mon" in
        "*") mon_text="" ;;
        "1") mon_text="Jan" ;;
        "2") mon_text="Feb" ;;
        "3") mon_text="Mar" ;;
        "4") mon_text="Apr" ;;
        "5") mon_text="May" ;;
        "6") mon_text="Jun" ;;
        "7") mon_text="Jul" ;;
        "8") mon_text="Aug" ;;
        "9") mon_text="Sep" ;;
        "10") mon_text="Oct" ;;
        "11") mon_text="Nov" ;;
        "12") mon_text="Dec" ;;
        "1-12") mon_text="" ;;
        *) mon_text="$mon" ;;
    esac
    
    local dom_text=""
    case "$dom" in
        "*") dom_text="" ;;
        "1") dom_text="1st" ;;
        "2") dom_text="2nd" ;;
        "3") dom_text="3rd" ;;
        [4-9]|1[0-9]|2[0-9]|3[0-1]) dom_text="${dom}th" ;;
        *) dom_text="$dom" ;;
    esac
    
    if [ "$dom" = "*" ] && [ "$mon" = "*" ]; then
        if [ -n "$dow_text" ]; then
            result="$dow_text at $time_str"
        else
            result="Daily at $time_str"
        fi
    elif [ "$dow" = "*" ]; then
        [ -n "$mon_text" ] && result="$dom_text $mon_text at $time_str" || result="$dom_text of month at $time_str"
    elif [ "$mon" = "*" ]; then
        result="$dow_text ($dom_text) at $time_str"
    else
        result="$dow_text $dom_text/$mon_text at $time_str"
    fi
    
    echo "$result"
}

echo -e "${CYAN}╠══════════════════════════════════════════════════════════════════╣${NC}"
echo -e "${CYAN}║${NC}  ${WHITE}⏰ CRON JOBS${NC} (${CRON_COUNT} scheduled)${CYAN}                                      ║${NC}"

if [ "$CRON_COUNT" -gt 0 ] && [ -n "$CRON_JOBS" ]; then
    echo -e "${CYAN}║${NC}  ${MAGENTA}┌──────────────────────────┬──────────────────────────────────────┐${NC}${CYAN}  ║${NC}"
    echo -e "${CYAN}║${NC}  ${MAGENTA}│${NC} ${DIM}WHEN                     ${NC}│${NC} ${DIM}COMMAND                                 ${NC}│${NC} ${MAGENTA}║${NC}"
    echo -e "${CYAN}║${NC}  ${MAGENTA}├──────────────────────────┼──────────────────────────────────────┤${NC}${CYAN}  ║${NC}"
    
    while IFS= read -r line; do
        [ -z "$line" ] && continue
        min=$(echo "$line" | awk '{print $1}')
        hour=$(echo "$line" | awk '{print $2}')
        dom=$(echo "$line" | awk '{print $3}')
        mon=$(echo "$line" | awk '{print $4}')
        dow=$(echo "$line" | awk '{print $5}')
        cmd=$(echo "$line" | awk '{for(i=6;i<=NF;i++) printf $i" "; print ""}' | sed 's/ *$//')
        
        human_readable=$(translate_cron "$min" "$hour" "$dom" "$mon" "$dow")
        when_display=$(echo "$human_readable" | cut -c1-25)
        cmd_display=$(echo "$cmd" | cut -c1-42)
        
        printf "${CYAN}║${NC}  ${MAGENTA}│${NC} ${WHITE}%-25s${NC}│${NC} %-42s${NC}│${NC} ${MAGENTA}║${NC}\n" "$when_display" "$cmd_display"
    done <<< "$CRON_JOBS"
    
    echo -e "${CYAN}║${NC}  ${MAGENTA}└──────────────────────────┴──────────────────────────────────────┘${NC}${CYAN}  ║${NC}"
else
    echo -e "${CYAN}║${NC}  ${DIM}No cron jobs configured${NC}${CYAN}                                            ║${NC}"
fi

has_orphans=0
[ "${ORPHAN_CONTAINER_COUNT:-0}" -gt 0 ] 2>/dev/null && has_orphans=1
[ "${ORPHAN_IMAGE_COUNT:-0}" -gt 0 ] 2>/dev/null && has_orphans=1
[ "${ORPHAN_VOLUME_COUNT:-0}" -gt 0 ] 2>/dev/null && has_orphans=1

if [ "$has_orphans" -eq 1 ]; then
    echo -e "${CYAN}╠══════════════════════════════════════════════════════════════════╣${NC}"
    echo -e "${CYAN}║${NC}  ${YELLOW}🧹 CLEANUP SUGGESTIONS${NC}${CYAN}                                        ║${NC}"
    echo -e "${CYAN}║${NC}  Run on server:${NC}${CYAN}                                                 ║${NC}"
    [ "${ORPHAN_CONTAINER_COUNT:-0}" -gt 0 ] 2>/dev/null && echo -e "${CYAN}║${NC}    ${DIM}docker container prune -f${NC}${CYAN}                                  ║${NC}"
    [ "${ORPHAN_IMAGE_COUNT:-0}" -gt 0 ] 2>/dev/null && echo -e "${CYAN}║${NC}    ${DIM}docker image prune -a${NC}${CYAN}                                      ║${NC}"
    [ "${ORPHAN_VOLUME_COUNT:-0}" -gt 0 ] 2>/dev/null && echo -e "${CYAN}║${NC}    ${DIM}docker volume prune${NC}${CYAN}                                        ║${NC}"
fi

echo -e "${CYAN}╚══════════════════════════════════════════════════════════════════╝${NC}"
echo ""