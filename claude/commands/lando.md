---
name: lando
description: Crear configuración Lando para PrestaShop
version: "0.1.0"
---

# Crear configuración Lando para PrestaShop

Genera la configuración `.lando.yml` y configura `app/config/parameters.php` para el proyecto PrestaShop actual.

## Argumento

`$ARGUMENTS` — Puerto MySQL para portforward (ej: 33405, 3401, 3402...)

## Instrucciones

1. Detecta el nombre del proyecto a partir del directorio actual (último segmento del path).
2. Detecta la versión de PrestaShop leyendo `app/AppKernel.php` (busca `MAJOR_VERSION`):
   - Si MAJOR_VERSION = 8 → PHP 8.1
   - Si MAJOR_VERSION < 8 → PHP 7.4
3. Crea el archivo `.lando.yml` en la raíz del proyecto con este contenido exacto:

```yaml
name: NOMBRE_PROYECTO
recipe: lamp
config:
  php: "VERSION_PHP"
  webroot: ./
  database: mysql
  xdebug: false
services:
  database:
    portforward: PUERTO_ARGUMENTO
    type: mysql
  mail:
    type: mailhog
    hogfrom:
      - appserver
```

Reemplaza:
- `NOMBRE_PROYECTO` por el nombre del directorio del proyecto
- `VERSION_PHP` por la versión detectada (7.4 o 8.1)
- `PUERTO_ARGUMENTO` por el puerto pasado como argumento (`$ARGUMENTS`)

4. Modifica `app/config/parameters.php` cambiando SOLO los valores de conexión a base de datos y mailer, conservando todo lo demás intacto (secret, cookies, keys, locale, prefijo, etc.):
   - `database_host` → `'database'`
   - `database_port` → `''`
   - `database_name` → `'lamp'`
   - `database_user` → `'lamp'`
   - `database_password` => `'lamp'`
   - `mailer_transport` → `'smtp'`
   - `mailer_host` → `'mail'`
   - `mailer_user` → `NULL`
   - `mailer_password` → `NULL`

5. NO crees archivos adicionales (ni php.ini, ni my-custom.cnf, ni carpeta .lando/).
6. Muestra un resumen con:
   - La configuración `.lando.yml` generada
   - Los cambios realizados en `parameters.php`
   - Las credenciales por defecto (user: `lamp`, pass: `lamp`, db: `lamp`, host: `database`)
   - La URL: `https://NOMBRE_PROYECTO.lndo.site/`
   - El puerto MySQL externo: `127.0.0.1:PUERTO`
   - Mailhog: SMTP en `mail:1025`
   - Recordatorio: `lando start` para levantar, `lando rebuild -y` si se cambia .lando.yml
