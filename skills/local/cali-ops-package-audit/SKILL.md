---
name: cali-ops-package-audit
description: "Run supply chain security scans before installing packages or before releases. Triggers when: user installs a package (npm, pip, go get, brew), user asks to 'scan dependencies', 'check vulnerabilities', 'supply chain', 'security audit', 'run trivy', 'run socket', or before any release/deployment. Also triggers on mentions of: socket.dev, trivy, OSV-scanner, dotenvx, CVE, dependency audit. Covers all four tools with concrete commands."

metadata:
  frequency: weekly
  category: infra
  context-cost: low
---

# Supply Chain Security

Run before installing any package. Run again before every release.

## Tool Selection

| Tool | Purpose | When |
|------|---------|------|
| **Socket.dev** | Behavioral malware scanning | Before `npm install`, `pip install` |
| **Trivy** | CVE + IaC + secrets | Before releases, full audits |
| **OSV-Scanner** | Precision CVE (commit-hash) | When Trivy has false positives |
| **dotenvx** | Encrypted env vars | When managing `.env` files |

## 1. Socket.dev — Malware Detection

Detects obfuscated code, network access in install scripts, typosquatting.

```bash
# Session setup (run once per session)
socket wrapper on

# CI gate — blocks on malicious packages
socket ci

# Manual scan with report
socket scan create --report
```

**When to run:**
- Before `npm install` or `pip install` in any project
- When adding new dependencies
- In CI pipelines as a gate

## 2. Trivy — CVE + IaC + Secrets

Scans dependencies, infrastructure-as-code, secrets, containers. No account needed.

```bash
# Quick scan — high and critical only
trivy fs --severity HIGH,CRITICAL --exit-code 1 .

# Full scan with all detectors
trivy fs --scanners vuln,secret,config --severity HIGH,CRITICAL .

# Scan specific path
trivy fs --severity HIGH,CRITICAL ./package.json
```

**When to run:**
- Before every release
- On CI for every PR
- When user asks "any vulnerabilities?"

## 3. OSV-Scanner — Precision CVE

Google scanner — commit-hash matching for fewer false positives than Trivy.

```bash
# Scan project
osv-scanner scan -r .

# Guided remediation
osv-scanner fix -M package.json -L package-lock.json
```

**When to run:**
- When Trivy reports a CVE but you suspect false positive
- When user wants precise commit-level matching

## 4. dotenvx — Encrypted Environment Variables

```bash
# Inject envs from .env into command
dotenvx run -- <command>

# Set a value
dotenvx set KEY value

# Encrypt .env file (for safe commit)
dotenvx encrypt

# Decrypt .env file
dotenvx decrypt
```

**When to run:**
- When project has `.env` with secrets
- Before committing code that references env vars

## Workflow

1. **Identify context**: What is the user doing? Installing? Releasing? Auditing?
2. **Select tool(s)**: Match tool to context using table above
3. **Run with user confirmation**: Show command before executing
4. **Report results**: Summarize findings, suggest fixes

## Examples

### Example 1: Scan before npm install

**Input:** "I'm about to install stripe as a dependency"

**Steps:**
1. Run: `socket wrapper on`
2. Run: `npm install stripe`
3. Socket automatically blocks malicious packages

**Output:** "Socket scanned stripe — 0 issues found. Safe to use."

### Example 2: Pre-release audit

**Input:** "We're about to release v0.3.0, run security checks"

**Steps:**
1. Run: `trivy fs --severity HIGH,CRITICAL --exit-code 1 .`
2. If CVEs found: `osv-scanner scan -r .` to verify
3. Report results

**Output:** "Trivy found 0 HIGH/CRITICAL CVEs. Release is safe."

### Example 3: False positive investigation

**Input:** "Trivy flagged a CVE in lodash but I think it's a false positive"

**Steps:**
1. Run: `osv-scanner scan -r .` for commit-hash precision
2. Compare results

**Output:** "OSV-Scanner confirms false positive — your lodash version is not affected."

## Edge Cases

### Tool not installed
- Check: `which socket` / `which trivy` / `which osv-scanner`
- If missing: offer to install via `brew install <tool>` or skip with user consent
- Never silently skip a scan

### Multiple tools report different results
- Trivy says vulnerable, OSV-Scanner says safe → trust OSV-Scanner (commit-hash precision)
- Socket says malicious, but you're sure it's safe → document accepted risk, add to `.socketrc.json` ignore list

### No lock file present
- Trivy/OSV need lock files to scan accurately
- If no lock file: run `npm install` (or equivalent) first, then scan
- Warn user that scanning without lock file may miss transitive dependencies

## Test Cases

### Should activate
- "Scan my dependencies before install"
- "Run trivy on the project"
- "Check for vulnerabilities"
- "We're about to release, run security audit"
- "Install socket and scan"

### Should NOT activate
- "Install lodash" (just installing, no security request)
- "What is trivy?" (general question, not a scan request)
- "Fix the failing test" (unrelated task)

## References

- `references/socket-setup.md` — Socket.dev detailed setup and CI integration
- `references/trivy-recipes.md` — Trivy scan recipes for different stacks
