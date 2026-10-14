-- 第 6 期：排班考勤 / 人力资源（精简）
-- RBAC 权限码预留（与收银权限分离，本阶段接口仍走登录鉴权）

INSERT INTO permissions (code, name, module)
SELECT v.code, v.name, v.module
FROM (VALUES
    ('hr:employee:read', '员工查看', 'hr'),
    ('hr:employee:write', '员工管理', 'hr'),
    ('hr:shift:write', '排班维护', 'hr'),
    ('hr:attendance:write', '考勤打卡', 'hr')
) AS v(code, name, module)
WHERE NOT EXISTS (SELECT 1 FROM permissions p WHERE p.code = v.code AND p.deleted_at IS NULL);

INSERT INTO role_permissions (role_id, permission_id)
SELECT 1, p.id FROM permissions p
WHERE p.code LIKE 'hr:%' AND p.deleted_at IS NULL
  AND NOT EXISTS (
      SELECT 1 FROM role_permissions rp
      WHERE rp.role_id = 1 AND rp.permission_id = p.id AND rp.deleted_at IS NULL
  );

CREATE TABLE IF NOT EXISTS employees (
    id         BIGSERIAL PRIMARY KEY,
    store_id   BIGINT       NOT NULL,
    user_id    BIGINT       NULL,
    emp_no     VARCHAR(64)  NOT NULL,
    name       VARCHAR(128) NOT NULL,
    phone      VARCHAR(32)  NOT NULL DEFAULT '',
    job_title  VARCHAR(64)  NOT NULL DEFAULT '',
    status     BOOLEAN      NOT NULL DEFAULT TRUE,
    hired_at   DATE         NULL,
    remark     VARCHAR(512) NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ  NULL,
    CONSTRAINT fk_employees_store_id FOREIGN KEY (store_id) REFERENCES stores (id),
    CONSTRAINT fk_employees_user_id FOREIGN KEY (user_id) REFERENCES users (id)
);

COMMENT ON TABLE employees IS '员工';
COMMENT ON COLUMN employees.id IS '主键';
COMMENT ON COLUMN employees.store_id IS '门店ID';
COMMENT ON COLUMN employees.user_id IS '关联登录用户ID（可选）';
COMMENT ON COLUMN employees.emp_no IS '工号（门店内唯一）';
COMMENT ON COLUMN employees.name IS '姓名';
COMMENT ON COLUMN employees.phone IS '手机号';
COMMENT ON COLUMN employees.job_title IS '岗位';
COMMENT ON COLUMN employees.status IS '是否在职启用';
COMMENT ON COLUMN employees.hired_at IS '入职日期';
COMMENT ON COLUMN employees.remark IS '备注';
COMMENT ON COLUMN employees.created_at IS '创建时间';
COMMENT ON COLUMN employees.updated_at IS '更新时间';
COMMENT ON COLUMN employees.deleted_at IS '删除时间';

CREATE UNIQUE INDEX uk_employees_store_id_emp_no ON employees (store_id, emp_no) WHERE deleted_at IS NULL;
CREATE INDEX idx_employees_store_id ON employees (store_id);
CREATE INDEX idx_employees_user_id ON employees (user_id);
CREATE INDEX idx_employees_status ON employees (status);
CREATE INDEX idx_employees_deleted_at ON employees (deleted_at);

CREATE TABLE IF NOT EXISTS shifts (
    id          BIGSERIAL PRIMARY KEY,
    store_id    BIGINT       NOT NULL,
    employee_id BIGINT       NOT NULL,
    shift_date  DATE         NOT NULL,
    start_time  TIME         NOT NULL,
    end_time    TIME         NOT NULL,
    shift_name  VARCHAR(64)  NOT NULL DEFAULT '',
    status      VARCHAR(32)  NOT NULL DEFAULT 'scheduled',
    remark      VARCHAR(255) NOT NULL DEFAULT '',
    created_by  BIGINT       NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ  NULL,
    CONSTRAINT fk_shifts_store_id FOREIGN KEY (store_id) REFERENCES stores (id),
    CONSTRAINT fk_shifts_employee_id FOREIGN KEY (employee_id) REFERENCES employees (id),
    CONSTRAINT ck_shifts_status CHECK (status IN ('scheduled', 'cancelled'))
);

COMMENT ON TABLE shifts IS '排班';
COMMENT ON COLUMN shifts.id IS '主键';
COMMENT ON COLUMN shifts.store_id IS '门店ID';
COMMENT ON COLUMN shifts.employee_id IS '员工ID';
COMMENT ON COLUMN shifts.shift_date IS '排班日期';
COMMENT ON COLUMN shifts.start_time IS '开始时间';
COMMENT ON COLUMN shifts.end_time IS '结束时间';
COMMENT ON COLUMN shifts.shift_name IS '班次名称';
COMMENT ON COLUMN shifts.status IS '状态：scheduled/cancelled';
COMMENT ON COLUMN shifts.remark IS '备注';
COMMENT ON COLUMN shifts.created_by IS '创建人';
COMMENT ON COLUMN shifts.created_at IS '创建时间';
COMMENT ON COLUMN shifts.updated_at IS '更新时间';
COMMENT ON COLUMN shifts.deleted_at IS '删除时间';

CREATE INDEX idx_shifts_store_id ON shifts (store_id);
CREATE INDEX idx_shifts_employee_id ON shifts (employee_id);
CREATE INDEX idx_shifts_shift_date ON shifts (shift_date);
CREATE INDEX idx_shifts_store_date ON shifts (store_id, shift_date);
CREATE INDEX idx_shifts_deleted_at ON shifts (deleted_at);

CREATE TABLE IF NOT EXISTS attendances (
    id            BIGSERIAL PRIMARY KEY,
    store_id      BIGINT       NOT NULL,
    employee_id   BIGINT       NOT NULL,
    work_date     DATE         NOT NULL,
    clock_in_at   TIMESTAMPTZ  NULL,
    clock_out_at  TIMESTAMPTZ  NULL,
    source        VARCHAR(32)  NOT NULL DEFAULT 'punch',
    status        VARCHAR(32)  NOT NULL DEFAULT 'normal',
    remark        VARCHAR(255) NOT NULL DEFAULT '',
    created_by    BIGINT       NOT NULL DEFAULT 0,
    updated_by    BIGINT       NOT NULL DEFAULT 0,
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at    TIMESTAMPTZ  NULL,
    CONSTRAINT fk_attendances_store_id FOREIGN KEY (store_id) REFERENCES stores (id),
    CONSTRAINT fk_attendances_employee_id FOREIGN KEY (employee_id) REFERENCES employees (id),
    CONSTRAINT ck_attendances_source CHECK (source IN ('punch', 'makeup')),
    CONSTRAINT ck_attendances_status CHECK (status IN ('normal', 'makeup'))
);

COMMENT ON TABLE attendances IS '考勤';
COMMENT ON COLUMN attendances.id IS '主键';
COMMENT ON COLUMN attendances.store_id IS '门店ID';
COMMENT ON COLUMN attendances.employee_id IS '员工ID';
COMMENT ON COLUMN attendances.work_date IS '出勤日';
COMMENT ON COLUMN attendances.clock_in_at IS '上班打卡时间';
COMMENT ON COLUMN attendances.clock_out_at IS '下班打卡时间';
COMMENT ON COLUMN attendances.source IS '来源：punch/makeup';
COMMENT ON COLUMN attendances.status IS '状态：normal/makeup';
COMMENT ON COLUMN attendances.remark IS '备注';
COMMENT ON COLUMN attendances.created_by IS '创建人';
COMMENT ON COLUMN attendances.updated_by IS '更新人';
COMMENT ON COLUMN attendances.created_at IS '创建时间';
COMMENT ON COLUMN attendances.updated_at IS '更新时间';
COMMENT ON COLUMN attendances.deleted_at IS '删除时间';

CREATE UNIQUE INDEX uk_attendances_store_emp_date
    ON attendances (store_id, employee_id, work_date) WHERE deleted_at IS NULL;
CREATE INDEX idx_attendances_store_id ON attendances (store_id);
CREATE INDEX idx_attendances_employee_id ON attendances (employee_id);
CREATE INDEX idx_attendances_work_date ON attendances (work_date);
CREATE INDEX idx_attendances_deleted_at ON attendances (deleted_at);
