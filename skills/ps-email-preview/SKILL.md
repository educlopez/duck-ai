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
  variables in HTML attributes highlighted with dashed outline + badge after element.
  Use this skill whenever the user wants to preview, inspect, or redesign PrestaShop
  email templates, asks "where do email templates go in my theme", wants to see what
  emails look like, mentions "email preview", "ver templates de email",
  "previsualizar emails", "email viewer", or starts working on transactional emails
  in a PS project. Trigger proactively when the user is about to redesign emails.
version: "4.0.0"
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
- **Navegación con teclado** — flechas ↑↓ navegan entre templates cuando el buscador no está enfocado; Escape cierra el drawer
- **Jerarquía de resolución** — replica el comportamiento de PS: child theme → parent theme → core

---

## Paso 1 — Detectar configuración del proyecto

Ejecuta estos comandos para recopilar los datos que necesitas:

```bash
# 1. Child theme: busca theme.yml con campo "parent:"
grep -rl "^parent:" themes/*/config/theme.yml 2>/dev/null

# 2. URL Lando: extrae el nombre del proyecto
grep "^name:" .lando.yml | head -1
```

Con esos resultados:
- **`CHILD_THEME`** = nombre de la carpeta del child theme (ej: `mitienda`)
- **`LANDO_URL`** = `https://{nombre-lando}.lndo.site`
  - El nombre Lando tiene guiones en vez de puntos y sin versión con puntos.
    Ej: `.lando.yml name: mitienda-v9.1` → URL base `mitienda-v91.lndo.site`
    (los puntos del número de versión se eliminan: `v9.1` → `v91`)
  - Si `lando info` está disponible, úsalo para obtener la URL exacta del servicio `appserver_nginx`

El idioma y el logo se detectan automáticamente en runtime:
- **Idioma** — `detectLanguages()` escanea `mails/*/` y filtra carpetas de 2 letras que contienen `.html`
- **Logo** — `resolveShopLogo()` lee `PS_LOGO_MAIL` de la DB vía `app/config/parameters.php`; fallback a `/img/logo.jpg`

---

## Paso 2 — Crear carpeta del child theme

```bash
mkdir -p themes/{CHILD_THEME}/mails/es/
```

Esta carpeta es donde van los templates personalizados. PS los resuelve en orden:
1. `themes/{child}/mails/{lang}/` ← **aquí van los overrides del proyecto**
2. `themes/{parent}/mails/{lang}/` ← parent theme (ej: panda)
3. `mails/{lang}/` ← core PS (fallback)

---

## Paso 3 — Generar email-preview.php

Lee el template desde `assets/email-preview.php` (junto a este SKILL.md) y
reemplaza estos 2 marcadores con los valores detectados:

| Marcador | Reemplazar con |
|---|---|
| `%%LANDO_URL%%` | URL completa sin trailing slash (ej: `https://mitienda-v91.lndo.site`) |
| `%%CHILD_THEME%%` | Nombre del child theme (ej: `mitienda`) |

Escribe el resultado en `email-preview.php` en la raíz del proyecto PS.

---

## Paso 4 — Verificar y comunicar al usuario

Confirma al usuario:
- La URL de acceso: `{LANDO_URL}/email-preview.php`
- Qué child theme se detectó y si la carpeta de mails se creó
- Dónde deben ir sus templates personalizados: `themes/{CHILD_THEME}/mails/{lang}/`
- Cómo funcionan los dots del sidebar: azul = child | gris = parent | claro = core

Si no se detecta un child theme (solo hay `classic` u otros sin `parent:`), usar
`mails/{lang}/` como única fuente y omitir la resolución de overrides.
Si no hay Lando (proyecto sin `.lando.yml`), dejar `%%LANDO_URL%%` vacío o con la
URL de producción; las rutas de mock funcionarán igualmente con rutas relativas.

---

## Personalizar datos mock

Los datos mock están en `$mock` dentro del PHP. Son genéricos (Ana García, Mi Tienda Online, euros).
El usuario puede editarlos directamente en `email-preview.php` para que reflejen su tienda real.

---

## Notas importantes

- El archivo `email-preview.php` es solo para desarrollo local — no subir a producción.
- Si el proyecto tiene idioma diferente al español, los `$friendlyNames` del PHP
  estarán en español (es solo el label del sidebar, no afecta los templates).
- El selector de idioma aparece automáticamente si hay más de un idioma en `mails/`.
- El logo se resuelve desde la DB en cada request; si la DB no está disponible, cae
  silenciosamente al logo por defecto `/img/logo.jpg`.
