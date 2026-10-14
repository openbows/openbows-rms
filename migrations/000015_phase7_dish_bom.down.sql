ALTER TABLE payments DROP CONSTRAINT IF EXISTS ck_payments_channel;
ALTER TABLE payments ADD CONSTRAINT ck_payments_channel
    CHECK (channel IN ('cash', 'wechat', 'alipay', 'other'));

DROP TABLE IF EXISTS dish_materials;
