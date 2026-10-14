-- 第 5 期：库存管理
-- 结存仅能通过 stock_docs 审核/冲销变更，禁止无单据直接改 stock_balances

CREATE TABLE IF NOT EXISTS materials (
    id         BIGSERIAL PRIMARY KEY,
    store_id   BIGINT       NOT NULL,
    code       VARCHAR(64)  NOT NULL,
    name       VARCHAR(128) NOT NULL,
    unit       VARCHAR(32)  NOT NULL DEFAULT '个',
    category   VARCHAR(64)  NOT NULL DEFAULT '',
    status     BOOLEAN      NOT NULL DEFAULT TRUE,
    remark     VARCHAR(512) NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ  NULL,
    CONSTRAINT fk_materials_store_id FOREIGN KEY (store_id) REFERENCES stores (id)
);

COMMENT ON TABLE materials IS '物料/原材料';
COMMENT ON COLUMN materials.id IS '主键';
COMMENT ON COLUMN materials.store_id IS '门店ID';
COMMENT ON COLUMN materials.code IS '物料编码（门店内唯一）';
COMMENT ON COLUMN materials.name IS '物料名称';
COMMENT ON COLUMN materials.unit IS '计量单位';
COMMENT ON COLUMN materials.category IS '分类';
COMMENT ON COLUMN materials.status IS '是否启用';
COMMENT ON COLUMN materials.remark IS '备注';
COMMENT ON COLUMN materials.created_at IS '创建时间';
COMMENT ON COLUMN materials.updated_at IS '更新时间';
COMMENT ON COLUMN materials.deleted_at IS '删除时间';

CREATE UNIQUE INDEX uk_materials_store_id_code ON materials (store_id, code) WHERE deleted_at IS NULL;
CREATE INDEX idx_materials_store_id ON materials (store_id);
CREATE INDEX idx_materials_name ON materials (name);
CREATE INDEX idx_materials_deleted_at ON materials (deleted_at);

CREATE TABLE IF NOT EXISTS stock_balances (
    id          BIGSERIAL PRIMARY KEY,
    store_id    BIGINT         NOT NULL,
    material_id BIGINT         NOT NULL,
    qty         NUMERIC(12, 3) NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ    NULL,
    CONSTRAINT fk_stock_balances_store_id FOREIGN KEY (store_id) REFERENCES stores (id),
    CONSTRAINT fk_stock_balances_material_id FOREIGN KEY (material_id) REFERENCES materials (id),
    CONSTRAINT ck_stock_balances_qty CHECK (qty >= 0)
);

COMMENT ON TABLE stock_balances IS '库存结存（仅单据审核/冲销可变更）';
COMMENT ON COLUMN stock_balances.id IS '主键';
COMMENT ON COLUMN stock_balances.store_id IS '门店ID';
COMMENT ON COLUMN stock_balances.material_id IS '物料ID';
COMMENT ON COLUMN stock_balances.qty IS '结存数量';
COMMENT ON COLUMN stock_balances.created_at IS '创建时间';
COMMENT ON COLUMN stock_balances.updated_at IS '更新时间';
COMMENT ON COLUMN stock_balances.deleted_at IS '删除时间';

CREATE UNIQUE INDEX uk_stock_balances_store_material ON stock_balances (store_id, material_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_stock_balances_store_id ON stock_balances (store_id);
CREATE INDEX idx_stock_balances_material_id ON stock_balances (material_id);
CREATE INDEX idx_stock_balances_deleted_at ON stock_balances (deleted_at);

-- doc_type: inbound入库 / outbound出库 / adjust盘点调整
-- status: draft草稿 → posted已审核生效 → voided已作废(冲销)
CREATE TABLE IF NOT EXISTS stock_docs (
    id           BIGSERIAL PRIMARY KEY,
    store_id     BIGINT       NOT NULL,
    doc_no       VARCHAR(64)  NOT NULL,
    doc_type     VARCHAR(32)  NOT NULL,
    status       VARCHAR(32)  NOT NULL DEFAULT 'draft',
    remark       VARCHAR(512) NOT NULL DEFAULT '',
    void_reason  VARCHAR(255) NOT NULL DEFAULT '',
    created_by   BIGINT       NOT NULL DEFAULT 0,
    posted_by    BIGINT       NOT NULL DEFAULT 0,
    voided_by    BIGINT       NOT NULL DEFAULT 0,
    posted_at    TIMESTAMPTZ  NULL,
    voided_at    TIMESTAMPTZ  NULL,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at   TIMESTAMPTZ  NULL,
    CONSTRAINT fk_stock_docs_store_id FOREIGN KEY (store_id) REFERENCES stores (id),
    CONSTRAINT ck_stock_docs_doc_type CHECK (doc_type IN ('inbound', 'outbound', 'adjust')),
    CONSTRAINT ck_stock_docs_status CHECK (status IN ('draft', 'posted', 'voided'))
);

COMMENT ON TABLE stock_docs IS '库存单据';
COMMENT ON COLUMN stock_docs.id IS '主键';
COMMENT ON COLUMN stock_docs.store_id IS '门店ID';
COMMENT ON COLUMN stock_docs.doc_no IS '单据号';
COMMENT ON COLUMN stock_docs.doc_type IS '类型：inbound/outbound/adjust';
COMMENT ON COLUMN stock_docs.status IS '状态：draft/posted/voided';
COMMENT ON COLUMN stock_docs.remark IS '备注';
COMMENT ON COLUMN stock_docs.void_reason IS '作废/冲销原因';
COMMENT ON COLUMN stock_docs.created_by IS '创建人';
COMMENT ON COLUMN stock_docs.posted_by IS '审核人';
COMMENT ON COLUMN stock_docs.voided_by IS '作废人';
COMMENT ON COLUMN stock_docs.posted_at IS '审核时间';
COMMENT ON COLUMN stock_docs.voided_at IS '作废时间';
COMMENT ON COLUMN stock_docs.created_at IS '创建时间';
COMMENT ON COLUMN stock_docs.updated_at IS '更新时间';
COMMENT ON COLUMN stock_docs.deleted_at IS '删除时间';

CREATE UNIQUE INDEX uk_stock_docs_store_id_doc_no ON stock_docs (store_id, doc_no) WHERE deleted_at IS NULL;
CREATE INDEX idx_stock_docs_store_id ON stock_docs (store_id);
CREATE INDEX idx_stock_docs_doc_type ON stock_docs (doc_type);
CREATE INDEX idx_stock_docs_status ON stock_docs (status);
CREATE INDEX idx_stock_docs_deleted_at ON stock_docs (deleted_at);

CREATE TABLE IF NOT EXISTS stock_doc_items (
    id          BIGSERIAL PRIMARY KEY,
    store_id    BIGINT         NOT NULL,
    doc_id      BIGINT         NOT NULL,
    material_id BIGINT         NOT NULL,
    qty         NUMERIC(12, 3) NOT NULL,
    unit_cost   NUMERIC(10, 2) NOT NULL DEFAULT 0,
    remark      VARCHAR(255)   NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ    NULL,
    CONSTRAINT fk_stock_doc_items_store_id FOREIGN KEY (store_id) REFERENCES stores (id),
    CONSTRAINT fk_stock_doc_items_doc_id FOREIGN KEY (doc_id) REFERENCES stock_docs (id),
    CONSTRAINT fk_stock_doc_items_material_id FOREIGN KEY (material_id) REFERENCES materials (id),
    CONSTRAINT ck_stock_doc_items_qty_nonzero CHECK (qty <> 0)
);

COMMENT ON TABLE stock_doc_items IS '库存单据明细';
COMMENT ON COLUMN stock_doc_items.id IS '主键';
COMMENT ON COLUMN stock_doc_items.store_id IS '门店ID';
COMMENT ON COLUMN stock_doc_items.doc_id IS '单据ID';
COMMENT ON COLUMN stock_doc_items.material_id IS '物料ID';
COMMENT ON COLUMN stock_doc_items.qty IS '数量：入库/出库为正；盘点调整可为正负（增减量）';
COMMENT ON COLUMN stock_doc_items.unit_cost IS '单位成本（可选）';
COMMENT ON COLUMN stock_doc_items.remark IS '备注';
COMMENT ON COLUMN stock_doc_items.created_at IS '创建时间';
COMMENT ON COLUMN stock_doc_items.updated_at IS '更新时间';
COMMENT ON COLUMN stock_doc_items.deleted_at IS '删除时间';

CREATE INDEX idx_stock_doc_items_store_id ON stock_doc_items (store_id);
CREATE INDEX idx_stock_doc_items_doc_id ON stock_doc_items (doc_id);
CREATE INDEX idx_stock_doc_items_material_id ON stock_doc_items (material_id);
CREATE INDEX idx_stock_doc_items_deleted_at ON stock_doc_items (deleted_at);
