---
name: ps-watch
description: >
  Sets up BrowserSync live-reload watcher for PrestaShop Panda child theme development
  in Lando. Auto-detects child theme name and Lando proxy URL. Watches child theme
  CSS/JS/TPL files and reloads the browser automatically on save. Use when working
  on PS frontend/maquetación in Lando and the user wants auto-reload, live reload,
  watcher, or is tired of reloading the browser manually after CSS changes.
version: "0.1.0"
---

# PS Watch — BrowserSync para Panda child theme en Lando

Auto-reload en el browser cada vez que guardas CSS, JS o TPL del child theme.
No toca Panda parent. Funciona fuera del contenedor Lando.

---

## Setup (ejecutar una vez por proyecto)

### Paso 1 — Detectar child theme y URL de Lando

```bash
# Child theme = cualquier carpeta en themes/ que no sea panda ni core
ls themes/ | grep -vE "^(panda|core|\.)"
```

```bash
# URL de Lando — leer el campo name del .lando.yml
grep "^name:" .lando.yml | awk '{print $2}'
# URL resultado: http://<name>.lndo.site
# Usar HTTP, no HTTPS — evita errores de cert self-signed de Lando
```

### Paso 2 — Crear `bs-config.js` en el root del proyecto

Sustituir `CHILD_THEME` por el nombre detectado y `LANDO_URL` por `http://<name>.lndo.site`:

```js
module.exports = {
  proxy: "http://LANDO_URL",
  files: [
    "themes/CHILD_THEME/assets/css/**/*.css",
    "themes/CHILD_THEME/assets/js/**/*.js",
    "themes/CHILD_THEME/templates/**/*.tpl",
    "modules/*/views/templates/**/*.tpl",
    "modules/*/views/css/**/*.css",
  ],
  open: false,
  notify: false,
  reloadDelay: 150,
  logLevel: "info",
};
```

### Paso 3 — Crear o actualizar `package.json` en el root

Si ya existe `package.json`, añadir solo las entradas que falten.
Si no existe, crear completo:

```json
{
  "private": true,
  "scripts": {
    "watch": "browser-sync start --config bs-config.js"
  },
  "devDependencies": {
    "browser-sync": "^3"
  }
}
```

### Paso 4 — Instalar dependencias

```bash
npm install
```

---

## Uso diario

```bash
# Arranca el watcher (con Lando ya corriendo)
npm run watch
```

BrowserSync abre en `http://localhost:3000` (proxy de Lando).
Abre esa URL en el browser en vez de la URL de Lando directamente.
Cada vez que guardas un fichero vigilado → el browser recarga solo.

---

## Notas importantes

- **Usar HTTP, no HTTPS** para el proxy: Lando sirve ambos pero HTTPS tiene cert self-signed que BrowserSync rechaza. HTTP funciona igual para desarrollo local.
- **Child theme nunca es `panda`**: Panda es el parent. El child theme varía por proyecto (`milagros`, `podiatech`, etc.).
- **`open: false`** evita que BrowserSync abra otra pestaña al arrancar.
- `modules/*/views/` cubre módulos custom del proyecto — si el proyecto no tiene módulos custom con CSS/TPL, no pasa nada (el watcher ignora patterns vacíos).
- Si el puerto 3000 está ocupado, BrowserSync usa 3001, 3002, etc. — lo indica en la terminal.

## `.gitignore` (añadir si no están)

```
node_modules/
```

---

## Checklist

- [ ] Child theme detectado (`ls themes/ | grep -vE "^(panda|core)"`)
- [ ] URL Lando obtenida (`grep "^name:" .lando.yml`)  
- [ ] `bs-config.js` creado con child theme name y URL correctos
- [ ] `package.json` creado/actualizado con script `watch` y dep `browser-sync`
- [ ] `npm install` ejecutado
- [ ] `node_modules/` en `.gitignore`
- [ ] `npm run watch` arranca y browser recarga al guardar un CSS
