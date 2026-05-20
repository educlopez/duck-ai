# ps-demo-user — Usuario demo PrestaShop 8 (Lando)

Crea un cliente de prueba con historial completo para testear la UI del panel de usuario.
Reutilizable en cualquier proyecto PS8 con Lando.

## Credenciales (siempre las mismas)

| Campo    | Valor        |
|----------|--------------|
| Email    | `edu@demo.com` |
| Password | `edudemo123` |
| Nombre   | Edu Demo     |

## Qué incluye

- 2 direcciones (casa + trabajo con empresa y NIF)
- 3 pedidos con estados distintos:
  - `DEMO00001` — **Entregado** (hace 60 días) + factura descargable + tracking GLS
  - `DEMO00002` — **Enviado** (hace 30 días) + factura + tracking GLS
  - `DEMO00003` — **En preparación** (hace 5 días)
- Pagos registrados por cada pedido
- Historial completo de cambios de estado
- Transportista con número de seguimiento ficticio

## Instrucciones de uso

### Paso 1 — Obtener datos del proyecto

```sql
-- Carrier activo
SELECT id_carrier, name FROM {prefix}carrier WHERE active=1 AND deleted=0 LIMIT 5;

-- Productos con precio para usar en order_detail
SELECT id_product, reference, price FROM {prefix}product WHERE active=1 AND price > 0 LIMIT 5;

-- Prefijo
-- Ver en app/config/parameters.php → database_prefix
```

### Paso 2 — Copiar y adaptar el script base

El script base está en `demo-user-setup.sql` (proyecto vives-8).

Cambios obligatorios para otros proyectos:
1. Reemplazar prefijo `vivesps_` → el del proyecto
2. Cambiar `@carrier_id` por el carrier activo
3. Cambiar los `product_id` en la sección ORDER DETAILS por productos reales
4. Si se quiere cambiar la contraseña: regenerar hash con:
   ```bash
   lando php -r "echo password_hash('nueva_clave', PASSWORD_BCRYPT);"
   ```
   Y actualizar el campo `passwd` en la sección CUSTOMER.

### Paso 3 — Ejecutar

```bash
lando mysql lamp -u lamp -plamp < demo-user-setup.sql
```

> El script incluye `SET sql_mode = ''` para evitar errores de fecha cero en Lando.

### Paso 4 — Verificar

Entrar en `https://{proyecto}.lndo.site/es/login` con `edu@demo.com` / `edudemo123`

O verificar por SQL:
```sql
SELECT id_customer, email FROM {prefix}customer WHERE email='edu@demo.com';
SELECT reference, current_state FROM {prefix}orders WHERE id_customer={id} LIMIT 5;
```

## Estructura del script

```
SET sql_mode = '';
1. CUSTOMER        — cuenta con bcrypt hash
2. ADDRESSES       — 2 direcciones (casa + trabajo)
3. CARTS           — 3 carritos (FK obligatoria de orders)
4. ORDERS          — 3 pedidos (entregado / enviado / preparación)
5. ORDER_DETAIL    — líneas de producto por pedido
6. ORDER_PAYMENT   — registro de pago por pedido
7. ORDER_INVOICE   — facturas (pedidos 1 y 2)
8. ORDER_HISTORY   — log de cambios de estado
9. ORDER_CARRIER   — transportista + tracking por pedido
```

## Gotchas PS8

| Problema | Solución |
|----------|----------|
| Error fecha `0000-00-00` | Añadir `SET sql_mode = ''` al inicio |
| `birthday = 0000-00-00` da error en backoffice | Usar `NULL` en lugar de `'0000-00-00'` — PS8 lo rechaza al editar el cliente |
| `secure_key = '-1'` da CustomerException | Debe ser un MD5 de 32 chars — usar `MD5(RAND())` en el INSERT |
| `newsletter_date_add = 0000-00-00` da error | Usar `NULL` en lugar de fecha cero |
| `id_cart` obligatorio en orders | Crear carts antes que orders |
| `invoice_address`/`delivery_address` no existen | Eliminados en PS8 — no incluir |
| `unit_price_tax_incl_after_specific_price` no existe | Eliminado en PS8 — no incluir |
| Password MD5 antiguo no funciona | PS8 usa bcrypt — regenerar hash |

## Eliminar usuario demo

```sql
SET sql_mode='';
SELECT @id := id_customer FROM {prefix}customer WHERE email='edu@demo.com';
DELETE oc FROM {prefix}order_carrier oc JOIN {prefix}orders o ON oc.id_order=o.id_order WHERE o.id_customer=@id;
DELETE oh FROM {prefix}order_history oh JOIN {prefix}orders o ON oh.id_order=o.id_order WHERE o.id_customer=@id;
DELETE oi FROM {prefix}order_invoice oi JOIN {prefix}orders o ON oi.id_order=o.id_order WHERE o.id_customer=@id;
DELETE op FROM {prefix}order_payment op JOIN {prefix}orders o ON op.order_reference=o.reference WHERE o.id_customer=@id;
DELETE od FROM {prefix}order_detail od JOIN {prefix}orders o ON od.id_order=o.id_order WHERE o.id_customer=@id;
DELETE FROM {prefix}orders WHERE id_customer=@id;
DELETE FROM {prefix}cart WHERE id_customer=@id;
DELETE FROM {prefix}address WHERE id_customer=@id;
DELETE FROM {prefix}customer WHERE email='edu@demo.com';
```
