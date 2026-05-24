---
name: ps-security-audit
description: >
  Sets up weekly automated security scanning for PrestaShop 8 projects hosted on GitLab.com.
  Checks: installed modules vs Friends of Presta advisory database, PrestaShop core version,
  and PHP/Composer dependencies via Trivy. Sends weekly HTML email report every Monday.
  Use ONLY for Cinetic PrestaShop projects on GitLab.com. Triggers when user asks for
  security scanning, vulnerability alerts, module CVE check, or Friends of Presta integration
  on a PrestaShop project.
version: "0.1.0"
---

# PrestaShop Security Audit

Weekly automated security scan for PrestaShop 8 projects on GitLab.com Free tier.

Covers:
1. **Friends of Presta (FoP)** module advisory check — cross-references installed modules against known CVEs
2. **PrestaShop core version** — detects outdated core vs latest stable
3. **Trivy** — scans PHP/Composer dependencies for CVEs and leaked secrets
4. **Weekly HTML email** — Monday 8am Madrid, same format as other Cinetic projects

---

## What's unique about PrestaShop security

The biggest attack vector in recent PS vulnerabilities is **third-party modules** (marketplace and premium themes) with SQL injection, path traversal, and file upload flaws. Friends of Presta maintains the authoritative CVE list at:
- GitHub: `https://github.com/friends-of-presta/security-advisories`
- Site: `https://security.friendsofpresta.org`

Most attacks exploit `id_cart`, `id_product`, `id_address` parameters without `pSQL()` sanitization in module controllers. Trivy does NOT catch these — only FoP cross-referencing does.

---

## Step 1 — GitLab CI Job

Create or update `.gitlab-ci.yml` in the project root:

```yaml
ps-security-audit:
  image:
    name: aquasec/trivy:latest
    entrypoint: [""]
  before_script:
    - apk add --no-cache curl python3 py3-packaging py3-yaml tar
  script:
    # --- Trivy scan (PHP/Composer deps + secrets) ---
    - trivy fs --exit-code 0 --scanners vuln,secret --format json -o trivy-report.json . 2>/dev/null
    - trivy fs --exit-code 0 --scanners vuln,secret --format template --template "@/contrib/html.tpl" -o trivy-report.html . 2>/dev/null || true

    # --- Main analysis: FoP module check + PS core version + email ---
    - |
      python3 << 'PYEOF'
      import json, os, re, glob, tarfile, urllib.request, yaml
      from packaging.version import Version, InvalidVersion

      # ── 1. PrestaShop core version ────────────────────────────────────
      ps_version = "unknown"
      defines_file = "config/defines.inc.php"
      if os.path.exists(defines_file):
          with open(defines_file, encoding="utf-8", errors="ignore") as f:
              m = re.search(r"_PS_VERSION_['\"]?\s*,\s*['\"]([^'\"]+)['\"]", f.read())
              if m:
                  ps_version = m.group(1)

      # Fetch latest PS stable from GitHub API
      ps_latest = "unknown"
      try:
          req = urllib.request.Request(
              "https://api.github.com/repos/PrestaShop/PrestaShop/releases/latest",
              headers={"User-Agent": "ps-security-audit-ci"}
          )
          with urllib.request.urlopen(req, timeout=10) as r:
              release = json.loads(r.read().decode())
              ps_latest = release.get("tag_name", "unknown").lstrip("v")
      except Exception as e:
          print(f"Warning: could not fetch PS latest version: {e}")

      ps_outdated = False
      try:
          ps_outdated = Version(ps_version) < Version(ps_latest)
      except InvalidVersion:
          pass

      print(f"PS core: {ps_version} | Latest: {ps_latest} | Outdated: {ps_outdated}")

      # ── 2. Installed modules + versions ──────────────────────────────
      installed = {}
      for mod_dir in glob.glob("modules/*/"):
          mod_name = os.path.basename(mod_dir.rstrip("/"))
          php_file = os.path.join(mod_dir, f"{mod_name}.php")
          if os.path.exists(php_file):
              with open(php_file, encoding="utf-8", errors="ignore") as f:
                  content = f.read()
              m = re.search(r"\$this->version\s*=\s*['\"]([^'\"]+)['\"]", content)
              installed[mod_name] = m.group(1) if m else "unknown"

      print(f"Found {len(installed)} modules installed")

      # ── 3. Fetch Friends of Presta advisories ────────────────────────
      advisories = {}
      try:
          urllib.request.urlretrieve(
              "https://github.com/friends-of-presta/security-advisories/archive/refs/heads/main.tar.gz",
              "fop.tar.gz"
          )
          with tarfile.open("fop.tar.gz") as tar:
              for member in tar.getmembers():
                  if "/_modules/" in member.name and member.name.endswith(".md"):
                      f = tar.extractfile(member)
                      if not f:
                          continue
                      content = f.read().decode("utf-8", errors="ignore")
                      if not content.startswith("---"):
                          continue
                      parts = content.split("---", 2)
                      if len(parts) < 3:
                          continue
                      try:
                          meta = yaml.safe_load(parts[1])
                          if not isinstance(meta, dict):
                              continue
                          mod_name = meta.get("module", "")
                          if mod_name:
                              advisories.setdefault(mod_name, []).append(meta)
                      except Exception:
                          pass
          print(f"Loaded {len(advisories)} FoP advisories")
      except Exception as e:
          print(f"Warning: could not fetch FoP advisories: {e}")

      # ── 4. Cross-reference modules vs advisories ──────────────────────
      def is_affected(installed_ver, affected_versions):
          try:
              iv = Version(str(installed_ver))
              for spec in (affected_versions or []):
                  spec = str(spec).strip()
                  if spec.startswith("<= "):
                      if iv <= Version(spec[3:].strip()): return True
                  elif spec.startswith("< "):
                      if iv < Version(spec[2:].strip()): return True
                  elif spec.startswith(">= "):
                      if iv >= Version(spec[3:].strip()): return True
                  elif spec.startswith("== ") or spec.startswith("= "):
                      if iv == Version(spec.lstrip("=").strip()): return True
          except InvalidVersion:
              pass
          return False

      severity_order = {"critical": 0, "high": 1, "medium": 2, "low": 3, "unknown": 4}

      vulnerable_modules = []
      for mod_name, installed_ver in installed.items():
          if mod_name not in advisories:
              continue
          for advisory in advisories[mod_name]:
              affected = advisory.get("affected_versions", [])
              if not affected or is_affected(installed_ver, affected):
                  fixed = advisory.get("fixed_versions", [])
                  fixed_str = ", ".join(str(v) for v in fixed) if fixed else "No fix yet"
                  cves = advisory.get("additional_cve") or []
                  if isinstance(cves, str):
                      cves = [cves]
                  sev = str(advisory.get("severity", "unknown")).lower()
                  vulnerable_modules.append({
                      "module": mod_name,
                      "installed": installed_ver,
                      "fixed": fixed_str,
                      "severity": sev,
                      "cvss": advisory.get("cvss_base_score", ""),
                      "cves": [str(c) for c in cves],
                      "title": advisory.get("title", ""),
                  })

      vulnerable_modules.sort(key=lambda x: severity_order.get(x["severity"], 4))
      print(f"Vulnerable modules found: {len(vulnerable_modules)}")

      # ── 5. Parse Trivy results ────────────────────────────────────────
      trivy_vulns = []
      try:
          with open("trivy-report.json") as f:
              trivy_data = json.load(f)

          from packaging.version import Version as V, InvalidVersion as IE

          def best_fix(fixed_str):
              if not fixed_str or fixed_str == "N/A":
                  return None
              parts = [p.strip() for p in fixed_str.split(",") if p.strip()]
              parsed = []
              for p in parts:
                  try: parsed.append((V(p), p))
                  except IE: parsed.append((V("0"), p))
              return max(parsed, key=lambda x: x[0])[1] if parsed else None

          grouped = {}
          trivy_sev_order = {"CRITICAL": 0, "HIGH": 1, "MEDIUM": 2, "LOW": 3, "UNKNOWN": 4}
          for result in trivy_data.get("Results", []):
              for v in result.get("Vulnerabilities", []):
                  key = (v.get("PkgName", ""), v.get("InstalledVersion", ""))
                  sev = v.get("Severity", "UNKNOWN")
                  cve = v.get("VulnerabilityID", "")
                  fix = best_fix(v.get("FixedVersion", ""))
                  if key not in grouped:
                      grouped[key] = {"pkg": key[0], "installed": key[1], "severity": sev, "cves": [], "fixes": []}
                  e = grouped[key]
                  if trivy_sev_order.get(sev, 5) < trivy_sev_order.get(e["severity"], 5):
                      e["severity"] = sev
                  if cve: e["cves"].append(cve)
                  if fix: e["fixes"].append(fix)

          for entry in grouped.values():
              parsed = []
              for v in entry["fixes"]:
                  try: parsed.append((V(v), v))
                  except IE: pass
              entry["best_fix"] = max(parsed, key=lambda x: x[0])[1] if parsed else None
              trivy_vulns.append(entry)

          trivy_vulns.sort(key=lambda x: trivy_sev_order.get(x["severity"], 5))
      except Exception as e:
          print(f"Warning: could not parse Trivy results: {e}")

      # ── 6. Build email ────────────────────────────────────────────────
      fop_colors = {
          "critical": ("#ffeef0", "#d73a49"),
          "high": ("#fff3cd", "#856404"),
          "medium": ("#e8f4fd", "#0c5460"),
          "low": ("#f0f0f0", "#555"),
          "unknown": ("#f0f0f0", "#555"),
      }
      trivy_colors = {
          "CRITICAL": ("#ffeef0", "#d73a49"),
          "HIGH": ("#fff3cd", "#856404"),
          "MEDIUM": ("#e8f4fd", "#0c5460"),
          "LOW": ("#f0f0f0", "#555"),
          "UNKNOWN": ("#f0f0f0", "#555"),
      }

      project = os.environ.get("CI_PROJECT_NAME", "")
      branch = os.environ.get("CI_COMMIT_REF_NAME", "")
      sha = os.environ.get("CI_COMMIT_SHORT_SHA", "")
      pipeline_url = os.environ.get("CI_PIPELINE_URL", "#")
      gmail_user = os.environ.get("GMAIL_USER", "")

      # PS core section
      ps_badge_color = "#ffeef0" if ps_outdated else "#e6ffed"
      ps_badge_text_color = "#d73a49" if ps_outdated else "#22863a"
      ps_status = f"⚠️ OUTDATED — latest is {ps_latest}" if ps_outdated else f"✓ Up to date ({ps_latest})"
      ps_section = f"""
        <div style='background:{ps_badge_color};border:1px solid {ps_badge_text_color};border-radius:6px;padding:12px 16px;margin-bottom:20px'>
          <strong>PrestaShop Core:</strong> {ps_version}
          &nbsp;|&nbsp;
          <strong style='color:{ps_badge_text_color}'>{ps_status}</strong>
        </div>"""

      # FoP module table
      fop_rows = ""
      for v in vulnerable_modules:
          bg, fg = fop_colors.get(v["severity"], ("#fff", "#333"))
          cves_str = ", ".join(v["cves"][:3]) + (f" +{len(v['cves'])-3} more" if len(v["cves"]) > 3 else "")
          fop_rows += f"""<tr>
            <td style='padding:8px;border-bottom:1px solid #eee'><code>{v['module']}</code></td>
            <td style='padding:8px;border-bottom:1px solid #eee'>{v['installed']}</td>
            <td style='padding:8px;border-bottom:1px solid #eee'><strong>{v['fixed']}</strong></td>
            <td style='padding:8px;border-bottom:1px solid #eee'>
              <span style='background:{bg};color:{fg};padding:2px 8px;border-radius:4px;font-weight:bold;font-size:12px'>{v['severity'].upper()}</span>
              {f"<small style='color:#888;margin-left:4px'>CVSS {v['cvss']}</small>" if v['cvss'] else ""}
            </td>
            <td style='padding:8px;border-bottom:1px solid #eee;font-size:11px;color:#555'>{cves_str}</td>
          </tr>"""

      fop_table = "" if not vulnerable_modules else f"""
        <h3 style='color:#24292e;margin-top:24px'>Vulnerable Modules (Friends of Presta)</h3>
        <table style='width:100%;border-collapse:collapse;margin-top:10px'>
          <thead>
            <tr style='background:#f6f8fa'>
              <th style='padding:10px;text-align:left;border-bottom:2px solid #e1e4e8'>Module</th>
              <th style='padding:10px;text-align:left;border-bottom:2px solid #e1e4e8'>Installed</th>
              <th style='padding:10px;text-align:left;border-bottom:2px solid #e1e4e8'>Fix version</th>
              <th style='padding:10px;text-align:left;border-bottom:2px solid #e1e4e8'>Severity</th>
              <th style='padding:10px;text-align:left;border-bottom:2px solid #e1e4e8'>CVE</th>
            </tr>
          </thead>
          <tbody>{fop_rows}</tbody>
        </table>"""

      fop_ok = "" if vulnerable_modules else "<p style='color:green;font-weight:bold'>✓ No vulnerable modules found.</p>"

      # Trivy table
      trivy_counts = {s: sum(1 for v in trivy_vulns if v["severity"] == s) for s in ["CRITICAL", "HIGH", "MEDIUM", "LOW"]}
      trivy_rows = ""
      for v in trivy_vulns:
          bg, fg = trivy_colors.get(v["severity"], ("#fff", "#333"))
          fixed = v["best_fix"] if v["best_fix"] else "<em style='color:#888'>No fix yet</em>"
          cves_str = ", ".join(v["cves"][:3]) + (f" +{len(v['cves'])-3} more" if len(v["cves"]) > 3 else "")
          trivy_rows += f"""<tr>
            <td style='padding:8px;border-bottom:1px solid #eee'><code>{v['pkg']}</code></td>
            <td style='padding:8px;border-bottom:1px solid #eee'>{v['installed']}</td>
            <td style='padding:8px;border-bottom:1px solid #eee'><strong>{fixed}</strong></td>
            <td style='padding:8px;border-bottom:1px solid #eee'>
              <span style='background:{bg};color:{fg};padding:2px 8px;border-radius:4px;font-weight:bold;font-size:12px'>{v['severity']}</span>
            </td>
            <td style='padding:8px;border-bottom:1px solid #eee;font-size:11px;color:#555'>{cves_str}</td>
          </tr>"""

      trivy_badges = "".join([
          f"<span style='display:inline-block;padding:6px 12px;border-radius:6px;font-weight:bold;margin-right:8px;background:{trivy_colors[s][0]};color:{trivy_colors[s][1]};border:1px solid {trivy_colors[s][1]}'>{s}: {trivy_counts[s]}</span>"
          for s in ["CRITICAL", "HIGH", "MEDIUM", "LOW"] if trivy_counts.get(s, 0) > 0
      ])

      trivy_section = "" if not trivy_vulns else f"""
        <h3 style='color:#24292e;margin-top:24px'>PHP/Composer Dependencies (Trivy)</h3>
        <p>{trivy_badges or "<span style='color:green;font-weight:bold'>✓ No dependency vulnerabilities found.</span>"}</p>
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
          <tbody>{trivy_rows}</tbody>
        </table>"""

      # Markdown block for Claude
      md_lines = [
          f"# PS Security Scan — {project}",
          f"Branch: {branch} | Commit: {sha}",
          f"PrestaShop core: {ps_version} | Latest: {ps_latest}",
          "",
          "## Vulnerable Modules (Friends of Presta)",
          "",
          "| Module | Installed | Fix | Severity | CVEs |",
          "|--------|-----------|-----|----------|------|",
      ]
      for v in vulnerable_modules:
          cves = ", ".join(v["cves"])
          md_lines.append(f"| {v['module']} | {v['installed']} | {v['fixed']} | {v['severity'].upper()} | {cves} |")

      md_lines += [
          "",
          "## PHP/Composer Dependencies (Trivy)",
          "",
          "| Package | Installed | Fix | Severity | CVEs |",
          "|---------|-----------|-----|----------|------|",
      ]
      for v in trivy_vulns:
          fix = v["best_fix"] if v["best_fix"] else "No fix yet"
          cves = ", ".join(v["cves"])
          md_lines.append(f"| {v['pkg']} | {v['installed']} | {fix} | {v['severity']} | {cves} |")

      md_lines += [
          "",
          "## Task",
          "Review PS security findings. For vulnerable modules: check if update is available in marketplace or if module should be disabled. For Composer deps: provide exact update command.",
      ]
      md_content = "\n".join(md_lines)

      with open("ps-security-report.md", "w") as f:
          f.write(md_content)

      md_escaped = md_content.replace("&", "&amp;").replace("<", "&lt;").replace(">", "&gt;")
      md_section = f"""
        <hr style='margin:30px 0;border:none;border-top:1px solid #e1e4e8'>
        <h3 style='color:#24292e'>Paste to Claude</h3>
        <p style='font-size:13px;color:#555'>Copia el bloque y pégalo en Claude para análisis y fixes:</p>
        <pre style='background:#f6f8fa;padding:15px;border-radius:6px;font-size:11px;white-space:pre-wrap;border:1px solid #e1e4e8'>{md_escaped}</pre>"""

      fop_count = len(vulnerable_modules)
      trivy_critical = trivy_counts.get("CRITICAL", 0)
      trivy_high = trivy_counts.get("HIGH", 0)
      subject = f"[{project}] PS Security — {fop_count} módulos vulnerables, {trivy_critical} CVE críticos"

      html = f"""From: {gmail_user}\r\nTo: eduardo@cineticdigital.com, dgalera@cineticdigital.com\r\nSubject: {subject}\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n
      <!DOCTYPE html><html><head></head><body style='font-family:Arial,sans-serif;max-width:900px;margin:0 auto;padding:20px;color:#333'>
        <h2 style='color:#24292e'>PrestaShop Security Scan — {project}</h2>
        <p>Branch: <strong>{branch}</strong> &nbsp;|&nbsp; Commit: <code>{sha}</code></p>
        {ps_section}
        <p><strong>FoP modules vulnerables: {fop_count}</strong></p>
        {fop_ok}
        {fop_table}
        {trivy_section}
        {md_section}
        <p style='margin-top:20px'>
          <a href='{pipeline_url}' style='display:inline-block;padding:10px 20px;background:#1f75cb;color:white;text-decoration:none;border-radius:4px'>
            Ver Pipeline y Descargar Artefactos
          </a>
        </p>
        <p style='font-size:12px;color:#888;margin-top:20px'>Generado por Trivy + Friends of Presta</p>
      </body></html>"""

      with open("email_body.txt", "w") as f:
          f.write(html)

      print(f"Report ready: {fop_count} vulnerable modules, {trivy_critical} CRITICAL / {trivy_high} HIGH PHP deps")
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
      - ps-security-report.md
    expire_in: 30 days
  rules:
    - if: $CI_PIPELINE_SOURCE == "schedule"
```

---

## Step 2 — GitLab CI/CD Variables

**Settings → CI/CD → Variables** en el proyecto:

| Variable | Value | Masked |
|----------|-------|--------|
| `GMAIL_USER` | `cuenta@gmail.com` | No |
| `GMAIL_APP_PASS` | App password Google (16 chars) | Yes |

**Cómo obtener App Password Gmail:**
1. Google Account → Security → 2-Step Verification (debe estar ON)
2. Buscar "App passwords" → Create → nombre "GitLab CI"
3. Copiar los 16 caracteres → pegar como `GMAIL_APP_PASS`

---

## Step 3 — Pipeline Schedule (Monday 8am Madrid)

Via GitLab API (una sola vez por proyecto):

```bash
curl --request POST \
  --header "PRIVATE-TOKEN: <tu-gitlab-token>" \
  "https://gitlab.com/api/v4/projects/<PROJECT_ID>/pipeline_schedules" \
  --form "description=Weekly PS security scan" \
  --form "ref=main" \
  --form "cron=0 7 * * 1" \
  --form "cron_timezone=Europe/Madrid"
```

O via UI: **CI/CD → Schedules → New schedule**

**Test manual (tras crear el schedule):**

```bash
curl --request POST \
  --header "PRIVATE-TOKEN: <token>" \
  "https://gitlab.com/api/v4/projects/<PROJECT_ID>/pipeline_schedules/<SCHEDULE_ID>/play"
```

---

## Step 4 — PS project structure assumptions

El job asume estructura estándar PS8:
- `modules/<name>/<name>.php` — módulos instalados (lee `$this->version`)
- `config/defines.inc.php` — versión del core (`_PS_VERSION_`)
- `vendor/` — dependencias Composer (Trivy lo escanea automáticamente)

Si los módulos están en path distinto, ajustar el glob en el script Python.

---

## Interpreting results

**FoP module hits:**
- Si `fixed_versions` existe → actualizar módulo en PS backoffice o via ZIP desde proveedor
- Si `fixed_versions` está vacío → módulo abandonado, evaluar desactivar y buscar alternativa
- Si versión instalada no aparece en `affected_versions` → FoP lo lista pero revisar manualmente

**Trivy PHP hits:**
- `symfony/*` → `composer update "symfony/*" --with-all-dependencies`
- Paquete específico → `composer update <vendor/package>`
- `composer audit` para detalle local

**PS Core outdated:**
- Seguir guía oficial de actualización: backup DB + files → actualizar vía 1-click o CLI

---

## Checklist para nuevo proyecto PS

- [ ] `.gitlab-ci.yml` con job `ps-security-audit` (Step 1)
- [ ] Variables GitLab CI/CD: `GMAIL_USER`, `GMAIL_APP_PASS`
- [ ] Pipeline schedule creado (lunes 07:00 UTC)
- [ ] Test manual → email recibido con secciones: Core version, FoP modules, Trivy deps
- [ ] Verificar que `modules/` existe en el repositorio (no en `.gitignore`)

---

## Notes

- `entrypoint: [""]` en la imagen Trivy es **obligatorio** — la imagen no tiene shell por defecto
- `--exit-code 0` en Trivy — pipeline nunca falla, siempre envía email
- FoP descarga ~50MB de tar.gz en cada run; normal, no hay API paginada más ligera
- PyYAML viene pre-instalado en Alpine Python (`py3-yaml`), no necesita pip
- Módulos sin `$this->version` en su PHP (raro) reportan `unknown` — no se cruzan con FoP
