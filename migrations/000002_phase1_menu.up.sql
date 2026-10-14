-- 第 1 期：菜单主数据（档口 / 分类 / 菜品 / 规格 / 加料）

CREATE TABLE IF NOT EXISTS stations (
    id         BIGSERIAL PRIMARY KEY,
    store_id   BIGINT       NOT NULL,
    code       VARCHAR(64)  NOT NULL,
    name       VARCHAR(128) NOT NULL,
    sort_no    INT          NOT NULL DEFAULT 0,
    status     BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ  NULL,
    CONSTRAINT fk_stations_store_id FOREIGN KEY (store_id) REFERENCES stores (id)
);
COMMENT ON TABLE stations IS '档口';
COMMENT ON COLUMN stations.id IS '主键';
COMMENT ON COLUMN stations.store_id IS '门店ID';
COMMENT ON COLUMN stations.code IS '档口编码';
COMMENT ON COLUMN stations.name IS '档口名称';
COMMENT ON COLUMN stations.sort_no IS '排序号';
COMMENT ON COLUMN stations.status IS '是否启用';
COMMENT ON COLUMN stations.created_at IS '创建时间';
COMMENT ON COLUMN stations.updated_at IS '更新时间';
COMMENT ON COLUMN stations.deleted_at IS '删除时间';
CREATE UNIQUE INDEX uk_stations_store_id_code ON stations (store_id, code) WHERE deleted_at IS NULL;
CREATE INDEX idx_stations_store_id ON stations (store_id);
CREATE INDEX idx_stations_deleted_at ON stations (deleted_at);

CREATE TABLE IF NOT EXISTS menu_categories (
    id         BIGSERIAL PRIMARY KEY,
    store_id   BIGINT       NOT NULL,
    name       VARCHAR(128) NOT NULL,
    sort_no    INT          NOT NULL DEFAULT 0,
    status     BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ  NULL,
    CONSTRAINT fk_menu_categories_store_id FOREIGN KEY (store_id) REFERENCES stores (id)
);
COMMENT ON TABLE menu_categories IS '菜单分类';
COMMENT ON COLUMN menu_categories.id IS '主键';
COMMENT ON COLUMN menu_categories.store_id IS '门店ID';
COMMENT ON COLUMN menu_categories.name IS '分类名称';
COMMENT ON COLUMN menu_categories.sort_no IS '排序号';
COMMENT ON COLUMN menu_categories.status IS '是否启用';
COMMENT ON COLUMN menu_categories.created_at IS '创建时间';
COMMENT ON COLUMN menu_categories.updated_at IS '更新时间';
COMMENT ON COLUMN menu_categories.deleted_at IS '删除时间';
CREATE UNIQUE INDEX uk_menu_categories_store_id_name ON menu_categories (store_id, name) WHERE deleted_at IS NULL;
CREATE INDEX idx_menu_categories_store_id ON menu_categories (store_id);
CREATE INDEX idx_menu_categories_deleted_at ON menu_categories (deleted_at);

CREATE TABLE IF NOT EXISTS dishes (
    id          BIGSERIAL PRIMARY KEY,
    store_id    BIGINT         NOT NULL,
    category_id BIGINT         NOT NULL,
    station_id  BIGINT         NULL,
    name        VARCHAR(128)   NOT NULL,
    description VARCHAR(512)   NOT NULL DEFAULT '',
    base_price  NUMERIC(10, 2) NOT NULL DEFAULT 0,
    is_on_sale  BOOLEAN        NOT NULL DEFAULT TRUE,
    sort_no     INT            NOT NULL DEFAULT 0,
    image_url   VARCHAR(512)   NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ    NULL,
    CONSTRAINT fk_dishes_store_id FOREIGN KEY (store_id) REFERENCES stores (id),
    CONSTRAINT fk_dishes_category_id FOREIGN KEY (category_id) REFERENCES menu_categories (id),
    CONSTRAINT fk_dishes_station_id FOREIGN KEY (station_id) REFERENCES stations (id)
);
COMMENT ON TABLE dishes IS '菜品';
COMMENT ON COLUMN dishes.id IS '主键';
COMMENT ON COLUMN dishes.store_id IS '门店ID';
COMMENT ON COLUMN dishes.category_id IS '分类ID';
COMMENT ON COLUMN dishes.station_id IS '档口ID';
COMMENT ON COLUMN dishes.name IS '菜品名称';
COMMENT ON COLUMN dishes.description IS '菜品描述';
COMMENT ON COLUMN dishes.base_price IS '基础价';
COMMENT ON COLUMN dishes.is_on_sale IS '是否上架';
COMMENT ON COLUMN dishes.sort_no IS '排序号';
COMMENT ON COLUMN dishes.image_url IS '图片地址';
COMMENT ON COLUMN dishes.created_at IS '创建时间';
COMMENT ON COLUMN dishes.updated_at IS '更新时间';
COMMENT ON COLUMN dishes.deleted_at IS '删除时间';
CREATE UNIQUE INDEX uk_dishes_store_id_name ON dishes (store_id, name) WHERE deleted_at IS NULL;
CREATE INDEX idx_dishes_store_id ON dishes (store_id);
CREATE INDEX idx_dishes_category_id ON dishes (category_id);
CREATE INDEX idx_dishes_station_id ON dishes (station_id);
CREATE INDEX idx_dishes_is_on_sale ON dishes (is_on_sale);
CREATE INDEX idx_dishes_deleted_at ON dishes (deleted_at);

CREATE TABLE IF NOT EXISTS dish_specs (
    id         BIGSERIAL PRIMARY KEY,
    store_id   BIGINT         NOT NULL,
    dish_id    BIGINT         NOT NULL,
    name       VARCHAR(128)   NOT NULL,
    price      NUMERIC(10, 2) NOT NULL DEFAULT 0,
    sort_no    INT            NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ    NULL,
    CONSTRAINT fk_dish_specs_store_id FOREIGN KEY (store_id) REFERENCES stores (id),
    CONSTRAINT fk_dish_specs_dish_id FOREIGN KEY (dish_id) REFERENCES dishes (id)
);
COMMENT ON TABLE dish_specs IS '菜品规格';
COMMENT ON COLUMN dish_specs.id IS '主键';
COMMENT ON COLUMN dish_specs.store_id IS '门店ID';
COMMENT ON COLUMN dish_specs.dish_id IS '菜品ID';
COMMENT ON COLUMN dish_specs.name IS '规格名称';
COMMENT ON COLUMN dish_specs.price IS '规格价格';
COMMENT ON COLUMN dish_specs.sort_no IS '排序号';
COMMENT ON COLUMN dish_specs.created_at IS '创建时间';
COMMENT ON COLUMN dish_specs.updated_at IS '更新时间';
COMMENT ON COLUMN dish_specs.deleted_at IS '删除时间';
CREATE UNIQUE INDEX uk_dish_specs_dish_id_name ON dish_specs (dish_id, name) WHERE deleted_at IS NULL;
CREATE INDEX idx_dish_specs_store_id ON dish_specs (store_id);
CREATE INDEX idx_dish_specs_dish_id ON dish_specs (dish_id);
CREATE INDEX idx_dish_specs_deleted_at ON dish_specs (deleted_at);

CREATE TABLE IF NOT EXISTS dish_options (
    id         BIGSERIAL PRIMARY KEY,
    store_id   BIGINT         NOT NULL,
    dish_id    BIGINT         NOT NULL,
    name       VARCHAR(128)   NOT NULL,
    price      NUMERIC(10, 2) NOT NULL DEFAULT 0,
    sort_no    INT            NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ    NULL,
    CONSTRAINT fk_dish_options_store_id FOREIGN KEY (store_id) REFERENCES stores (id),
    CONSTRAINT fk_dish_options_dish_id FOREIGN KEY (dish_id) REFERENCES dishes (id)
);
COMMENT ON TABLE dish_options IS '菜品加料';
COMMENT ON COLUMN dish_options.id IS '主键';
COMMENT ON COLUMN dish_options.store_id IS '门店ID';
COMMENT ON COLUMN dish_options.dish_id IS '菜品ID';
COMMENT ON COLUMN dish_options.name IS '加料名称';
COMMENT ON COLUMN dish_options.price IS '加料价格';
COMMENT ON COLUMN dish_options.sort_no IS '排序号';
COMMENT ON COLUMN dish_options.created_at IS '创建时间';
COMMENT ON COLUMN dish_options.updated_at IS '更新时间';
COMMENT ON COLUMN dish_options.deleted_at IS '删除时间';
CREATE UNIQUE INDEX uk_dish_options_dish_id_name ON dish_options (dish_id, name) WHERE deleted_at IS NULL;
CREATE INDEX idx_dish_options_store_id ON dish_options (store_id);
CREATE INDEX idx_dish_options_dish_id ON dish_options (dish_id);
CREATE INDEX idx_dish_options_deleted_at ON dish_options (deleted_at);
