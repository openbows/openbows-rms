-- 第 0 期：平台底座（门店 / 账号 / RBAC）

CREATE TABLE IF NOT EXISTS stores (
    id         BIGSERIAL PRIMARY KEY,
    code       VARCHAR(64)  NOT NULL,
    name       VARCHAR(128) NOT NULL,
    status     BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ  NULL
);
COMMENT ON TABLE stores IS '门店';
COMMENT ON COLUMN stores.id IS '主键';
COMMENT ON COLUMN stores.code IS '门店编码';
COMMENT ON COLUMN stores.name IS '门店名称';
COMMENT ON COLUMN stores.status IS '是否启用';
COMMENT ON COLUMN stores.created_at IS '创建时间';
COMMENT ON COLUMN stores.updated_at IS '更新时间';
COMMENT ON COLUMN stores.deleted_at IS '删除时间';
CREATE UNIQUE INDEX uk_stores_code ON stores (code) WHERE deleted_at IS NULL;
CREATE INDEX idx_stores_deleted_at ON stores (deleted_at);

CREATE TABLE IF NOT EXISTS users (
    id            BIGSERIAL PRIMARY KEY,
    store_id      BIGINT       NOT NULL,
    username      VARCHAR(64)  NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    phone         VARCHAR(32)  NULL,
    display_name  VARCHAR(64)  NOT NULL DEFAULT '',
    status        BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at    TIMESTAMPTZ  NULL,
    CONSTRAINT fk_users_store_id FOREIGN KEY (store_id) REFERENCES stores (id)
);
COMMENT ON TABLE users IS '账号';
COMMENT ON COLUMN users.id IS '主键';
COMMENT ON COLUMN users.store_id IS '门店ID';
COMMENT ON COLUMN users.username IS '登录用户名';
COMMENT ON COLUMN users.password_hash IS '密码哈希';
COMMENT ON COLUMN users.phone IS '手机号';
COMMENT ON COLUMN users.display_name IS '显示名称';
COMMENT ON COLUMN users.status IS '是否启用';
COMMENT ON COLUMN users.created_at IS '创建时间';
COMMENT ON COLUMN users.updated_at IS '更新时间';
COMMENT ON COLUMN users.deleted_at IS '删除时间';
CREATE UNIQUE INDEX uk_users_username ON users (username) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_store_id ON users (store_id);
CREATE INDEX idx_users_deleted_at ON users (deleted_at);

CREATE TABLE IF NOT EXISTS roles (
    id         BIGSERIAL PRIMARY KEY,
    store_id   BIGINT       NOT NULL,
    code       VARCHAR(64)  NOT NULL,
    name       VARCHAR(128) NOT NULL,
    remark     VARCHAR(255) NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ  NULL,
    CONSTRAINT fk_roles_store_id FOREIGN KEY (store_id) REFERENCES stores (id)
);
COMMENT ON TABLE roles IS '角色';
COMMENT ON COLUMN roles.id IS '主键';
COMMENT ON COLUMN roles.store_id IS '门店ID';
COMMENT ON COLUMN roles.code IS '角色编码';
COMMENT ON COLUMN roles.name IS '角色名称';
COMMENT ON COLUMN roles.remark IS '备注';
COMMENT ON COLUMN roles.created_at IS '创建时间';
COMMENT ON COLUMN roles.updated_at IS '更新时间';
COMMENT ON COLUMN roles.deleted_at IS '删除时间';
CREATE UNIQUE INDEX uk_roles_store_id_code ON roles (store_id, code) WHERE deleted_at IS NULL;
CREATE INDEX idx_roles_store_id ON roles (store_id);
CREATE INDEX idx_roles_deleted_at ON roles (deleted_at);

CREATE TABLE IF NOT EXISTS permissions (
    id         BIGSERIAL PRIMARY KEY,
    code       VARCHAR(128) NOT NULL,
    name       VARCHAR(128) NOT NULL,
    module     VARCHAR(64)  NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ  NULL
);
COMMENT ON TABLE permissions IS '权限点';
COMMENT ON COLUMN permissions.id IS '主键';
COMMENT ON COLUMN permissions.code IS '权限编码';
COMMENT ON COLUMN permissions.name IS '权限名称';
COMMENT ON COLUMN permissions.module IS '所属模块';
COMMENT ON COLUMN permissions.created_at IS '创建时间';
COMMENT ON COLUMN permissions.updated_at IS '更新时间';
COMMENT ON COLUMN permissions.deleted_at IS '删除时间';
CREATE UNIQUE INDEX uk_permissions_code ON permissions (code) WHERE deleted_at IS NULL;
CREATE INDEX idx_permissions_deleted_at ON permissions (deleted_at);

CREATE TABLE IF NOT EXISTS role_permissions (
    id            BIGSERIAL PRIMARY KEY,
    role_id       BIGINT      NOT NULL,
    permission_id BIGINT      NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at    TIMESTAMPTZ NULL,
    CONSTRAINT fk_role_permissions_role_id FOREIGN KEY (role_id) REFERENCES roles (id),
    CONSTRAINT fk_role_permissions_permission_id FOREIGN KEY (permission_id) REFERENCES permissions (id)
);
COMMENT ON TABLE role_permissions IS '角色-权限';
COMMENT ON COLUMN role_permissions.id IS '主键';
COMMENT ON COLUMN role_permissions.role_id IS '角色ID';
COMMENT ON COLUMN role_permissions.permission_id IS '权限ID';
COMMENT ON COLUMN role_permissions.created_at IS '创建时间';
COMMENT ON COLUMN role_permissions.updated_at IS '更新时间';
COMMENT ON COLUMN role_permissions.deleted_at IS '删除时间';
CREATE UNIQUE INDEX uk_role_permissions_role_perm ON role_permissions (role_id, permission_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_role_permissions_role_id ON role_permissions (role_id);
CREATE INDEX idx_role_permissions_permission_id ON role_permissions (permission_id);
CREATE INDEX idx_role_permissions_deleted_at ON role_permissions (deleted_at);

CREATE TABLE IF NOT EXISTS user_roles (
    id         BIGSERIAL PRIMARY KEY,
    user_id    BIGINT      NOT NULL,
    role_id    BIGINT      NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ NULL,
    CONSTRAINT fk_user_roles_user_id FOREIGN KEY (user_id) REFERENCES users (id),
    CONSTRAINT fk_user_roles_role_id FOREIGN KEY (role_id) REFERENCES roles (id)
);
COMMENT ON TABLE user_roles IS '用户-角色';
COMMENT ON COLUMN user_roles.id IS '主键';
COMMENT ON COLUMN user_roles.user_id IS '用户ID';
COMMENT ON COLUMN user_roles.role_id IS '角色ID';
COMMENT ON COLUMN user_roles.created_at IS '创建时间';
COMMENT ON COLUMN user_roles.updated_at IS '更新时间';
COMMENT ON COLUMN user_roles.deleted_at IS '删除时间';
CREATE UNIQUE INDEX uk_user_roles_user_role ON user_roles (user_id, role_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_user_roles_user_id ON user_roles (user_id);
CREATE INDEX idx_user_roles_role_id ON user_roles (role_id);
CREATE INDEX idx_user_roles_deleted_at ON user_roles (deleted_at);

-- seed：默认门店
INSERT INTO stores (id, code, name, status)
VALUES (1, 'DEFAULT', '默认门店', TRUE);

-- seed：权限点
INSERT INTO permissions (id, code, name, module) VALUES
(1, 'store:read',   '门店查看', 'store'),
(2, 'store:write',  '门店管理', 'store'),
(3, 'user:read',    '用户查看', 'user'),
(4, 'user:write',   '用户管理', 'user'),
(5, 'role:read',    '角色查看', 'role'),
(6, 'role:write',   '角色管理', 'role');

-- seed：管理员角色
INSERT INTO roles (id, store_id, code, name, remark)
VALUES (1, 1, 'admin', '管理员', '默认门店管理员');

INSERT INTO role_permissions (role_id, permission_id)
SELECT 1, id FROM permissions WHERE deleted_at IS NULL;

-- seed：admin / Admin@123（bcrypt）
-- hash 由 scripts/gen_bcrypt.go 生成后写入；此处使用预置哈希
INSERT INTO users (id, store_id, username, password_hash, phone, display_name, status)
VALUES (
    1,
    1,
    'admin',
    '$2a$10$AmEtJ/zsM/Bl5LEq1pbXpeHzVUkZaV8SRB6jYxYJoYPIsFtjZRQ92',
    NULL,
    '系统管理员',
    TRUE
);

INSERT INTO user_roles (user_id, role_id) VALUES (1, 1);

SELECT setval(pg_get_serial_sequence('stores', 'id'), (SELECT MAX(id) FROM stores));
SELECT setval(pg_get_serial_sequence('users', 'id'), (SELECT MAX(id) FROM users));
SELECT setval(pg_get_serial_sequence('roles', 'id'), (SELECT MAX(id) FROM roles));
SELECT setval(pg_get_serial_sequence('permissions', 'id'), (SELECT MAX(id) FROM permissions));
