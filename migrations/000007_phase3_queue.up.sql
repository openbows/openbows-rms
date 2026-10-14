-- 第 3 期：排队叫号
-- 状态：waiting → called → seated；waiting/called → cancelled

CREATE TABLE IF NOT EXISTS queue_tickets (
    id            BIGSERIAL PRIMARY KEY,
    store_id      BIGINT       NOT NULL,
    business_day  DATE         NOT NULL,
    ticket_no     INT          NOT NULL,
    guest_count   INT          NOT NULL DEFAULT 1,
    phone         VARCHAR(32)  NOT NULL DEFAULT '',
    status        VARCHAR(32)  NOT NULL DEFAULT 'waiting',
    order_id      BIGINT       NULL,
    table_id      BIGINT       NULL,
    remark        VARCHAR(255) NOT NULL DEFAULT '',
    called_at     TIMESTAMPTZ  NULL,
    seated_at     TIMESTAMPTZ  NULL,
    cancelled_at  TIMESTAMPTZ  NULL,
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at    TIMESTAMPTZ  NULL,
    CONSTRAINT fk_queue_tickets_store_id FOREIGN KEY (store_id) REFERENCES stores (id),
    CONSTRAINT fk_queue_tickets_order_id FOREIGN KEY (order_id) REFERENCES orders (id),
    CONSTRAINT fk_queue_tickets_table_id FOREIGN KEY (table_id) REFERENCES tables (id),
    CONSTRAINT ck_queue_tickets_status CHECK (status IN ('waiting', 'called', 'seated', 'cancelled')),
    CONSTRAINT ck_queue_tickets_guest_count CHECK (guest_count > 0),
    CONSTRAINT ck_queue_tickets_ticket_no CHECK (ticket_no > 0)
);

COMMENT ON TABLE queue_tickets IS '排队叫号票';
COMMENT ON COLUMN queue_tickets.id IS '主键';
COMMENT ON COLUMN queue_tickets.store_id IS '门店ID';
COMMENT ON COLUMN queue_tickets.business_day IS '营业日';
COMMENT ON COLUMN queue_tickets.ticket_no IS '排队号（门店+营业日内唯一）';
COMMENT ON COLUMN queue_tickets.guest_count IS '就餐人数';
COMMENT ON COLUMN queue_tickets.phone IS '联系电话';
COMMENT ON COLUMN queue_tickets.status IS '状态：waiting/called/seated/cancelled';
COMMENT ON COLUMN queue_tickets.order_id IS '关联订单ID（可选）';
COMMENT ON COLUMN queue_tickets.table_id IS '入座桌台ID（可选）';
COMMENT ON COLUMN queue_tickets.remark IS '备注';
COMMENT ON COLUMN queue_tickets.called_at IS '叫号时间';
COMMENT ON COLUMN queue_tickets.seated_at IS '入座时间';
COMMENT ON COLUMN queue_tickets.cancelled_at IS '取消时间';
COMMENT ON COLUMN queue_tickets.created_at IS '创建时间';
COMMENT ON COLUMN queue_tickets.updated_at IS '更新时间';
COMMENT ON COLUMN queue_tickets.deleted_at IS '删除时间';

-- 关键：排队号在「门店 + 营业日」内唯一（软删除部分唯一索引）
CREATE UNIQUE INDEX uk_queue_tickets_store_day_no
    ON queue_tickets (store_id, business_day, ticket_no)
    WHERE deleted_at IS NULL;

CREATE INDEX idx_queue_tickets_store_id ON queue_tickets (store_id);
CREATE INDEX idx_queue_tickets_business_day ON queue_tickets (business_day);
CREATE INDEX idx_queue_tickets_status ON queue_tickets (status);
CREATE INDEX idx_queue_tickets_store_day_status ON queue_tickets (store_id, business_day, status);
CREATE INDEX idx_queue_tickets_deleted_at ON queue_tickets (deleted_at);
