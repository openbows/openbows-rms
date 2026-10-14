-- 第 3 期：厨房显示 KDS（方案 B：独立厨单表 + 回写 order_items.kitchen_status）

ALTER TABLE order_items
    ADD COLUMN IF NOT EXISTS kitchen_status VARCHAR(32) NOT NULL DEFAULT 'pending';
COMMENT ON COLUMN order_items.kitchen_status IS '出品状态：pending/cooking/done';
ALTER TABLE order_items DROP CONSTRAINT IF EXISTS ck_order_items_kitchen_status;
ALTER TABLE order_items
    ADD CONSTRAINT ck_order_items_kitchen_status CHECK (kitchen_status IN ('pending', 'cooking', 'done'));
CREATE INDEX IF NOT EXISTS idx_order_items_kitchen_status ON order_items (kitchen_status);

-- ticket status: pending → cooking → done
CREATE TABLE IF NOT EXISTS kds_tickets (
    id         BIGSERIAL PRIMARY KEY,
    store_id   BIGINT       NOT NULL,
    station_id BIGINT       NOT NULL,
    order_id   BIGINT       NOT NULL,
    order_no   VARCHAR(64)  NOT NULL,
    table_id   BIGINT       NOT NULL,
    status     VARCHAR(32)  NOT NULL DEFAULT 'pending',
    started_at TIMESTAMPTZ  NULL,
    done_at    TIMESTAMPTZ  NULL,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ  NULL,
    CONSTRAINT fk_kds_tickets_store_id FOREIGN KEY (store_id) REFERENCES stores (id),
    CONSTRAINT fk_kds_tickets_station_id FOREIGN KEY (station_id) REFERENCES stations (id),
    CONSTRAINT fk_kds_tickets_order_id FOREIGN KEY (order_id) REFERENCES orders (id),
    CONSTRAINT fk_kds_tickets_table_id FOREIGN KEY (table_id) REFERENCES tables (id),
    CONSTRAINT ck_kds_tickets_status CHECK (status IN ('pending', 'cooking', 'done'))
);
COMMENT ON TABLE kds_tickets IS 'KDS厨单';
COMMENT ON COLUMN kds_tickets.id IS '主键';
COMMENT ON COLUMN kds_tickets.store_id IS '门店ID';
COMMENT ON COLUMN kds_tickets.station_id IS '档口ID';
COMMENT ON COLUMN kds_tickets.order_id IS '订单ID';
COMMENT ON COLUMN kds_tickets.order_no IS '订单号快照';
COMMENT ON COLUMN kds_tickets.table_id IS '桌台ID';
COMMENT ON COLUMN kds_tickets.status IS '厨单状态：pending/cooking/done';
COMMENT ON COLUMN kds_tickets.started_at IS '开始制作时间';
COMMENT ON COLUMN kds_tickets.done_at IS '出餐完成时间';
COMMENT ON COLUMN kds_tickets.created_at IS '创建时间';
COMMENT ON COLUMN kds_tickets.updated_at IS '更新时间';
COMMENT ON COLUMN kds_tickets.deleted_at IS '删除时间';
CREATE INDEX idx_kds_tickets_store_id ON kds_tickets (store_id);
CREATE INDEX idx_kds_tickets_station_id ON kds_tickets (station_id);
CREATE INDEX idx_kds_tickets_order_id ON kds_tickets (order_id);
CREATE INDEX idx_kds_tickets_status ON kds_tickets (status);
CREATE INDEX idx_kds_tickets_deleted_at ON kds_tickets (deleted_at);

CREATE TABLE IF NOT EXISTS kds_ticket_items (
    id            BIGSERIAL PRIMARY KEY,
    store_id      BIGINT       NOT NULL,
    ticket_id     BIGINT       NOT NULL,
    order_item_id BIGINT       NOT NULL,
    dish_name     VARCHAR(128) NOT NULL,
    spec_name     VARCHAR(128) NOT NULL DEFAULT '',
    qty           INT          NOT NULL DEFAULT 1,
    status        VARCHAR(32)  NOT NULL DEFAULT 'pending',
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at    TIMESTAMPTZ  NULL,
    CONSTRAINT fk_kds_ticket_items_store_id FOREIGN KEY (store_id) REFERENCES stores (id),
    CONSTRAINT fk_kds_ticket_items_ticket_id FOREIGN KEY (ticket_id) REFERENCES kds_tickets (id),
    CONSTRAINT fk_kds_ticket_items_order_item_id FOREIGN KEY (order_item_id) REFERENCES order_items (id),
    CONSTRAINT ck_kds_ticket_items_status CHECK (status IN ('pending', 'cooking', 'done'))
);
COMMENT ON TABLE kds_ticket_items IS 'KDS厨单明细';
COMMENT ON COLUMN kds_ticket_items.id IS '主键';
COMMENT ON COLUMN kds_ticket_items.store_id IS '门店ID';
COMMENT ON COLUMN kds_ticket_items.ticket_id IS '厨单ID';
COMMENT ON COLUMN kds_ticket_items.order_item_id IS '订单明细ID';
COMMENT ON COLUMN kds_ticket_items.dish_name IS '菜品名称快照';
COMMENT ON COLUMN kds_ticket_items.spec_name IS '规格名称快照';
COMMENT ON COLUMN kds_ticket_items.qty IS '数量';
COMMENT ON COLUMN kds_ticket_items.status IS '出品状态：pending/cooking/done';
COMMENT ON COLUMN kds_ticket_items.created_at IS '创建时间';
COMMENT ON COLUMN kds_ticket_items.updated_at IS '更新时间';
COMMENT ON COLUMN kds_ticket_items.deleted_at IS '删除时间';
CREATE INDEX idx_kds_ticket_items_store_id ON kds_ticket_items (store_id);
CREATE INDEX idx_kds_ticket_items_ticket_id ON kds_ticket_items (ticket_id);
CREATE INDEX idx_kds_ticket_items_order_item_id ON kds_ticket_items (order_item_id);
CREATE INDEX idx_kds_ticket_items_deleted_at ON kds_ticket_items (deleted_at);
