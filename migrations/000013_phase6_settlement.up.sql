-- 第 6 期：支付对账 / 日结

CREATE TABLE IF NOT EXISTS settlements (
    id              BIGSERIAL PRIMARY KEY,
    store_id        BIGINT         NOT NULL,
    business_day    DATE           NOT NULL,
    order_count     INT            NOT NULL DEFAULT 0,
    payment_count   INT            NOT NULL DEFAULT 0,
    total_amount    NUMERIC(10, 2) NOT NULL DEFAULT 0,
    cash_amount     NUMERIC(10, 2) NOT NULL DEFAULT 0,
    wechat_amount   NUMERIC(10, 2) NOT NULL DEFAULT 0,
    alipay_amount   NUMERIC(10, 2) NOT NULL DEFAULT 0,
    other_amount    NUMERIC(10, 2) NOT NULL DEFAULT 0,
    status          VARCHAR(32)    NOT NULL DEFAULT 'closed',
    remark          VARCHAR(512)   NOT NULL DEFAULT '',
    created_by      BIGINT         NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ    NULL,
    CONSTRAINT fk_settlements_store_id FOREIGN KEY (store_id) REFERENCES stores (id),
    CONSTRAINT ck_settlements_status CHECK (status IN ('closed'))
);

COMMENT ON TABLE settlements IS '日结汇总';
COMMENT ON COLUMN settlements.id IS '主键';
COMMENT ON COLUMN settlements.store_id IS '门店ID';
COMMENT ON COLUMN settlements.business_day IS '营业日';
COMMENT ON COLUMN settlements.order_count IS '关联订单数';
COMMENT ON COLUMN settlements.payment_count IS '支付笔数';
COMMENT ON COLUMN settlements.total_amount IS '支付合计';
COMMENT ON COLUMN settlements.cash_amount IS '现金合计';
COMMENT ON COLUMN settlements.wechat_amount IS '微信合计';
COMMENT ON COLUMN settlements.alipay_amount IS '支付宝合计';
COMMENT ON COLUMN settlements.other_amount IS '其他渠道合计';
COMMENT ON COLUMN settlements.status IS '状态：closed';
COMMENT ON COLUMN settlements.remark IS '备注';
COMMENT ON COLUMN settlements.created_by IS '日结人';
COMMENT ON COLUMN settlements.created_at IS '创建时间';
COMMENT ON COLUMN settlements.updated_at IS '更新时间';
COMMENT ON COLUMN settlements.deleted_at IS '删除时间';

-- 关键：同门店同营业日不可重复日结
CREATE UNIQUE INDEX uk_settlements_store_business_day
    ON settlements (store_id, business_day) WHERE deleted_at IS NULL;
CREATE INDEX idx_settlements_store_id ON settlements (store_id);
CREATE INDEX idx_settlements_business_day ON settlements (business_day);
CREATE INDEX idx_settlements_deleted_at ON settlements (deleted_at);

CREATE TABLE IF NOT EXISTS settlement_items (
    id            BIGSERIAL PRIMARY KEY,
    store_id      BIGINT         NOT NULL,
    settlement_id BIGINT         NOT NULL,
    payment_id    BIGINT         NOT NULL,
    order_id      BIGINT         NOT NULL,
    channel       VARCHAR(32)    NOT NULL,
    amount        NUMERIC(10, 2) NOT NULL,
    paid_at       TIMESTAMPTZ    NOT NULL,
    created_at    TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    deleted_at    TIMESTAMPTZ    NULL,
    CONSTRAINT fk_settlement_items_store_id FOREIGN KEY (store_id) REFERENCES stores (id),
    CONSTRAINT fk_settlement_items_settlement_id FOREIGN KEY (settlement_id) REFERENCES settlements (id),
    CONSTRAINT fk_settlement_items_payment_id FOREIGN KEY (payment_id) REFERENCES payments (id),
    CONSTRAINT fk_settlement_items_order_id FOREIGN KEY (order_id) REFERENCES orders (id)
);

COMMENT ON TABLE settlement_items IS '日结明细（追溯支付/订单）';
COMMENT ON COLUMN settlement_items.id IS '主键';
COMMENT ON COLUMN settlement_items.store_id IS '门店ID';
COMMENT ON COLUMN settlement_items.settlement_id IS '日结ID';
COMMENT ON COLUMN settlement_items.payment_id IS '支付ID';
COMMENT ON COLUMN settlement_items.order_id IS '订单ID';
COMMENT ON COLUMN settlement_items.channel IS '支付渠道';
COMMENT ON COLUMN settlement_items.amount IS '支付金额';
COMMENT ON COLUMN settlement_items.paid_at IS '支付时间';
COMMENT ON COLUMN settlement_items.created_at IS '创建时间';
COMMENT ON COLUMN settlement_items.updated_at IS '更新时间';
COMMENT ON COLUMN settlement_items.deleted_at IS '删除时间';

CREATE INDEX idx_settlement_items_store_id ON settlement_items (store_id);
CREATE INDEX idx_settlement_items_settlement_id ON settlement_items (settlement_id);
CREATE INDEX idx_settlement_items_payment_id ON settlement_items (payment_id);
CREATE INDEX idx_settlement_items_order_id ON settlement_items (order_id);
CREATE INDEX idx_settlement_items_deleted_at ON settlement_items (deleted_at);

CREATE TABLE IF NOT EXISTS reconciliations (
    id             BIGSERIAL PRIMARY KEY,
    store_id       BIGINT         NOT NULL,
    business_day   DATE           NOT NULL,
    channel        VARCHAR(32)    NOT NULL,
    system_amount  NUMERIC(10, 2) NOT NULL DEFAULT 0,
    channel_amount NUMERIC(10, 2) NOT NULL DEFAULT 0,
    diff_amount    NUMERIC(10, 2) NOT NULL DEFAULT 0,
    system_count   INT            NOT NULL DEFAULT 0,
    channel_count  INT            NOT NULL DEFAULT 0,
    status         VARCHAR(32)    NOT NULL DEFAULT 'matched',
    remark         VARCHAR(512)   NOT NULL DEFAULT '',
    created_by     BIGINT         NOT NULL DEFAULT 0,
    created_at     TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    deleted_at     TIMESTAMPTZ    NULL,
    CONSTRAINT fk_reconciliations_store_id FOREIGN KEY (store_id) REFERENCES stores (id),
    CONSTRAINT ck_reconciliations_channel CHECK (channel IN ('cash', 'wechat', 'alipay', 'other')),
    CONSTRAINT ck_reconciliations_status CHECK (status IN ('matched', 'mismatched'))
);

COMMENT ON TABLE reconciliations IS '渠道对账';
COMMENT ON COLUMN reconciliations.id IS '主键';
COMMENT ON COLUMN reconciliations.store_id IS '门店ID';
COMMENT ON COLUMN reconciliations.business_day IS '营业日';
COMMENT ON COLUMN reconciliations.channel IS '支付渠道';
COMMENT ON COLUMN reconciliations.system_amount IS '系统支付合计';
COMMENT ON COLUMN reconciliations.channel_amount IS '渠道账单合计';
COMMENT ON COLUMN reconciliations.diff_amount IS '差异金额（渠道-系统）';
COMMENT ON COLUMN reconciliations.system_count IS '系统笔数';
COMMENT ON COLUMN reconciliations.channel_count IS '渠道笔数';
COMMENT ON COLUMN reconciliations.status IS '状态：matched/mismatched';
COMMENT ON COLUMN reconciliations.remark IS '备注';
COMMENT ON COLUMN reconciliations.created_by IS '对账人';
COMMENT ON COLUMN reconciliations.created_at IS '创建时间';
COMMENT ON COLUMN reconciliations.updated_at IS '更新时间';
COMMENT ON COLUMN reconciliations.deleted_at IS '删除时间';

CREATE INDEX idx_reconciliations_store_id ON reconciliations (store_id);
CREATE INDEX idx_reconciliations_business_day ON reconciliations (business_day);
CREATE INDEX idx_reconciliations_channel ON reconciliations (channel);
CREATE INDEX idx_reconciliations_status ON reconciliations (status);
CREATE INDEX idx_reconciliations_deleted_at ON reconciliations (deleted_at);

-- diff_type: match / amount_mismatch / missing_in_channel / missing_in_system / total_diff
CREATE TABLE IF NOT EXISTS reconciliation_items (
    id               BIGSERIAL PRIMARY KEY,
    store_id         BIGINT         NOT NULL,
    reconciliation_id BIGINT        NOT NULL,
    payment_id       BIGINT         NULL,
    order_id         BIGINT         NULL,
    out_trade_no     VARCHAR(128)   NOT NULL DEFAULT '',
    system_amount    NUMERIC(10, 2) NOT NULL DEFAULT 0,
    channel_amount   NUMERIC(10, 2) NOT NULL DEFAULT 0,
    diff_amount      NUMERIC(10, 2) NOT NULL DEFAULT 0,
    diff_type        VARCHAR(32)    NOT NULL,
    remark           VARCHAR(255)   NOT NULL DEFAULT '',
    created_at       TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    deleted_at       TIMESTAMPTZ    NULL,
    CONSTRAINT fk_reconciliation_items_store_id FOREIGN KEY (store_id) REFERENCES stores (id),
    CONSTRAINT fk_reconciliation_items_recon_id FOREIGN KEY (reconciliation_id) REFERENCES reconciliations (id),
    CONSTRAINT fk_reconciliation_items_payment_id FOREIGN KEY (payment_id) REFERENCES payments (id),
    CONSTRAINT fk_reconciliation_items_order_id FOREIGN KEY (order_id) REFERENCES orders (id),
    CONSTRAINT ck_reconciliation_items_diff_type CHECK (diff_type IN ('match', 'amount_mismatch', 'missing_in_channel', 'missing_in_system', 'total_diff'))
);

COMMENT ON TABLE reconciliation_items IS '对账明细/差异';
COMMENT ON COLUMN reconciliation_items.id IS '主键';
COMMENT ON COLUMN reconciliation_items.store_id IS '门店ID';
COMMENT ON COLUMN reconciliation_items.reconciliation_id IS '对账ID';
COMMENT ON COLUMN reconciliation_items.payment_id IS '系统支付ID';
COMMENT ON COLUMN reconciliation_items.order_id IS '订单ID';
COMMENT ON COLUMN reconciliation_items.out_trade_no IS '第三方交易号';
COMMENT ON COLUMN reconciliation_items.system_amount IS '系统金额';
COMMENT ON COLUMN reconciliation_items.channel_amount IS '渠道金额';
COMMENT ON COLUMN reconciliation_items.diff_amount IS '差异金额';
COMMENT ON COLUMN reconciliation_items.diff_type IS '差异类型';
COMMENT ON COLUMN reconciliation_items.remark IS '备注';
COMMENT ON COLUMN reconciliation_items.created_at IS '创建时间';
COMMENT ON COLUMN reconciliation_items.updated_at IS '更新时间';
COMMENT ON COLUMN reconciliation_items.deleted_at IS '删除时间';

CREATE INDEX idx_reconciliation_items_store_id ON reconciliation_items (store_id);
CREATE INDEX idx_reconciliation_items_recon_id ON reconciliation_items (reconciliation_id);
CREATE INDEX idx_reconciliation_items_diff_type ON reconciliation_items (diff_type);
CREATE INDEX idx_reconciliation_items_payment_id ON reconciliation_items (payment_id);
CREATE INDEX idx_reconciliation_items_deleted_at ON reconciliation_items (deleted_at);
