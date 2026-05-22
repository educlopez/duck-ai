# lando-setup

## Trigger

**Full setup** (project not yet configured locally):
"monta el proyecto", "setup lando", "prepara en local", "levanta el lando de X", "configura X en local"
→ Run all steps in order.

**DB refresh only** (Lando already running, just need fresh DB):
"actualízame la DB", "refresca la DB de X", "bájame la DB de X", "sync DB", "quiero la DB actualizada de X"
→ Skip to [DB Refresh Mode](#db-refresh-mode) below.

## Context
- Projects live at ~/developer/cinetic/{project-name}/
- SSH hosts in ~/.ssh/config follow pattern: {alias}-prod, {alias}-pre
- All projects are PrestaShop 8 or 9
- DB credentials in app/config/parameters.php (on the SERVER)
- Existing port range in use: 33290–33409 (pick next free one)

## SSH Authentication

### Cinetic SSH Keys (Eduardo's keys)
- `~/.ssh/id_rsa_cinetic` — RSA, `eduardo@cineticdigital.com` — **primary Cinetic key**
- `~/.ssh/id_ed25519_cinetic` — Ed25519, `eduardocalvolopez@Mac-m1-Edu.local` — modern alternative

Always use `-i ~/.ssh/id_rsa_cinetic` explicitly on all SSH/rsync commands to avoid key ambiguity.

### Auth modes

**Key-based (most projects):** Cinetic public key is on the server. No password needed.

**Password-based (some projects):** Key not on server. Use `sshpass` via stdin to avoid password in process list:
```bash
export SSHPASS='{password}'
sshpass -e ssh -i ~/.ssh/id_rsa_cinetic {user}@{ip} "command"
sshpass -e ssh -i ~/.ssh/id_rsa_cinetic {user}@{ip} "mysqldump ..." > /tmp/dump.sql
unset SSHPASS
```
Install if missing: `brew install sshpass`
**Never put the password directly in the command string** — use `-e` (reads from `$SSHPASS` env var) instead of `-p`.

**Adding new host to ~/.ssh/config** (do this when user gives raw `user@IP`):
```
Host {alias}-{env}
  HostName {IP}
  User {user}
  IdentityFile ~/.ssh/id_rsa_cinetic
```
Ask user if they want to save it for future use. If yes, append to ~/.ssh/config.

## Step 1 — Confirm inputs

Ask user (if not already provided):
- Project folder name (e.g. `ps8-pincolor`)
- SSH connection: existing alias OR `user@IP`
- Environment: **prod** or **pre** — ALWAYS ask, never assume
- Auth method: key (default) or password? If password, ask for it now (will store in env var, not shell string).
- Web server: **nginx** (lemp) or **Apache** (lamp)?
  - Default: **nginx** for PS9 on Ploi servers; **Apache** for PS8 or legacy setups.
  - If unsure: check server — Ploi = nginx, cPanel/Plesk = Apache.

Derive:
- `PROJECT_DIR=~/developer/cinetic/{project-name}`
- `SSH_TARGET={alias}-{env}` or `{user}@{ip}`
- `LANDO_NAME={project-name}` (same as folder)
- `LOCAL_URL={lando-name}.lndo.site`
- `LANDO_RECIPE=lemp` (nginx) or `lamp` (Apache)

**Test SSH connectivity before proceeding:**
```bash
ssh -i ~/.ssh/id_rsa_cinetic -o ConnectTimeout=5 -o BatchMode=yes {SSH_TARGET} "echo OK" 2>&1
```
If output is not `OK`, stop and diagnose auth failure with user before continuing.

## Step 2 — Detect PrestaShop version

```bash
grep -A2 '"prestashop/prestashop"' ~/developer/cinetic/{project-name}/composer.json
```

Or check:
```bash
grep "_PS_VERSION_" ~/developer/cinetic/{project-name}/config/defines.inc.php 2>/dev/null | head -1
```

- Version 8.x → PS8, PHP 8.1
- Version 9.x → PS9, PHP 8.3

If version is not clearly readable from files, ask user: "¿PS8 o PS9?" — do not guess from folder name.

## Step 3 — Find remote webroot

```bash
ssh -i ~/.ssh/id_rsa_cinetic {SSH_TARGET} "find /var/www -name 'parameters.php' -path '*/app/config/*' 2>/dev/null"
```

If multiple results appear, show them to user and ask which is the correct one.
Derive `REMOTE_ROOT` by stripping `/app/config/parameters.php` from the chosen path.
Example: `/var/www/html/app/config/parameters.php` → `REMOTE_ROOT=/var/www/html`

## Step 4 — Sync missing files from server

These files/folders are NOT in git and must be copied from the server via rsync.

### 4a — vendor/ (root, required)
```bash
rsync -avz --progress -e "ssh -i ~/.ssh/id_rsa_cinetic" \
  {SSH_TARGET}:{REMOTE_ROOT}/vendor/ ~/developer/cinetic/{project}/vendor/
```

### 4b — translations/ (required for PS9)
Without this folder, PS9 throws DirectoryNotFoundException on boot.
```bash
rsync -avz --progress -e "ssh -i ~/.ssh/id_rsa_cinetic" \
  {SSH_TARGET}:{REMOTE_ROOT}/translations/ ~/developer/cinetic/{project}/translations/
```

### 4c — app/config/ files (required)
If `~/developer/cinetic/{project}/app/config/parameters.php` already exists locally, warn user:
> "app/config/ exists locally. Sync from server will overwrite parameters.php. Continue? (y/n)"

```bash
rsync -avz --progress -e "ssh -i ~/.ssh/id_rsa_cinetic" \
  {SSH_TARGET}:{REMOTE_ROOT}/app/config/ ~/developer/cinetic/{project}/app/config/
```

### 4d — modules/ (required)
Sync the full modules directory. Modules not in git (theme modules, standard PS modules) won't exist locally otherwise.
```bash
rsync -avz --progress -e "ssh -i ~/.ssh/id_rsa_cinetic" \
  {SSH_TARGET}:{REMOTE_ROOT}/modules/ ~/developer/cinetic/{project}/modules/
```
This is simpler and safer than syncing individual vendor/ subdirs — covers theme modules (e.g. Panda) and any PS module not committed to git.

### 4e — img/ folder (OPTIONAL — ask user first)
The img/ folder can be several GB. Ask user:
> "¿Descargo la carpeta img/ del servidor o usamos placeholder de imágenes? (recomendado: placeholder)"

- If placeholder → trigger `lando-img-placeholder` skill after setup (it handles both lamp and lemp automatically)
- If download → `rsync -avz --progress -e "ssh -i ~/.ssh/id_rsa_cinetic" {SSH_TARGET}:{REMOTE_ROOT}/img/ ~/developer/cinetic/{project}/img/`

**Note for nginx (lemp) projects**: placeholder rules are built into `.lando/nginx-site.conf` — no separate Apache file needed. The `lando-img-placeholder` skill detects the recipe and acts accordingly.

**Regardless of choice (placeholder or download), always create required subdirs:**
```bash
mkdir -p ~/developer/cinetic/{project}/img/{c,p,m,st,cms,l,tmp,os,s,co,su}
```
Without these, `ps_mainmenu` crashes calling `scandir('/app/img/c/')` on boot.

## Step 5 — Update parameters.php with Lando credentials

After syncing app/config/ from server, parameters.php points to the remote DB → instant 500.
Update it with local Lando values:

```bash
sed -i '' \
  -e "s/'database_host' => '[^']*'/'database_host' => 'database'/" \
  -e "s/'database_name' => '[^']*'/'database_name' => 'lamp'/" \
  -e "s/'database_user' => '[^']*'/'database_user' => 'lamp'/" \
  -e "s/'database_password' => '[^']*'/'database_password' => 'lamp'/" \
  -e "s/'mailer_host' => '[^']*'/'mailer_host' => 'sendmailhog'/" \
  ~/developer/cinetic/{project}/app/config/parameters.php
```

Lando default DB credentials: host=`database`, name=`lamp`, user=`lamp`, password=`lamp`, mailer=`sendmailhog`.
Verify the file was updated correctly before proceeding.

## Step 6 — Get remote DB credentials

At this point `app/config/parameters.php` should be synced locally. Read it:
```bash
grep -E "database_(name|user|password|prefix)" ~/developer/cinetic/{project}/app/config/parameters.php
```

Extract:
- `database_name` → `DB_NAME`
- `database_user` → `DB_USER`
- `database_password` → `DB_PASS`
- `database_prefix` → `DB_PREFIX` (default `ps_` but MUST read actual value)

## Step 7 — Dump DB from server

Key-based:
```bash
ssh -i ~/.ssh/id_rsa_cinetic {SSH_TARGET} \
  "mysqldump -h 127.0.0.1 -u'${DB_USER}' -p'${DB_PASS}' '${DB_NAME}'" > ~/developer/cinetic/{project-name}/dump.sql
```

Password-based (SSH password):
```bash
export SSHPASS='{ssh_password}'
sshpass -e ssh -i ~/.ssh/id_rsa_cinetic {user}@{ip} \
  "mysqldump -h 127.0.0.1 -u'${DB_USER}' -p'${DB_PASS}' '${DB_NAME}'" > ~/developer/cinetic/{project-name}/dump.sql
unset SSHPASS
```

Verify dump:
```bash
wc -c ~/developer/cinetic/{project-name}/dump.sql
```
If file is under 100KB, likely failed or empty — check for errors before continuing.

## Step 8 — Pick available port

```bash
grep -r "portforward" ~/developer/cinetic/*/.lando.yml 2>/dev/null
```

Pick the next free port above the highest found (start at 33290 if none found).

## Step 9 — Generate .lando.yml

Create `~/developer/cinetic/{project-name}/.lando.yml`:

**If this project was previously set up with a different MySQL version**, the old Docker volume may be corrupt (error: "Upgrade is not supported after a crash or shutdown with innodb_fast_shutdown = 2"). Fix: run `lando destroy -y` before `lando start`. This deletes the local DB volume — safe since we're importing a fresh dump anyway.

First, create the PHP config override (prevents `max_execution_time` errors during theme install/parsing):
```bash
mkdir -p ~/developer/cinetic/{project-name}/.lando
```

Create `~/developer/cinetic/{project-name}/.lando/php.ini`:
```ini
max_execution_time = 300
max_input_time = 300
memory_limit = 512M
post_max_size = 64M
upload_max_filesize = 64M
```

**For PS8 (PHP 8.1):**
```yaml
name: {lando-name}
recipe: lamp
config:
  php: "8.1"
  webroot: ./
  database: mysql
  xdebug: false
services:
  appserver:
    config:
      php: .lando/php.ini
  database:
    portforward: {next-free-port}
    type: mysql
  mail:
    type: mailhog
    hogfrom:
      - appserver
```

**For PS9 with Apache (PHP 8.3):**
```yaml
name: {lando-name}
recipe: lamp
config:
  php: "8.3"
  webroot: ./
  database: mysql
  xdebug: false
services:
  appserver:
    config:
      php: .lando/php.ini
  database:
    portforward: {next-free-port}
    type: mysql
  mail:
    type: mailhog
    hogfrom:
      - appserver
```

**For PS9 with nginx (PHP 8.3) — default for Ploi servers:**

First create `.lando/nginx-site.conf`:
```nginx
server {
    listen 80;
    root /app;
    index index.php index.html;
    client_max_body_size 128M;

    charset utf-8;

    # Placeholder — friendly URLs: /{id}-{type}_default/...
    location ~* ^/[0-9]+-[a-z]+_default.*\.(jpg|jpeg|png|gif|webp)$ {
        try_files $uri /img/placeholder-dev.jpg;
    }

    # Placeholder — classic PS paths: /img/p/, /img/c/, /img/m/, /img/st/
    location ~* ^/img/(p|c|m|st)/.*\.(jpg|jpeg|png|gif|webp)$ {
        try_files $uri /img/placeholder-dev.jpg;
    }

    # PS9 double-admin-path fix (legacy module redirect bug with Symfony router)
    rewrite ^/(admin[-_\w]*)/.+/admin[-_\w]*/index\.php$ /$1/index.php last;

    # Product image SEO URL → real disk path
    rewrite ^/([0-9])(\-[\w-]+)?/.+\.(jpg|jpeg|png|gif|webp)$           /img/p/$1/$1$2.$3 last;
    rewrite ^/([0-9])([0-9])(\-[\w-]+)?/.+\.(jpg|jpeg|png|gif|webp)$    /img/p/$1/$2/$1$2$3.$4 last;
    rewrite ^/([0-9])([0-9])([0-9])(\-[\w-]+)?/.+\.(jpg|jpeg|png|gif|webp)$ /img/p/$1/$2/$3/$1$2$3$4.$5 last;
    rewrite ^/([0-9])([0-9])([0-9])([0-9])(\-[\w-]+)?/.+\.(jpg|jpeg|png|gif|webp)$ /img/p/$1/$2/$3/$4/$1$2$3$4$5.$6 last;
    rewrite ^/([0-9])([0-9])([0-9])([0-9])([0-9])(\-[\w-]+)?/.+\.(jpg|jpeg|png|gif|webp)$ /img/p/$1/$2/$3/$4/$5/$1$2$3$4$5$6.$7 last;
    rewrite ^/([0-9])([0-9])([0-9])([0-9])([0-9])([0-9])(\-[\w-]+)?/.+\.(jpg|jpeg|png|gif|webp)$ /img/p/$1/$2/$3/$4/$5/$6/$1$2$3$4$5$6$7.$8 last;

    # Category images
    rewrite ^/c/([0-9]+\-[\w-]+)\.(jpg|jpeg|png|gif|webp)$ /img/c/$1.$2 last;

    # Sitemap
    rewrite ^/(\w+)-sitemap\.xml$ /sitemap.php?lang=$1 last;

    # Block sensitive directories
    location ~* ^/(config|app|bin|src|var|vendor)(/|$) { deny all; }
    location ~* /\.env { deny all; }
    location ~* ^/(composer\.(json|lock)|package(-lock)?\.json|yarn\.lock|Makefile)$ { deny all; }
    location ~* \.(bak|sql|log|twig)$ { deny all; }
    location ~* ^/(upload|img)/.*\.php[0-9]?$ { deny all; }

    # Admin: non-index PHP files (filemanager/dialog.php, etc.)
    location ~* ^/admin[-_\w]*/(?!index\.php).+\.php(/|$) {
        fastcgi_pass fpm:9000;
        fastcgi_split_path_info ^(.+\.php)(/.*)$;
        fastcgi_buffers 32 32k;
        fastcgi_buffer_size 32k;
        fastcgi_index index.php;
        fastcgi_param SCRIPT_FILENAME $document_root$fastcgi_script_name;
        fastcgi_param DOCUMENT_ROOT $document_root;
        fastcgi_read_timeout 300;
        include fastcgi_params;
    }

    # Admin: clean URLs (Symfony routes) → index.php
    location ~* ^/(admin[-_\w]*)/(?!index\.php)(.+)$ {
        try_files $uri $uri/ /$1/index.php/$2$is_args$args;
    }

    # Admin: index.php (PS9 Symfony routing entry point)
    location ~* ^/admin[-_\w]*/index\.php(/|$) {
        fastcgi_pass fpm:9000;
        fastcgi_split_path_info ^(.+\.php)(/.*)$;
        fastcgi_buffers 32 32k;
        fastcgi_buffer_size 32k;
        fastcgi_index index.php;
        fastcgi_param SCRIPT_FILENAME $document_root$fastcgi_script_name;
        fastcgi_param DOCUMENT_ROOT $document_root;
        fastcgi_read_timeout 300;
        include fastcgi_params;
    }

    # PS9 admin-api
    location /admin-api {
        try_files $uri $uri/ /admin-api/index.php$is_args$args;
    }

    # Static assets
    location ~* \.(gif|jpe?g|png|ico|svg|webp|css|js|woff2?|ttf|eot|otf)$ {
        try_files $uri /index.php?$query_string;
        expires 1d;
        access_log off;
        log_not_found off;
    }

    # Front → PS router
    location / {
        try_files $uri $uri/ /index.php?$query_string;
    }

    # PHP via FPM
    location ~ \.php$ {
        try_files $uri /index.php =404;
        fastcgi_pass fpm:9000;
        fastcgi_split_path_info ^(.+\.php)(/.+)$;
        fastcgi_buffers 32 32k;
        fastcgi_buffer_size 32k;
        fastcgi_index index.php;
        fastcgi_param SCRIPT_FILENAME $document_root$fastcgi_script_name;
        fastcgi_param DOCUMENT_ROOT $document_root;
        fastcgi_param HTTP_PROXY "";
        fastcgi_read_timeout 300;
        include fastcgi_params;
    }
}
```

Then create `.lando.yml`:
```yaml
name: {lando-name}
recipe: lemp
config:
  php: "8.3"
  webroot: ./
  database: mysql
  xdebug: false
services:
  appserver:
    config:
      vhosts: .lando/nginx-site.conf
      php: .lando/php.ini
  database:
    portforward: {next-free-port}
    type: mysql
  mail:
    type: mailhog
    hogfrom:
      - appserver
```

## Step 10 — Ensure OrbStack is running

OrbStack is the Docker runtime. If it's not running, Lando will fail or open Docker Desktop (not configured).

Check if OrbStack is running:
```bash
pgrep -x "OrbStack" > /dev/null && echo "RUNNING" || echo "STOPPED"
```

If STOPPED, launch it and wait for Docker to be ready:
```bash
open -a OrbStack
```
Then poll until Docker responds (up to 30s):
```bash
for i in $(seq 1 15); do
  docker info > /dev/null 2>&1 && echo "Docker ready" && break
  echo "Waiting for OrbStack... ($i)"
  sleep 2
done
```
If Docker is still not ready after 30s, stop and ask user to check OrbStack manually.

## Step 11 — Start Lando

```bash
cd ~/developer/cinetic/{project-name} && lando start
```

After start, verify PHP version matches expected:
```bash
cd ~/developer/cinetic/{project-name} && lando php -v
```
If version doesn't match (e.g. shows 8.1 but expected 8.3), stop and fix `.lando.yml` before importing DB.

## Step 12 — Import DB

```bash
cd ~/developer/cinetic/{project-name} && lando db-import dump.sql
```

After import, verify tables exist:
```bash
cd ~/developer/cinetic/{project-name} && lando mysql -e "SELECT COUNT(*) FROM ${DB_PREFIX}configuration;" 2>/dev/null
```
If this returns 0 or errors, DO NOT delete the dump file — report failure to user and stop.

## Step 13 — Fix URLs in DB

Use the actual `DB_PREFIX` read from parameters.php (Step 6), not hardcoded `ps_`:

```bash
cd ~/developer/cinetic/{project-name} && lando mysql -e "
UPDATE ${DB_PREFIX}configuration SET value = '{lando-name}.lndo.site' WHERE name = 'PS_SHOP_DOMAIN';
UPDATE ${DB_PREFIX}configuration SET value = '{lando-name}.lndo.site' WHERE name = 'PS_SHOP_DOMAIN_SSL';
UPDATE ${DB_PREFIX}shop_url SET domain = '{lando-name}.lndo.site', domain_ssl = '{lando-name}.lndo.site';
UPDATE ${DB_PREFIX}configuration SET value = 0 WHERE name = 'PS_SSL_ENABLED';
UPDATE ${DB_PREFIX}configuration SET value = 0 WHERE name = 'PS_SSL_ENABLED_EVERYWHERE';
"
```

## Step 14 — Clear cache

```bash
cd ~/developer/cinetic/{project-name} && lando php bin/console cache:clear
```

If that fails, ask user before deleting:
> "cache:clear falló. ¿Borro var/cache/ manualmente? (y/n)"

If yes:
```bash
rm -rf ~/developer/cinetic/{project-name}/var/cache/*
```

## Step 15 — Done

Clean dump only after successful import (verified in Step 12):
```bash
rm ~/developer/cinetic/{project-name}/dump.sql
```

Report to user:
- Local URL: https://{lando-name}.lndo.site
- Admin URL: https://{lando-name}.lndo.site/admin{suffix} (check admin folder name in project root)
- DB port: {port} (for TablePlus/phpMyAdmin)

---

## DB Refresh Mode

Use when Lando is already set up and running. Only refreshes the database.

### Confirm inputs
- Project folder (e.g. `ps8-pincolor`)
- SSH host (check ~/.ssh/config — use existing alias if available)
- Environment: **prod** or **pre** — ALWAYS ask, never assume

### R0 — Read local credentials
Check parameters.php exists:
```bash
test -f ~/developer/cinetic/{project}/app/config/parameters.php || echo "MISSING"
```
If MISSING: stop and tell user to run full setup or sync app/config from server first.

Read DB credentials and prefix:
```bash
grep -E "database_(name|user|password|prefix)" ~/developer/cinetic/{project}/app/config/parameters.php
```

### R1 — Test SSH + Dump from server

Test connectivity first:
```bash
ssh -i ~/.ssh/id_rsa_cinetic -o ConnectTimeout=5 -o BatchMode=yes {SSH_TARGET} "echo OK"
```

Then dump (key-based):
```bash
ssh -i ~/.ssh/id_rsa_cinetic {SSH_TARGET} \
  "mysqldump -h 127.0.0.1 -u'${DB_USER}' -p'${DB_PASS}' '${DB_NAME}'" > ~/developer/cinetic/{project-name}/dump.sql
```

Verify: `wc -c ~/developer/cinetic/{project-name}/dump.sql` — must be > 100KB.

### R2 — Import (drop + recreate)

```bash
cd ~/developer/cinetic/{project-name} && lando db-import dump.sql
```

Verify tables after import:
```bash
cd ~/developer/cinetic/{project-name} && lando mysql -e "SELECT COUNT(*) FROM ${DB_PREFIX}configuration;"
```
If fails, keep dump and report to user.

### R3 — Fix URLs

Use actual `DB_PREFIX` from parameters.php:
```bash
cd ~/developer/cinetic/{project-name} && lando mysql -e "
UPDATE ${DB_PREFIX}configuration SET value = '{lando-name}.lndo.site' WHERE name = 'PS_SHOP_DOMAIN';
UPDATE ${DB_PREFIX}configuration SET value = '{lando-name}.lndo.site' WHERE name = 'PS_SHOP_DOMAIN_SSL';
UPDATE ${DB_PREFIX}shop_url SET domain = '{lando-name}.lndo.site', domain_ssl = '{lando-name}.lndo.site';
UPDATE ${DB_PREFIX}configuration SET value = 0 WHERE name = 'PS_SSL_ENABLED';
UPDATE ${DB_PREFIX}configuration SET value = 0 WHERE name = 'PS_SSL_ENABLED_EVERYWHERE';
"
```

### R4 — Clear cache

```bash
cd ~/developer/cinetic/{project-name} && lando php bin/console cache:clear
```

### R5 — Done
Delete dump (only if import verified): `rm ~/developer/cinetic/{project-name}/dump.sql`. Report URL.

---

## Notes

- ALWAYS confirm prod vs pre with user before SSHing anywhere
- Never run destructive SQL (DROP, TRUNCATE) without explicit user confirmation
- Large DBs (>500MB) can take 5+ min to import — normal
- If shop shows wrong theme/images: check module status or trigger `lando-img-placeholder` skill
- If `lando db-import` hangs > 30min: check disk space with `lando exec appserver df -h`
- `lando db-import` only accepts paths inside the project directory — always dump to project root as `dump.sql`
