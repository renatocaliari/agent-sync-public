---
name: cali-ops-docker-server-dashboard
description: >
  [Cali] -
  INTERACTIVE SKILL: Discover servers from ~/.ssh/config, auto-detect hosts
  (filtering out non-server entries like github.com), prompt user to pick one
  via question tool, then SSH into the chosen server and render a real-time
  ASCII dashboard with Docker containers, images, volumes, routes, cron,
  orphaned resources, and cleanup suggestions. REQUIRES question tool, SSH
  config parsing, and shell execution.
---

# Docker Server Dashboard Skill

## 🔴 CRITICAL: THIS IS AN INTERACTIVE AUTO-DISCOVERY SKILL

When this skill is loaded, you MUST follow these steps **in order** without asking the user what to do next:

### STEP 1 — Discover servers
- Read `~/.ssh/config`
- Extract every `Host` entry
- **FILTER OUT** hosts where `HostName` contains `github.com` or `User` is `git`
- The remaining entries are your server candidates

### STEP 2 — Ask user which server (MANDATORY)
Use the `question` tool to let the user pick a server:
```js
question({
  questions: [{
    header: "Selecionar Servidor",
    question: "Qual servidor você quer monitorar?",
    options: [
      // Populate dynamically from ~/.ssh/config
      // Example:
      { label: "server.renatocaliari.com (100.120.175.47)", description: "Servidor Ubuntu com Docker" },
    ]
    // "Type your own answer" comes automatically
  }]
})
```

### STEP 3 — Run the dashboard
```bash
DEPLOY_SERVER=<user-selected-server> bash /Users/cali/.agents/skills/cali-docker-server-dashboard/references/dashboard.sh
```

---

## Direct execution (without agent)

```bash
DEPLOY_SERVER=root@your-server.com ~/.agents/skills/cali-docker-server-dashboard/references/dashboard.sh
```

---

## What the dashboard shows

| Section | Information |
|---------|-------------|
| **Server Info** | Hostname, OS, Docker version, uptime |
| **Disk Usage** | Total, used, available with progress bar |
| **Memory Usage** | Total, used, available with progress bar |
| **Containers** | Name, image, status (🟢/🟡/🔴), ports |
| **Rotas (Caddy)** | Serviço, tipo (path/porta), URL completa com Tailscale |
| **Images** | Repository, tag, size |
| **Volumes** | Name, linked container, mount path |
| **Network** | Container IPs and network names |
| **Cron Jobs** | Scheduled tasks with human-readable schedule |
| **Orphaned Resources** | Stopped containers, unused images, dangling volumes |
| **Cleanup Suggestions** | Commands to remove orphaned resources |

## Features

- ✅ **CONTAINS 3 MANDATORY STEPS ABOVE** — do not skip, do not ask user what to do next
- ✅ Beautiful ASCII tables with colors
- ✅ Health status indicators (🟢 healthy, 🟡 starting, 🔴 unhealthy)
- ✅ Visual progress bars for disk/memory
- ✅ Warnings when usage >80%
- ✅ **Server auto-discovery** — reads `~/.ssh/config`, filters non-servers
- ✅ **Routing table auto-discovery** — parses Caddyfile and cross-references with Docker containers
- ✅ **Tailscale hostname dynamic** — auto-discovers via `tailscale status --json`
- ✅ **Single SSH connection** — all data collected in 1 connection (was 18+ before)
- ✅ URLs copiáveis prontas
