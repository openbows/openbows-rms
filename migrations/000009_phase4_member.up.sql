-- 第 4 期：会员管理（余额/积分变动必须记流水）

CREATE TABLE IF NOT EXISTS members (
    id           BIGSERIAL PRIMARY KEY,
    store_id     BIGINT       NOT NULL,
    phone        VARCHAR(32)  NOT NULL,
    name         VARCHAR(128) NOT NULL DEFAULT '',
    gender       VARCHAR(16)  NOT NULL DEFAULT 'unknown',
    birthday     DATE         NULL,
    level        VARCHAR(32)  NOT NULL DEFAULT 'normal',
    status       BOOLEAN      NOT NULL DEFAULT TRUE,
    remark       VARCHAR(512) NOT NULL DEFAULT '',
    created_by   BIGINT       NOT NULL DEFAULT 0,
    updated_by   BIGINT       NOT NULL DEFAULT 0,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at   TIMESTAMPTZ  NULL,
    CONSTRAINT fk_members_store_id FOREIGN KEY (store_id) REFERENCES stores (id),
    CONSTRAINT ck_members_gender CHECK (gender IN ('unknown', 'male', 'female'))
);

COMMENT ON TABLE members IS '会员';
COMMENT ON COLUMN members.id IS '主键';
COMMENT ON COLUMN members.store_id IS '门店ID';
COMMENT ON COLUMN members.phone IS '手机号（门店内唯一）';
COMMENT ON COLUMN members.name IS '姓名';
COMMENT ON COLUMN members.gender IS '性别';
COMMENT ON COLUMN members.birthday IS '生日';
COMMENT ON COLUMN members.level IS '会员等级';
COMMENT ON COLUMN members.status IS '是否启用';
COMMENT ON COLUMN members.remark IS '备注';
COMMENT ON COLUMN members.created_by IS '创建人';
COMMENT ON COLUMN members.updated_by IS '更新人';
COMMENT ON COLUMN members.created_at IS '创建时间';
COMMENT ON COLUMN members.updated_at IS '更新时间';
COMMENT ON COLUMN members.deleted_at IS '删除时间';

CREATE UNIQUE INDEX uk_members_store_id_phone ON members (store_id, phone) WHERE deleted_at IS NULL;
CREATE INDEX idx_members_store_id ON members (store_id);
CREATE INDEX idx_members_phone ON members (phone);
CREATE INDEX idx_members_status ON members (status);
CREATE INDEX idx_members_deleted_at ON members (deleted_at);

CREATE TABLE IF NOT EXISTS member_wallets (
    id         BIGSERIAL PRIMARY KEY,
    store_id   BIGINT         NOT NULL,
    member_id  BIGINT         NOT NULL,
    balance    NUMERIC(10, 2) NOT NULL DEFAULT 0,
    points     INT            NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ    NULL,
    CONSTRAINT fk_member_wallets_store_id FOREIGN KEY (store_id) REFERENCES stores (id),
    CONSTRAINT fk_member_wallets_member_id FOREIGN KEY (member_id) REFERENCES members (id),
    CONSTRAINT ck_member_wallets_balance CHECK (balance >= 0),
    CONSTRAINT ck_member_wallets_points CHECK (points >= 0)
);

COMMENT ON TABLE member_wallets IS '会员钱包';
COMMENT ON COLUMN member_wallets.id IS '主键';
COMMENT ON COLUMN member_wallets.store_id IS '门店ID';
COMMENT ON COLUMN member_wallets.member_id IS '会员ID';
COMMENT ON COLUMN member_wallets.balance IS '余额';
COMMENT ON COLUMN member_wallets.points IS '积分';
COMMENT ON COLUMN member_wallets.created_at IS '创建时间';
COMMENT ON COLUMN member_wallets.updated_at IS '更新时间';
COMMENT ON COLUMN member_wallets.deleted_at IS '删除时间';

CREATE UNIQUE INDEX uk_member_wallets_member_id ON member_wallets (member_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_member_wallets_store_id ON member_wallets (store_id);
CREATE INDEX idx_member_wallets_deleted_at ON member_wallets (deleted_at);

CREATE TABLE IF NOT EXISTS member_wallet_logs (
    id             BIGSERIAL PRIMARY KEY,
    store_id       BIGINT         NOT NULL,
    member_id      BIGINT         NOT NULL,
    wallet_id      BIGINT         NOT NULL,
    change_type    VARCHAR(32)    NOT NULL,
    amount         NUMERIC(10, 2) NOT NULL,
    balance_before NUMERIC(10, 2) NOT NULL,
    balance_after  NUMERIC(10, 2) NOT NULL,
    biz_no         VARCHAR(64)    NOT NULL DEFAULT '',
    order_id       BIGINT         NULL,
    remark         VARCHAR(255)   NOT NULL DEFAULT '',
    operator_id    BIGINT         NOT NULL DEFAULT 0,
    created_at     TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    deleted_at     TIMESTAMPTZ    NULL,
    CONSTRAINT fk_member_wallet_logs_store_id FOREIGN KEY (store_id) REFERENCES stores (id),
    CONSTRAINT fk_member_wallet_logs_member_id FOREIGN KEY (member_id) REFERENCES members (id),
    CONSTRAINT fk_member_wallet_logs_wallet_id FOREIGN KEY (wallet_id) REFERENCES member_wallets (id),
    CONSTRAINT ck_member_wallet_logs_change_type CHECK (change_type IN ('recharge', 'consume', 'adjust')),
    CONSTRAINT ck_member_wallet_logs_amount CHECK (amount > 0)
);

COMMENT ON TABLE member_wallet_logs IS '会员余额流水';
COMMENT ON COLUMN member_wallet_logs.id IS '主键';
COMMENT ON COLUMN member_wallet_logs.store_id IS '门店ID';
COMMENT ON COLUMN member_wallet_logs.member_id IS '会员ID';
COMMENT ON COLUMN member_wallet_logs.wallet_id IS '钱包ID';
COMMENT ON COLUMN member_wallet_logs.change_type IS '变动类型：recharge/consume/adjust';
COMMENT ON COLUMN member_wallet_logs.amount IS '变动金额（正数）';
COMMENT ON COLUMN member_wallet_logs.balance_before IS '变动前余额';
COMMENT ON COLUMN member_wallet_logs.balance_after IS '变动后余额';
COMMENT ON COLUMN member_wallet_logs.biz_no IS '业务单号';
COMMENT ON COLUMN member_wallet_logs.order_id IS '关联订单ID（可选）';
COMMENT ON COLUMN member_wallet_logs.remark IS '备注';
COMMENT ON COLUMN member_wallet_logs.operator_id IS '操作人ID';
COMMENT ON COLUMN member_wallet_logs.created_at IS '创建时间';
COMMENT ON COLUMN member_wallet_logs.updated_at IS '更新时间';
COMMENT ON COLUMN member_wallet_logs.deleted_at IS '删除时间';

CREATE INDEX idx_member_wallet_logs_store_id ON member_wallet_logs (store_id);
CREATE INDEX idx_member_wallet_logs_member_id ON member_wallet_logs (member_id);
CREATE INDEX idx_member_wallet_logs_wallet_id ON member_wallet_logs (wallet_id);
CREATE INDEX idx_member_wallet_logs_change_type ON member_wallet_logs (change_type);
CREATE INDEX idx_member_wallet_logs_created_at ON member_wallet_logs (created_at);
CREATE INDEX idx_member_wallet_logs_deleted_at ON member_wallet_logs (deleted_at);

CREATE TABLE IF NOT EXISTS member_point_logs (
    id            BIGSERIAL PRIMARY KEY,
    store_id      BIGINT       NOT NULL,
    member_id     BIGINT       NOT NULL,
    wallet_id     BIGINT       NOT NULL,
    change_type   VARCHAR(32)  NOT NULL,
    points        INT          NOT NULL,
    points_before INT          NOT NULL,
    points_after  INT          NOT NULL,
    biz_no        VARCHAR(64)  NOT NULL DEFAULT '',
    order_id      BIGINT       NULL,
    remark        VARCHAR(255) NOT NULL DEFAULT '',
    operator_id   BIGINT       NOT NULL DEFAULT 0,
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at    TIMESTAMPTZ  NULL,
    CONSTRAINT fk_member_point_logs_store_id FOREIGN KEY (store_id) REFERENCES stores (id),
    CONSTRAINT fk_member_point_logs_member_id FOREIGN KEY (member_id) REFERENCES members (id),
    CONSTRAINT fk_member_point_logs_wallet_id FOREIGN KEY (wallet_id) REFERENCES member_wallets (id),
    CONSTRAINT ck_member_point_logs_change_type CHECK (change_type IN ('earn', 'spend', 'adjust')),
    CONSTRAINT ck_member_point_logs_points CHECK (points > 0)
);

COMMENT ON TABLE member_point_logs IS '会员积分流水';
COMMENT ON COLUMN member_point_logs.id IS '主键';
COMMENT ON COLUMN member_point_logs.store_id IS '门店ID';
COMMENT ON COLUMN member_point_logs.member_id IS '会员ID';
COMMENT ON COLUMN member_point_logs.wallet_id IS '钱包ID';
COMMENT ON COLUMN member_point_logs.change_type IS '变动类型：earn/spend/adjust';
COMMENT ON COLUMN member_point_logs.points IS '变动积分（正数）';
COMMENT ON COLUMN member_point_logs.points_before IS '变动前积分';
COMMENT ON COLUMN member_point_logs.points_after IS '变动后积分';
COMMENT ON COLUMN member_point_logs.biz_no IS '业务单号';
COMMENT ON COLUMN member_point_logs.order_id IS '关联订单ID（可选）';
COMMENT ON COLUMN member_point_logs.remark IS '备注';
COMMENT ON COLUMN member_point_logs.operator_id IS '操作人ID';
COMMENT ON COLUMN member_point_logs.created_at IS '创建时间';
COMMENT ON COLUMN member_point_logs.updated_at IS '更新时间';
COMMENT ON COLUMN member_point_logs.deleted_at IS '删除时间';

CREATE INDEX idx_member_point_logs_store_id ON member_point_logs (store_id);
CREATE INDEX idx_member_point_logs_member_id ON member_point_logs (member_id);
CREATE INDEX idx_member_point_logs_wallet_id ON member_point_logs (wallet_id);
CREATE INDEX idx_member_point_logs_change_type ON member_point_logs (change_type);
CREATE INDEX idx_member_point_logs_created_at ON member_point_logs (created_at);
CREATE INDEX idx_member_point_logs_deleted_at ON member_point_logs (deleted_at);
