---
name: gitlab-security-setup
description: >
  Sets up a full dependency vulnerability scanning stack on projects hosted on
  GitLab.com Free tier (Laravel, Astro, TanStack, etc.). Use when the user asks
  to add dependency scanning, vulnerability alerts, security setup, Trivy, pnpm
  supply chain protection, or wants weekly email reports of vulnerabilities on a
  GitLab-hosted project. Do NOT use for GitHub-hosted projects or PrestaShop.
---

# GitLab Security Setup

Full security stack for Cinetic projects on GitLab.com Free tier.
Covers: pnpm 11 supply chain, Trivy weekly scan, HTML email reports via Gmail.

## What gets set up

1. **pnpm 11** with supply chain protection (`minimumReleaseAge`, overrides)
2. **Trivy** vulnerability + secret scanner via GitLab CI
3. **Weekly scheduled pipeline** (Monday 8am Madrid) with HTML email report
4. **Composer audit** for PHP/Laravel projects
5. **Gmail SMTP** delivery via GitLab CI/CD variables

---

## Step 1 — pnpm 11 Supply Chain

### `pnpm-workspace.yaml` (create or update)

```yaml
minimumReleaseAge: 2880   # packages must be 48h old before install (minutes)

overrides:
  form-data: ">=4.0.4"
  axios: ">=1.15.2"
  lodash: ">=4.18.0"
  picomatch: ">=4.0.4"
  qs: ">=6.14.2"
```

**Rules:**
- `minimumReleaseAge` is in **minutes** (2880 = 48h). Blocks supply chain attacks via typosquatting/fast-publish.
- `overrides` pins known vulnerable transitive deps. Add new entries as CVEs appear.
- Do NOT put `minimumReleaseAge` in `.npmrc` — pnpm 11 reads it from `pnpm-workspace.yaml` only.

### `package.json` additions

```json
{
  "packageManager": "pnpm@11.x.x"
}
```

Remove any `overrides` or `pnpm.overrides` blocks from `package.json` — they belong in `pnpm-workspace.yaml` for pnpm 11.

### `publicar` deploy script (if project has one)

```bash
#!/bin/bash
php artisan migrate --force
pnpm build
```

Ensure it uses `pnpm`, not `npm run`.

### GitHub Actions lint workflow (if exists)

Replace `npm ci` / `npm install` / `npm run` with:
```yaml
- run: npm install -g pnpm
- run: pnpm install --frozen-lockfile
- run: pnpm run format
- run: pnpm run lint
```

---

## Step 2 — GitLab CI Trivy Scan

Create or update `.gitlab-ci.yml`:

```yaml
dependency-scan:
  image:
    name: aquasec/trivy:latest
    entrypoint: [""]         # required — Trivy image has no shell by default
  before_script:
    - apk add --no-cache curl python3 py3-packaging
  script:
    # JSON report (structured data) — vuln + secret scan
    - trivy fs --exit-code 0 --scanners vuln,secret --format json -o trivy-report.json . 2>/dev/null

    # HTML report artifact
    - trivy fs --exit-code 0 --scanners vuln,secret --format template --template "@/contrib/html.tpl" -o trivy-report.html . 2>/dev/null || true

    # Parse JSON → build HTML email
    - |
      python3 << 'PYEOF'
      import json, os

      with open("trivy-report.json") as f:
          data = json.load(f)

      from packaging.version import Version, InvalidVersion

      def highest_fix(fixed_str):
          if not fixed_str or fixed_str == "N/A":
              return None
          parts = [p.strip() for p in fixed_str.split(",") if p.strip()]
          parsed = []
          for p in parts:
              try:
                  parsed.append((Version(p), p))
              except InvalidVersion:
                  parsed.append((Version("0"), p))
          return max(parsed, key=lambda x: x[0])[1] if parsed else None

      grouped = {}
      severity_order = {"CRITICAL": 0, "HIGH": 1, "MEDIUM": 2, "LOW": 3, "UNKNOWN": 4}

      for result in data.get("Results", []):
          for v in result.get("Vulnerabilities", []):
              key = (v.get("PkgName", ""), v.get("InstalledVersion", ""))
              fixed_raw = v.get("FixedVersion", "")
              sev = v.get("Severity", "UNKNOWN")
              cve = v.get("VulnerabilityID", "")
              fix = highest_fix(fixed_raw)

              if key not in grouped:
                  grouped[key] = {"pkg": key[0], "installed": key[1], "severity": sev, "cves": [], "fixes": []}

              entry = grouped[key]
              if severity_order.get(sev, 5) < severity_order.get(entry["severity"], 5):
                  entry["severity"] = sev
              if cve:
                  entry["cves"].append(cve)
              if fix:
                  entry["fixes"].append(fix)

      def max_version(versions):
          parsed = []
          for v in versions:
              try:
                  parsed.append((Version(v), v))
              except InvalidVersion:
                  pass
          return max(parsed, key=lambda x: x[0])[1] if parsed else None

      vulns = list(grouped.values())
      for entry in vulns:
          entry["best_fix"] = max_version(entry["fixes"])

      vulns.sort(key=lambda x: severity_order.get(x["severity"], 5))

      counts = {s: sum(1 for v in vulns if v["severity"] == s) for s in ["CRITICAL", "HIGH", "MEDIUM", "LOW"]}

      colors = {
          "CRITICAL": ("#ffeef0", "#d73a49"),
          "HIGH": ("#fff3cd", "#856404"),
          "MEDIUM": ("#e8f4fd", "#0c5460"),
          "LOW": ("#f0f0f0", "#555"),
      }

      rows = ""
      for v in vulns:
          bg, fg = colors.get(v["severity"], ("#fff", "#333"))
          fixed = v["best_fix"] if v["best_fix"] else "<em style='color:#888'>No fix yet</em>"
          cves_str = ", ".join(v["cves"][:3]) + (f" +{len(v['cves'])-3} more" if len(v["cves"]) > 3 else "")
          rows += f"""<tr>
            <td style='padding:8px;border-bottom:1px solid #eee'><code>{v['pkg']}</code></td>
            <td style='padding:8px;border-bottom:1px solid #eee'>{v['installed']}</td>
            <td style='padding:8px;border-bottom:1px solid #eee'><strong>{fixed}</strong></td>
            <td style='padding:8px;border-bottom:1px solid #eee'>
              <span style='background:{bg};color:{fg};padding:2px 8px;border-radius:4px;font-weight:bold;font-size:12px'>{v['severity']}</span>
            </td>
            <td style='padding:8px;border-bottom:1px solid #eee;font-size:11px;color:#555'>{cves_str}</td>
          </tr>"""

      badges = "".join([
          f"<span style='display:inline-block;padding:8px 16px;border-radius:6px;font-weight:bold;margin-right:10px;background:{colors[s][0]};color:{colors[s][1]};border:1px solid {colors[s][1]}'>{s}: {counts[s]}</span>"
          for s in ["CRITICAL", "HIGH", "MEDIUM", "LOW"] if counts[s] > 0
      ])

      project = os.environ.get("CI_PROJECT_NAME", "")
      branch = os.environ.get("CI_COMMIT_REF_NAME", "")
      sha = os.environ.get("CI_COMMIT_SHORT_SHA", "")
      pipeline_url = os.environ.get("CI_PIPELINE_URL", "#")
      gmail_user = os.environ.get("GMAIL_USER", "")

      table = "" if not vulns else f"""
      <table style='width:100%;border-collapse:collapse;margin-top:10px'>
        <thead>
          <tr style='background:#f6f8fa'>
            <th style='padding:10px;text-align:left;border-bottom:2px solid #e1e4e8'>Package</th>
            <th style='padding:10px;text-align:left;border-bottom:2px solid #e1e4e8'>Installed</th>
            <th style='padding:10px;text-align:left;border-bottom:2px solid #e1e4e8'>Fix version</th>
            <th style='padding:10px;text-align:left;border-bottom:2px solid #e1e4e8'>Severity</th>
            <th style='padding:10px;text-align:left;border-bottom:2px solid #e1e4e8'>CVE</th>
          </tr>
        </thead>
        <tbody>{rows}</tbody>
      </table>"""

      no_vulns = "<p style='color:green;font-weight:bold'>No vulnerabilities found.</p>" if not vulns else ""

      # Markdown block for Claude analysis
      md_lines = [
          f"# Security Scan — {project}",
          f"Branch: {branch} | Commit: {sha}",
          "",
          "## Vulnerabilities",
          "",
          "| Package | Installed | Fix version | Severity | CVEs |",
          "|---------|-----------|-------------|----------|------|",
      ]
      for v in vulns:
          fix = v["best_fix"] if v["best_fix"] else "No fix yet"
          cves = ", ".join(v["cves"])
          md_lines.append(f"| {v['pkg']} | {v['installed']} | {fix} | {v['severity']} | {cves} |")

      md_lines += [
          "",
          "## Task",
          "Review these vulnerabilities and suggest how to fix them in the codebase.",
          "For each package, check if it's a direct or transitive dependency and provide the exact command to update it.",
      ]
      md_content = "\n".join(md_lines)

      with open("trivy-report.md", "w") as f:
          f.write(md_content)

      md_escaped = md_content.replace("&", "&amp;").replace("<", "&lt;").replace(">", "&gt;")
      md_section = f"""
        <hr style='margin:30px 0;border:none;border-top:1px solid #e1e4e8'>
        <h3 style='color:#24292e'>Paste to Claude</h3>
        <p style='font-size:13px;color:#555'>Copia el bloque de abajo y pégalo en Claude para que analice y proponga los fixes:</p>
        <pre style='background:#f6f8fa;padding:15px;border-radius:6px;font-size:11px;white-space:pre-wrap;border:1px solid #e1e4e8'>{md_escaped}</pre>
      """

      html = f"""From: {gmail_user}\r\nTo: eduardo@cineticdigital.com, dgalera@cineticdigital.com\r\nSubject: [{project}] Security Scan — {counts.get('CRITICAL',0)} critical, {counts.get('HIGH',0)} high\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n
      <!DOCTYPE html><html><head></head><body style='font-family:Arial,sans-serif;max-width:900px;margin:0 auto;padding:20px;color:#333'>
        <h2 style='color:#24292e'>Security Scan — {project}</h2>
        <p>Branch: <strong>{branch}</strong> &nbsp;|&nbsp; Commit: <code>{sha}</code></p>
        <p>{badges}</p>
        {no_vulns}
        {table}
        {md_section}
        <p style='margin-top:20px'>
          <a href='{pipeline_url}' style='display:inline-block;padding:10px 20px;background:#1f75cb;color:white;text-decoration:none;border-radius:4px'>
            View Pipeline &amp; Download Artifacts
          </a>
        </p>
        <p style='font-size:12px;color:#888;margin-top:20px'>Generated by Trivy</p>
      </body></html>"""

      with open("email_body.txt", "w") as f:
          f.write(html)

      print(f"Vulnerabilities found: {len(vulns)} (CRITICAL: {counts['CRITICAL']}, HIGH: {counts['HIGH']})")
      PYEOF

    - |
      curl --url "smtps://smtp.gmail.com:465" \
        --ssl-reqd \
        --mail-from "$GMAIL_USER" \
        --mail-rcpt "eduardo@cineticdigital.com" \
        --mail-rcpt "dgalera@cineticdigital.com" \
        --user "$GMAIL_USER:$GMAIL_APP_PASS" \
        -T email_body.txt
  artifacts:
    paths:
      - trivy-report.html
      - trivy-report.json
      - trivy-report.md
    expire_in: 30 days
  rules:
    - if: $CI_PIPELINE_SOURCE == "schedule"
```

**Key implementation details:**
- `entrypoint: [""]` — mandatory; Trivy Docker image has no shell otherwise (exit code 127)
- `--exit-code 0` — never fail the pipeline; email even when clean
- `--scanners vuln,secret` — covers both dependency CVEs and leaked secrets
- `rules: schedule` — only runs on scheduled pipelines, not every push
- Vulnerabilities grouped by `(package, installed_version)` — one row per package, showing highest severity and best fix version
- `py3-packaging` via apk — not pip (pip is blocked in Alpine CI)

---

## Step 3 — Gmail CI/CD Variables

In GitLab project: **Settings → CI/CD → Variables**

| Variable | Value | Protected | Masked |
|----------|-------|-----------|--------|
| `GMAIL_USER` | `your-account@gmail.com` | No | No |
| `GMAIL_APP_PASS` | App password from Google | No | Yes |

**Getting Gmail App Password:**
1. Google Account → Security → 2-Step Verification (must be ON)
2. Search "App passwords" → Create → name it "GitLab CI"
3. Copy the 16-char password → paste as `GMAIL_APP_PASS`

---

## Step 4 — Weekly Scheduled Pipeline

Create via GitLab API (run once per project):

```bash
curl --request POST \
  --header "PRIVATE-TOKEN: <your-gitlab-token>" \
  "https://gitlab.com/api/v4/projects/<PROJECT_ID>/pipeline_schedules" \
  --form "description=Weekly security scan" \
  --form "ref=main" \
  --form "cron=0 7 * * 1" \
  --form "cron_timezone=Europe/Madrid"
```

- `0 7 * * 1` = Monday 07:00 UTC = 08:00/09:00 Madrid (winter/summer)
- `ref` = default branch (`main` or `develop`)
- `PROJECT_ID` = GitLab project → Settings → General

Or via UI: **CI/CD → Schedules → New schedule**

**Trigger manually to test:**

```bash
curl --request POST \
  --header "PRIVATE-TOKEN: <token>" \
  "https://gitlab.com/api/v4/projects/<PROJECT_ID>/pipeline_schedules/<SCHEDULE_ID>/play"
```

---

## Step 5 — PHP/Composer Projects (Laravel)

Add `composer audit` to the CI job's script block (before trivy):

```yaml
    - |
      if [ -f composer.json ]; then
        composer audit --format=plain 2>/dev/null || true
      fi
```

For fixing PHP vulnerabilities locally:

```bash
# Update specific packages
composer update "symfony/*" --with-all-dependencies

# Update all PHP deps (careful — test after)
composer update
```

Common PHP transitive dep CVEs — update these when flagged:
- `symfony/*` — update to latest patch on your major (e.g. `7.4.x`)
- `phpunit/phpunit` + `pestphp/pest` — must update together: `composer update phpunit/phpunit pestphp/pest --with-all-dependencies`
- `league/commonmark`, `psy/psysh` — `composer update <package>`

---

## Updating overrides for new CVEs

When the scan reports a fixable HIGH/CRITICAL on a transitive dep:

1. Check if it's in `pnpm-workspace.yaml` `overrides` already → bump version
2. If new package → add entry: `package-name: ">=fixed-version"`
3. Run `pnpm install` to regenerate lockfile
4. Commit + push → next weekly scan should show it resolved

**Skip alpha/RC fixes:** If the only fix is an alpha (e.g. `8.0.0-alpha.17`), skip — wait for stable release.

---

## Adapting for different project types

| Project type | Notes |
|---|---|
| **Laravel** | Include `composer audit` step; `publicar` = `php artisan migrate --force && pnpm build` |
| **Astro** | No composer step; `publicar` = `pnpm build` |
| **TanStack / pure frontend** | No composer step; check `pnpm-workspace.yaml` at root |
| **No Node** | Skip pnpm setup; Trivy still scans PHP deps |

---

## Checklist for new project

- [ ] `pnpm-workspace.yaml` created with `minimumReleaseAge: 2880` + overrides
- [ ] `package.json` has `"packageManager": "pnpm@11.x.x"`, no `overrides` block
- [ ] `.gitlab-ci.yml` has `dependency-scan` job (Step 2)
- [ ] GitLab CI/CD variables set: `GMAIL_USER`, `GMAIL_APP_PASS`
- [ ] Pipeline schedule created (Monday 8am Madrid)
- [ ] Manual trigger test → email received by both `eduardo@cineticdigital.com` and `dgalera@cineticdigital.com`
- [ ] `pnpm install` runs clean locally
