-- 第 7 期：菜品配方 BOM + 支付渠道 member

CREATE TABLE IF NOT EXISTS dish_materials (
    id          BIGSERIAL PRIMARY KEY,
    store_id    BIGINT         NOT NULL,
    dish_id     BIGINT         NOT NULL,
    material_id BIGINT         NOT NULL,
    qty         NUMERIC(12, 3) NOT NULL,
    remark      VARCHAR(255)   NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ    NULL,
    CONSTRAINT fk_dish_materials_store_id FOREIGN KEY (store_id) REFERENCES stores (id),
    CONSTRAINT fk_dish_materials_dish_id FOREIGN KEY (dish_id) REFERENCES dishes (id),
    CONSTRAINT fk_dish_materials_material_id FOREIGN KEY (material_id) REFERENCES materials (id),
    CONSTRAINT ck_dish_materials_qty CHECK (qty > 0)
);

COMMENT ON TABLE dish_materials IS '菜品配方（BOM）';
COMMENT ON COLUMN dish_materials.id IS '主键';
COMMENT ON COLUMN dish_materials.store_id IS '门店ID';
COMMENT ON COLUMN dish_materials.dish_id IS '菜品ID';
COMMENT ON COLUMN dish_materials.material_id IS '物料ID';
COMMENT ON COLUMN dish_materials.qty IS '单份用量';
COMMENT ON COLUMN dish_materials.remark IS '备注';
COMMENT ON COLUMN dish_materials.created_at IS '创建时间';
COMMENT ON COLUMN dish_materials.updated_at IS '更新时间';
COMMENT ON COLUMN dish_materials.deleted_at IS '删除时间';

CREATE UNIQUE INDEX uk_dish_materials_dish_material
    ON dish_materials (dish_id, material_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_dish_materials_store_id ON dish_materials (store_id);
CREATE INDEX idx_dish_materials_dish_id ON dish_materials (dish_id);
CREATE INDEX idx_dish_materials_material_id ON dish_materials (material_id);
CREATE INDEX idx_dish_materials_deleted_at ON dish_materials (deleted_at);

-- 支付渠道增加会员余额
ALTER TABLE payments DROP CONSTRAINT IF EXISTS ck_payments_channel;
ALTER TABLE payments ADD CONSTRAINT ck_payments_channel
    CHECK (channel IN ('cash', 'wechat', 'alipay', 'other', 'member'));
COMMENT ON COLUMN payments.channel IS '支付渠道：cash/wechat/alipay/other/member';
