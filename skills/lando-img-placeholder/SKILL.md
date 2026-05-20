---
name: lando-img-placeholder
description: Sets up a static image placeholder for local Lando PrestaShop development so broken/missing product images show a nice placeholder instead of broken icons. Use this skill when the user says "placeholder de imagenes en lando", "lando image placeholder", "imagenes rotas en lando", "broken images in lando", or wants to avoid downloading hundreds of product images for local development. Trigger proactively whenever the user is working on a Lando PrestaShop project and mentions broken images or missing product photos.
version: "0.1.0"
---

# Lando Image Placeholder for PrestaShop

Intercepts missing product image requests via Apache rewrite rules and serves a static placeholder image. No PHP scripts, no DB changes — pure Apache mod_rewrite.

## How it works

A custom Apache VirtualHost config intercepts image requests where the file doesn't exist on disk and rewrites them to a single placeholder JPEG. This covers all PrestaShop image URL patterns.

## Steps

### 1. Create the `.lando` directory and Apache config

Create `.lando/apache-placeholder.conf` replacing `{project}` with the actual Lando project name (found in `.lando.yml` as the `name:` field, e.g. `vives-8`):

```apache
<VirtualHost *:80>
  ServerName {project}.lndo.site
  DocumentRoot /app
  <Directory /app>
    Options Indexes FollowSymLinks
    AllowOverride All
    Require all granted
  </Directory>
  RewriteEngine On
  # PS friendly image URLs: /{id}-{type}_default/{name}.jpg (product image types only)
  # Pattern requires "_default" suffix — matches home_default, large_default, etc.
  # This avoids catching slider/module images that also use numeric IDs in URLs.
  RewriteCond /app%{REQUEST_URI} !-f
  RewriteCond %{REQUEST_URI} ^/[0-9]+-[a-z]+_default
  RewriteCond %{REQUEST_URI} \.(jpg|jpeg|png|gif|webp)$
  RewriteRule ^ /img/placeholder-dev.jpg [L]
  # PS traditional image paths: /img/p/, /img/c/, /img/m/, /img/st/
  RewriteCond /app%{REQUEST_URI} !-f
  RewriteCond %{REQUEST_URI} ^/img/(p|c|m|st)/
  RewriteCond %{REQUEST_URI} \.(jpg|jpeg|png|gif|webp)$
  RewriteRule ^ /img/placeholder-dev.jpg [L]
  # NOTE: /stupload/ is intentionally excluded — slider and module images live there.
</VirtualHost>
```

**Critical**: The `RewriteCond /app%{REQUEST_URI} !-f` line checks the actual filesystem path inside the container (`/app` is the DocumentRoot inside Lando). This ensures the rule only fires for truly missing files.

### 2. Download the placeholder image

Download a visually appealing placeholder (misty mountain landscape from Unsplash) to `img/placeholder-dev.jpg`:

```bash
curl -L "https://images.unsplash.com/photo-1611572789411-6240f6cea970?q=80&w=500&h=500&fit=crop" -o img/placeholder-dev.jpg
```

### 3. Update `.lando.yml`

Add the vhosts reference under `services.appserver.config`. Read the existing `.lando.yml` first and merge — don't overwrite. The key addition:

```yaml
services:
  appserver:
    config:
      vhosts: .lando/apache-placeholder.conf
```

If `services.appserver` already exists, just add the `config.vhosts` key. If `.lando.yml` doesn't have a `services` section yet, add the whole block.

### 4. Rebuild Lando

```bash
lando rebuild -y
```

This restarts the appserver with the new Apache config. Takes ~1-2 minutes.

## Verification

After rebuild, test that a missing image returns the placeholder:

```bash
curl -I "http://{project}.lndo.site/img/p/1/test.jpg"
# Should return HTTP/1.1 200 OK
```

## Notes

- The placeholder image is a real static JPEG — no PHP involved. This avoids conflicts with PrestaShop's `.htaccess` rewrite rules.
- The VirtualHost wrapping is required. A bare `<Directory>` block without `<VirtualHost>` causes 403 errors on the entire site.
- `/img/st/` covers the store thumbnail path used by some PS versions.
- The friendly URL pattern `/[0-9]+-[a-z]+_default` (no trailing slash) is intentionally strict but broad enough to also cover retina variants like `home_default_2x` — PrestaShop image types always end in `_default` (home_default, large_default, etc.). This avoids matching slider/module URLs that use numeric IDs but aren't product images.
- `/stupload/` is excluded — sliders and modules store images there and they exist on disk, but intercepting it breaks slider images.
- `.lando.yml` is typically gitignored in PS projects — safe to modify locally without committing.
