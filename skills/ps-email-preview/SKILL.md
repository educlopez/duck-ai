---
name: ps-email-preview
description: >
  Generates a full-featured email template viewer for PrestaShop projects, deployed
  as email-preview/index.php (subfolder, ROOT = dirname(__DIR__)). Features: Geist
  font + Geist Mono UI matching Vercel design system; PS admin authentication via
  Cookie class (always denies on failure — never allows by default); CSRF protection
  on all write endpoints; sandboxed iframe (sandbox="allow-scripts"); shop logo and
  favicon resolved dynamically from PS DB; source version pill dropdown (custom,
  not native select) showing child/parent/core with colored dots and count badge;
  language pill dropdown (same custom pattern); sidebar with real-time search (F
  shortcut), Geist-style tabs in drawer and preview bar, source stack dots per
  template; "↑ child" button with Geist confirmation modal to copy template to child
  theme override; "Eliminar override" button with Geist destructive modal (red band,
  irreversibility warning) to delete the child override and revert to parent/core;
  Variables inspector drawer (tabs: Variables with clipboard badges + Diff vs core);
  auto-reload Watch mode polling filemtime; email client-style header (subject,
  avatar with Gravatar fallback, De/Para fields); subject line from static map;
  plain-text TXT preview tab; keyboard shortcuts (v/d/r/w/f/↑↓/Esc); sidebar footer
  showing logged-in PS employee with Gravatar avatar; "Datos reales" mock toggle
  substitutes variables in both text and attributes (links/images work in mock mode);
  variables in HTML attributes highlighted with dashed outline + badge after element;
  "Enviar test" button in topbar opens a Geist modal to send the current template
  with mock data to any real email address using PS Mail::Send() — requires Mailhog
  SMTP configured in PS backoffice.
  Use this skill whenever the user wants to preview, inspect, or redesign PrestaShop
  email templates, asks "where do email templates go in my theme", wants to see what
  emails look like, mentions "email preview", "ver templates de email",
  "previsualizar emails", "email viewer", or starts working on transactional emails
  in a PS project. Trigger proactively when the user is about to redesign emails.
version: "5.0.0"
author: Eduardo Calvo
---

# PS Email Preview

Genera un visor de email templates para PrestaShop en un solo archivo PHP.
Sirve para ver todos los templates renderizados, con variables resaltadas o
sustituidas por datos mock, y entiende la jerarquía de overrides de PS.

## Características del visor

- **UI con Geist font** — diseño limpio estilo Vercel
- **Logo de tienda en sidebar** — resuelto dinámicamente desde la DB PS (`PS_LOGO_MAIL`); si falla, muestra un icono con la inicial del child theme
- **Badges de fuente** — cada template muestra si viene de `child` (azul), `parent` (gris borde) o `core` (gris claro)
- **Buscador en el sidebar** — filtra templates en tiempo real por nombre o clave, ocultando grupos vacíos
- **Toggle de ancho de preview** — botones 600 / 800 / full para testear el layout en anchos típicos de email
- **Botón "↑ child"** — copia el template seleccionado al child theme override con un clic; solo aparece cuando el template no es ya del child
- **Drawer de Variables (Inspector)** — panel lateral con dos pestañas:
  - **Variables** — variables agrupadas en "Con dato mock" (azul) y "Sin dato mock" (gris); clic en cualquier badge copia `{variable}` al portapapeles con toast de confirmación
  - **Diff vs core** — muestra las líneas que difieren entre el override activo y el template core original; si no hay override muestra un aviso
- **Selector de idioma** — detectado automáticamente de `mails/*/` (solo visible si hay más de un idioma)
- **Toggle "Datos reales"** — sustituye variables por datos mock realistas; variables sin mock aparecen resaltadas en rojo
- **"Enviar test"** — botón en el topbar que abre un modal Geist para enviar el template actual con datos mock a cualquier email real; usa `Mail::Send()` de PS con resolución automática de rutas (child theme primero); recuerda la última dirección usada en `localStorage`
- **Navegación con teclado** — flechas ↑↓ navegan entre templates cuando el buscador no está enfocado; Escape cierra el drawer
- **Jerarquía de resolución** — replica el comportamiento de PS: child theme → parent theme → core

---

## Paso 1 — Detectar configuración del proyecto

Ejecuta estos comandos para recopilar los datos que necesitas:

```bash
# 1. Child theme: busca theme.yml con campo "parent:"
grep -rl "^parent:" themes/*/config/theme.yml 2>/dev/null

# 2. Nombre del proyecto Lando
grep "^name:" .lando.yml | head -1

# 3. Hostname interno de Mailhog (para configurar SMTP en PS)
lando info 2>/dev/null | grep -A5 '"mail"' | grep 'host'
```

Con esos resultados:
- **`CHILD_THEME`** = nombre de la carpeta del child theme (ej: `mitienda`)
- **`LANDO_URL`** = `https://{nombre-lando}.lndo.site`
  - El nombre Lando tiene guiones en vez de puntos y sin versión con puntos.
    Ej: `.lando.yml name: mitienda-v9.1` → URL base `mitienda-v91.lndo.site`
    (los puntos del número de versión se eliminan: `v9.1` → `v91`)
  - Si `lando info` está disponible, úsalo para obtener la URL exacta del servicio `appserver_nginx`
- **`MAILHOG_HOST`** = hostname interno del servicio mail, formato `mail.{lando-slug}.internal`
  - Ej: proyecto `mitienda-v91` → host `mail.mitiendav91.internal`
  - Puerto siempre `1025`
  - Mailhog web UI siempre en `http://localhost:{puerto-externo}` (ver `lando info`)

El idioma y el logo se detectan automáticamente en runtime:
- **Idioma** — `detectLanguages()` escanea `mails/*/` y filtra carpetas de 2 letras que contienen `.html`
- **Logo** — `resolveShopLogo()` lee `PS_LOGO_MAIL` de la DB vía `app/config/parameters.php`; fallback a `/img/logo.jpg`

---

## Paso 2 — Configurar Mailhog SMTP en PS

Para que "Enviar test" funcione, PS tiene que estar apuntando a Mailhog.
Hazlo **vía base de datos** (más rápido que el backoffice):

```bash
lando mysql {DB_NAME} --user={DB_USER} --password={DB_PASS} -e "
UPDATE {PREFIX}configuration SET value='2' WHERE name='PS_MAIL_METHOD';
UPDATE {PREFIX}configuration SET value='{MAILHOG_HOST}' WHERE name='PS_MAIL_SERVER';
UPDATE {PREFIX}configuration SET value='1025' WHERE name='PS_MAIL_SMTP_PORT';
UPDATE {PREFIX}configuration SET value='off' WHERE name='PS_MAIL_SMTP_ENCRYPTION';
UPDATE {PREFIX}configuration SET value='' WHERE name='PS_MAIL_USER';
UPDATE {PREFIX}configuration SET value='' WHERE name='PS_MAIL_PASSWD';
"
```

Donde:
- `{DB_NAME}` = base de datos del proyecto (ej: `lemp`)
- `{DB_USER}` / `{DB_PASS}` = credenciales DB (Lando por defecto: `lamp` / `lamp`)
- `{PREFIX}` = prefijo de tablas PS (ej: `ps_` o el personalizado del proyecto)
- `{MAILHOG_HOST}` = host interno detectado en el paso anterior (ej: `mail.mitiendav91.internal`)

Verificar que quedó bien:
```bash
lando mysql {DB_NAME} --user={DB_USER} --password={DB_PASS} -e "
SELECT name, value FROM {PREFIX}configuration
WHERE name IN ('PS_MAIL_METHOD','PS_MAIL_SERVER','PS_MAIL_SMTP_PORT','PS_MAIL_SMTP_ENCRYPTION');
"
```

Resultado esperado:
| name | value |
|------|-------|
| PS_MAIL_METHOD | 2 |
| PS_MAIL_SERVER | mail.{slug}.internal |
| PS_MAIL_SMTP_PORT | 1025 |
| PS_MAIL_SMTP_ENCRYPTION | off |

> **Alternativa backoffice**: Parámetros Avanzados → Email → activar SMTP y rellenar los mismos campos.

---

## Paso 3 — Crear carpeta del child theme y copiar .txt

PS requiere **tanto `.html` como `.txt`** cuando `PS_MAIL_TYPE = 3` (HTML + texto).
Si el child theme solo tiene `.html`, el envío fallará con "template missing".

```bash
mkdir -p themes/{CHILD_THEME}/mails/es/

# Copiar los .txt del core para todos los templates que el child tenga en .html
for f in themes/{CHILD_THEME}/mails/es/*.html; do
  key=$(basename "$f" .html)
  txt="mails/es/${key}.txt"
  dest="themes/{CHILD_THEME}/mails/es/${key}.txt"
  [ -f "$txt" ] && [ ! -f "$dest" ] && cp "$txt" "$dest" && echo "✓ $key.txt"
done
```

Esta carpeta es donde van los templates personalizados. PS los resuelve en orden:
1. `themes/{child}/mails/{lang}/` ← **aquí van los overrides del proyecto**
2. `themes/{parent}/mails/{lang}/` ← parent theme (ej: panda)
3. `mails/{lang}/` ← core PS (fallback)

---

## Paso 4 — Generar email-preview/index.php

Lee el template desde `assets/email-preview.php` (junto a este SKILL.md) y
reemplaza estos 2 marcadores con los valores detectados:

| Marcador | Reemplazar con |
|---|---|
| `%%LANDO_URL%%` | URL completa sin trailing slash (ej: `https://mitienda-v91.lndo.site`) |
| `%%CHILD_THEME%%` | Nombre del child theme (ej: `mitienda`) |

Escribe el resultado en **`email-preview/index.php`** (subcarpeta, no raíz):

```bash
mkdir -p email-preview/
# escribir el archivo como email-preview/index.php
```

El archivo usa `ROOT = dirname(__DIR__)` para subir un nivel y acceder a PS.

---

## Paso 5 — Verificar y comunicar al usuario

Confirma al usuario:
- La URL de acceso: `{LANDO_URL}/email-preview/`
- Qué child theme se detectó y si la carpeta de mails se creó
- Dónde deben ir sus templates personalizados: `themes/{CHILD_THEME}/mails/{lang}/`
- Cómo funcionan los dots del sidebar: azul = child | gris = parent | claro = core
- Mailhog web UI para ver emails de prueba: `http://localhost:{puerto-externo}`
- Botón "Enviar test" en el topbar para enviar cualquier template con datos mock

Si no se detecta un child theme (solo hay `classic` u otros sin `parent:`), usar
`mails/{lang}/` como única fuente y omitir la resolución de overrides.
Si no hay Lando (proyecto sin `.lando.yml`), dejar `%%LANDO_URL%%` vacío o con la
URL de producción; las rutas de mock funcionarán igualmente con rutas relativas.

---

## Personalizar datos mock

Los datos mock están en `$mock` dentro del PHP. Son genéricos (Ana García, Mi Tienda Online, euros).
El usuario puede editarlos directamente en `email-preview/index.php` para que reflejen su tienda real.

---

## Notas importantes

- El archivo `email-preview/index.php` es solo para desarrollo local — no subir a producción.
- Si el proyecto tiene idioma diferente al español, los `$friendlyNames` del PHP
  estarán en español (es solo el label del sidebar, no afecta los templates).
- El selector de idioma aparece automáticamente si hay más de un idioma en `mails/`.
- El logo se resuelve desde la DB en cada request; si la DB no está disponible, cae
  silenciosamente al logo por defecto `/img/logo.jpg`.
- **`PS_MAIL_TYPE = 3`** (HTML+TXT) es el default de PS — siempre copiar los `.txt` del core
  al child theme junto a los `.html`, o el envío fallará en silencio.
- El subject de los emails se pre-procesa sustituyendo las vars mock antes de llamar
  a `Mail::Send()`, ya que PS no sustituye `$templateVars` en el asunto.
