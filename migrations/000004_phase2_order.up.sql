-- 第 2 期：点餐收银 / 订单

-- status: pending_confirm → cooking → pending_pay → completed | cancelled
CREATE TABLE IF NOT EXISTS orders (
    id              BIGSERIAL PRIMARY KEY,
    store_id        BIGINT         NOT NULL,
    table_id        BIGINT         NOT NULL,
    order_no        VARCHAR(64)    NOT NULL,
    status          VARCHAR(32)    NOT NULL DEFAULT 'pending_confirm',
    guest_count     INT            NOT NULL DEFAULT 1,
    subtotal_amount NUMERIC(10, 2) NOT NULL DEFAULT 0,
    fee_amount      NUMERIC(10, 2) NOT NULL DEFAULT 0,
    discount_amount NUMERIC(10, 2) NOT NULL DEFAULT 0,
    payable_amount  NUMERIC(10, 2) NOT NULL DEFAULT 0,
    remark          VARCHAR(512)   NOT NULL DEFAULT '',
    created_by      BIGINT         NOT NULL DEFAULT 0,
    updated_by      BIGINT         NOT NULL DEFAULT 0,
    opened_at       TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    closed_at       TIMESTAMPTZ    NULL,
    created_at      TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ    NULL,
    CONSTRAINT fk_orders_store_id FOREIGN KEY (store_id) REFERENCES stores (id),
    CONSTRAINT fk_orders_table_id FOREIGN KEY (table_id) REFERENCES tables (id),
    CONSTRAINT ck_orders_status CHECK (status IN ('pending_confirm', 'cooking', 'pending_pay', 'completed', 'cancelled'))
);
COMMENT ON TABLE orders IS '订单';
COMMENT ON COLUMN orders.id IS '主键';
COMMENT ON COLUMN orders.store_id IS '门店ID';
COMMENT ON COLUMN orders.table_id IS '桌台ID';
COMMENT ON COLUMN orders.order_no IS '订单号';
COMMENT ON COLUMN orders.status IS '状态：pending_confirm/cooking/pending_pay/completed/cancelled';
COMMENT ON COLUMN orders.guest_count IS '就餐人数';
COMMENT ON COLUMN orders.subtotal_amount IS '菜品小计';
COMMENT ON COLUMN orders.fee_amount IS '费用合计（服务费/税等）';
COMMENT ON COLUMN orders.discount_amount IS '优惠合计';
COMMENT ON COLUMN orders.payable_amount IS '应付金额';
COMMENT ON COLUMN orders.remark IS '备注';
COMMENT ON COLUMN orders.created_by IS '创建人';
COMMENT ON COLUMN orders.updated_by IS '更新人';
COMMENT ON COLUMN orders.opened_at IS '开台时间';
COMMENT ON COLUMN orders.closed_at IS '关台/完结时间';
COMMENT ON COLUMN orders.created_at IS '创建时间';
COMMENT ON COLUMN orders.updated_at IS '更新时间';
COMMENT ON COLUMN orders.deleted_at IS '删除时间';
CREATE UNIQUE INDEX uk_orders_store_id_order_no ON orders (store_id, order_no) WHERE deleted_at IS NULL;
CREATE INDEX idx_orders_store_id ON orders (store_id);
CREATE INDEX idx_orders_table_id ON orders (table_id);
CREATE INDEX idx_orders_status ON orders (status);
CREATE INDEX idx_orders_deleted_at ON orders (deleted_at);

-- item status: normal / returned
CREATE TABLE IF NOT EXISTS order_items (
    id            BIGSERIAL PRIMARY KEY,
    store_id      BIGINT         NOT NULL,
    order_id      BIGINT         NOT NULL,
    dish_id       BIGINT         NOT NULL,
    dish_name     VARCHAR(128)   NOT NULL,
    dish_spec_id  BIGINT         NULL,
    spec_name     VARCHAR(128)   NOT NULL DEFAULT '',
    unit_price    NUMERIC(10, 2) NOT NULL DEFAULT 0,
    qty           INT            NOT NULL DEFAULT 1,
    amount        NUMERIC(10, 2) NOT NULL DEFAULT 0,
    status        VARCHAR(32)    NOT NULL DEFAULT 'normal',
    remark        VARCHAR(255)   NOT NULL DEFAULT '',
    sort_no       INT            NOT NULL DEFAULT 0,
    created_at    TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    deleted_at    TIMESTAMPTZ    NULL,
    CONSTRAINT fk_order_items_store_id FOREIGN KEY (store_id) REFERENCES stores (id),
    CONSTRAINT fk_order_items_order_id FOREIGN KEY (order_id) REFERENCES orders (id),
    CONSTRAINT fk_order_items_dish_id FOREIGN KEY (dish_id) REFERENCES dishes (id),
    CONSTRAINT ck_order_items_status CHECK (status IN ('normal', 'returned')),
    CONSTRAINT ck_order_items_qty CHECK (qty > 0)
);
COMMENT ON TABLE order_items IS '订单菜品明细';
COMMENT ON COLUMN order_items.id IS '主键';
COMMENT ON COLUMN order_items.store_id IS '门店ID';
COMMENT ON COLUMN order_items.order_id IS '订单ID';
COMMENT ON COLUMN order_items.dish_id IS '菜品ID';
COMMENT ON COLUMN order_items.dish_name IS '菜品名称快照';
COMMENT ON COLUMN order_items.dish_spec_id IS '规格ID';
COMMENT ON COLUMN order_items.spec_name IS '规格名称快照';
COMMENT ON COLUMN order_items.unit_price IS '单价（含加料前）';
COMMENT ON COLUMN order_items.qty IS '数量';
COMMENT ON COLUMN order_items.amount IS '行金额';
COMMENT ON COLUMN order_items.status IS '状态：normal/returned';
COMMENT ON COLUMN order_items.remark IS '备注';
COMMENT ON COLUMN order_items.sort_no IS '排序号';
COMMENT ON COLUMN order_items.created_at IS '创建时间';
COMMENT ON COLUMN order_items.updated_at IS '更新时间';
COMMENT ON COLUMN order_items.deleted_at IS '删除时间';
CREATE INDEX idx_order_items_store_id ON order_items (store_id);
CREATE INDEX idx_order_items_order_id ON order_items (order_id);
CREATE INDEX idx_order_items_deleted_at ON order_items (deleted_at);

CREATE TABLE IF NOT EXISTS order_item_modifiers (
    id            BIGSERIAL PRIMARY KEY,
    store_id      BIGINT         NOT NULL,
    order_item_id BIGINT         NOT NULL,
    option_id     BIGINT         NULL,
    option_name   VARCHAR(128)   NOT NULL,
    price         NUMERIC(10, 2) NOT NULL DEFAULT 0,
    qty           INT            NOT NULL DEFAULT 1,
    created_at    TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    deleted_at    TIMESTAMPTZ    NULL,
    CONSTRAINT fk_order_item_modifiers_store_id FOREIGN KEY (store_id) REFERENCES stores (id),
    CONSTRAINT fk_order_item_modifiers_order_item_id FOREIGN KEY (order_item_id) REFERENCES order_items (id)
);
COMMENT ON TABLE order_item_modifiers IS '订单明细加料';
COMMENT ON COLUMN order_item_modifiers.id IS '主键';
COMMENT ON COLUMN order_item_modifiers.store_id IS '门店ID';
COMMENT ON COLUMN order_item_modifiers.order_item_id IS '订单明细ID';
COMMENT ON COLUMN order_item_modifiers.option_id IS '加料选项ID';
COMMENT ON COLUMN order_item_modifiers.option_name IS '加料名称快照';
COMMENT ON COLUMN order_item_modifiers.price IS '加料单价';
COMMENT ON COLUMN order_item_modifiers.qty IS '数量';
COMMENT ON COLUMN order_item_modifiers.created_at IS '创建时间';
COMMENT ON COLUMN order_item_modifiers.updated_at IS '更新时间';
COMMENT ON COLUMN order_item_modifiers.deleted_at IS '删除时间';
CREATE INDEX idx_order_item_modifiers_store_id ON order_item_modifiers (store_id);
CREATE INDEX idx_order_item_modifiers_order_item_id ON order_item_modifiers (order_item_id);
CREATE INDEX idx_order_item_modifiers_deleted_at ON order_item_modifiers (deleted_at);

CREATE TABLE IF NOT EXISTS order_status_logs (
    id            BIGSERIAL PRIMARY KEY,
    store_id      BIGINT       NOT NULL,
    order_id      BIGINT       NOT NULL,
    from_status   VARCHAR(32)  NOT NULL DEFAULT '',
    to_status     VARCHAR(32)  NOT NULL,
    operator_id   BIGINT       NOT NULL DEFAULT 0,
    operator_name VARCHAR(64)  NOT NULL DEFAULT '',
    remark        VARCHAR(255) NOT NULL DEFAULT '',
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at    TIMESTAMPTZ  NULL,
    CONSTRAINT fk_order_status_logs_store_id FOREIGN KEY (store_id) REFERENCES stores (id),
    CONSTRAINT fk_order_status_logs_order_id FOREIGN KEY (order_id) REFERENCES orders (id)
);
COMMENT ON TABLE order_status_logs IS '订单状态流水';
COMMENT ON COLUMN order_status_logs.id IS '主键';
COMMENT ON COLUMN order_status_logs.store_id IS '门店ID';
COMMENT ON COLUMN order_status_logs.order_id IS '订单ID';
COMMENT ON COLUMN order_status_logs.from_status IS '原状态';
COMMENT ON COLUMN order_status_logs.to_status IS '新状态';
COMMENT ON COLUMN order_status_logs.operator_id IS '操作人ID';
COMMENT ON COLUMN order_status_logs.operator_name IS '操作人名称';
COMMENT ON COLUMN order_status_logs.remark IS '备注';
COMMENT ON COLUMN order_status_logs.created_at IS '创建时间';
COMMENT ON COLUMN order_status_logs.updated_at IS '更新时间';
COMMENT ON COLUMN order_status_logs.deleted_at IS '删除时间';
CREATE INDEX idx_order_status_logs_store_id ON order_status_logs (store_id);
CREATE INDEX idx_order_status_logs_order_id ON order_status_logs (order_id);
CREATE INDEX idx_order_status_logs_deleted_at ON order_status_logs (deleted_at);

-- fee_type: discount / service / tax / other
CREATE TABLE IF NOT EXISTS order_fee_items (
    id         BIGSERIAL PRIMARY KEY,
    store_id   BIGINT         NOT NULL,
    order_id   BIGINT         NOT NULL,
    fee_type   VARCHAR(32)    NOT NULL,
    name       VARCHAR(128)   NOT NULL,
    amount     NUMERIC(10, 2) NOT NULL DEFAULT 0,
    remark     VARCHAR(255)   NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ    NULL,
    CONSTRAINT fk_order_fee_items_store_id FOREIGN KEY (store_id) REFERENCES stores (id),
    CONSTRAINT fk_order_fee_items_order_id FOREIGN KEY (order_id) REFERENCES orders (id),
    CONSTRAINT ck_order_fee_items_fee_type CHECK (fee_type IN ('discount', 'service', 'tax', 'other'))
);
COMMENT ON TABLE order_fee_items IS '订单费用分项';
COMMENT ON COLUMN order_fee_items.id IS '主键';
COMMENT ON COLUMN order_fee_items.store_id IS '门店ID';
COMMENT ON COLUMN order_fee_items.order_id IS '订单ID';
COMMENT ON COLUMN order_fee_items.fee_type IS '类型：discount/service/tax/other';
COMMENT ON COLUMN order_fee_items.name IS '分项名称';
COMMENT ON COLUMN order_fee_items.amount IS '金额（优惠为负或正由业务约定，本系统优惠存正数）';
COMMENT ON COLUMN order_fee_items.remark IS '备注';
COMMENT ON COLUMN order_fee_items.created_at IS '创建时间';
COMMENT ON COLUMN order_fee_items.updated_at IS '更新时间';
COMMENT ON COLUMN order_fee_items.deleted_at IS '删除时间';
CREATE INDEX idx_order_fee_items_store_id ON order_fee_items (store_id);
CREATE INDEX idx_order_fee_items_order_id ON order_fee_items (order_id);
CREATE INDEX idx_order_fee_items_deleted_at ON order_fee_items (deleted_at);
