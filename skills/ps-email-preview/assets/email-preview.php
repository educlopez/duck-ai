<?php
/**
 * PrestaShop Email Template Viewer — Milagros
 * Access: https://milagros-colombia-b2b-v91.lndo.site/email-preview.php
 *
 * Jerarquía de resolución (igual que PS en producción):
 *   1. themes/milagros/mails/{lang}/   ← child theme override
 *   2. themes/panda/mails/{lang}/      ← parent theme override
 *   3. mails/{lang}/                   ← core / fallback
 */

define('ROOT', dirname(__DIR__));   // PS root — one level up from email-preview/
define('LANDO_URL', (isset($_SERVER['HTTPS']) && $_SERVER['HTTPS'] !== 'off' ? 'https' : 'http') . '://' . $_SERVER['HTTP_HOST']);
define('CHILD_THEME', '%%CHILD_THEME%%');

// ── Protección: solo empleados logueados en el backoffice ─────────────────────

function denyAccess(string $title, string $msg, int $code = 403): never {
    http_response_code($code);
    echo '<!DOCTYPE html><html lang="es"><head><meta charset="UTF-8"><link href="https://fonts.googleapis.com/css2?family=Geist:wght@400;500;600&display=swap" rel="stylesheet"><style>*{box-sizing:border-box;margin:0;padding:0}body{font-family:Geist,-apple-system,sans-serif;background:#fafafa;display:flex;align-items:center;justify-content:center;min-height:100vh}.card{background:#fff;border:1px solid #eaeaea;border-radius:12px;padding:40px 48px;text-align:center;max-width:360px;width:100%}.icon{margin-bottom:20px;display:flex;justify-content:center}h1{font-size:15px;font-weight:600;color:#000;margin-bottom:8px}p{font-size:13px;color:#666;line-height:1.6}</style></head><body><div class="card"><div class="icon"><svg width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="#bbb" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="11" width="18" height="11" rx="2"/><path d="M7 11V7a5 5 0 0 1 10 0v4"/></svg></div><h1>' . htmlspecialchars($title) . '</h1><p>' . htmlspecialchars($msg) . '</p></div></body></html>';
    exit;
}

function requirePsAdminAuth(): ?array {
    $psConfig = ROOT . '/config/config.inc.php';
    if (!file_exists($psConfig)) denyAccess('No disponible', 'PrestaShop no encontrado en esta ruta.', 503);
    try {
        ob_start();
        @require_once $psConfig;
        ob_end_clean();
        if (!class_exists('Cookie')) denyAccess('No disponible', 'El sistema no pudo iniciarse. Comprueba que la BD esté activa.', 503);
        $cookieName = defined('_COOKIE_ADMIN_') ? _COOKIE_ADMIN_ : 'psAdmin';
        $cookie = new Cookie($cookieName);
        if (empty($cookie->id_employee)) {
            $adminDir = 'admin';
            foreach (glob(ROOT . '/admin*', GLOB_ONLYDIR) as $d) {
                $name = basename($d);
                if (preg_match('/^admin[a-z0-9]*$/i', $name)) { $adminDir = $name; break; }
            }
            $loginUrl = LANDO_URL . '/' . $adminDir . '/index.php';
            http_response_code(403);
            echo '<!DOCTYPE html><html lang="es"><head><meta charset="UTF-8">'
               . '<title>Acceso restringido</title>'
               . '<link rel="preconnect" href="https://fonts.googleapis.com">'
               . '<link href="https://fonts.googleapis.com/css2?family=Geist:wght@400;500;600&display=swap" rel="stylesheet">'
               . '<style>*{box-sizing:border-box;margin:0;padding:0}'
               . 'body{font-family:Geist,-apple-system,sans-serif;background:#fafafa;display:flex;align-items:center;justify-content:center;min-height:100vh}'
               . '.card{background:#fff;border:1px solid #eaeaea;border-radius:12px;padding:40px 48px;text-align:center;max-width:360px;width:100%}'
               . '.icon{margin-bottom:20px;display:flex;justify-content:center}'
               . 'h1{font-size:16px;font-weight:600;color:#000;margin-bottom:8px}'
               . 'p{font-size:13px;color:#888;line-height:1.6;margin-bottom:24px}'
               . 'a{display:inline-flex;align-items:center;gap:6px;height:36px;padding:0 18px;background:#000;color:#fff;border-radius:7px;font-size:13px;font-weight:500;text-decoration:none}'
               . 'a:hover{background:#333}'
               . '</style></head><body>'
               . '<div class="card">'
               . '<div class="icon"><svg width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="#bbb" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="11" width="18" height="11" rx="2"/><path d="M7 11V7a5 5 0 0 1 10 0v4"/></svg></div>'
               . '<h1>Acceso restringido</h1>'
               . '<p>Necesitas estar logueado en el backoffice de PrestaShop para ver el visor de email templates.</p>'
               . '<a href="' . htmlspecialchars($loginUrl) . '">Ir al backoffice →</a>'
               . '</div></body></html>';
            exit;
        }
        // Fetch employee data
        $params = @include ROOT . '/app/config/parameters.php';
        if ($params) {
            $p = $params['parameters'];
            $pdo = new PDO(
                'mysql:host=' . $p['database_host'] . ';dbname=' . $p['database_name'] . ';charset=utf8',
                $p['database_user'], $p['database_password'], [PDO::ATTR_ERRMODE => PDO::ERRMODE_SILENT]
            );
            $stmt = $pdo->prepare("SELECT id_employee, firstname, lastname, email FROM `{$p['database_prefix']}employee` WHERE id_employee = ? LIMIT 1");
            $stmt->execute([(int)$cookie->id_employee]);
            $emp = $stmt->fetch(PDO::FETCH_ASSOC);
            if ($emp) return $emp;
        }
    } catch (Throwable $e) {
        ob_end_clean();
        denyAccess('No disponible', 'El sistema no pudo iniciarse. Comprueba que la BD esté activa.', 503);
    }
    return null; // unreachable but satisfies return type
}

$psEmployee = requirePsAdminAuth();

// ── CSRF token ────────────────────────────────────────────────────────────────
if (session_status() === PHP_SESSION_NONE) session_start();
if (empty($_SESSION['ep_csrf'])) $_SESSION['ep_csrf'] = bin2hex(random_bytes(32));
$csrfToken = $_SESSION['ep_csrf'];

function verifyCsrf(): void {
    $token = $_GET['csrf'] ?? $_POST['csrf'] ?? '';
    if (!hash_equals($_SESSION['ep_csrf'] ?? '', $token)) {
        header('Content-Type: application/json');
        http_response_code(403);
        echo json_encode(['ok' => false, 'error' => 'CSRF token inválido']);
        exit;
    }
}

// ── Detección de idiomas ──────────────────────────────────────────────────────

function detectLanguages(string $root): array {
    $langs = [];
    $coreMailsDir = $root . '/mails/';
    if (!is_dir($coreMailsDir)) return ['es'];
    foreach (glob($coreMailsDir . '*', GLOB_ONLYDIR) as $d) {
        $l = basename($d);
        if (preg_match('/^[a-z]{2}$/', $l) && !empty(glob($d . '/*.html'))) {
            $langs[$l] = true;
        }
    }
    ksort($langs);
    return array_keys($langs) ?: ['es'];
}

$availableLangs = detectLanguages(ROOT);
$defaultLang    = in_array('es', $availableLangs) ? 'es' : ($availableLangs[0] ?? 'es');
$lang = isset($_GET['lang']) && in_array($_GET['lang'], $availableLangs, true)
    ? $_GET['lang'] : $defaultLang;

$langNames = [
    'es' => 'Español', 'en' => 'English', 'fr' => 'Français',
    'pt' => 'Português', 'de' => 'Deutsch', 'it' => 'Italiano',
    'nl' => 'Nederlands', 'pl' => 'Polski',
];

$mailDirs = [
    'child'  => ROOT . '/themes/' . CHILD_THEME . '/mails/' . $lang . '/',
    'parent' => ROOT . '/themes/panda/mails/' . $lang . '/',
    'core'   => ROOT . '/mails/' . $lang . '/',
];

// ── Asuntos conocidos de PS (no están en los .html — se pasan por código) ─────
$subjectMap = [
    'account'                => 'Bienvenido/a a {shop_name}',
    'backoffice_order'       => 'Nuevo pedido de {firstname} {lastname}',
    'bankwire'               => 'Pedido {order_name} · Instrucciones de pago',
    'cheque'                 => 'Pedido {order_name} · Pago con cheque',
    'contact'                => 'Re: {subject}',
    'contact_form'           => 'Mensaje de {firstname} {lastname}',
    'credit_slip'            => 'Nota de crédito {credit_slip_number}',
    'download_product'       => 'Tu descarga · {product_name}',
    'employee_password'      => 'Tu contraseña en {shop_name}',
    'forward_msg'            => 'Fwd: {subject}',
    'guest_to_customer'      => 'Tu cuenta en {shop_name}',
    'import'                 => 'Importación completada',
    'in_transit'             => 'Pedido {order_name} en tránsito',
    'log_alert'              => 'Alerta de sistema · {shop_name}',
    'newsletter'             => 'Confirmación de suscripción newsletter',
    'order_canceled'         => 'Pedido {order_name} cancelado',
    'order_changed'          => 'Pedido {order_name} modificado',
    'order_conf'             => 'Confirmación de pedido {order_name}',
    'order_customer_comment' => 'Nuevo comentario en pedido {order_name}',
    'order_merchant_comment' => 'Nota del vendedor · pedido {order_name}',
    'order_return_state'     => 'Estado de tu devolución: {return_state}',
    'outofstock'             => 'Alerta de stock · {product_name}',
    'password'               => 'Tu nueva contraseña en {shop_name}',
    'password_query'         => 'Recupera tu contraseña en {shop_name}',
    'payment'                => 'Pago confirmado · pedido {order_name}',
    'payment_error'          => 'Error en el pago de tu pedido',
    'preparation'            => 'Pedido {order_name} en preparación',
    'productoutofstock'      => 'Producto agotado: {product_name}',
    'refund'                 => 'Reembolso de {refund_total}',
    'reply_msg'              => 'Re: {subject}',
    'shipped'                => 'Tu pedido {order_name} ha sido enviado',
    'test'                   => 'Email de prueba · {shop_name}',
    'voucher'                => 'Tu vale {voucher_num} para {shop_name}',
    'voucher_new'            => '¡Tienes un nuevo vale para {shop_name}!',
];

$friendlyNames = [
    'account'                => 'Cuenta creada',
    'backoffice_order'       => 'Backoffice · Nuevo pedido',
    'bankwire'               => 'Pedido · Transferencia bancaria',
    'cheque'                 => 'Pedido · Pago con cheque',
    'contact'                => 'Mensaje de contacto (respuesta)',
    'contact_form'           => 'Formulario de contacto',
    'credit_slip'            => 'Nota de crédito',
    'download_product'       => 'Descarga de producto virtual',
    'employee_password'      => 'Contraseña de empleado',
    'forward_msg'            => 'Mensaje reenviado',
    'guest_to_customer'      => 'Invitado → Cuenta',
    'import'                 => 'Importación completada',
    'in_transit'             => 'Pedido en tránsito',
    'log_alert'              => 'Alerta de sistema',
    'newsletter'             => 'Confirmación newsletter',
    'order_canceled'         => 'Pedido cancelado',
    'order_changed'          => 'Pedido modificado',
    'order_conf'             => 'Confirmación de pedido',
    'order_customer_comment' => 'Comentario del cliente',
    'order_merchant_comment' => 'Comentario del vendedor',
    'order_return_state'     => 'Estado de devolución',
    'outofstock'             => 'Sin stock (alerta)',
    'password'               => 'Contraseña restablecida',
    'password_query'         => 'Recuperar contraseña',
    'payment'                => 'Pago confirmado',
    'payment_error'          => 'Error en pago',
    'preparation'            => 'Pedido en preparación',
    'productoutofstock'      => 'Producto agotado',
    'refund'                 => 'Reembolso',
    'reply_msg'              => 'Respuesta a mensaje',
    'shipped'                => 'Pedido enviado',
    'test'                   => 'Email de prueba',
    'voucher'                => 'Vale/Cupón aplicado',
    'voucher_new'            => 'Nuevo vale/Cupón',
];

$groups = [
    'Cuenta'       => ['account', 'guest_to_customer', 'password', 'password_query', 'newsletter'],
    'Pedidos'      => ['order_conf', 'bankwire', 'cheque', 'payment', 'payment_error', 'preparation', 'in_transit', 'shipped', 'order_changed', 'order_canceled'],
    'Devoluciones' => ['refund', 'credit_slip', 'order_return_state'],
    'Mensajes'     => ['contact', 'contact_form', 'forward_msg', 'reply_msg', 'order_customer_comment', 'order_merchant_comment'],
    'Productos'    => ['outofstock', 'productoutofstock', 'download_product', 'voucher', 'voucher_new'],
    'Sistema'      => ['backoffice_order', 'employee_password', 'import', 'log_alert', 'test'],
];

// ── Datos mock ────────────────────────────────────────────────────────────────

$mockProductsHtml = '
<tr>
  <td style="border:1px solid #D6D4D4;padding:8px;font-family:Open-sans,sans-serif;font-size:13px;color:#555454;">REF-MIL-0042</td>
  <td style="border:1px solid #D6D4D4;padding:8px;font-family:Open-sans,sans-serif;font-size:13px;color:#555454;"><strong>Sérum Vitamina C Intensivo 30ml</strong><br><small>x 12 unidades</small></td>
  <td style="border:1px solid #D6D4D4;padding:8px;font-family:Open-sans,sans-serif;font-size:13px;color:#555454;text-align:right;">142.500,00 EUR</td>
  <td style="border:1px solid #D6D4D4;padding:8px;font-family:Open-sans,sans-serif;font-size:13px;color:#555454;text-align:right;">1.710,00 EUR</td>
</tr>
<tr>
  <td style="border:1px solid #D6D4D4;padding:8px;font-family:Open-sans,sans-serif;font-size:13px;color:#555454;">REF-MIL-0091</td>
  <td style="border:1px solid #D6D4D4;padding:8px;font-family:Open-sans,sans-serif;font-size:13px;color:#555454;"><strong>Crema Hidratante FPS 50 · 50ml</strong><br><small>x 6 unidades</small></td>
  <td style="border:1px solid #D6D4D4;padding:8px;font-family:Open-sans,sans-serif;font-size:13px;color:#555454;text-align:right;">89.000,00 EUR</td>
  <td style="border:1px solid #D6D4D4;padding:8px;font-family:Open-sans,sans-serif;font-size:13px;color:#555454;text-align:right;">534.000,00 EUR</td>
</tr>';

$mockDeliveryHtml = '<div style="font-family:Open-sans,sans-serif;font-size:13px;line-height:1.6;color:#363A41;">
  <strong>Empresa Ejemplo · Tienda Centro</strong><br>
  María Camila Rodríguez<br>Calle Mayor 10, Local 5<br>Madrid, España<br>+34 600 123 456
</div>';

$mockInvoiceHtml = '<div style="font-family:Open-sans,sans-serif;font-size:13px;line-height:1.6;color:#363A41;">
  <strong>Empresa Ejemplo S.L.</strong><br>
  B-12345678<br>Calle Mayor 10, Local 5<br>Madrid, España
</div>';

function resolveShopLogo(string $baseUrl): string {
    $params = @include ROOT . '/app/config/parameters.php';
    if (!$params) return $baseUrl . '/img/logo.jpg';
    $p = $params['parameters'];
    try {
        $pdo  = new PDO(
            'mysql:host=' . $p['database_host'] . ';dbname=' . $p['database_name'] . ';charset=utf8',
            $p['database_user'], $p['database_password'], [PDO::ATTR_ERRMODE => PDO::ERRMODE_SILENT]
        );
        $file = $pdo->query(
            "SELECT value FROM `{$p['database_prefix']}configuration` WHERE name='PS_LOGO_MAIL' LIMIT 1"
        )->fetchColumn();
        return $file ? $baseUrl . '/img/' . $file : $baseUrl . '/img/logo.jpg';
    } catch (Exception $e) {
        return $baseUrl . '/img/logo.jpg';
    }
}

$shopLogoUrl = resolveShopLogo(LANDO_URL);

function resolveShopFavicon(string $baseUrl): string {
    $params = @include ROOT . '/app/config/parameters.php';
    if (!$params) return $baseUrl . '/favicon.ico';
    $p = $params['parameters'];
    try {
        $pdo  = new PDO(
            'mysql:host=' . $p['database_host'] . ';dbname=' . $p['database_name'] . ';charset=utf8',
            $p['database_user'], $p['database_password'], [PDO::ATTR_ERRMODE => PDO::ERRMODE_SILENT]
        );
        $file = $pdo->query(
            "SELECT value FROM `{$p['database_prefix']}configuration` WHERE name='PS_FAVICON' LIMIT 1"
        )->fetchColumn();
        return $file ? $baseUrl . '/img/' . $file : $baseUrl . '/favicon.ico';
    } catch (Exception $e) {
        return $baseUrl . '/favicon.ico';
    }
}

$shopFaviconUrl = resolveShopFavicon(LANDO_URL);

$mock = [
    'shop_name'               => 'Mi Tienda Online',
    'shop_url'                => LANDO_URL,
    'shop_logo'               => $shopLogoUrl,
    'firstname'               => 'María Camila',
    'lastname'                => 'Rodríguez',
    'email'                   => 'cliente@example.com',
    'guest_email'             => 'cliente@example.com',
    'order_name'              => 'ORD-2026-04824',
    'date'                    => '28/05/2026 10:32',
    'payment'                 => 'Transferencia bancaria',
    'carrier'                 => 'Correos Express',
    'followup'                => LANDO_URL . '/rastreo/ORD-2026-04824',
    'products'                => $mockProductsHtml,
    'discounts'               => '',
    'total_products'          => '2.244.000,00 EUR',
    'total_discounts'         => '0,00 EUR',
    'total_discounts_tax_incl'=> '0,00 EUR',
    'total_discounts_tax_excl'=> '0,00 EUR',
    'total_shipping'          => '85.000,00 EUR',
    'total_shipping_tax_incl' => '85.000,00 EUR',
    'total_shipping_tax_excl' => '71.429,00 EUR',
    'total_tax_paid'          => '202.671,00 EUR',
    'total_tax'               => '202.671,00 EUR',
    'total_wrapping'          => '0,00 EUR',
    'total_paid'              => '2.329,00 EUR',
    'total_paid_real'         => '2.329,00 EUR',
    'delivery_block_html'     => $mockDeliveryHtml,
    'invoice_block_html'      => $mockInvoiceHtml,
    'delivery_block_txt'      => 'Empresa Ejemplo · Tienda Centro, Cra 15 #45-23, Bogotá',
    'invoice_block_txt'       => 'Empresa Ejemplo S.L. · B-12345678',
    'bankwire_owner'          => 'Mi Empresa S.L.',
    'bankwire_details'        => 'Banco Ejemplo · Cta Corriente · 0049-0000-00',
    'bankwire_address'        => 'B-87654321 · Av. Principal 47, Barcelona, España',
    'check_name'              => 'Mi Empresa S.L.',
    'check_address_html'      => '<strong>Mi Empresa S.L.</strong><br>B-87654321<br>Av. Principal 47, Barcelona, España',
    'history_url'             => LANDO_URL . '/es/historial-pedidos',
    'guest_tracking_url'      => LANDO_URL . '/rastreo',
    'order_url'               => LANDO_URL . '/es/pedido/12345',
    'new_passwd'              => '● ● ● ● ● ● ● ●',
    'passwd_url'              => LANDO_URL . '/es/recuperar-contrasena?token=abc123',
    'token'                   => 'abc123xyz456',
    'voucher_num'             => 'BIENVENIDA20',
    'voucher_url'             => LANDO_URL . '/es/descuento/BIENVENIDA20',
    'amount'                  => '20%',
    'gift_message'            => '¡Gracias por confiar en Mi Tienda Online!',
    'id_order_return'         => 'DEV-2026-0012',
    'return_state'            => 'Aprobada',
    'message'                 => 'Buenos días, necesito información sobre el estado de mi pedido ORD-2026-04824.',
    'reply'                   => 'Estimada María Camila, su pedido fue despachado el 27/05 y llegará el 29/05.',
    'subject'                 => 'Consulta pedido ORD-2026-04824',
    'id_order'                => '12345',
    'order_state'             => 'En camino',
    'product_name'            => 'Sérum Vitamina C Intensivo 30ml',
    'product_ref'             => 'REF-MIL-0042',
    'credit_slip_number'      => 'NC-2026-0007',
    'credit_slip_url'         => LANDO_URL . '/es/nota-credito/NC-2026-0007',
    'refund_total'            => '142.500,00 EUR',
    'url'                     => LANDO_URL,
    'name'                    => 'María Camila Rodríguez',
    'employee_lastname'       => 'Gómez',
    'employee_firstname'      => 'Laura',
    'msg_guest'               => 'Hola, quisiera saber disponibilidad del Sérum Vitamina C en 50ml.',
    // Backoffice order
    'order_link'              => '#/orders/12345',
    // Contact / forward
    'attached_file'           => '',
    'employee'                => 'Laura Gómez',
    'messages'                => 'Hola, quisiera saber disponibilidad del Sérum Vitamina C en 50ml.',
    'comment'                 => 'Cliente VIP — prioridad alta.',
    // Download product
    'virtualProducts'         => '<tr><td>Sérum Vitamina C Intensivo 30ml</td><td><a href="#">Descargar</a></td></tr>',
    // Import
    'filename'                => 'productos_importacion_2026.csv',
    // In transit
    'meta_products'           => 'Sérum Vitamina C Intensivo 30ml x12, Crema Hidratante FPS50 x6',
    // Order conf
    'recycled_packaging_label'=> '',
    // Order return state
    'state_order_return'      => 'Aprobada y reembolsada',
    // Product out of stock
    'product'                 => 'Sérum Vitamina C Intensivo 30ml',
    'last_qty'                => '2',
    'qty'                     => '0',
    // Reply msg
    'link'                    => '#/mensajes/12345',
    // Voucher
    'voucher_amount'          => '20.000,00 EUR',
];

// ── Helpers ───────────────────────────────────────────────────────────────────

function resolveTemplate(string $key, array $dirs): array {
    foreach ($dirs as $source => $dir) {
        $path = $dir . $key . '.html';
        if (file_exists($path)) return ['path' => $path, 'source' => $source];
    }
    return ['path' => null, 'source' => null];
}

function buildTemplateList(array $dirs, array $names): array {
    $all = [];
    foreach (glob($dirs['core'] . '*.html') as $f) {
        $key      = basename($f, '.html');
        $resolved = resolveTemplate($key, $dirs);
        preg_match_all('/\{([a-zA-Z0-9_]+)\}/', file_get_contents($resolved['path']), $m);
        $vars = array_values(array_unique($m[1]));
        sort($vars);
        // Detect all sources that have this template
        $sources = [];
        foreach ($dirs as $srcName => $dir) {
            if (file_exists($dir . $key . '.html')) $sources[] = $srcName;
        }
        $all[$key] = ['name' => $names[$key] ?? $key, 'vars' => $vars, 'source' => $resolved['source'], 'sources' => $sources];
    }
    ksort($all);
    return $all;
}

// Simple line diff — returns array of ['type' => same|add|del, 'line' => string]
function lineDiff(string $a, string $b): array {
    $la = explode("\n", $a);
    $lb = explode("\n", $b);
    $n  = max(count($la), count($lb));
    $out = [];
    for ($i = 0; $i < $n; $i++) {
        $lineA = $la[$i] ?? null;
        $lineB = $lb[$i] ?? null;
        if ($lineA === $lineB)       { $out[] = ['type' => 'same', 'line' => $lineA ?? '']; }
        elseif ($lineA === null)     { $out[] = ['type' => 'add',  'line' => $lineB]; }
        elseif ($lineB === null)     { $out[] = ['type' => 'del',  'line' => $lineA]; }
        else { $out[] = ['type' => 'del', 'line' => $lineA]; $out[] = ['type' => 'add', 'line' => $lineB]; }
    }
    return $out;
}

// ── Acción: mtime del template activo (para auto-reload) ─────────────────────

if (isset($_GET['mtime']) && isset($_GET['t'])) {
    $key      = preg_replace('/[^a-z0-9_]/', '', $_GET['t']);
    $resolved = resolveTemplate($key, $mailDirs);
    header('Content-Type: application/json');
    echo json_encode(['mtime' => $resolved['path'] ? filemtime($resolved['path']) : 0]);
    exit;
}

// ── Acción: plain-text companion ──────────────────────────────────────────────

if (isset($_GET['txt']) && isset($_GET['t'])) {
    $key = preg_replace('/[^a-z0-9_]/', '', $_GET['t']);
    $txtPath = null;
    foreach ($mailDirs as $dir) {
        $p = $dir . $key . '.txt';
        if (file_exists($p)) { $txtPath = $p; break; }
    }
    header('Content-Type: text/html; charset=UTF-8');
    if (!$txtPath) {
        echo '<html><body style="font:13px/1.8 var(--font-mono);padding:32px;color:#888">No hay versión .txt para este template.</body></html>';
        exit;
    }
    $content = htmlspecialchars(file_get_contents($txtPath));
    echo '<html><head><link href="https://fonts.googleapis.com/css2?family=Geist+Mono:wght@400&display=swap" rel="stylesheet"></head><body style="font:13px/1.8 \'Geist Mono\',ui-monospace,monospace;padding:32px 40px;max-width:640px;margin:0 auto;color:#333;white-space:pre-wrap;word-break:break-word">' . $content . '</body></html>';
    exit;
}

// ── Acción: copiar al child theme ─────────────────────────────────────────────

if (isset($_GET['copy_to_child']) && isset($_GET['t'])) {
    verifyCsrf();
    $key      = preg_replace('/[^a-z0-9_]/', '', $_GET['t']);
    $resolved = resolveTemplate($key, $mailDirs);
    $destDir  = ROOT . '/themes/' . CHILD_THEME . '/mails/' . $lang . '/';
    $destFile = $destDir . $key . '.html';
    if (!$resolved['path']) { echo json_encode(['ok' => false, 'error' => 'Template not found']); exit; }
    if (!is_dir($destDir)) mkdir($destDir, 0755, true);
    $ok = copy($resolved['path'], $destFile);
    header('Content-Type: application/json');
    echo json_encode(['ok' => $ok, 'path' => 'themes/' . CHILD_THEME . '/mails/' . $lang . '/' . $key . '.html']);
    exit;
}

// ── Acción: eliminar override del child theme ─────────────────────────────────

if (isset($_GET['delete_override']) && isset($_GET['t'])) {
    verifyCsrf();
    $key      = preg_replace('/[^a-z0-9_]/', '', $_GET['t']);
    $filePath = ROOT . '/themes/' . CHILD_THEME . '/mails/' . $lang . '/' . $key . '.html';
    header('Content-Type: application/json');
    if (!file_exists($filePath)) {
        echo json_encode(['ok' => false, 'error' => 'El archivo no existe']); exit;
    }
    $ok = unlink($filePath);
    echo json_encode(['ok' => $ok, 'path' => 'themes/' . CHILD_THEME . '/mails/' . $lang . '/' . $key . '.html']);
    exit;
}

// ── Acción: diff vs core ──────────────────────────────────────────────────────

if (isset($_GET['diff']) && isset($_GET['t'])) {
    $key      = preg_replace('/[^a-z0-9_]/', '', $_GET['t']);
    $resolved = resolveTemplate($key, $mailDirs);
    $corePath = $mailDirs['core'] . $key . '.html';
    header('Content-Type: application/json');
    if (!$resolved['path'] || $resolved['source'] === 'core' || !file_exists($corePath)) {
        echo json_encode(['same' => true]); exit;
    }
    $diff = lineDiff(file_get_contents($corePath), file_get_contents($resolved['path']));
    echo json_encode(['same' => false, 'source' => $resolved['source'], 'diff' => $diff]);
    exit;
}

// ── Modo render (iframe) ──────────────────────────────────────────────────────

if (isset($_GET['t']) && isset($_GET['render'])) {
    $key      = preg_replace('/[^a-z0-9_]/', '', $_GET['t']);
    $useMock  = !empty($_GET['mock']);
    $forceSrc = isset($_GET['forcesrc']) && isset($mailDirs[$_GET['forcesrc']]) ? $_GET['forcesrc'] : null;
    if ($forceSrc) {
        $forcedPath = $mailDirs[$forceSrc] . $key . '.html';
        $resolved   = file_exists($forcedPath) ? ['path' => $forcedPath, 'source' => $forceSrc] : resolveTemplate($key, $mailDirs);
    } else {
        $resolved = resolveTemplate($key, $mailDirs);
    }
    if (!$resolved['path']) { http_response_code(404); exit('Template not found'); }
    $html = file_get_contents($resolved['path']);
    $monoStack = "'Geist Mono',ui-monospace,'SF Mono',monospace";
    $badgeMock = fn(string $raw) => '<span style="display:inline-flex;align-items:center;gap:5px;background:#fef2f2;color:#b91c1c;border:1px solid #fca5a5;border-radius:999px;padding:2px 10px 2px 3px;font-weight:500;font-family:' . $monoStack . ';font-size:11px;line-height:1.2;vertical-align:middle;"><span style="width:16px;height:16px;border-radius:50%;background:#dc2626;color:#fff;display:inline-flex;align-items:center;justify-content:center;font-size:8px;font-weight:800;flex-shrink:0;line-height:1;">{}</span>' . htmlspecialchars($raw) . '</span>';
    $badgeVar  = fn(string $raw) => '<span style="display:inline-flex;align-items:center;gap:5px;background:#eff6ff;color:#1e40af;border:1px solid #bfdbfe;border-radius:999px;padding:2px 10px 2px 3px;font-weight:500;font-family:' . $monoStack . ';font-size:11px;line-height:1.2;vertical-align:middle;"><span style="width:16px;height:16px;border-radius:50%;background:#2563eb;color:#fff;display:inline-flex;align-items:center;justify-content:center;font-size:8px;font-weight:800;flex-shrink:0;line-height:1;">{}</span>' . htmlspecialchars($raw) . '</span>';

    if ($useMock) {
        // Inside tags: substitute mock values so links/images work; outside tags: badge or substitute
        $html = preg_replace_callback('/(<[^>]*>)|\{([a-zA-Z0-9_]+)\}/', function($m) use ($mock, $badgeMock) {
            if ($m[1] !== '') {
                // Substitute variables inside attributes with mock values (makes links/imgs work)
                return preg_replace_callback('/\{([a-zA-Z0-9_]+)\}/', fn($v) => $mock[$v[1]] ?? $v[0], $m[1]);
            }
            $key = $m[2]; $val = $mock[$key] ?? null;
            return $val !== null ? $val : ($badgeMock)('{' . $key . '}');
        }, $html);
    } else {
        // Inside tags: mark <a> and <img> that have variables so we can show badges after them
        // Outside tags: show badge inline
        $html = preg_replace_callback('/(<[^>]*>)|\{([a-zA-Z0-9_]+)\}/', function($m) use ($badgeVar) {
            if ($m[1] !== '') {
                $tag = $m[1];
                if (preg_match('/^<(a|img)\b/i', $tag)) {
                    preg_match_all('/\{([a-zA-Z0-9_]+)\}/', $tag, $av);
                    if (!empty($av[1])) {
                        $vars = implode(',', array_unique($av[1]));
                        $tag  = preg_replace('/(\s*\/?>)$/', " data-ps-vars=\"{$vars}\"$1", $tag);
                    }
                }
                return $tag;
            }
            return ($badgeVar)('{' . $m[2] . '}');
        }, $html);

        // Inject script that inserts badges after <a data-ps-vars> and <img data-ps-vars>
        $inject = '<style>[data-ps-vars]{outline:1.5px dashed #93c5fd;outline-offset:2px;border-radius:2px;}</style>'
            . '<script>document.addEventListener("DOMContentLoaded",function(){'
            . 'document.querySelectorAll("[data-ps-vars]").forEach(function(el){'
            . 'var vars=el.getAttribute("data-ps-vars").split(",");'
            . 'var w=document.createElement("span");'
            . 'w.style.cssText="display:inline-flex;gap:3px;vertical-align:middle;margin-left:4px;flex-wrap:wrap;";'
            . 'vars.forEach(function(v){'
            . 'var b=document.createElement("span");'
            . 'b.style.cssText="display:inline-flex;align-items:center;gap:3px;background:#eff6ff;color:#1e40af;border:1px solid #bfdbfe;border-radius:999px;padding:1px 8px 1px 2px;font-size:10px;font-family:monospace;line-height:1.4;white-space:nowrap;";'
            . 'b.innerHTML="<span style=\'width:13px;height:13px;border-radius:50%;background:#2563eb;color:#fff;display:inline-flex;align-items:center;justify-content:center;font-size:7px;font-weight:800;flex-shrink:0;\'>{}</span>{"+v+"}";'
            . 'w.appendChild(b);});'
            . 'el.insertAdjacentElement("afterend",w);'
            . '});});</script>';
        $html = str_replace('</body>', $inject . '</body>', $html);
        if (strpos($html, '</body>') === false) $html .= $inject;
    }
    header('Content-Type: text/html; charset=UTF-8');
    echo $html;
    exit;
}

// ── Viewer principal ──────────────────────────────────────────────────────────

$templates = buildTemplateList($mailDirs, $friendlyNames);
$selected  = isset($_GET['t']) ? preg_replace('/[^a-z0-9_]/', '', $_GET['t']) : 'order_conf';
if (!isset($templates[$selected])) $selected = array_key_first($templates);
$selData   = $templates[$selected];
$useMock   = !empty($_GET['mock']);

$sourceLabels = [
    'child'  => ['label' => 'child · ' . CHILD_THEME, 'cls' => 'src-prod'],
    'parent' => ['label' => 'parent · panda',          'cls' => 'src-prev'],
    'core'   => ['label' => 'core / fallback',         'cls' => 'src-dev'],
];
$src          = $sourceLabels[$selData['source']] ?? $sourceLabels['core'];
$childInitial = strtoupper(substr(CHILD_THEME, 0, 1));
$multiLang    = count($availableLangs) > 1;
$isOverride   = $selData['source'] !== 'core';
$srcIcons     = ['child' => '↑', 'parent' => '◆', 'core' => '○'];

function renderNavItemInner(string $key, array $tpl, string $lang, bool $useMock): string {
    $srcOrder  = ['child','parent','core'];
    $srcLabels = ['child' => 'child', 'parent' => 'parent', 'core' => 'core'];
    $multi     = count($tpl['sources']) > 1;

    // Stacked dots
    $dots = '<span class="src-stack">';
    foreach ($srcOrder as $s) {
        if (in_array($s, $tpl['sources'])) {
            $dim   = ($s !== $tpl['source']) ? ' dim' : '';
            $dots .= '<span class="src-stack-dot' . $dim . '" data-src="' . $s . '"></span>';
        }
    }
    $dots .= '</span>';

    return $dots . '<span class="nav-label">' . htmlspecialchars($tpl['name']) . '</span>';
}
$srcIcon      = $srcIcons[$selData['source']] ?? '○';
$selSubject   = $subjectMap[$selected] ?? null;
// Check .txt companion exists in any of the dirs
$hasTxt = false;
foreach ($mailDirs as $dir) {
    if (file_exists($dir . $selected . '.txt')) { $hasTxt = true; break; }
}
?>
<!DOCTYPE html>
<html lang="es">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Email Templates — <?= htmlspecialchars(CHILD_THEME) ?></title>
  <link rel="icon" href="<?= htmlspecialchars($shopFaviconUrl) ?>">
  <link rel="preconnect" href="https://fonts.googleapis.com">
  <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
  <link href="https://fonts.googleapis.com/css2?family=Geist:wght@300;400;500;600;700&family=Geist+Mono:wght@400;500;600&display=swap" rel="stylesheet">
  <style>
    *, *::before, *::after { box-sizing: border-box; margin: 0; padding: 0; }
    :root {
      --sb-w: 248px; --tb-h: 48px;
      --bg: #fff; --sb-bg: #fafafa; --bd: #eaeaea;
      --tx: #000; --tx-2: #444; --tx-3: #666; --tx-4: #999;
      --hover: #fafafa; --sel: #f2f2f2; --blue: #0070f3;
      --green: #22c55e; --red: #dc2626;
      --font: 'Geist', -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
      --font-mono: 'Geist Mono', ui-monospace, 'SF Mono', monospace;
    }
    html, body { height: 100%; overflow: hidden; font-family: var(--font); background: #fafafa; color: var(--tx); }
    .layout { display: flex; height: 100vh; }

    /* ── Sidebar ── */
    .sidebar { width: var(--sb-w); min-width: var(--sb-w); background: #fafafa; border-right: 1px solid var(--bd); display: flex; flex-direction: column; overflow: hidden; }

    /* Header: logo left, count badge right */
    .sb-header { padding: 14px 16px 13px; border-bottom: 1px solid var(--bd); flex-shrink: 0; display: flex; align-items: center; justify-content: space-between; gap: 10px; min-height: 52px; }
    .sb-logo-img { max-height: 24px; max-width: 120px; object-fit: contain; display: block; }
    .sb-icon { background: var(--blue); color: #fff; border-radius: 6px; width: 28px; height: 28px; display: inline-flex; align-items: center; justify-content: center; font-size: 12px; font-weight: 700; letter-spacing: -.02em; flex-shrink: 0; }
    .sb-count { background: var(--bg); color: var(--tx-2); border: 1px solid var(--bd); border-radius: 999px; padding: 2px 9px; font-size: 11.5px; font-weight: 500; font-family: var(--font); white-space: nowrap; flex-shrink: 0; }

    /* Search box */
    .sb-search-wrap { padding: 10px 12px; border-bottom: 1px solid var(--bd); flex-shrink: 0; }
    .sb-search-box { display: flex; align-items: center; gap: 8px; border: 1px solid var(--bd); border-radius: 8px; padding: 0 10px; height: 34px; background: var(--bg); transition: border-color .15s; }
    .sb-search-box:focus-within { border-color: var(--blue); }
    .sb-search-ico { color: var(--tx-4); display: flex; align-items: center; flex-shrink: 0; }
    .sb-search { flex: 1; border: none; outline: none; font-size: 13px; font-family: var(--font); color: var(--tx); background: transparent; }
    .sb-search::placeholder { color: var(--tx-4); }
    .sb-search-kbd { background: var(--sb-bg); border: 1px solid var(--bd); border-radius: 4px; padding: 1px 5px; font-size: 10.5px; font-family: var(--font); color: var(--tx-4); flex-shrink: 0; line-height: 1.6; }

    /* Nav */
    .sb-nav { flex: 1; overflow-y: auto; padding: 6px 0 20px; scrollbar-width: thin; scrollbar-color: var(--bd) transparent; }
    .nav-group { padding: 16px 16px 4px; font-size: 11px; font-weight: 500; color: var(--tx-4); font-family: var(--font); letter-spacing: 0.04em; text-transform: uppercase; }
    .nav-group.hidden { display: none; }
    .nav-item { display: flex; align-items: center; gap: 9px; height: 36px; padding: 0 10px; margin: 0 6px; border-radius: 7px; font-size: 13px; color: var(--tx-2); text-decoration: none; transition: background .1s, color .1s; cursor: pointer; font-family: var(--font); overflow: hidden; }
    .nav-item:hover  { background: var(--hover); color: var(--tx); }
    .nav-item.active { background: var(--sel); color: var(--tx); font-weight: 500; }
    .nav-item.hidden { display: none; }
    .nav-label { flex: 1; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
    .nav-dot { width: 6px; height: 6px; border-radius: 50%; flex-shrink: 0; }
    .nav-count { background: var(--sb-bg); border: 1px solid var(--bd); border-radius: 999px; padding: 0 6px; font-size: 10.5px; font-weight: 500; color: var(--tx-4); flex-shrink: 0; line-height: 1.8; font-family: var(--font); }
    .nav-item.active .nav-count { background: var(--bg); }
    .sb-no-results { display: none; padding: 16px 16px; font-size: 12px; color: var(--tx-4); }

    /* Source stack dots — decorative only */
    .src-stack { display:inline-flex;align-items:center;flex-shrink:0; }
    .src-stack-dot { width:9px;height:9px;border-radius:50%;border:1.5px solid var(--bg);margin-left:-3px;flex-shrink:0;transition:opacity .12s; }
    .src-stack-dot:first-child { margin-left:0; }
    .src-stack-dot[data-src="child"]  { background:#2563eb; }
    .src-stack-dot[data-src="parent"] { background:#50E3C2; }
    .src-stack-dot[data-src="core"]   { background:#F97316; }
    .src-stack-dot.dim { opacity:.3; }
    .nav-item.active .src-stack-dot { border-color:var(--sel); }

    /* Source version pill — topbar */
    /* Source version pill — topbar */
    .src-pill { position:relative;display:inline-flex;align-items:center;gap:6px;border:1px solid #eaeaea;border-radius:8px;padding:0 8px 0 6px;background:#fff;flex-shrink:0;transition:border-color .12s,box-shadow .12s;cursor:pointer;height:28px;user-select:none;outline:none; }
    .src-pill:hover { border-color:#c4c4c4; }
    .src-pill.open { border-color:#a0a0a0;box-shadow:0 0 0 3px rgba(0,0,0,.06); }
    .src-pill.hidden { display:none; }
    .src-pill-dots { display:inline-flex;align-items:center;flex-shrink:0; }
    .src-pill-dot { width:9px;height:9px;border-radius:50%;border:1.5px solid #fff;margin-left:-3px;flex-shrink:0; }
    .src-pill-dot:first-child { margin-left:0; }
    .src-pill-dot[data-src="child"]  { background:#2563eb; }
    .src-pill-dot[data-src="parent"] { background:#50E3C2; }
    .src-pill-dot[data-src="core"]   { background:#F97316; }
    .src-pill-label { font-size:12px;font-weight:400;font-family:var(--font);color:#111;white-space:nowrap; }
    .src-pill-count { background:var(--sel);border-radius:999px;padding:1px 6px;font-size:10.5px;font-weight:500;color:var(--tx-3);line-height:1.6;flex-shrink:0; }
    .src-pill-chevron { display:inline-flex;align-items:center;color:var(--tx-4);flex-shrink:0;transition:transform .15s;margin-left:-2px; }
    .src-pill.open .src-pill-chevron { transform:rotate(180deg); }

    /* Source dropdown popover */
    .src-dropdown { position:absolute;top:calc(100% + 6px);left:50%;transform:translateX(-50%) translateY(-4px);min-width:160px;background:#fff;border:1px solid #eaeaea;border-radius:8px;box-shadow:0 4px 16px rgba(0,0,0,.08),0 1px 4px rgba(0,0,0,.04);z-index:200;overflow:hidden;opacity:0;pointer-events:none;transition:opacity .15s,transform .15s; }
    .src-dropdown.open { opacity:1;pointer-events:all;transform:translateX(-50%) translateY(0); }
    .src-dropdown-item { display:flex;align-items:center;gap:9px;padding:7px 12px;cursor:pointer;transition:background .1s;font-size:13px;font-weight:400;font-family:var(--font);color:#444;position:relative; }
    .src-dropdown-item:hover { background:#fafafa; }
    .src-dropdown-item.active { color:#000;font-weight:500; }
    .src-dropdown-dot { width:10px;height:10px;border-radius:50%;flex-shrink:0; }
    .src-dropdown-dot[data-src="child"]  { background:#2563eb; }
    .src-dropdown-dot[data-src="parent"] { background:#50E3C2; }
    .src-dropdown-dot[data-src="core"]   { background:#F97316; }
    .src-dropdown-check { margin-left:auto;color:var(--tx);font-size:11px;opacity:0; }
    .src-dropdown-item.active .src-dropdown-check { opacity:1; }
    .src-dropdown-meta { font-size:11px;color:var(--tx-4);margin-left:auto;font-family:var(--font-mono); }
    .sb-footer { padding:10px 14px;border-top:1px solid var(--bd);flex-shrink:0;display:flex;align-items:center;gap:10px; }
    .sb-emp-avatar { width:28px;height:28px;border-radius:50%;object-fit:cover;flex-shrink:0;background:var(--sel);display:flex;align-items:center;justify-content:center;font-size:11px;font-weight:600;color:var(--tx-3);overflow:hidden; }
    .sb-emp-avatar img { width:100%;height:100%;object-fit:cover;border-radius:50%;display:block; }
    .sb-emp-name { font-size:12.5px;font-weight:500;color:var(--tx-2);flex:1;white-space:nowrap;overflow:hidden;text-overflow:ellipsis; }
    .sb-emp-email { font-size:11px;color:var(--tx-4);white-space:nowrap;overflow:hidden;text-overflow:ellipsis;margin-top:1px; }

    /* ── Main ── */
    .main { flex: 1; display: flex; flex-direction: column; min-width: 0; overflow: hidden; }
    .topbar { height: var(--tb-h); min-height: var(--tb-h); background: #fafafa; border-bottom: 1px solid var(--bd); display: flex; align-items: center; gap: 7px; padding: 0 14px; flex-shrink: 0; }
    .tb-name { font-size: 14px; font-weight: 600; color: var(--tx); white-space: nowrap; flex-shrink: 0; letter-spacing: -.02em; }
    .tb-file { font-size: 11px; color: var(--tx-4); font-family: var(--font-mono); white-space: nowrap; flex-shrink: 0; }

    /* Source badges — Vercel pill style */
    .src-prod, .src-prev, .src-dev {
      display:inline-flex;align-items:center;gap:6px;border-radius:999px;
      padding:3px 10px 3px 3px;font-size:11.5px;font-weight:500;
      white-space:nowrap;flex-shrink:0;font-family:var(--font);border:1px solid;
    }
    .src-prod { background:#eff6ff;border-color:#bfdbfe;color:#1e40af; }
    .src-prev { background:#f0fdfb;border-color:#99f6e4;color:#0f766e; }
    .src-dev  { background:#fff7ed;border-color:#fed7aa;color:#c2410c; }
    .src-icon { width:18px;height:18px;border-radius:50%;display:inline-flex;align-items:center;justify-content:center;font-size:10px;font-weight:700;flex-shrink:0;line-height:1; }
    .src-prod .src-icon { background:#2563eb;color:#fff; }
    .src-prev .src-icon { background:#50E3C2;color:#134e4a; }
    .src-dev  .src-icon { background:#F97316;color:#fff; }
    .sep { flex: 1; min-width: 4px; }

    /* Shared control base — all topbar controls inherit this look */
    .tb-ctrl {
      flex-shrink:0;display:inline-flex;align-items:center;height:28px;
      border:1px solid var(--bd);border-radius:6px;background:var(--bg);
      font-size:12px;font-weight:400;font-family:var(--font);
      color:var(--tx-3);white-space:nowrap;transition:border-color .12s,color .12s,background .12s;
    }
    .tb-ctrl:hover { border-color:#c4c4c4;color:var(--tx-2); }

    /* Width toggle — Geist default tabs style, right-aligned in preview bar */
    .width-toggle { display:inline-flex;align-items:stretch;flex-shrink:0;margin-bottom:-1px; }
    .w-btn { display:inline-flex;align-items:center;padding:0 8px;height:36px;border:none;border-bottom:2px solid transparent;background:transparent;font-size:12.5px;font-weight:400;font-family:var(--font);color:#999;cursor:pointer;transition:color .12s,border-color .12s;white-space:nowrap; }
    .w-btn.active { color:#000;border-bottom-color:#000;font-weight:500; }
    .w-btn:hover:not(.active) { color:#444; }

    /* Lang pill dropdown */
    .lang-pill { position:relative;display:inline-flex;align-items:center;gap:5px;border:1px solid #eaeaea;border-radius:8px;padding:0 8px 0 10px;background:#fff;flex-shrink:0;transition:border-color .12s,box-shadow .12s;cursor:pointer;height:28px;user-select:none;outline:none; }
    .lang-pill:hover { border-color:#c4c4c4; }
    .lang-pill.open { border-color:#a0a0a0;box-shadow:0 0 0 3px rgba(0,0,0,.06); }
    .lang-pill-label { font-size:12px;font-weight:400;font-family:var(--font);color:#111;white-space:nowrap; }
    .lang-pill-chevron { display:inline-flex;align-items:center;color:var(--tx-4);flex-shrink:0;transition:transform .15s; }
    .lang-pill.open .lang-pill-chevron { transform:rotate(180deg); }
    .lang-dropdown { position:absolute;top:calc(100% + 6px);left:50%;transform:translateX(-50%) translateY(-4px);min-width:140px;background:#fff;border:1px solid #eaeaea;border-radius:8px;box-shadow:0 4px 16px rgba(0,0,0,.08),0 1px 4px rgba(0,0,0,.04);z-index:200;overflow:hidden;opacity:0;pointer-events:none;transition:opacity .15s,transform .15s; }
    .lang-dropdown.open { opacity:1;pointer-events:all;transform:translateX(-50%) translateY(0); }
    .lang-dropdown-item { display:flex;align-items:center;gap:9px;padding:7px 12px;cursor:pointer;transition:background .1s;font-size:13px;font-weight:400;font-family:var(--font);color:#444; }
    .lang-dropdown-item:hover { background:#fafafa; }
    .lang-dropdown-item.active { color:#000;font-weight:500; }
    .lang-dropdown-code { font-size:10.5px;font-family:var(--font-mono);color:var(--tx-4);background:var(--sb-bg);border:1px solid var(--bd);border-radius:4px;padding:1px 5px;flex-shrink:0; }
    .lang-dropdown-check { margin-left:auto;color:var(--tx);font-size:11px;opacity:0; }
    .lang-dropdown-item.active .lang-dropdown-check { opacity:1; }

    /* Topbar buttons */
    .tb-btn { flex-shrink:0;display:inline-flex;align-items:center;gap:6px;height:28px;padding:0 10px;border-radius:6px;font-size:13px;font-weight:400;cursor:pointer;border:1px solid var(--bd);background:var(--bg);color:var(--tx-3);transition:border-color .12s,color .12s,background .12s;white-space:nowrap;font-family:var(--font); }
    .tb-btn:hover { border-color:#c4c4c4;color:var(--tx-2); }
    .tb-btn.active { background:#000;color:#fff;border-color:#000; }
    .tb-btn .dot { display:none; }
    .tb-btn .count { background:var(--blue);color:#fff;border-radius:999px;padding:1px 6px;font-size:10px;font-weight:600;line-height:1.5; }

    /* Copy-to-child */
    .copy-btn { flex-shrink:0;display:inline-flex;align-items:center;gap:5px;height:28px;padding:0 10px;border-radius:6px;font-size:13px;font-weight:400;cursor:pointer;border:1px solid var(--bd);background:var(--bg);color:var(--tx-3);transition:border-color .12s,color .12s;white-space:nowrap;font-family:var(--font); }
    .copy-btn:hover { border-color:var(--blue);color:var(--blue); }
    .copy-btn.done { border-color:#16a34a;color:#16a34a;pointer-events:none; }
    .copy-btn.hidden { display:none; }

    /* Preview area */
    .iframe-wrap { flex: 1; background: #f2f2f2; overflow: hidden; display: flex; align-items: stretch; justify-content: center; transition: all .2s; }
    .iframe-wrap iframe { flex: 1; border: none; display: block; background: #fff; max-width: var(--preview-w, 100%); transition: max-width .2s, box-shadow .2s; }
    .iframe-wrap.constrained iframe { box-shadow: 0 0 0 1px rgba(0,0,0,.08), 0 4px 24px rgba(0,0,0,.08); }

    .footer { height: 36px; min-height: 36px; background: #fafafa; border-top: 1px solid var(--bd); display: flex; align-items: center; padding: 0 14px; gap: 10px; font-size: 12px; color: var(--tx-3); flex-shrink: 0; font-family: var(--font); }
    .ft-sep { width:1px; height:14px; background:var(--bd); flex-shrink:0; }
    .leg-s { display:inline-flex;align-items:center;gap:4px;background:#eff6ff;color:#1e40af;border:1px solid #bfdbfe;border-radius:999px;padding:2px 9px 2px 3px;font-family:var(--font-mono);font-weight:500;font-size:10.5px;flex-shrink:0; }
    .leg-s .li { width:14px;height:14px;border-radius:50%;background:#2563eb;color:#fff;display:inline-flex;align-items:center;justify-content:center;font-size:7px;font-weight:800;flex-shrink:0; }
    .leg-s.red { background:#fef2f2;color:#b91c1c;border-color:#fca5a5; }
    .leg-s.red .li { background:#dc2626; }

    /* ── Drawer ── */
    .drawer-overlay { position:fixed;inset:0;background:rgba(0,0,0,.08);z-index:100;opacity:0;pointer-events:none;transition:opacity .2s; }
    .drawer-overlay.open { opacity:1;pointer-events:all; }
    .drawer { position:fixed;top:10px;right:10px;bottom:10px;width:320px;background:var(--bg);border:1px solid var(--bd);border-radius:14px;z-index:101;transform:translateX(calc(100% + 20px));transition:transform .22s cubic-bezier(.4,0,.2,1);display:flex;flex-direction:column;overflow:hidden;box-shadow:0 8px 32px rgba(0,0,0,.1),0 2px 8px rgba(0,0,0,.06); }
    .drawer.open { transform:translateX(0); }
    .drawer-header { padding:14px 16px;border-bottom:1px solid var(--bd);display:flex;align-items:center;justify-content:space-between;flex-shrink:0;gap:10px; }
    .drawer-title-wrap { display:flex;align-items:center;gap:8px; }
    .drawer-title { font-size:14px;font-weight:600;color:var(--tx);letter-spacing:-.01em;font-family:var(--font); }
    .drawer-tpl { font-size:11px;color:var(--tx-3);font-family:var(--font-mono);white-space:nowrap;overflow:hidden;text-overflow:ellipsis;max-width:120px; }
    .drawer-close { background:none;border:1px solid var(--bd);border-radius:999px;cursor:pointer;color:var(--tx-3);font-size:14px;line-height:1;width:26px;height:26px;display:inline-flex;align-items:center;justify-content:center;transition:all .15s;flex-shrink:0; }
    .drawer-close:hover { color:var(--tx);border-color:#aaa;background:var(--hover); }

    /* Drawer tabs — Geist default tabs */
    .drawer-tabs { display:flex;align-items:stretch;border-bottom:1px solid #eaeaea;flex-shrink:0;padding:0 16px;gap:0; }
    .d-tab { height:40px;padding:0 4px;margin-right:16px;border:none;border-bottom:2px solid transparent;background:transparent;font-size:13px;font-weight:400;font-family:var(--font);color:#999;cursor:pointer;transition:color .12s,border-color .12s;margin-bottom:-1px;white-space:nowrap; }
    .d-tab:last-child { margin-right:0; }
    .d-tab.active { color:#000;border-bottom-color:#000;font-weight:500; }
    .d-tab:hover:not(.active) { color:#444; }

    .drawer-body { flex:1;overflow-y:auto;padding:14px 16px;scrollbar-width:thin;scrollbar-color:var(--bd) transparent; }
    .drawer-pane { display:none; }
    .drawer-pane.active { display:block; }
    .drawer-section { margin-bottom:16px; }
    .drawer-section-label { font-size:11px;font-weight:500;letter-spacing:.05em;text-transform:uppercase;color:var(--tx-4);margin-bottom:8px;font-family:var(--font); }
    .drawer-vars { display:flex;flex-wrap:wrap;gap:5px; }

    /* Variable badges — clickable */
    .var-badge { display:inline-flex;align-items:center;gap:5px;background:var(--blue);color:#fff;border-radius:999px;padding:3px 11px 3px 3px;font-family:var(--font-mono);font-size:11px;font-weight:500;white-space:nowrap;letter-spacing:-.01em;cursor:pointer;transition:opacity .15s,transform .1s;user-select:none; }
    .var-badge:hover { opacity:.85; }
    .var-badge:active { transform:scale(.95); }
    .var-badge.no-mock { background:#6b7280; }
    .var-icon { width:18px;height:18px;border-radius:50%;background:rgba(255,255,255,.22);display:inline-flex;align-items:center;justify-content:center;font-size:9px;font-weight:800;flex-shrink:0;letter-spacing:-.05em;font-family:var(--font-mono); }

    /* Diff view */
    .diff-wrap { font-family:var(--font-mono);font-size:11px;line-height:1.6;overflow-x:auto; }
    .diff-line { display:flex;gap:0;min-width:0; }
    .diff-sign { flex-shrink:0;width:16px;color:var(--tx-4);text-align:center; }
    .diff-line.add { background:#f0fdf4; }
    .diff-line.add .diff-sign { color:var(--green); }
    .diff-line.del { background:#fef2f2; }
    .diff-line.del .diff-sign { color:var(--red); }
    .diff-text { flex:1;white-space:pre-wrap;word-break:break-all;color:var(--tx-2);padding:0 4px; }
    .diff-line.add .diff-text { color:#166534; }
    .diff-line.del .diff-text { color:var(--red); }
    .diff-same { color:var(--tx-4); }
    .diff-empty { font-size:12px;color:var(--tx-3);padding:4px 0; }
    .diff-loading { font-size:12px;color:var(--tx-3);padding:4px 0; }

    /* Email client header */
    .subject-bar { border-bottom:1px solid var(--bd);background:var(--bg);padding:14px 18px 13px;flex-shrink:0; }
    .eml-subject { font-size:15px;font-weight:600;color:var(--tx);letter-spacing:-.025em;margin-bottom:10px;white-space:nowrap;overflow:hidden;text-overflow:ellipsis; }
    .eml-subject .sv { color:var(--blue);font-family:var(--font-mono);font-size:12px;font-weight:500; }
    .eml-from-row { display:flex;align-items:flex-start;gap:10px; }
    .eml-avatar { width:32px;height:32px;border-radius:50%;background:var(--blue);color:#fff;display:flex;align-items:center;justify-content:center;font-size:13px;font-weight:600;flex-shrink:0;letter-spacing:-.02em;user-select:none; }
    .eml-from-info { flex:1;min-width:0; }
    .eml-from-name { font-size:13px;font-weight:500;color:var(--tx);display:flex;align-items:baseline;gap:5px;flex-wrap:wrap; }
    .eml-from-addr { font-size:11.5px;font-weight:400;color:var(--tx-3); }
    .eml-to-row { font-size:11.5px;color:var(--tx-3);margin-top:3px; }
    .eml-to-val { color:var(--tx-2); }
    .eml-to-val .sv { color:var(--blue);font-family:var(--font-mono);font-size:10.5px; }
    .eml-time { font-size:11.5px;color:var(--tx-4);flex-shrink:0;padding-top:2px; }

    /* Preview mode bar (HTML / TXT) — Geist default tabs */
    .preview-bar { height:36px;border-bottom:1px solid #eaeaea;background:#fafafa;display:flex;align-items:stretch;padding:0 14px;gap:0;flex-shrink:0; }
    .pv-btn { display:inline-flex;align-items:center;padding:0 6px;margin-right:12px;border:none;border-bottom:2px solid transparent;background:transparent;font-size:12.5px;font-weight:400;font-family:var(--font);color:#999;cursor:pointer;transition:color .12s,border-color .12s;margin-bottom:-1px;white-space:nowrap; }
    .pv-btn:last-child { margin-right:0; }
    .pv-btn.active { color:#000;border-bottom-color:#000;font-weight:500; }
    .pv-btn:hover:not(.active) { color:#444; }
    .pv-sep { display:none; }

    /* Watch button */
    .watch-btn { display:inline-flex;align-items:center;gap:5px;background:none;border:1px solid var(--bd);border-radius:6px;padding:0 10px;height:24px;font-size:12px;font-weight:400;font-family:var(--font);color:var(--tx-3);cursor:pointer;transition:all .15s;flex-shrink:0; }
    .watch-btn:hover { border-color:#c4c4c4;color:var(--tx-2); }
    .watch-btn.active { border-color:var(--green);color:#16a34a;background:#f0fdf4; }
    .watch-dot { width:5px;height:5px;border-radius:50%;background:currentColor;flex-shrink:0; }
    .watch-btn.active .watch-dot { animation:watchRing 1.4s ease-out infinite; }
    @keyframes watchRing {
      0%   { box-shadow:0 0 0 0 var(--green); opacity:1; }
      60%  { box-shadow:0 0 0 4px transparent; opacity:.7; }
      100% { box-shadow:0 0 0 0 transparent; opacity:1; }
    }

    /* Keyboard hints */
    .kbd { display:inline-flex;align-items:center;justify-content:center;min-width:18px;height:18px;background:#fafafa;border:1px solid #e2e2e5;border-radius:5px;padding:0 4px;font-size:10px;font-family:var(--font-mono);color:var(--tx-3);line-height:1; }

    /* Clipboard toast */
    .toast { position:fixed;bottom:20px;left:50%;transform:translateX(-50%) translateY(10px);background:var(--tx);color:#fff;font-size:12px;font-family:var(--font);font-weight:500;padding:7px 14px;border-radius:8px;opacity:0;pointer-events:none;transition:opacity .2s,transform .2s;z-index:400;white-space:nowrap; }
    .toast.show { opacity:1;transform:translateX(-50%) translateY(0); }

    /* Geist Modal */
    .modal-overlay { position:fixed;inset:0;background:rgba(0,0,0,.15);z-index:300;opacity:0;pointer-events:none;transition:opacity .15s;display:flex;align-items:center;justify-content:center; }
    .modal-overlay.open { opacity:1;pointer-events:all; }
    .modal { background:#fff;border-radius:12px;box-shadow:0 8px 32px rgba(0,0,0,.12),0 2px 8px rgba(0,0,0,.06);width:480px;max-width:calc(100vw - 32px);opacity:0;transform:scale(.96) translateY(6px);transition:transform .15s cubic-bezier(.4,0,.2,1),opacity .15s; }
    .modal-overlay.open .modal { opacity:1;transform:scale(1) translateY(0); }
    .modal-body { padding:28px 28px 20px; }
    .modal-title { font-size:20px;font-weight:600;color:#000;letter-spacing:-.025em;margin-bottom:8px;font-family:var(--font); }
    .modal-desc { font-size:14px;color:#666;line-height:1.6;font-family:var(--font); }
    .modal-path { display:inline-block;background:#fafafa;border:1px solid #eaeaea;border-radius:6px;padding:7px 12px;font-family:var(--font-mono);font-size:12px;color:#444;margin-top:14px;word-break:break-all; }
    .modal-footer { padding:16px 28px 24px;display:flex;align-items:center;justify-content:flex-end;gap:8px;border-top:1px solid #eaeaea; }
    .modal-cancel { height:36px;padding:0 16px;border:1px solid #eaeaea;border-radius:6px;background:#fff;font-size:13px;font-weight:500;font-family:var(--font);color:#444;cursor:pointer;transition:border-color .12s,color .12s; }
    .modal-cancel:hover { border-color:#c4c4c4;color:#000; }
    .modal-confirm { height:36px;padding:0 16px;border:none;border-radius:6px;background:#000;font-size:13px;font-weight:500;font-family:var(--font);color:#fff;cursor:pointer;transition:background .12s; }
    .modal-confirm:hover { background:#333; }
    .modal-confirm:disabled { background:#666;cursor:not-allowed; }

    /* Delete override button */
    .del-btn { flex-shrink:0;display:none;align-items:center;gap:5px;height:28px;padding:0 10px;border-radius:6px;font-size:13px;font-weight:400;cursor:pointer;border:1px solid #fca5a5;background:#fff;color:#dc2626;transition:border-color .12s,background .12s;white-space:nowrap;font-family:var(--font); }
    .del-btn.visible { display:inline-flex; }
    .del-btn:hover { background:#fef2f2;border-color:#f87171; }

    /* Destructive modal additions */
    .modal-warning { display:flex;align-items:flex-start;gap:10px;background:#fff0f0;border:1px solid #fca5a5;border-radius:8px;padding:10px 14px;margin-top:16px; }
    .modal-warning svg { flex-shrink:0;margin-top:1px; }
    .modal-warning-text { font-size:13px;color:#b91c1c;line-height:1.5;font-family:var(--font); }
    .modal-confirm.danger { background:#dc2626; }
    .modal-confirm.danger:hover { background:#b91c1c; }
    .modal-confirm.danger:disabled { background:#f87171;cursor:not-allowed; }
  </style>
</head>
<body>
<div class="layout">

  <aside class="sidebar">
    <div class="sb-header">
      <img src="<?= htmlspecialchars($shopLogoUrl) ?>" alt="<?= htmlspecialchars(CHILD_THEME) ?>"
           class="sb-logo-img"
           onerror="this.style.display='none';document.getElementById('sb-icon-fallback').style.display='inline-flex';">
      <span class="sb-icon" id="sb-icon-fallback" style="display:none"><?= htmlspecialchars($childInitial) ?></span>
      <span class="sb-count"><?= count($templates) ?></span>
    </div>
    <div class="sb-search-wrap">
      <div class="sb-search-box">
        <span class="sb-search-ico">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><circle cx="11" cy="11" r="8"/><path d="m21 21-4.35-4.35"/></svg>
        </span>
        <input type="text" class="sb-search" id="sb-search" placeholder="Find…" autocomplete="off" spellcheck="false">
        <span class="sb-search-kbd">F</span>
      </div>
    </div>
    <nav class="sb-nav" id="nav">
      <?php
      $rendered  = [];
      $dotColors = ['child' => '#0070f3', 'parent' => '#888', 'core' => '#ccc'];
      foreach ($groups as $groupLabel => $keys):
        $validKeys = array_filter($keys, fn($k) => isset($templates[$k]));
        if (empty($validKeys)) continue;
      ?>
        <div class="nav-group" data-group="<?= htmlspecialchars($groupLabel) ?>"><?= htmlspecialchars($groupLabel) ?></div>
        <?php foreach ($validKeys as $key):
          $rendered[] = $key;
          $tpl = $templates[$key];
          $dc  = $dotColors[$tpl['source']] ?? '#ccc';
        ?>
          <a class="nav-item<?= $key === $selected ? ' active' : '' ?>"
             href="?t=<?= urlencode($key) ?>&lang=<?= urlencode($lang) ?><?= $useMock ? '&mock=1' : '' ?>"
             data-key="<?= $key ?>"
             data-name="<?= strtolower(htmlspecialchars($tpl['name'])) ?>"
             data-group="<?= htmlspecialchars($groupLabel) ?>"
             title="<?= htmlspecialchars($key) ?>.html · <?= $tpl['source'] ?>">
            <?php echo renderNavItemInner($key, $tpl, $lang, $useMock); ?>
          </a>
        <?php endforeach; ?>
      <?php endforeach; ?>
      <?php $ungrouped = array_diff(array_keys($templates), $rendered);
      if (!empty($ungrouped)): ?>
        <div class="nav-group" data-group="Otros">Otros</div>
        <?php foreach ($ungrouped as $key):
          $tpl = $templates[$key];
        ?>
          <a class="nav-item<?= $key === $selected ? ' active' : '' ?>"
             href="?t=<?= urlencode($key) ?>&lang=<?= urlencode($lang) ?><?= $useMock ? '&mock=1' : '' ?>"
             data-key="<?= $key ?>"
             data-name="<?= strtolower(htmlspecialchars($tpl['name'])) ?>"
             data-group="Otros">
            <?php echo renderNavItemInner($key, $tpl, $lang, $useMock); ?>
          </a>
        <?php endforeach; ?>
      <?php endif; ?>
      <div class="sb-no-results" id="no-results">Sin resultados</div>
    </nav>
    <?php if ($psEmployee): ?>
    <div class="sb-footer">
      <div class="sb-emp-avatar">
<?php
$empInitials  = strtoupper(substr($psEmployee['firstname'],0,1).substr($psEmployee['lastname'],0,1));
$gravatarUrl  = 'https://www.gravatar.com/avatar/' . md5(strtolower(trim($psEmployee['email']))) . '?s=56&d=404';
$psAvatarUrl  = LANDO_URL . '/img/employees/' . (int)$psEmployee['id_employee'] . '.jpg';
?>
        <img src="<?= $gravatarUrl ?>"
             alt=""
             onerror="this.src='<?= $psAvatarUrl ?>'; this.onerror=function(){this.style.display='none';this.parentElement.textContent='<?= htmlspecialchars($empInitials) ?>';}">
      </div>
      <div style="flex:1;min-width:0">
        <div class="sb-emp-name"><?= htmlspecialchars($psEmployee['firstname'] . ' ' . $psEmployee['lastname']) ?></div>
        <div class="sb-emp-email"><?= htmlspecialchars($psEmployee['email']) ?></div>
      </div>
    </div>
    <?php endif; ?>
  </aside>

  <div class="main">
    <div class="topbar">
      <span class="tb-name" id="tpl-name"><?= htmlspecialchars($selData['name']) ?></span>
      <span class="tb-file" id="tpl-file"><?= htmlspecialchars($selected) ?>.html</span>
      <span class="<?= $src['cls'] ?>" id="src-badge"><span class="src-icon"><?= $srcIcon ?></span><?= htmlspecialchars($src['label']) ?></span>
      <?php
      $srcOrder = ['child','parent','core'];
      $srcLabels = ['child'=>'child','parent'=>'parent','core'=>'core'];
      $multiSrc  = count($selData['sources']) > 1;
      ?>
      <?php
      $availSrc  = array_values(array_filter($srcOrder, fn($s) => in_array($s, $selData['sources'])));
      $activeSrcIdx = array_search($selData['source'], $availSrc);
      ?>
      <div class="src-pill<?= $multiSrc ? '' : ' hidden' ?>" id="src-pill" role="button" tabindex="0">
        <span class="src-pill-dots" id="src-pill-dots">
          <?php foreach ($availSrc as $s): ?>
            <span class="src-pill-dot" data-src="<?= $s ?>"></span>
          <?php endforeach; ?>
        </span>
        <span class="src-pill-label" id="src-pill-label"><?= htmlspecialchars($selData['source']) ?></span>
        <span class="src-pill-count" id="src-pill-count"><?= ($activeSrcIdx + 1) ?>/<?= count($availSrc) ?></span>
        <span class="src-pill-chevron">
          <svg width="10" height="6" viewBox="0 0 10 6" fill="none"><path d="M1 1l4 4 4-4" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/></svg>
        </span>
        <div class="src-dropdown" id="src-dropdown">
          <?php
          $srcDescriptions = ['child' => CHILD_THEME, 'parent' => 'panda', 'core' => 'fallback'];
          foreach ($availSrc as $s):
          ?>
            <div class="src-dropdown-item<?= $s === $selData['source'] ? ' active' : '' ?>" data-src="<?= $s ?>">
              <span class="src-dropdown-dot" data-src="<?= $s ?>"></span>
              <span><?= $srcLabels[$s] ?></span>
              <span class="src-dropdown-meta"><?= $srcDescriptions[$s] ?></span>
              <span class="src-dropdown-check">
                <svg width="12" height="9" viewBox="0 0 12 9" fill="none"><path d="M1 4l3.5 3.5L11 1" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"/></svg>
              </span>
            </div>
          <?php endforeach; ?>
        </div>
      </div>
      <div class="sep"></div>

      <?php if ($multiLang): ?>
      <div class="lang-pill" id="lang-pill" role="button" tabindex="0">
        <span class="lang-pill-label" id="lang-pill-label"><?= htmlspecialchars($langNames[$lang] ?? strtoupper($lang)) ?></span>
        <span class="lang-pill-chevron">
          <svg width="10" height="6" viewBox="0 0 10 6" fill="none"><path d="M1 1l4 4 4-4" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/></svg>
        </span>
        <div class="lang-dropdown" id="lang-dropdown">
          <?php foreach ($availableLangs as $l): ?>
            <div class="lang-dropdown-item<?= $l === $lang ? ' active' : '' ?>" data-lang="<?= htmlspecialchars($l) ?>" data-url="<?= htmlspecialchars('?t=' . urlencode($selected) . '&lang=' . urlencode($l) . ($useMock ? '&mock=1' : '')) ?>">
              <span><?= htmlspecialchars($langNames[$l] ?? $l) ?></span>
              <span class="lang-dropdown-code"><?= htmlspecialchars($l) ?></span>
              <span class="lang-dropdown-check">
                <svg width="12" height="9" viewBox="0 0 12 9" fill="none"><path d="M1 4l3.5 3.5L11 1" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"/></svg>
              </span>
            </div>
          <?php endforeach; ?>
        </div>
      </div>
      <?php endif; ?>

      <button class="tb-btn" id="vars-btn">
        Variables <span class="count" id="vars-count"><?= count($selData['vars']) ?></span>
      </button>

      <button class="copy-btn<?= $isOverride ? ' hidden' : '' ?>" id="copy-btn" title="Copiar al child theme para editar">
        ↑ child
      </button>
      <button class="del-btn<?= $isOverride ? ' visible' : '' ?>" id="del-btn" title="Eliminar override del child theme">
        <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="3 6 5 6 21 6"/><path d="M19 6l-1 14a2 2 0 0 1-2 2H8a2 2 0 0 1-2-2L5 6"/><path d="M10 11v6M14 11v6"/><path d="M9 6V4h6v2"/></svg>
        Eliminar override
      </button>

      <button class="tb-btn<?= $useMock ? ' active' : '' ?>" id="mock-btn">Datos reales</button>
    </div>

    <?php if ($selSubject): ?>
    <div class="subject-bar" id="subject-bar">
      <div class="eml-subject" id="subject-text"><?= htmlspecialchars($selSubject) ?></div>
      <div class="eml-from-row">
        <div class="eml-avatar" id="eml-avatar">
          <img src="<?= htmlspecialchars($shopFaviconUrl) ?>" alt=""
               style="width:32px;height:32px;border-radius:50%;object-fit:cover;display:block;"
               onerror="this.style.display='none';this.nextElementSibling.style.display='inline';">
          <span style="display:none"><?= htmlspecialchars($childInitial) ?></span>
        </div>
        <div class="eml-from-info">
          <div class="eml-from-name">
            <?= htmlspecialchars($mock['shop_name']) ?>
            <span class="eml-from-addr">&lt;noreply@<?= strtolower(str_replace(' ', '', CHILD_THEME)) ?>.com&gt;</span>
          </div>
          <div class="eml-to-row">Para: <span class="eml-to-val" id="eml-to-val">{firstname} {lastname}</span></div>
        </div>
        <div class="eml-time" id="eml-time"><?= date('d M. Y, H:i') ?></div>
      </div>
    </div>
    <?php endif; ?>

    <div class="preview-bar">
      <button class="pv-btn active" data-pv="html">HTML</button>
      <?php if ($hasTxt): ?>
        <button class="pv-btn" data-pv="txt">TXT</button>
      <?php endif; ?>
      <span style="flex:1"></span>
      <div class="width-toggle" title="Ancho de preview">
        <button class="w-btn active" data-w="full" title="Ancho completo">Full</button>
        <button class="w-btn" data-w="800" title="800px">800</button>
        <button class="w-btn" data-w="600" title="600px — email estándar">600</button>
      </div>
    </div>

    <div class="iframe-wrap" id="iframe-wrap">
      <iframe id="frame" src="?t=<?= urlencode($selected) ?>&lang=<?= urlencode($lang) ?>&render=1<?= $useMock ? '&mock=1' : '' ?>" sandbox="allow-scripts"></iframe>
    </div>

    <div class="footer">
      <?php if ($useMock): ?>
        <span class="leg-s red"><span class="li">{}</span>{sin_mock}</span>
        <span style="font-size:11.5px;color:var(--tx-3)">= variable sin dato mock</span>
      <?php else: ?>
        <span class="leg-s"><span class="li">{}</span>{variable}</span>
        <span style="font-size:11.5px;color:var(--tx-3)">= variable PS</span>
      <?php endif; ?>

      <span style="flex:1"></span>

      <!-- kbd shortcuts -->
      <span style="display:inline-flex;align-items:center;gap:5px;font-size:11px;color:var(--tx-4);">
        <kbd class="kbd">v</kbd><span>vars</span>
        <span style="color:var(--bd);margin:0 2px">·</span>
        <kbd class="kbd">d</kbd><span>diff</span>
        <span style="color:var(--bd);margin:0 2px">·</span>
        <kbd class="kbd">r</kbd><span>reload</span>
        <span style="color:var(--bd);margin:0 2px">·</span>
        <kbd class="kbd">↑↓</kbd><span>nav</span>
        <span style="color:var(--bd);margin:0 2px">·</span>
        <kbd class="kbd">f</kbd><span>buscar</span>
      </span>

      <span class="ft-sep"></span>
      <button class="watch-btn" id="watch-btn" title="Auto-reload cuando el archivo cambia (W)"><span class="watch-dot"></span>Watch</button>
      <span class="ft-sep"></span>

      <!-- file info -->
      <span id="footer-info" style="font-size:11.5px;color:var(--tx-2);font-family:var(--font-mono)"><?= htmlspecialchars($selected) ?>.html</span>
      <span style="font-size:11.5px;color:var(--tx-4)"><?= count($selData['vars']) ?> vars</span>
      <span style="font-size:11.5px;color:var(--tx-4)">·</span>
      <span style="font-size:11.5px;color:var(--tx-3)"><?= htmlspecialchars($src['label']) ?></span>
    </div>
  </div>
</div>

<!-- Variables Drawer -->
<div class="drawer-overlay" id="drawer-overlay"></div>
<div class="drawer" id="drawer" role="dialog" aria-modal="true" aria-label="Inspector">
  <div class="drawer-header">
    <div class="drawer-title-wrap">
      <span class="drawer-title">Inspector</span>
      <span class="drawer-tpl" id="drawer-tpl"><?= htmlspecialchars($selected) ?>.html</span>
    </div>
    <button class="drawer-close" id="drawer-close">✕</button>
  </div>
  <div class="drawer-tabs">
    <button class="d-tab active" data-tab="vars">Variables</button>
    <button class="d-tab" data-tab="diff">Diff vs core</button>
  </div>
  <div id="drawer-subject" style="padding:8px 16px 0;display:none">
    <div style="font-size:10px;font-weight:600;letter-spacing:.06em;text-transform:uppercase;color:var(--tx-4);margin-bottom:3px;font-family:var(--font)">Asunto</div>
    <div id="drawer-subject-text" style="font-size:12px;color:var(--tx-2);font-family:var(--font);padding-bottom:10px;border-bottom:1px solid var(--bd)"></div>
  </div>
  <div class="drawer-body">
    <div class="drawer-pane active" id="pane-vars"></div>
    <div class="drawer-pane" id="pane-diff"><span class="diff-loading">Cargando diff…</span></div>
  </div>
</div>

<div class="modal-overlay" id="copy-modal-overlay" role="dialog" aria-modal="true">
  <div class="modal">
    <div class="modal-body">
      <div class="modal-title">Crear override en child theme</div>
      <div class="modal-desc">Se copiará este template al directorio del child theme para que puedas personalizarlo. A partir de ese momento, tu versión tendrá prioridad sobre el template original.</div>
      <code class="modal-path" id="modal-path-display"></code>
    </div>
    <div class="modal-footer">
      <button class="modal-cancel" id="modal-cancel-btn">Cancelar</button>
      <button class="modal-confirm" id="modal-confirm-btn">Crear override</button>
    </div>
  </div>
</div>

<div class="modal-overlay" id="del-modal-overlay" role="dialog" aria-modal="true">
  <div class="modal">
    <div class="modal-body">
      <div class="modal-title">Eliminar override</div>
      <div class="modal-desc">Se eliminará tu versión personalizada de este template del child theme. El sistema volverá a usar el template del parent theme o core.</div>
      <div class="modal-warning">
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="#dc2626" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/></svg>
        <span class="modal-warning-text">Esta acción eliminará el archivo del servidor. Puedes volver a crear el override cuando quieras.</span>
      </div>
      <code class="modal-path" id="del-modal-path-display"></code>
    </div>
    <div class="modal-footer">
      <button class="modal-cancel" id="del-modal-cancel-btn">Cancelar</button>
      <button class="modal-confirm danger" id="del-modal-confirm-btn">Eliminar override</button>
    </div>
  </div>
</div>

<div class="toast" id="toast"></div>

<script>
const frame      = document.getElementById('frame');
const nameEl     = document.getElementById('tpl-name');
const fileEl     = document.getElementById('tpl-file');
const srcEl      = document.getElementById('src-badge');
const mockBtn    = document.getElementById('mock-btn');
const footEl     = document.getElementById('footer-info');
const varsBtn    = document.getElementById('vars-btn');
const varsCnt    = document.getElementById('vars-count');
const copyBtn    = document.getElementById('copy-btn');
const drawer     = document.getElementById('drawer');
const overlay    = document.getElementById('drawer-overlay');
const drawerTpl  = document.getElementById('drawer-tpl');
const paneVars   = document.getElementById('pane-vars');
const paneDiff   = document.getElementById('pane-diff');
const iframeWrap = document.getElementById('iframe-wrap');
const toast      = document.getElementById('toast');
const searchEl   = document.getElementById('sb-search');
const watchBtn    = document.getElementById('watch-btn');
const srcPill      = document.getElementById('src-pill');
const srcDropdown  = document.getElementById('src-dropdown');
const srcPillDots  = document.getElementById('src-pill-dots');
const srcPillCnt   = document.getElementById('src-pill-count');
const srcPillLabel = document.getElementById('src-pill-label');
let   pillOpen     = false;
const subjectBar = document.getElementById('subject-bar');
const subjectEl  = document.getElementById('subject-text');
const drawerSubjectWrap = document.getElementById('drawer-subject');
const drawerSubjectText = document.getElementById('drawer-subject-text');

const CSRF      = <?= json_encode($csrfToken) ?>;
const meta      = <?= json_encode($templates, JSON_UNESCAPED_UNICODE) ?>;
const srcMap    = { child:'src-prod', parent:'src-prev', core:'src-dev' };
const labelMap  = <?= json_encode(array_map(fn($v) => $v['label'], $sourceLabels), JSON_UNESCAPED_UNICODE) ?>;
const srcIcons  = { child:'↑', parent:'◆', core:'○' };
function srcBadgeHtml(src) {
  const icons = { child:'↑', parent:'◆', core:'○' };
  return `<span class="src-icon">${icons[src] || '○'}</span>${labelMap[src] || src}`;
}
const mockKeys  = <?= json_encode(array_keys($mock)) ?>;
const mockData  = <?= json_encode($mock, JSON_UNESCAPED_UNICODE) ?>;
const subjectMap = <?= json_encode($subjectMap, JSON_UNESCAPED_UNICODE) ?>;

let curKey       = <?= json_encode($selected) ?>;
let curMock      = <?= $useMock ? 'true' : 'false' ?>;
let curLang      = <?= json_encode($lang) ?>;
let curPv        = 'html';
let curForceSrc  = null; // null = auto priority, or 'child'|'parent'|'core'
let activeTab    = 'vars';
let diffCache    = {};
let watchMode    = false;
let watchTimer   = null;
let lastMtime    = 0;

// ── Subject helpers ───────────────────────────────────────────────────────────
const emlToVal  = document.getElementById('eml-to-val');

function resolveSubject(key, mock) {
  const tpl = subjectMap[key];
  if (!tpl) return null;
  if (!mock) return tpl;
  return tpl.replace(/\{([a-zA-Z0-9_]+)\}/g, (_, v) => mockData[v] ?? '{' + v + '}');
}

function highlightVars(str) {
  return str.replace(/\{([a-zA-Z0-9_]+)\}/g, m => `<span class="sv">${m}</span>`);
}

function renderSubject(key, mock) {
  const subject = subjectMap[key];
  if (!subject) {
    if (subjectBar) subjectBar.style.display = 'none';
    if (drawerSubjectWrap) drawerSubjectWrap.style.display = 'none';
    return;
  }
  if (subjectBar && subjectEl) {
    subjectBar.style.display = '';
    subjectEl.innerHTML = mock
      ? resolveSubject(key, true)
      : highlightVars(subject);
  }
  // To: field
  if (emlToVal) {
    emlToVal.innerHTML = mock
      ? (mockData['firstname'] + ' ' + mockData['lastname'])
      : highlightVars('{firstname} {lastname}');
  }
  // Drawer
  if (drawerSubjectWrap && drawerSubjectText) {
    drawerSubjectWrap.style.display = 'block';
    drawerSubjectText.textContent = resolveSubject(key, mock) || subject;
  }
}

// ── Preview mode (HTML / TXT) ─────────────────────────────────────────────────
function setPreviewMode(pv, key, lang, mock) {
  curPv = pv;
  document.querySelectorAll('.pv-btn').forEach(b => b.classList.toggle('active', b.dataset.pv === pv));
  if (pv === 'txt') {
    frame.src = '?txt=1&t=' + encodeURIComponent(key) + '&lang=' + encodeURIComponent(lang);
  } else {
    frame.src = '?t=' + encodeURIComponent(key) + '&lang=' + encodeURIComponent(lang) + '&render=1' + (mock ? '&mock=1' : '');
  }
}

document.querySelectorAll('.pv-btn').forEach(btn => {
  btn.addEventListener('click', () => setPreviewMode(btn.dataset.pv, curKey, curLang, curMock));
});

// ── Watch / auto-reload ───────────────────────────────────────────────────────
function startWatch() {
  watchMode = true;
  watchBtn.classList.add('active');
  // Get initial mtime
  fetch('?mtime=1&t=' + encodeURIComponent(curKey) + '&lang=' + encodeURIComponent(curLang))
    .then(r => r.json()).then(d => { lastMtime = d.mtime; });
  watchTimer = setInterval(() => {
    fetch('?mtime=1&t=' + encodeURIComponent(curKey) + '&lang=' + encodeURIComponent(curLang))
      .then(r => r.json())
      .then(d => {
        if (lastMtime && d.mtime > lastMtime) {
          diffCache = {}; // invalidate diff
          if (curPv === 'html') {
            frame.src = '?t=' + encodeURIComponent(curKey) + '&lang=' + encodeURIComponent(curLang) + '&render=1' + (curMock ? '&mock=1' : '') + '&_=' + Date.now();
          }
          showToast('Template recargado');
        }
        lastMtime = d.mtime;
      });
  }, 1500);
}

function stopWatch() {
  watchMode = false;
  watchBtn.classList.remove('active');
  clearInterval(watchTimer);
  watchTimer = null;
}

watchBtn.addEventListener('click', () => watchMode ? stopWatch() : startWatch());

// ── Toast ─────────────────────────────────────────────────────────────────────
let toastTimer;
function showToast(msg) {
  toast.textContent = msg;
  toast.classList.add('show');
  clearTimeout(toastTimer);
  toastTimer = setTimeout(() => toast.classList.remove('show'), 1800);
}

// ── Sidebar search ────────────────────────────────────────────────────────────
searchEl?.addEventListener('input', () => {
  const q = searchEl.value.trim().toLowerCase();
  const items = document.querySelectorAll('.nav-item');
  const groups = {};
  items.forEach(el => {
    const name = el.dataset.name || '';
    const key  = el.dataset.key  || '';
    const show = !q || name.includes(q) || key.includes(q);
    el.classList.toggle('hidden', !show);
    const g = el.dataset.group;
    if (!groups[g]) groups[g] = 0;
    if (show) groups[g]++;
  });
  document.querySelectorAll('.nav-group').forEach(el => {
    const g = el.dataset.group;
    el.classList.toggle('hidden', groups[g] === 0);
  });
  const anyVisible = Object.values(groups).some(v => v > 0);
  document.getElementById('no-results').style.display = anyVisible ? 'none' : 'block';
});

// ── Width toggle ──────────────────────────────────────────────────────────────
document.querySelectorAll('.w-btn').forEach(btn => {
  btn.addEventListener('click', () => {
    document.querySelectorAll('.w-btn').forEach(b => b.classList.remove('active'));
    btn.classList.add('active');
    const w = btn.dataset.w;
    if (w === 'full') {
      iframeWrap.style.setProperty('--preview-w', '100%');
      iframeWrap.classList.remove('constrained');
    } else {
      iframeWrap.style.setProperty('--preview-w', w + 'px');
      iframeWrap.classList.add('constrained');
    }
  });
});

// ── Drawer tabs ───────────────────────────────────────────────────────────────
document.querySelectorAll('.d-tab').forEach(tab => {
  tab.addEventListener('click', () => {
    document.querySelectorAll('.d-tab').forEach(t => t.classList.remove('active'));
    tab.classList.add('active');
    activeTab = tab.dataset.tab;
    document.querySelectorAll('.drawer-pane').forEach(p => p.classList.remove('active'));
    document.getElementById('pane-' + activeTab).classList.add('active');
    if (activeTab === 'diff') loadDiff(curKey);
  });
});

// ── Diff ──────────────────────────────────────────────────────────────────────
function loadDiff(key) {
  if (diffCache[key] !== undefined) { renderDiff(diffCache[key]); return; }
  paneDiff.innerHTML = '<span class="diff-loading">Cargando diff…</span>';
  fetch('?diff=1&t=' + encodeURIComponent(key) + '&lang=' + encodeURIComponent(curLang))
    .then(r => r.json())
    .then(data => { diffCache[key] = data; renderDiff(data); })
    .catch(() => { paneDiff.innerHTML = '<span class="diff-empty">Error al cargar el diff.</span>'; });
}

function renderDiff(data) {
  if (data.same) {
    paneDiff.innerHTML = '<span class="diff-empty">Este template no tiene override — se usa directamente el core.</span>';
    return;
  }
  const changed = data.diff.filter(l => l.type !== 'same');
  if (!changed.length) {
    paneDiff.innerHTML = '<span class="diff-empty">El override es idéntico al core.</span>';
    return;
  }
  const srcLabel = data.source === 'child' ? 'child theme' : 'parent theme';
  let html = '<div class="diff-wrap">';
  let prev = null;
  data.diff.forEach(l => {
    if (l.type === 'same') {
      if (prev !== 'same') html += '<div class="diff-line same"><span class="diff-sign"></span><span class="diff-text diff-same">···</span></div>';
    } else {
      html += `<div class="diff-line ${l.type}"><span class="diff-sign">${l.type === 'add' ? '+' : '−'}</span><span class="diff-text">${escHtml(l.line)}</span></div>`;
    }
    prev = l.type;
  });
  html += '</div>';
  html = `<p style="font-size:11px;color:var(--tx-3);margin-bottom:10px;font-family:var(--font);">Override en <strong style="color:var(--tx-2)">${srcLabel}</strong> vs core</p>` + html;
  paneDiff.innerHTML = html;
}

function escHtml(s) {
  return s.replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;');
}

// ── Variables drawer ──────────────────────────────────────────────────────────
function renderVars(key) {
  const vars    = meta[key]?.vars || [];
  const hasMock = vars.filter(v => mockKeys.includes(v));
  const noMock  = vars.filter(v => !mockKeys.includes(v));
  let html = '';
  if (hasMock.length) {
    html += '<div class="drawer-section"><div class="drawer-section-label">Con dato mock</div><div class="drawer-vars">';
    html += hasMock.map(v => `<span class="var-badge has-mock" data-var="${v}"><span class="var-icon">{}</span>{${v}}</span>`).join('');
    html += '</div></div>';
  }
  if (noMock.length) {
    html += '<div class="drawer-section"><div class="drawer-section-label">Sin dato mock</div><div class="drawer-vars">';
    html += noMock.map(v => `<span class="var-badge no-mock" data-var="${v}"><span class="var-icon">{}</span>{${v}}</span>`).join('');
    html += '</div></div>';
  }
  if (!vars.length) html = '<p style="font-size:12px;color:var(--tx-3);padding:4px 0">Sin variables en este template.</p>';
  paneVars.innerHTML = html;
  // Clipboard on badge click
  paneVars.querySelectorAll('.var-badge').forEach(badge => {
    badge.addEventListener('click', () => {
      const text = '{' + badge.dataset.var + '}';
      navigator.clipboard?.writeText(text).then(() => showToast('Copiado: ' + text)).catch(() => showToast(text));
    });
  });
}

function openDrawer() {
  drawerTpl.textContent = curKey + '.html';
  renderVars(curKey);
  if (activeTab === 'diff') loadDiff(curKey);
  drawer.classList.add('open');
  overlay.classList.add('open');
}

function closeDrawer() {
  drawer.classList.remove('open');
  overlay.classList.remove('open');
}

// ── Copy to child modal ───────────────────────────────────────────────────────
const delBtn           = document.getElementById('del-btn');
const delModalOverlay  = document.getElementById('del-modal-overlay');
const delModalPath     = document.getElementById('del-modal-path-display');
const delModalCancel   = document.getElementById('del-modal-cancel-btn');
const delModalConfirm  = document.getElementById('del-modal-confirm-btn');
const copyModalOverlay = document.getElementById('copy-modal-overlay');
const copyModalPath    = document.getElementById('modal-path-display');
const modalCancelBtn   = document.getElementById('modal-cancel-btn');
const modalConfirmBtn  = document.getElementById('modal-confirm-btn');

function openCopyModal() {
  const destPath = 'themes/<?= CHILD_THEME ?>/' + 'mails/' + curLang + '/' + curKey + '.html';
  if (copyModalPath) copyModalPath.textContent = destPath;
  copyModalOverlay?.classList.add('open');
}
function closeCopyModal() {
  copyModalOverlay?.classList.remove('open');
}

copyBtn?.addEventListener('click', openCopyModal);
modalCancelBtn?.addEventListener('click', closeCopyModal);
copyModalOverlay?.addEventListener('click', e => { if (e.target === copyModalOverlay) closeCopyModal(); });
document.addEventListener('keydown', e => { if (e.key === 'Escape' && copyModalOverlay?.classList.contains('open')) closeCopyModal(); });

modalConfirmBtn?.addEventListener('click', () => {
  modalConfirmBtn.disabled = true;
  modalConfirmBtn.textContent = 'Copiando…';
  fetch('?copy_to_child=1&csrf=' + encodeURIComponent(CSRF) + '&t=' + encodeURIComponent(curKey) + '&lang=' + encodeURIComponent(curLang))
    .then(r => r.json())
    .then(data => {
      closeCopyModal();
      modalConfirmBtn.disabled = false;
      modalConfirmBtn.textContent = 'Crear override';
      if (data.ok) {
        copyBtn.textContent = '✓ child';
        copyBtn.classList.add('done');
        showToast('Copiado a ' + data.path);
        meta[curKey].source = 'child';
        srcEl.className = 'src-prod';
        srcEl.innerHTML = srcBadgeHtml('child');
        diffCache = {};
      } else {
        showToast('Error al copiar el template');
      }
    })
    .catch(() => {
      closeCopyModal();
      modalConfirmBtn.disabled = false;
      modalConfirmBtn.textContent = 'Crear override';
      showToast('Error de red');
    });
});

// ── Main load function ────────────────────────────────────────────────────────
function buildUrl(key, mock, lang, forceSrc) {
  let url = '?t=' + encodeURIComponent(key) + '&lang=' + encodeURIComponent(lang) + (mock ? '&mock=1' : '');
  if (forceSrc) url += '&forcesrc=' + encodeURIComponent(forceSrc);
  return url;
}

function buildFrameUrl(key, lang, mock, forceSrc, pv) {
  if (pv === 'txt') return '?txt=1&t=' + encodeURIComponent(key) + '&lang=' + encodeURIComponent(lang);
  let url = '?t=' + encodeURIComponent(key) + '&lang=' + encodeURIComponent(lang) + '&render=1' + (mock ? '&mock=1' : '');
  if (forceSrc) url += '&forcesrc=' + encodeURIComponent(forceSrc);
  return url;
}

// Update source stack dots in the nav for a given key + active source
function updateSrcStack(key, activeSrc) {
  document.querySelectorAll(`.src-stack-dot[data-key="${key}"]`).forEach(dot => {
    dot.classList.toggle('dim', dot.dataset.src !== activeSrc);
  });
}

function load(key, mock, lang) {
  curKey = key; curMock = mock; curLang = lang;
  const d = meta[key], src = d.source || 'core';
  nameEl.textContent  = d.name;
  fileEl.textContent  = key + '.html';
  srcEl.className  = srcMap[src] || 'src-dev';
  srcEl.innerHTML  = srcBadgeHtml(src);
  varsCnt.textContent = d.vars.length;
  if (footEl) footEl.textContent = key + '.html';
  mockBtn.classList.toggle('active', mock);
  curPv = 'html';
  curForceSrc = null;
  document.querySelectorAll('.pv-btn').forEach(b => b.classList.toggle('active', b.dataset.pv === 'html'));
  frame.src = buildFrameUrl(key, lang, mock, null, 'html');
  updateSrcStack(key, d.source);
  // Update src pill
  const srcOrder = ['child','parent','core'];
  const available = d.sources || [];
  const multi = available.length > 1;
  if (srcPill) srcPill.classList.toggle('hidden', !multi);
  if (multi) {
    if (srcPillDots) srcPillDots.innerHTML = srcOrder.filter(s => available.includes(s))
      .map(s => `<span class="src-pill-dot" data-src="${s}"></span>`).join('');
    const idx = srcOrder.filter(s => available.includes(s)).indexOf(d.source);
    if (srcPillLabel) srcPillLabel.textContent = d.source;
    if (srcPillCnt) srcPillCnt.textContent = (idx + 1) + '/' + available.length;
    if (srcDropdown) srcDropdown.innerHTML = srcOrder.filter(s => available.includes(s)).map(s => {
      const descs = {child:'<?= CHILD_THEME ?>',parent:'panda',core:'fallback'};
      return `<div class="src-dropdown-item${s===d.source?' active':''}" data-src="${s}"><span class="src-dropdown-dot" data-src="${s}"></span><span>${s}</span><span class="src-dropdown-meta">${descs[s]}</span><span class="src-dropdown-check"><svg width="12" height="9" viewBox="0 0 12 9" fill="none"><path d="M1 4l3.5 3.5L11 1" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"/></svg></span></div>`;
    }).join('');
  }
  history.replaceState(null, '', buildUrl(key, mock, lang));
  document.querySelectorAll('.nav-item').forEach(el => {
    el.classList.toggle('active', el.dataset.key === key);
    el.href = buildUrl(el.dataset.key, mock, lang);
  });
  if (copyBtn) {
    copyBtn.classList.toggle('hidden', src === 'child');
    copyBtn.textContent = '↑ child';
    copyBtn.classList.remove('done');
  }
  if (delBtn) delBtn.classList.toggle('visible', src === 'child');
  // Subject
  renderSubject(key, mock);
  // Reset watch mtime so it re-fetches for new template
  lastMtime = 0;
  if (drawer.classList.contains('open')) {
    drawerTpl.textContent = key + '.html';
    renderVars(key);
    if (activeTab === 'diff') loadDiff(key);
  }
}

// ── Event listeners ───────────────────────────────────────────────────────────
// ── Delete override modal ─────────────────────────────────────────────────────
function openDelModal() {
  const destPath = 'themes/<?= CHILD_THEME ?>/' + 'mails/' + curLang + '/' + curKey + '.html';
  if (delModalPath) delModalPath.textContent = destPath;
  delModalOverlay?.classList.add('open');
}
function closeDelModal() { delModalOverlay?.classList.remove('open'); }

delBtn?.addEventListener('click', openDelModal);
delModalCancel?.addEventListener('click', closeDelModal);
delModalOverlay?.addEventListener('click', e => { if (e.target === delModalOverlay) closeDelModal(); });
document.addEventListener('keydown', e => { if (e.key === 'Escape' && delModalOverlay?.classList.contains('open')) closeDelModal(); });

delModalConfirm?.addEventListener('click', () => {
  delModalConfirm.disabled = true;
  delModalConfirm.textContent = 'Eliminando…';
  fetch('?delete_override=1&csrf=' + encodeURIComponent(CSRF) + '&t=' + encodeURIComponent(curKey) + '&lang=' + encodeURIComponent(curLang))
    .then(r => r.json())
    .then(data => {
      closeDelModal();
      delModalConfirm.disabled = false;
      delModalConfirm.textContent = 'Eliminar override';
      if (data.ok) {
        showToast('Override eliminado');
        // Recalculate source — remove 'child' from sources, update badge
        if (meta[curKey]?.sources) {
          meta[curKey].sources = meta[curKey].sources.filter(s => s !== 'child');
          meta[curKey].source  = meta[curKey].sources[0] || 'core';
        }
        const newSrc = meta[curKey]?.source || 'core';
        srcEl.className = srcMap[newSrc] || 'src-dev';
        srcEl.innerHTML = srcBadgeHtml(newSrc);
        if (copyBtn) { copyBtn.classList.remove('hidden'); copyBtn.textContent = '↑ child'; copyBtn.classList.remove('done'); }
        if (delBtn) delBtn.classList.remove('visible');
        diffCache = {};
        // Reload iframe with new source
        frame.src = buildFrameUrl(curKey, curLang, curMock, null, curPv);
      } else {
        showToast('Error al eliminar: ' + (data.error || 'desconocido'));
      }
    })
    .catch(() => {
      closeDelModal();
      delModalConfirm.disabled = false;
      delModalConfirm.textContent = 'Eliminar override';
      showToast('Error de red');
    });
});

// ── Source pill dropdown ──────────────────────────────────────────────────────
function openPill()  { pillOpen = true;  srcPill?.classList.add('open');    srcDropdown?.classList.add('open'); }
function closePill() { pillOpen = false; srcPill?.classList.remove('open'); srcDropdown?.classList.remove('open'); }

function selectSource(src) {
  closePill();
  const srcOrder = ['child','parent','core'];
  const available = meta[curKey]?.sources || [];
  curForceSrc = src;
  frame.src = buildFrameUrl(curKey, curLang, curMock, src, curPv);
  srcEl.className = srcMap[src] || 'src-dev';
  srcEl.innerHTML = srcBadgeHtml(src);
  updateSrcStack(curKey, src);
  const idx = srcOrder.filter(s => available.includes(s)).indexOf(src);
  if (srcPillLabel) srcPillLabel.textContent = src;
  if (srcPillCnt)   srcPillCnt.textContent = (idx + 1) + '/' + available.length;
  srcDropdown?.querySelectorAll('.src-dropdown-item').forEach(el =>
    el.classList.toggle('active', el.dataset.src === src));
  diffCache = {};
}

srcPill?.addEventListener('click', e => {
  e.stopPropagation();
  pillOpen ? closePill() : openPill();
});
srcPill?.addEventListener('keydown', e => {
  if (e.key === 'Enter' || e.key === ' ') { e.preventDefault(); pillOpen ? closePill() : openPill(); }
  if (e.key === 'Escape') closePill();
});
srcDropdown?.addEventListener('click', e => {
  e.stopPropagation();
  const item = e.target.closest('.src-dropdown-item');
  if (item) selectSource(item.dataset.src);
});
document.addEventListener('click', () => { if (pillOpen) closePill(); });

document.getElementById('nav').addEventListener('click', e => {
  const i = e.target.closest('.nav-item');
  if (i) { e.preventDefault(); load(i.dataset.key, curMock, curLang); }
});
mockBtn.addEventListener('click', () => load(curKey, !curMock, curLang));
varsBtn.addEventListener('click', openDrawer);
overlay.addEventListener('click', closeDrawer);
document.getElementById('drawer-close').addEventListener('click', closeDrawer);
document.addEventListener('keydown', e => {
  if (e.key === 'Escape') { closeDrawer(); return; }
  if (document.activeElement === searchEl) return;
  if (e.key === 'ArrowDown' || e.key === 'ArrowUp') {
    e.preventDefault();
    const navKeys = Array.from(document.querySelectorAll('.nav-item:not(.hidden)')).map(el => el.dataset.key);
    const idx  = navKeys.indexOf(curKey);
    const next = e.key === 'ArrowDown' ? navKeys[idx + 1] : navKeys[idx - 1];
    if (next) {
      load(next, curMock, curLang);
      document.querySelector(`.nav-item[data-key="${next}"]`)?.scrollIntoView({ block: 'nearest' });
    }
    return;
  }
  if (e.metaKey || e.ctrlKey || e.altKey) return;
  switch (e.key.toLowerCase()) {
    case 'v':
      e.preventDefault();
      if (drawer.classList.contains('open') && activeTab === 'vars') { closeDrawer(); break; }
      activeTab = 'vars';
      document.querySelectorAll('.d-tab').forEach(t => t.classList.toggle('active', t.dataset.tab === 'vars'));
      document.querySelectorAll('.drawer-pane').forEach(p => p.classList.toggle('active', p.id === 'pane-vars'));
      openDrawer();
      break;
    case 'd':
      e.preventDefault();
      if (drawer.classList.contains('open') && activeTab === 'diff') { closeDrawer(); break; }
      activeTab = 'diff';
      document.querySelectorAll('.d-tab').forEach(t => t.classList.toggle('active', t.dataset.tab === 'diff'));
      document.querySelectorAll('.drawer-pane').forEach(p => p.classList.toggle('active', p.id === 'pane-diff'));
      openDrawer();
      loadDiff(curKey);
      break;
    case 'r':
      e.preventDefault();
      diffCache = {};
      setPreviewMode(curPv, curKey, curLang, curMock);
      showToast('Recargado');
      break;
    case 'w':
      e.preventDefault();
      watchMode ? stopWatch() : startWatch();
      break;
    case 'f':
      e.preventDefault();
      searchEl?.focus();
      searchEl?.select();
      break;
  }
});
// Init subject for the initially selected template
renderSubject(curKey, curMock);

<?php if ($multiLang): ?>
// ── Language pill dropdown ────────────────────────────────────────────────────
const langPill     = document.getElementById('lang-pill');
const langDropdown = document.getElementById('lang-dropdown');
const langPillLabel= document.getElementById('lang-pill-label');
let   langPillOpen = false;

function openLangPill()  { langPillOpen = true;  langPill?.classList.add('open');    langDropdown?.classList.add('open'); }
function closeLangPill() { langPillOpen = false; langPill?.classList.remove('open'); langDropdown?.classList.remove('open'); }

langPill?.addEventListener('click', e => {
  e.stopPropagation();
  langPillOpen ? closeLangPill() : openLangPill();
});
langPill?.addEventListener('keydown', e => {
  if (e.key === 'Enter' || e.key === ' ') { e.preventDefault(); langPillOpen ? closeLangPill() : openLangPill(); }
  if (e.key === 'Escape') closeLangPill();
});
langDropdown?.addEventListener('click', e => {
  e.stopPropagation();
  const item = e.target.closest('.lang-dropdown-item');
  if (item) window.location.href = item.dataset.url;
});
document.addEventListener('click', () => { if (langPillOpen) closeLangPill(); });
<?php endif; ?>
</script>
</body>
</html>
