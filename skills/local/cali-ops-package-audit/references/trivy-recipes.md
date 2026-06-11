# Trivy Scan Recipes

## Go Projects

```bash
# Dependencies only
trivy fs --scanners vuln --severity HIGH,CRITICAL .

# With secrets detection
trivy fs --scanners vuln,secret --severity HIGH,CRITICAL .

# Full (deps + config + secrets)
trivy fs --scanners vuln,secret,config --severity HIGH,CRITICAL .
```

## Node.js Projects

```bash
# package-lock.json
trivy fs --scanners vuln --severity HIGH,CRITICAL .

# Including .env files
trivy fs --scanners vuln,secret --severity HIGH,CRITICAL .
```

## Docker/Containers

```bash
# Scan built image
trivy image myapp:latest

# Scan Dockerfile for misconfig
trivy config Dockerfile
```

## CI Integration

```yaml
# GitHub Actions
- name: Run Trivy
  uses: aquasecurity/trivy-action@master
  with:
    scan-type: 'fs'
    scan-ref: '.'
    severity: 'HIGH,CRITICAL'
    exit-code: '1'
```

## Interpreting Results

- **HIGH/CRITICAL**: Fix immediately or document accepted risk
- **MEDIUM**: Review and schedule fix
- **LOW**: Note but don't block
- **Unknown severity**: Check NVD for CVSS score

## Remediation

```bash
# Update vulnerable dependency
go get -u <package>@latest    # Go
npm update <package>          # Node
pip install --upgrade <pkg>   # Python

# If no fix available:
# 1. Check if vulnerability applies to your usage
# 2. Document accepted risk in SECURITY.md
# 3. Monitor for fix
```
