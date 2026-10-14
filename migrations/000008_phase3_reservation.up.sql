-- 第 3 期：预订管理
-- 状态：pending_confirm → confirmed → seated；pending_confirm/confirmed → cancelled / no_show

CREATE TABLE IF NOT EXISTS reservations (
    id             BIGSERIAL PRIMARY KEY,
    store_id       BIGINT         NOT NULL,
    table_id       BIGINT         NULL,
    customer_name  VARCHAR(128)   NOT NULL,
    phone          VARCHAR(32)    NOT NULL DEFAULT '',
    guest_count    INT            NOT NULL DEFAULT 2,
    reserve_at     TIMESTAMPTZ    NOT NULL,
    channel        VARCHAR(32)    NOT NULL DEFAULT 'phone',
    status         VARCHAR(32)    NOT NULL DEFAULT 'pending_confirm',
    remark         VARCHAR(512)   NOT NULL DEFAULT '',
    deposit_amount NUMERIC(10, 2) NOT NULL DEFAULT 0,
    deposit_paid   BOOLEAN        NOT NULL DEFAULT FALSE,
    order_id       BIGINT         NULL,
    cancel_reason  VARCHAR(255)   NOT NULL DEFAULT '',
    confirmed_at   TIMESTAMPTZ    NULL,
    seated_at      TIMESTAMPTZ    NULL,
    cancelled_at   TIMESTAMPTZ    NULL,
    no_show_at     TIMESTAMPTZ    NULL,
    created_by     BIGINT         NOT NULL DEFAULT 0,
    updated_by     BIGINT         NOT NULL DEFAULT 0,
    created_at     TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    deleted_at     TIMESTAMPTZ    NULL,
    CONSTRAINT fk_reservations_store_id FOREIGN KEY (store_id) REFERENCES stores (id),
    CONSTRAINT fk_reservations_table_id FOREIGN KEY (table_id) REFERENCES tables (id),
    CONSTRAINT fk_reservations_order_id FOREIGN KEY (order_id) REFERENCES orders (id),
    CONSTRAINT ck_reservations_status CHECK (status IN ('pending_confirm', 'confirmed', 'seated', 'cancelled', 'no_show')),
    CONSTRAINT ck_reservations_channel CHECK (channel IN ('phone', 'walk_in', 'wechat', 'meituan', 'dianping', 'online', 'other')),
    CONSTRAINT ck_reservations_guest_count CHECK (guest_count > 0),
    CONSTRAINT ck_reservations_deposit_amount CHECK (deposit_amount >= 0)
);

COMMENT ON TABLE reservations IS '预订';
COMMENT ON COLUMN reservations.id IS '主键';
COMMENT ON COLUMN reservations.store_id IS '门店ID';
COMMENT ON COLUMN reservations.table_id IS '预订桌台ID';
COMMENT ON COLUMN reservations.customer_name IS '顾客姓名';
COMMENT ON COLUMN reservations.phone IS '联系电话';
COMMENT ON COLUMN reservations.guest_count IS '就餐人数';
COMMENT ON COLUMN reservations.reserve_at IS '预订到店时间';
COMMENT ON COLUMN reservations.channel IS '渠道：phone/walk_in/wechat/meituan/dianping/online/other';
COMMENT ON COLUMN reservations.status IS '状态：pending_confirm/confirmed/seated/cancelled/no_show';
COMMENT ON COLUMN reservations.remark IS '备注';
COMMENT ON COLUMN reservations.deposit_amount IS '订金金额（预留，本阶段不收款）';
COMMENT ON COLUMN reservations.deposit_paid IS '订金是否已付（预留）';
COMMENT ON COLUMN reservations.order_id IS '入座后关联订单ID';
COMMENT ON COLUMN reservations.cancel_reason IS '取消/违约原因';
COMMENT ON COLUMN reservations.confirmed_at IS '确认时间';
COMMENT ON COLUMN reservations.seated_at IS '入座时间';
COMMENT ON COLUMN reservations.cancelled_at IS '取消时间';
COMMENT ON COLUMN reservations.no_show_at IS '违约时间';
COMMENT ON COLUMN reservations.created_by IS '创建人';
COMMENT ON COLUMN reservations.updated_by IS '更新人';
COMMENT ON COLUMN reservations.created_at IS '创建时间';
COMMENT ON COLUMN reservations.updated_at IS '更新时间';
COMMENT ON COLUMN reservations.deleted_at IS '删除时间';

CREATE INDEX idx_reservations_store_id ON reservations (store_id);
CREATE INDEX idx_reservations_table_id ON reservations (table_id);
CREATE INDEX idx_reservations_status ON reservations (status);
CREATE INDEX idx_reservations_channel ON reservations (channel);
CREATE INDEX idx_reservations_reserve_at ON reservations (reserve_at);
CREATE INDEX idx_reservations_store_status ON reservations (store_id, status);
CREATE INDEX idx_reservations_phone ON reservations (phone);
CREATE INDEX idx_reservations_deleted_at ON reservations (deleted_at);
