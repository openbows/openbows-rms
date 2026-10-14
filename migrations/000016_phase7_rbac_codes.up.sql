-- 第 7 期：补齐业务权限码并赋给 admin 角色

INSERT INTO permissions (code, name, module)
SELECT v.code, v.name, v.module
FROM (VALUES
    ('menu:read', '菜单查看', 'menu'),
    ('menu:write', '菜单管理', 'menu'),
    ('table:read', '桌台查看', 'table'),
    ('table:write', '桌台管理', 'table'),
    ('order:read', '订单查看', 'order'),
    ('order:write', '订单操作', 'order'),
    ('payment:write', '收银支付', 'payment'),
    ('kds:read', '厨显查看', 'kds'),
    ('kds:write', '厨显操作', 'kds'),
    ('queue:write', '排队叫号', 'queue'),
    ('reservation:write', '预订管理', 'reservation'),
    ('member:read', '会员查看', 'member'),
    ('member:write', '会员管理', 'member'),
    ('stock:read', '库存查看', 'stock'),
    ('stock:write', '库存管理', 'stock'),
    ('purchase:read', '采购查看', 'purchase'),
    ('purchase:write', '采购管理', 'purchase'),
    ('finance:read', '财务查看', 'finance'),
    ('finance:write', '财务日结对账', 'finance')
) AS v(code, name, module)
WHERE NOT EXISTS (SELECT 1 FROM permissions p WHERE p.code = v.code AND p.deleted_at IS NULL);

INSERT INTO role_permissions (role_id, permission_id)
SELECT 1, p.id FROM permissions p
WHERE p.deleted_at IS NULL
  AND NOT EXISTS (
      SELECT 1 FROM role_permissions rp
      WHERE rp.role_id = 1 AND rp.permission_id = p.id AND rp.deleted_at IS NULL
  );
