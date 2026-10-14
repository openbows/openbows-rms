-- 第 5 期：采购管理（收货与库存入库同事务）

CREATE TABLE IF NOT EXISTS suppliers (
    id           BIGSERIAL PRIMARY KEY,
    store_id     BIGINT       NOT NULL,
    code         VARCHAR(64)  NOT NULL,
    name         VARCHAR(128) NOT NULL,
    contact_name VARCHAR(64)  NOT NULL DEFAULT '',
    phone        VARCHAR(32)  NOT NULL DEFAULT '',
    address      VARCHAR(255) NOT NULL DEFAULT '',
    status       BOOLEAN      NOT NULL DEFAULT TRUE,
    remark       VARCHAR(512) NOT NULL DEFAULT '',
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at   TIMESTAMPTZ  NULL,
    CONSTRAINT fk_suppliers_store_id FOREIGN KEY (store_id) REFERENCES stores (id)
);

COMMENT ON TABLE suppliers IS '供应商';
COMMENT ON COLUMN suppliers.id IS '主键';
COMMENT ON COLUMN suppliers.store_id IS '门店ID';
COMMENT ON COLUMN suppliers.code IS '供应商编码（门店内唯一）';
COMMENT ON COLUMN suppliers.name IS '供应商名称';
COMMENT ON COLUMN suppliers.contact_name IS '联系人';
COMMENT ON COLUMN suppliers.phone IS '电话';
COMMENT ON COLUMN suppliers.address IS '地址';
COMMENT ON COLUMN suppliers.status IS '是否启用';
COMMENT ON COLUMN suppliers.remark IS '备注';
COMMENT ON COLUMN suppliers.created_at IS '创建时间';
COMMENT ON COLUMN suppliers.updated_at IS '更新时间';
COMMENT ON COLUMN suppliers.deleted_at IS '删除时间';

CREATE UNIQUE INDEX uk_suppliers_store_id_code ON suppliers (store_id, code) WHERE deleted_at IS NULL;
CREATE INDEX idx_suppliers_store_id ON suppliers (store_id);
CREATE INDEX idx_suppliers_name ON suppliers (name);
CREATE INDEX idx_suppliers_deleted_at ON suppliers (deleted_at);

-- status: draft → submitted → partial_received → received；draft/submitted → cancelled
CREATE TABLE IF NOT EXISTS purchase_orders (
    id              BIGSERIAL PRIMARY KEY,
    store_id        BIGINT         NOT NULL,
    supplier_id     BIGINT         NOT NULL,
    order_no        VARCHAR(64)    NOT NULL,
    status          VARCHAR(32)    NOT NULL DEFAULT 'draft',
    total_amount    NUMERIC(10, 2) NOT NULL DEFAULT 0,
    remark          VARCHAR(512)   NOT NULL DEFAULT '',
    created_by      BIGINT         NOT NULL DEFAULT 0,
    submitted_by    BIGINT         NOT NULL DEFAULT 0,
    cancelled_by    BIGINT         NOT NULL DEFAULT 0,
    submitted_at    TIMESTAMPTZ    NULL,
    cancelled_at    TIMESTAMPTZ    NULL,
    created_at      TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ    NULL,
    CONSTRAINT fk_purchase_orders_store_id FOREIGN KEY (store_id) REFERENCES stores (id),
    CONSTRAINT fk_purchase_orders_supplier_id FOREIGN KEY (supplier_id) REFERENCES suppliers (id),
    CONSTRAINT ck_purchase_orders_status CHECK (status IN ('draft', 'submitted', 'partial_received', 'received', 'cancelled'))
);

COMMENT ON TABLE purchase_orders IS '采购订单';
COMMENT ON COLUMN purchase_orders.id IS '主键';
COMMENT ON COLUMN purchase_orders.store_id IS '门店ID';
COMMENT ON COLUMN purchase_orders.supplier_id IS '供应商ID';
COMMENT ON COLUMN purchase_orders.order_no IS '采购单号';
COMMENT ON COLUMN purchase_orders.status IS '状态：draft/submitted/partial_received/received/cancelled';
COMMENT ON COLUMN purchase_orders.total_amount IS '订单金额合计';
COMMENT ON COLUMN purchase_orders.remark IS '备注';
COMMENT ON COLUMN purchase_orders.created_by IS '创建人';
COMMENT ON COLUMN purchase_orders.submitted_by IS '提交人';
COMMENT ON COLUMN purchase_orders.cancelled_by IS '取消人';
COMMENT ON COLUMN purchase_orders.submitted_at IS '提交时间';
COMMENT ON COLUMN purchase_orders.cancelled_at IS '取消时间';
COMMENT ON COLUMN purchase_orders.created_at IS '创建时间';
COMMENT ON COLUMN purchase_orders.updated_at IS '更新时间';
COMMENT ON COLUMN purchase_orders.deleted_at IS '删除时间';

CREATE UNIQUE INDEX uk_purchase_orders_store_order_no ON purchase_orders (store_id, order_no) WHERE deleted_at IS NULL;
CREATE INDEX idx_purchase_orders_store_id ON purchase_orders (store_id);
CREATE INDEX idx_purchase_orders_supplier_id ON purchase_orders (supplier_id);
CREATE INDEX idx_purchase_orders_status ON purchase_orders (status);
CREATE INDEX idx_purchase_orders_deleted_at ON purchase_orders (deleted_at);

CREATE TABLE IF NOT EXISTS purchase_order_items (
    id           BIGSERIAL PRIMARY KEY,
    store_id     BIGINT         NOT NULL,
    order_id     BIGINT         NOT NULL,
    material_id  BIGINT         NOT NULL,
    qty          NUMERIC(12, 3) NOT NULL,
    unit_price   NUMERIC(10, 2) NOT NULL DEFAULT 0,
    amount       NUMERIC(10, 2) NOT NULL DEFAULT 0,
    received_qty NUMERIC(12, 3) NOT NULL DEFAULT 0,
    remark       VARCHAR(255)   NOT NULL DEFAULT '',
    created_at   TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    deleted_at   TIMESTAMPTZ    NULL,
    CONSTRAINT fk_purchase_order_items_store_id FOREIGN KEY (store_id) REFERENCES stores (id),
    CONSTRAINT fk_purchase_order_items_order_id FOREIGN KEY (order_id) REFERENCES purchase_orders (id),
    CONSTRAINT fk_purchase_order_items_material_id FOREIGN KEY (material_id) REFERENCES materials (id),
    CONSTRAINT ck_purchase_order_items_qty CHECK (qty > 0),
    CONSTRAINT ck_purchase_order_items_received_qty CHECK (received_qty >= 0)
);

COMMENT ON TABLE purchase_order_items IS '采购订单明细';
COMMENT ON COLUMN purchase_order_items.id IS '主键';
COMMENT ON COLUMN purchase_order_items.store_id IS '门店ID';
COMMENT ON COLUMN purchase_order_items.order_id IS '采购订单ID';
COMMENT ON COLUMN purchase_order_items.material_id IS '物料ID';
COMMENT ON COLUMN purchase_order_items.qty IS '采购数量';
COMMENT ON COLUMN purchase_order_items.unit_price IS '采购单价';
COMMENT ON COLUMN purchase_order_items.amount IS '行金额';
COMMENT ON COLUMN purchase_order_items.received_qty IS '已收货数量';
COMMENT ON COLUMN purchase_order_items.remark IS '备注';
COMMENT ON COLUMN purchase_order_items.created_at IS '创建时间';
COMMENT ON COLUMN purchase_order_items.updated_at IS '更新时间';
COMMENT ON COLUMN purchase_order_items.deleted_at IS '删除时间';

CREATE INDEX idx_purchase_order_items_store_id ON purchase_order_items (store_id);
CREATE INDEX idx_purchase_order_items_order_id ON purchase_order_items (order_id);
CREATE INDEX idx_purchase_order_items_material_id ON purchase_order_items (material_id);
CREATE INDEX idx_purchase_order_items_deleted_at ON purchase_order_items (deleted_at);

-- status: draft → posted
CREATE TABLE IF NOT EXISTS purchase_receipts (
    id                BIGSERIAL PRIMARY KEY,
    store_id          BIGINT       NOT NULL,
    purchase_order_id BIGINT       NOT NULL,
    supplier_id       BIGINT       NOT NULL,
    receipt_no        VARCHAR(64)  NOT NULL,
    status            VARCHAR(32)  NOT NULL DEFAULT 'draft',
    stock_doc_id      BIGINT       NULL,
    remark            VARCHAR(512) NOT NULL DEFAULT '',
    created_by        BIGINT       NOT NULL DEFAULT 0,
    posted_by         BIGINT       NOT NULL DEFAULT 0,
    posted_at         TIMESTAMPTZ  NULL,
    created_at        TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at        TIMESTAMPTZ  NULL,
    CONSTRAINT fk_purchase_receipts_store_id FOREIGN KEY (store_id) REFERENCES stores (id),
    CONSTRAINT fk_purchase_receipts_po_id FOREIGN KEY (purchase_order_id) REFERENCES purchase_orders (id),
    CONSTRAINT fk_purchase_receipts_supplier_id FOREIGN KEY (supplier_id) REFERENCES suppliers (id),
    CONSTRAINT fk_purchase_receipts_stock_doc_id FOREIGN KEY (stock_doc_id) REFERENCES stock_docs (id),
    CONSTRAINT ck_purchase_receipts_status CHECK (status IN ('draft', 'posted'))
);

COMMENT ON TABLE purchase_receipts IS '采购收货单';
COMMENT ON COLUMN purchase_receipts.id IS '主键';
COMMENT ON COLUMN purchase_receipts.store_id IS '门店ID';
COMMENT ON COLUMN purchase_receipts.purchase_order_id IS '采购订单ID';
COMMENT ON COLUMN purchase_receipts.supplier_id IS '供应商ID';
COMMENT ON COLUMN purchase_receipts.receipt_no IS '收货单号';
COMMENT ON COLUMN purchase_receipts.status IS '状态：draft/posted';
COMMENT ON COLUMN purchase_receipts.stock_doc_id IS '关联库存入库单ID';
COMMENT ON COLUMN purchase_receipts.remark IS '备注';
COMMENT ON COLUMN purchase_receipts.created_by IS '创建人';
COMMENT ON COLUMN purchase_receipts.posted_by IS '入库确认人';
COMMENT ON COLUMN purchase_receipts.posted_at IS '入库确认时间';
COMMENT ON COLUMN purchase_receipts.created_at IS '创建时间';
COMMENT ON COLUMN purchase_receipts.updated_at IS '更新时间';
COMMENT ON COLUMN purchase_receipts.deleted_at IS '删除时间';

CREATE UNIQUE INDEX uk_purchase_receipts_store_receipt_no ON purchase_receipts (store_id, receipt_no) WHERE deleted_at IS NULL;
CREATE INDEX idx_purchase_receipts_store_id ON purchase_receipts (store_id);
CREATE INDEX idx_purchase_receipts_po_id ON purchase_receipts (purchase_order_id);
CREATE INDEX idx_purchase_receipts_status ON purchase_receipts (status);
CREATE INDEX idx_purchase_receipts_stock_doc_id ON purchase_receipts (stock_doc_id);
CREATE INDEX idx_purchase_receipts_deleted_at ON purchase_receipts (deleted_at);

CREATE TABLE IF NOT EXISTS purchase_receipt_items (
    id                      BIGSERIAL PRIMARY KEY,
    store_id                BIGINT         NOT NULL,
    receipt_id              BIGINT         NOT NULL,
    purchase_order_item_id  BIGINT         NOT NULL,
    material_id             BIGINT         NOT NULL,
    qty                     NUMERIC(12, 3) NOT NULL,
    unit_price              NUMERIC(10, 2) NOT NULL DEFAULT 0,
    remark                  VARCHAR(255)   NOT NULL DEFAULT '',
    created_at              TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    deleted_at              TIMESTAMPTZ    NULL,
    CONSTRAINT fk_purchase_receipt_items_store_id FOREIGN KEY (store_id) REFERENCES stores (id),
    CONSTRAINT fk_purchase_receipt_items_receipt_id FOREIGN KEY (receipt_id) REFERENCES purchase_receipts (id),
    CONSTRAINT fk_purchase_receipt_items_poi_id FOREIGN KEY (purchase_order_item_id) REFERENCES purchase_order_items (id),
    CONSTRAINT fk_purchase_receipt_items_material_id FOREIGN KEY (material_id) REFERENCES materials (id),
    CONSTRAINT ck_purchase_receipt_items_qty CHECK (qty > 0)
);

COMMENT ON TABLE purchase_receipt_items IS '采购收货明细';
COMMENT ON COLUMN purchase_receipt_items.id IS '主键';
COMMENT ON COLUMN purchase_receipt_items.store_id IS '门店ID';
COMMENT ON COLUMN purchase_receipt_items.receipt_id IS '收货单ID';
COMMENT ON COLUMN purchase_receipt_items.purchase_order_item_id IS '采购订单明细ID';
COMMENT ON COLUMN purchase_receipt_items.material_id IS '物料ID';
COMMENT ON COLUMN purchase_receipt_items.qty IS '本次收货数量';
COMMENT ON COLUMN purchase_receipt_items.unit_price IS '单价快照';
COMMENT ON COLUMN purchase_receipt_items.remark IS '备注';
COMMENT ON COLUMN purchase_receipt_items.created_at IS '创建时间';
COMMENT ON COLUMN purchase_receipt_items.updated_at IS '更新时间';
COMMENT ON COLUMN purchase_receipt_items.deleted_at IS '删除时间';

CREATE INDEX idx_purchase_receipt_items_store_id ON purchase_receipt_items (store_id);
CREATE INDEX idx_purchase_receipt_items_receipt_id ON purchase_receipt_items (receipt_id);
CREATE INDEX idx_purchase_receipt_items_poi_id ON purchase_receipt_items (purchase_order_item_id);
CREATE INDEX idx_purchase_receipt_items_material_id ON purchase_receipt_items (material_id);
CREATE INDEX idx_purchase_receipt_items_deleted_at ON purchase_receipt_items (deleted_at);
