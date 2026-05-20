---
name: ps-url
description: Actualizar shop_url de PrestaShop para Lando
version: "0.1.0"
---

# Actualizar shop_url de PrestaShop para Lando

Actualiza la tabla `shop_url` de PrestaShop para que use la URL de Lando en lugar de la URL de producción.

## Argumento

`$ARGUMENTS` — No se requiere argumento. Si se pasa, se ignora.

## Instrucciones

1. Lee `app/config/parameters.php` para obtener `database_prefix` y `database_name`.
2. Lee `.lando.yml` para obtener el `name` del proyecto. El dominio Lando será `{name}.lndo.site`.
3. Ejecuta el UPDATE SQL via `lando mysql <database_name>` para actualizar todas las filas de `{prefix}shop_url`:
   - `domain` → `{name}.lndo.site`
   - `domain_ssl` → `{name}.lndo.site`
4. Muestra un resumen con:
   - El dominio anterior (haz un SELECT antes del UPDATE para mostrarlo)
   - El nuevo dominio: `{name}.lndo.site`
   - La URL final: `https://{name}.lndo.site/`
   - Recordatorio: limpiar caché con `lando ssh -c "rm -rf var/cache/*"` si hay problemas
