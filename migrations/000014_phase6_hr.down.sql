DROP TABLE IF EXISTS attendances;
DROP TABLE IF EXISTS shifts;
DROP TABLE IF EXISTS employees;

DELETE FROM role_permissions WHERE permission_id IN (SELECT id FROM permissions WHERE code LIKE 'hr:%');
DELETE FROM permissions WHERE code LIKE 'hr:%';
