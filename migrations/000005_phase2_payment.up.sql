-- 第 2 期：支付（订单侧）

-- channel: cash / wechat / alipay / other
-- status: success / failed
CREATE TABLE IF NOT EXISTS payments (
    id            BIGSERIAL PRIMARY KEY,
    store_id      BIGINT         NOT NULL,
    order_id      BIGINT         NOT NULL,
    payment_no    VARCHAR(64)    NOT NULL,
    channel       VARCHAR(32)    NOT NULL,
    amount        NUMERIC(10, 2) NOT NULL,
    status        VARCHAR(32)    NOT NULL DEFAULT 'success',
    paid_at       TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    out_trade_no  VARCHAR(128)   NOT NULL DEFAULT '',
    callback_raw  JSONB          NULL,
    operator_id   BIGINT         NOT NULL DEFAULT 0,
    remark        VARCHAR(255)   NOT NULL DEFAULT '',
    created_at    TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    deleted_at    TIMESTAMPTZ    NULL,
    CONSTRAINT fk_payments_store_id FOREIGN KEY (store_id) REFERENCES stores (id),
    CONSTRAINT fk_payments_order_id FOREIGN KEY (order_id) REFERENCES orders (id),
    CONSTRAINT ck_payments_channel CHECK (channel IN ('cash', 'wechat', 'alipay', 'other')),
    CONSTRAINT ck_payments_status CHECK (status IN ('success', 'failed'))
);
COMMENT ON TABLE payments IS '订单支付记录';
COMMENT ON COLUMN payments.id IS '主键';
COMMENT ON COLUMN payments.store_id IS '门店ID';
COMMENT ON COLUMN payments.order_id IS '订单ID';
COMMENT ON COLUMN payments.payment_no IS '支付单号';
COMMENT ON COLUMN payments.channel IS '支付渠道：cash/wechat/alipay/other';
COMMENT ON COLUMN payments.amount IS '支付金额';
COMMENT ON COLUMN payments.status IS '状态：success/failed';
COMMENT ON COLUMN payments.paid_at IS '支付时间';
COMMENT ON COLUMN payments.out_trade_no IS '第三方交易号（预留）';
COMMENT ON COLUMN payments.callback_raw IS '第三方回调原文（预留）';
COMMENT ON COLUMN payments.operator_id IS '操作人ID';
COMMENT ON COLUMN payments.remark IS '备注';
COMMENT ON COLUMN payments.created_at IS '创建时间';
COMMENT ON COLUMN payments.updated_at IS '更新时间';
COMMENT ON COLUMN payments.deleted_at IS '删除时间';
CREATE UNIQUE INDEX uk_payments_store_id_payment_no ON payments (store_id, payment_no) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX uk_payments_order_id_success ON payments (order_id) WHERE status = 'success' AND deleted_at IS NULL;
CREATE INDEX idx_payments_store_id ON payments (store_id);
CREATE INDEX idx_payments_order_id ON payments (order_id);
CREATE INDEX idx_payments_channel ON payments (channel);
CREATE INDEX idx_payments_deleted_at ON payments (deleted_at);
