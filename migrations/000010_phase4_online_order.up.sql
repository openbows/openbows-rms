-- 第 4 期：线上点餐（复用 orders，增加渠道字段 + online_order_extras）

ALTER TABLE orders
    ADD COLUMN IF NOT EXISTS channel VARCHAR(32) NOT NULL DEFAULT 'dine_in';
ALTER TABLE orders
    ADD COLUMN IF NOT EXISTS member_id BIGINT NULL;

COMMENT ON COLUMN orders.channel IS '下单渠道：dine_in/online_h5/online_scan/takeout';
COMMENT ON COLUMN orders.member_id IS '关联会员ID（可选）';

ALTER TABLE orders DROP CONSTRAINT IF EXISTS ck_orders_channel;
ALTER TABLE orders
    ADD CONSTRAINT ck_orders_channel CHECK (channel IN ('dine_in', 'online_h5', 'online_scan', 'takeout'));

ALTER TABLE orders DROP CONSTRAINT IF EXISTS fk_orders_member_id;
ALTER TABLE orders
    ADD CONSTRAINT fk_orders_member_id FOREIGN KEY (member_id) REFERENCES members (id);

CREATE INDEX IF NOT EXISTS idx_orders_channel ON orders (channel);
CREATE INDEX IF NOT EXISTS idx_orders_member_id ON orders (member_id);
CREATE INDEX IF NOT EXISTS idx_orders_store_channel ON orders (store_id, channel);

CREATE TABLE IF NOT EXISTS online_order_extras (
    id             BIGSERIAL PRIMARY KEY,
    store_id       BIGINT       NOT NULL,
    order_id       BIGINT       NOT NULL,
    contact_name   VARCHAR(128) NOT NULL DEFAULT '',
    contact_phone  VARCHAR(32)  NOT NULL DEFAULT '',
    access_token   VARCHAR(64)  NOT NULL,
    source         VARCHAR(32)  NOT NULL DEFAULT 'h5',
    created_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at     TIMESTAMPTZ  NULL,
    CONSTRAINT fk_online_order_extras_store_id FOREIGN KEY (store_id) REFERENCES stores (id),
    CONSTRAINT fk_online_order_extras_order_id FOREIGN KEY (order_id) REFERENCES orders (id),
    CONSTRAINT ck_online_order_extras_source CHECK (source IN ('h5', 'scan'))
);

COMMENT ON TABLE online_order_extras IS '线上订单扩展（顾客端凭证与联系人）';
COMMENT ON COLUMN online_order_extras.id IS '主键';
COMMENT ON COLUMN online_order_extras.store_id IS '门店ID';
COMMENT ON COLUMN online_order_extras.order_id IS '订单ID';
COMMENT ON COLUMN online_order_extras.contact_name IS '联系人姓名';
COMMENT ON COLUMN online_order_extras.contact_phone IS '联系人电话';
COMMENT ON COLUMN online_order_extras.access_token IS '顾客端查询凭证';
COMMENT ON COLUMN online_order_extras.source IS '来源：h5/scan';
COMMENT ON COLUMN online_order_extras.created_at IS '创建时间';
COMMENT ON COLUMN online_order_extras.updated_at IS '更新时间';
COMMENT ON COLUMN online_order_extras.deleted_at IS '删除时间';

CREATE UNIQUE INDEX uk_online_order_extras_order_id ON online_order_extras (order_id) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX uk_online_order_extras_access_token ON online_order_extras (access_token) WHERE deleted_at IS NULL;
CREATE INDEX idx_online_order_extras_store_id ON online_order_extras (store_id);
CREATE INDEX idx_online_order_extras_deleted_at ON online_order_extras (deleted_at);
