-- 第 1 期：桌台主数据（区域 / 桌台）

CREATE TABLE IF NOT EXISTS areas (
    id         BIGSERIAL PRIMARY KEY,
    store_id   BIGINT       NOT NULL,
    name       VARCHAR(128) NOT NULL,
    sort_no    INT          NOT NULL DEFAULT 0,
    status     BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ  NULL,
    CONSTRAINT fk_areas_store_id FOREIGN KEY (store_id) REFERENCES stores (id)
);
COMMENT ON TABLE areas IS '桌台区域';
COMMENT ON COLUMN areas.id IS '主键';
COMMENT ON COLUMN areas.store_id IS '门店ID';
COMMENT ON COLUMN areas.name IS '区域名称';
COMMENT ON COLUMN areas.sort_no IS '排序号';
COMMENT ON COLUMN areas.status IS '是否启用';
COMMENT ON COLUMN areas.created_at IS '创建时间';
COMMENT ON COLUMN areas.updated_at IS '更新时间';
COMMENT ON COLUMN areas.deleted_at IS '删除时间';
CREATE UNIQUE INDEX uk_areas_store_id_name ON areas (store_id, name) WHERE deleted_at IS NULL;
CREATE INDEX idx_areas_store_id ON areas (store_id);
CREATE INDEX idx_areas_deleted_at ON areas (deleted_at);

-- status: idle=空闲, occupied=占用, reserved=预留
CREATE TABLE IF NOT EXISTS tables (
    id               BIGSERIAL PRIMARY KEY,
    store_id         BIGINT       NOT NULL,
    area_id          BIGINT       NOT NULL,
    code             VARCHAR(64)  NOT NULL,
    name             VARCHAR(128) NOT NULL DEFAULT '',
    capacity         INT          NOT NULL DEFAULT 4,
    status           VARCHAR(32)  NOT NULL DEFAULT 'idle',
    sort_no          INT          NOT NULL DEFAULT 0,
    current_order_id BIGINT       NULL,
    created_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at       TIMESTAMPTZ  NULL,
    CONSTRAINT fk_tables_store_id FOREIGN KEY (store_id) REFERENCES stores (id),
    CONSTRAINT fk_tables_area_id FOREIGN KEY (area_id) REFERENCES areas (id),
    CONSTRAINT ck_tables_status CHECK (status IN ('idle', 'occupied', 'reserved'))
);
COMMENT ON TABLE tables IS '桌台';
COMMENT ON COLUMN tables.id IS '主键';
COMMENT ON COLUMN tables.store_id IS '门店ID';
COMMENT ON COLUMN tables.area_id IS '区域ID';
COMMENT ON COLUMN tables.code IS '桌台编号';
COMMENT ON COLUMN tables.name IS '桌台名称';
COMMENT ON COLUMN tables.capacity IS '可容纳人数';
COMMENT ON COLUMN tables.status IS '状态：idle空闲/occupied占用/reserved预留';
COMMENT ON COLUMN tables.sort_no IS '排序号';
COMMENT ON COLUMN tables.current_order_id IS '当前挂单订单ID（预留，订单期写入）';
COMMENT ON COLUMN tables.created_at IS '创建时间';
COMMENT ON COLUMN tables.updated_at IS '更新时间';
COMMENT ON COLUMN tables.deleted_at IS '删除时间';
CREATE UNIQUE INDEX uk_tables_store_id_code ON tables (store_id, code) WHERE deleted_at IS NULL;
CREATE INDEX idx_tables_store_id ON tables (store_id);
CREATE INDEX idx_tables_area_id ON tables (area_id);
CREATE INDEX idx_tables_status ON tables (status);
CREATE INDEX idx_tables_deleted_at ON tables (deleted_at);
