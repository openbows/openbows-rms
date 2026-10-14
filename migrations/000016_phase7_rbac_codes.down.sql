DELETE FROM role_permissions
WHERE permission_id IN (
    SELECT id FROM permissions WHERE code IN (
        'menu:read','menu:write','table:read','table:write',
        'order:read','order:write','payment:write',
        'kds:read','kds:write','queue:write','reservation:write',
        'member:read','member:write','stock:read','stock:write',
        'purchase:read','purchase:write','finance:read','finance:write'
    )
);
DELETE FROM permissions WHERE code IN (
    'menu:read','menu:write','table:read','table:write',
    'order:read','order:write','payment:write',
    'kds:read','kds:write','queue:write','reservation:write',
    'member:read','member:write','stock:read','stock:write',
    'purchase:read','purchase:write','finance:read','finance:write'
);
