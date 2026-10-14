# Openbows-Rms

单店 / 小连锁餐饮门店管理系统：后端 API（Go + Gin + GORM + PostgreSQL）+ Vue3 三端 Web（管理后台 / 收银 POS / 厨房 KDS）。

## 1. 项目解决什么问题

餐饮门店日常经营涉及开台点餐、厨显出品、结账支付、会员、库存采购、日结对账与排班考勤等多条链路。本项目用**单体后端 + 三端前端**把这些能力串成可演示、可落地的闭环：

- **前台**：桌台开台 → 点菜改单 → 厨显出品 → 结账（现金 / 微信 / 支付宝 / 其他 / 会员余额）
- **履约与客流**：排队叫号入座开台、预订确认入座、扫码 / H5 线上点餐 API
- **供应链**：物料与单据化出入库、菜品配方（BOM）结账自动出库、采购收货入库
- **经营与人力**：日结与渠道对账、员工排班考勤
- **权限**：门店隔离（`store_id`）+ JWT + RBAC 权限码强制鉴权

不做：微信小程序、真实第三方支付通道、完整会计总账 / 薪资核算。

## 2. 主要功能

| 域 | 能力 |
|----|------|
| 组织权限 | 门店、用户、角色、权限；登录签发 JWT（含权限码） |
| 菜单桌台 | 档口、分类、菜品（规格 / 加料 / 配方）、区域、桌台状态 |
| 点餐收银 | 订单全链路状态、退菜、费用分项；支付与防重复支付 |
| 厨房 KDS | 按档口厨单；开始 / 完成出品 |
| 排队预订 | 排队取号 / 叫号 / 入座开台；预订确认 / 入座 / 取消 / 违约 |
| 会员线上 | 会员钱包与积分流水；线上点餐 API（`access_token` 查单） |
| 库存采购 | 物料、出入库 / 盘点单据过账；供应商、采购单、收货入库 |
| 财务 HR | 营业日日结、渠道对账与差异；员工、排班、打卡 / 补卡 |
| 前端 | 管理后台、收银 POS、厨显 KDS（Vue3 + Element Plus） |

统一 API 前缀：`/apis/v1/{模块}`  
统一响应：`{"code":200,"msg":"操作成功","data":{}}`（业务错误多为 `code=400/401/403`，HTTP 仍为 200）。

## 3. 安装方法（含项目结构）

### 3.1 环境要求

- Go 1.24+
- PostgreSQL 14+
- Node.js 20+ / npm 10+（前端）
- 可选：Docker Compose

国内拉取 Go 依赖：

```bash
go env -w GOPROXY=https://goproxy.cn,direct
```

### 3.2 项目结构

```text
Openbows-Rms/
├── cmd/server/              # 后端入口
├── configs/                 # 配置样例（config.example.yaml）
├── docker/                  # Dockerfile + docker-compose
├── migrations/              # PostgreSQL 迁移 *.up.sql / *.down.sql
├── scripts/                 # 迁移脚本、联调测试脚本
├── internal/
│   ├── config/              # 配置加载
│   ├── middleware/          # JWT、RBAC、Recovery
│   ├── model/               # GORM 模型
│   ├── repository/          # 数据访问
│   ├── service/             # 业务逻辑
│   ├── controller/          # HTTP 处理器
│   ├── router/              # 路由注册
│   ├── response/            # 统一 JSON
│   └── pkg/                 # 内部复用（如 jwtutil）
├── web/                     # 前端 monorepo（npm workspaces）
│   ├── packages/shared/     # axios / 登录封装
│   ├── apps/admin/          # 管理后台 :5173
│   ├── apps/pos/            # 收银 POS :5174
│   └── apps/kds/            # 厨显 KDS :5175
├── go.mod
└── README.md
```

依赖方向：`router → controller → service → repository → model/db`。

### 3.3 获取代码与配置

```bash
cd Openbows-Rms
cp configs/config.example.yaml configs/config.yaml
# 按本机修改 database.dsn
```

可用环境变量覆盖：`DATABASE_DSN`、`JWT_SECRET`、`JWT_EXPIRE_HOURS`、`SERVER_ADDR`。

### 3.4 数据库迁移

先创建库 `openbows_rms`（UTF8），再按序号执行迁移：

```bash
# 新库：按顺序执行全部 up
# psql "$DATABASE_URL" -f migrations/000001_phase0_platform.up.sql
# ... 直至 000016_phase7_rbac_codes.up.sql

# 或使用仓库脚本（按历史模块 / 全量）
go run scripts/migrate_all.go
# 仅补第 7 期：go run scripts/apply_phase7_migration.go
```

种子账号（迁移内置）：

- 用户名：`admin`
- 密码：`Admin@123`
- 门店：`DEFAULT`

### 3.5 安装前端依赖

```bash
cd web
npm install
```

### 3.6 Docker（可选）

```bash
cd docker
docker compose up -d --build
```

包含 PostgreSQL、Redis（占位）、API。首次仍需对库执行 `migrations/*.up.sql`。

## 4. 使用方法

### 4.1 构建前端并启动后端（推荐：同一端口访问）

后端会托管 `web/apps/*/dist`：

| 地址 | 说明 |
|------|------|
| http://127.0.0.1:8080/ | 管理后台（登录页） |
| http://127.0.0.1:8080/pos/ | 收银 POS |
| http://127.0.0.1:8080/kds/ | 厨显 KDS |

```bash
# 仓库根目录
cd web && npm install && npm run build && cd ..
go run ./cmd/server -config configs/config.yaml
```

或：

```bash
go build -o bin/server.exe ./cmd/server
./bin/server.exe -config configs/config.yaml
```

浏览器打开 http://127.0.0.1:8080/ ，使用 `admin` / `Admin@123` 登录。

### 4.2 开发模式（前后端分离热更新）

后端需已在 `:8080`；Vite 将 `/apis` 代理到该地址。

```bash
cd web
npm run dev:admin   # http://127.0.0.1:5173  管理后台
npm run dev:pos     # http://127.0.0.1:5174  收银
npm run dev:kds     # http://127.0.0.1:5175  厨显
```

### 4.3 建议演示路径

1. **管理后台**：维护菜品配方、物料入库  
2. **POS**：选空闲桌台 → 开台点菜 → 推进至待结账 → 现金或会员余额结账  
3. **KDS**：对待制作厨单点「开始 / 完成」  
4. **管理后台**：查看出库单、库存余额、日结  

### 4.4 接口联调脚本（可选）

```powershell
# Windows PowerShell，后端需已启动
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/test_full_suite.ps1
```

## 5. 输入输出示例

### 5.1 登录

**请求**

```bash
curl -s -X POST http://127.0.0.1:8080/apis/v1/auth/login \
  -H "Content-Type: application/json" \
  -d "{\"username\":\"admin\",\"password\":\"Admin@123\"}"
```

**响应（节选）**

```json
{
  "code": 200,
  "msg": "操作成功",
  "data": {
    "token": "<JWT>",
    "user": {
      "id": 1,
      "store_id": 1,
      "username": "admin",
      "display_name": "系统管理员"
    }
  }
}
```

### 5.2 当前用户与权限

```bash
curl -s http://127.0.0.1:8080/apis/v1/auth/me \
  -H "Authorization: Bearer <JWT>"
```

```json
{
  "code": 200,
  "msg": "操作成功",
  "data": {
    "user": { "id": 1, "username": "admin", "store_id": 1 },
    "roles": [{ "code": "admin", "name": "管理员" }],
    "permission_codes": ["store:read", "order:write", "payment:write", "..."]
  }
}
```

### 5.3 开台下单

```bash
curl -s -X POST http://127.0.0.1:8080/apis/v1/orders \
  -H "Authorization: Bearer <JWT>" \
  -H "Content-Type: application/json" \
  -d "{\"table_id\":1,\"guest_count\":2,\"items\":[{\"dish_id\":1,\"qty\":1}]}"
```

```json
{
  "code": 200,
  "msg": "操作成功",
  "data": {
    "order": {
      "id": 1,
      "order_no": "O1...",
      "status": "pending_confirm",
      "payable_amount": "18.00"
    },
    "items": [{ "dish_name": "...", "qty": 1, "amount": "18.00" }]
  }
}
```

### 5.4 结账（现金）

订单需先推进到 `pending_pay`，金额须等于应付：

```bash
curl -s -X POST http://127.0.0.1:8080/apis/v1/orders/1/pay \
  -H "Authorization: Bearer <JWT>" \
  -H "Content-Type: application/json" \
  -d "{\"channel\":\"cash\",\"amount\":\"18.00\"}"
```

会员余额支付额外传 `member_id`，且 `channel` 为 `member`。有配方时同事务生成出库单并扣减结存。

### 5.5 错误示例

未登录：

```json
{ "code": 401, "msg": "未登录或令牌无效", "data": null }
```

无权限：

```json
{ "code": 403, "msg": "无权限", "data": null }
```

参数 / 业务失败：

```json
{ "code": 400, "msg": "支付金额须等于应付金额 18.00", "data": null }
```

---

## License

内部项目，按团队约定使用。
