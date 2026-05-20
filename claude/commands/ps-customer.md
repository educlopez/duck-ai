---
name: ps-customer
description: Crear cliente de prueba en PrestaShop
version: "0.1.0"
---

# Crear cliente de prueba en PrestaShop

Crea un cliente de prueba completo en PrestaShop con dirección y pedidos de ejemplo.

## Argumento

`$ARGUMENTS` — No se requiere argumento. Si se pasa, se ignora.

## Datos del cliente

- **Email:** `test@test.com`
- **Password:** `test1234`
- **Nombre:** `Test User`
- **Dirección:** Calle Test 123, 08001 Barcelona, España
- **3 pedidos** en distintos estados (entregado, enviado, pago aceptado)

## Instrucciones

Ejecuta un único script PHP via `lando php -r "..."` que haga todo. El script debe:

1. Cargar PrestaShop: `require '/app/config/config.inc.php';`
2. Usar `$prefix = _DB_PREFIX_`, `$db = Db::getInstance()`
3. **Limpiar** cualquier customer existente con email `test@test.com` (y sus address, cart, orders, order_detail, order_history, customer_group, cart_product)
4. **Crear customer** usando `PrestaShop\PrestaShop\Core\Crypto\Hashing` para el hash de `test1234`. Con `active=1, id_default_group=3, id_shop=1, id_lang=1`
5. **Insertar customer_group** para grupos 1, 2, 3
6. **Crear dirección** de envío: alias `Casa`, Calle Test 123, 08001 Barcelona, España (id_country=6), DNI 12345678A, teléfono 600000000
7. **Crear 3 pedidos** con productos reales de la tienda. Para cada pedido:
   - Buscar 3 productos activos que tengan combinaciones: `SELECT p.id_product, pl.name, pa.id_product_attribute, p.price FROM {prefix}product p JOIN {prefix}product_lang pl ON ... JOIN {prefix}product_attribute pa ON ... WHERE p.active=1 LIMIT 3`
   - Crear un `cart` y `cart_product`
   - Crear el `order` en `{prefix}orders` con referencia aleatoria (9 chars uppercase)
   - Crear `order_detail` con el producto
   - Crear `order_history`
   - Usar carrier=592, currency=1 (EUR), payment=Transferencia bancaria, module=ps_wirepayment
   - Los 3 pedidos en estados: 5 (Entregado, hace 45 días), 4 (Enviado, hace 15 días), 2 (Pago aceptado, hace 2 días)

8. IMPORTANTE: Escapar `$` como `\$` en el script PHP al ejecutarlo con `lando php -r`

9. Mostrar resumen:
   - Email: `test@test.com`
   - Password: `test1234`
   - Dirección creada
   - IDs y referencias de los 3 pedidos
