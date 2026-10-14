package router

import (
	"github.com/gin-gonic/gin"

	authctl "openbows-rms/internal/controller/auth"
	financectl "openbows-rms/internal/controller/finance"
	hrctl "openbows-rms/internal/controller/hr"
	kdsctl "openbows-rms/internal/controller/kds"
	memberctl "openbows-rms/internal/controller/member"
	menuctl "openbows-rms/internal/controller/menu"
	onlinectl "openbows-rms/internal/controller/online"
	orderctl "openbows-rms/internal/controller/order"
	paymentctl "openbows-rms/internal/controller/payment"
	purchasectl "openbows-rms/internal/controller/purchase"
	queuectl "openbows-rms/internal/controller/queue"
	reservationctl "openbows-rms/internal/controller/reservation"
	rolectl "openbows-rms/internal/controller/role"
	stockctl "openbows-rms/internal/controller/stock"
	storectl "openbows-rms/internal/controller/store"
	tablectl "openbows-rms/internal/controller/table"
	userctl "openbows-rms/internal/controller/user"
	"openbows-rms/internal/middleware"
	"openbows-rms/internal/pkg/jwtutil"
)

type Controllers struct {
	Auth        *authctl.Controller
	Store       *storectl.Controller
	User        *userctl.Controller
	Role        *rolectl.Controller
	Station     *menuctl.StationController
	Category    *menuctl.CategoryController
	Dish        *menuctl.DishController
	Area        *tablectl.AreaController
	Table       *tablectl.TableController
	Order       *orderctl.Controller
	Payment     *paymentctl.Controller
	Kds         *kdsctl.Controller
	Queue       *queuectl.Controller
	Reservation *reservationctl.Controller
	Member      *memberctl.Controller
	Online      *onlinectl.Controller
	Stock       *stockctl.Controller
	Purchase    *purchasectl.Controller
	Finance     *financectl.Controller
	HR          *hrctl.Controller
}

func Setup(jm *jwtutil.Manager, ctl Controllers) *gin.Engine {
	r := gin.New()
	r.Use(middleware.Recovery(), middleware.RequestID(), gin.Logger())

	v1 := r.Group("/apis/v1")
	{
		auth := v1.Group("/auth")
		{
			auth.POST("/login", ctl.Auth.Login)
			auth.GET("/me", middleware.JWTAuth(jm), ctl.Auth.Me)
		}

		// 顾客端线上点餐：无员工 JWT，靠 store_id+table_id / access_token
		online := v1.Group("/online")
		{
			online.POST("/orders", ctl.Online.Create)
			online.GET("/orders/:id", ctl.Online.Get)
		}

		authRequired := v1.Group("")
		authRequired.Use(middleware.JWTAuth(jm))
		{
			stores := authRequired.Group("/stores")
			{
				stores.GET("", middleware.RequirePermission("store:read", "store:write"), ctl.Store.List)
				stores.GET("/:id", middleware.RequirePermission("store:read", "store:write"), ctl.Store.Get)
				stores.POST("", middleware.RequirePermission("store:write"), ctl.Store.Create)
				stores.PUT("/:id", middleware.RequirePermission("store:write"), ctl.Store.Update)
				stores.DELETE("/:id", middleware.RequirePermission("store:write"), ctl.Store.Delete)
			}

			users := authRequired.Group("/users")
			{
				users.GET("", middleware.RequirePermission("user:read", "user:write"), ctl.User.List)
				users.GET("/:id", middleware.RequirePermission("user:read", "user:write"), ctl.User.Get)
				users.POST("", middleware.RequirePermission("user:write"), ctl.User.Create)
				users.PUT("/:id", middleware.RequirePermission("user:write"), ctl.User.Update)
				users.DELETE("/:id", middleware.RequirePermission("user:write"), ctl.User.Delete)
				users.PUT("/:id/roles", middleware.RequirePermission("user:write"), ctl.User.AssignRoles)
				users.GET("/:id/roles", middleware.RequirePermission("user:read", "user:write"), ctl.User.ListRoles)
			}

			roles := authRequired.Group("/roles")
			{
				roles.GET("", middleware.RequirePermission("role:read", "role:write"), ctl.Role.List)
				roles.GET("/:id", middleware.RequirePermission("role:read", "role:write"), ctl.Role.Get)
				roles.POST("", middleware.RequirePermission("role:write"), ctl.Role.Create)
				roles.PUT("/:id", middleware.RequirePermission("role:write"), ctl.Role.Update)
				roles.DELETE("/:id", middleware.RequirePermission("role:write"), ctl.Role.Delete)
				roles.PUT("/:id/permissions", middleware.RequirePermission("role:write"), ctl.Role.AssignPermissions)
				roles.GET("/:id/permissions", middleware.RequirePermission("role:read", "role:write"), ctl.Role.ListPermissionsOfRole)
			}

			authRequired.GET("/permissions", middleware.RequirePermission("role:read", "role:write"), ctl.Role.ListAllPermissions)

			stations := authRequired.Group("/stations")
			{
				stations.GET("", middleware.RequirePermission("menu:read", "menu:write"), ctl.Station.List)
				stations.GET("/:id", middleware.RequirePermission("menu:read", "menu:write"), ctl.Station.Get)
				stations.POST("", middleware.RequirePermission("menu:write"), ctl.Station.Create)
				stations.PUT("/:id", middleware.RequirePermission("menu:write"), ctl.Station.Update)
				stations.DELETE("/:id", middleware.RequirePermission("menu:write"), ctl.Station.Delete)
			}

			categories := authRequired.Group("/menu-categories")
			{
				categories.GET("", middleware.RequirePermission("menu:read", "menu:write"), ctl.Category.List)
				categories.GET("/:id", middleware.RequirePermission("menu:read", "menu:write"), ctl.Category.Get)
				categories.POST("", middleware.RequirePermission("menu:write"), ctl.Category.Create)
				categories.PUT("/:id", middleware.RequirePermission("menu:write"), ctl.Category.Update)
				categories.DELETE("/:id", middleware.RequirePermission("menu:write"), ctl.Category.Delete)
			}

			dishes := authRequired.Group("/dishes")
			{
				dishes.GET("", middleware.RequirePermission("menu:read", "menu:write", "order:write"), ctl.Dish.List)
				dishes.GET("/:id", middleware.RequirePermission("menu:read", "menu:write", "order:write"), ctl.Dish.Get)
				dishes.POST("", middleware.RequirePermission("menu:write"), ctl.Dish.Create)
				dishes.PUT("/:id", middleware.RequirePermission("menu:write"), ctl.Dish.Update)
				dishes.DELETE("/:id", middleware.RequirePermission("menu:write"), ctl.Dish.Delete)
				dishes.PUT("/:id/specs", middleware.RequirePermission("menu:write"), ctl.Dish.ReplaceSpecs)
				dishes.PUT("/:id/options", middleware.RequirePermission("menu:write"), ctl.Dish.ReplaceOptions)
				dishes.PUT("/:id/materials", middleware.RequirePermission("menu:write"), ctl.Dish.ReplaceMaterials)
			}

			areas := authRequired.Group("/areas")
			{
				areas.GET("", middleware.RequirePermission("table:read", "table:write", "order:write"), ctl.Area.List)
				areas.GET("/:id", middleware.RequirePermission("table:read", "table:write"), ctl.Area.Get)
				areas.POST("", middleware.RequirePermission("table:write"), ctl.Area.Create)
				areas.PUT("/:id", middleware.RequirePermission("table:write"), ctl.Area.Update)
				areas.DELETE("/:id", middleware.RequirePermission("table:write"), ctl.Area.Delete)
			}

			tables := authRequired.Group("/tables")
			{
				tables.GET("", middleware.RequirePermission("table:read", "table:write", "order:write"), ctl.Table.List)
				tables.GET("/:id", middleware.RequirePermission("table:read", "table:write", "order:write"), ctl.Table.Get)
				tables.POST("", middleware.RequirePermission("table:write"), ctl.Table.Create)
				tables.PUT("/:id", middleware.RequirePermission("table:write"), ctl.Table.Update)
				tables.PUT("/:id/status", middleware.RequirePermission("table:write", "order:write"), ctl.Table.ChangeStatus)
				tables.DELETE("/:id", middleware.RequirePermission("table:write"), ctl.Table.Delete)
			}

			orders := authRequired.Group("/orders")
			{
				orders.GET("", middleware.RequirePermission("order:read", "order:write"), ctl.Order.List)
				orders.GET("/:id", middleware.RequirePermission("order:read", "order:write"), ctl.Order.Get)
				orders.POST("", middleware.RequirePermission("order:write"), ctl.Order.Create)
				orders.POST("/:id/items", middleware.RequirePermission("order:write"), ctl.Order.AddItems)
				orders.POST("/:id/items/:item_id/return", middleware.RequirePermission("order:write"), ctl.Order.ReturnItem)
				orders.POST("/:id/advance", middleware.RequirePermission("order:write"), ctl.Order.Advance)
				orders.POST("/:id/cancel", middleware.RequirePermission("order:write"), ctl.Order.Cancel)
				orders.POST("/:id/pay", middleware.RequirePermission("payment:write"), ctl.Payment.Pay)
				orders.GET("/:id/payments", middleware.RequirePermission("payment:write", "order:read"), ctl.Payment.List)
			}

			kds := authRequired.Group("/kds")
			{
				kds.GET("/tickets", middleware.RequirePermission("kds:read", "kds:write"), ctl.Kds.List)
				kds.POST("/tickets/:id/start", middleware.RequirePermission("kds:write"), ctl.Kds.Start)
				kds.POST("/tickets/:id/done", middleware.RequirePermission("kds:write"), ctl.Kds.Done)
			}

			queues := authRequired.Group("/queues")
			{
				queues.POST("", middleware.RequirePermission("queue:write"), ctl.Queue.Take)
				queues.GET("", middleware.RequirePermission("queue:write"), ctl.Queue.List)
				queues.POST("/:id/call", middleware.RequirePermission("queue:write"), ctl.Queue.Call)
				queues.POST("/:id/seat", middleware.RequirePermission("queue:write"), ctl.Queue.Seat)
				queues.POST("/:id/cancel", middleware.RequirePermission("queue:write"), ctl.Queue.Cancel)
			}

			reservations := authRequired.Group("/reservations")
			{
				reservations.GET("", middleware.RequirePermission("reservation:write"), ctl.Reservation.List)
				reservations.GET("/:id", middleware.RequirePermission("reservation:write"), ctl.Reservation.Get)
				reservations.POST("", middleware.RequirePermission("reservation:write"), ctl.Reservation.Create)
				reservations.PUT("/:id", middleware.RequirePermission("reservation:write"), ctl.Reservation.Update)
				reservations.DELETE("/:id", middleware.RequirePermission("reservation:write"), ctl.Reservation.Delete)
				reservations.POST("/:id/confirm", middleware.RequirePermission("reservation:write"), ctl.Reservation.Confirm)
				reservations.POST("/:id/seat", middleware.RequirePermission("reservation:write"), ctl.Reservation.Seat)
				reservations.POST("/:id/cancel", middleware.RequirePermission("reservation:write"), ctl.Reservation.Cancel)
				reservations.POST("/:id/no-show", middleware.RequirePermission("reservation:write"), ctl.Reservation.NoShow)
			}

			members := authRequired.Group("/members")
			{
				members.GET("", middleware.RequirePermission("member:read", "member:write"), ctl.Member.List)
				members.GET("/:id", middleware.RequirePermission("member:read", "member:write"), ctl.Member.Get)
				members.POST("", middleware.RequirePermission("member:write"), ctl.Member.Create)
				members.PUT("/:id", middleware.RequirePermission("member:write"), ctl.Member.Update)
				members.DELETE("/:id", middleware.RequirePermission("member:write"), ctl.Member.Delete)
				members.POST("/:id/recharge", middleware.RequirePermission("member:write"), ctl.Member.Recharge)
				members.POST("/:id/consume", middleware.RequirePermission("member:write"), ctl.Member.Consume)
				members.POST("/:id/points/earn", middleware.RequirePermission("member:write"), ctl.Member.EarnPoints)
				members.POST("/:id/points/spend", middleware.RequirePermission("member:write"), ctl.Member.SpendPoints)
				members.GET("/:id/wallet-logs", middleware.RequirePermission("member:read", "member:write"), ctl.Member.ListWalletLogs)
				members.GET("/:id/point-logs", middleware.RequirePermission("member:read", "member:write"), ctl.Member.ListPointLogs)
			}

			materials := authRequired.Group("/materials")
			{
				materials.GET("", middleware.RequirePermission("stock:read", "stock:write"), ctl.Stock.ListMaterials)
				materials.GET("/:id", middleware.RequirePermission("stock:read", "stock:write"), ctl.Stock.GetMaterial)
				materials.POST("", middleware.RequirePermission("stock:write"), ctl.Stock.CreateMaterial)
				materials.PUT("/:id", middleware.RequirePermission("stock:write"), ctl.Stock.UpdateMaterial)
				materials.DELETE("/:id", middleware.RequirePermission("stock:write"), ctl.Stock.DeleteMaterial)
			}

			authRequired.GET("/stock-balances", middleware.RequirePermission("stock:read", "stock:write"), ctl.Stock.ListBalances)

			stockDocs := authRequired.Group("/stock-docs")
			{
				stockDocs.GET("", middleware.RequirePermission("stock:read", "stock:write"), ctl.Stock.ListDocs)
				stockDocs.GET("/:id", middleware.RequirePermission("stock:read", "stock:write"), ctl.Stock.GetDoc)
				stockDocs.POST("", middleware.RequirePermission("stock:write"), ctl.Stock.CreateDoc)
				stockDocs.POST("/:id/post", middleware.RequirePermission("stock:write"), ctl.Stock.PostDoc)
				stockDocs.POST("/:id/void", middleware.RequirePermission("stock:write"), ctl.Stock.VoidDoc)
			}

			suppliers := authRequired.Group("/suppliers")
			{
				suppliers.GET("", middleware.RequirePermission("purchase:read", "purchase:write"), ctl.Purchase.ListSuppliers)
				suppliers.GET("/:id", middleware.RequirePermission("purchase:read", "purchase:write"), ctl.Purchase.GetSupplier)
				suppliers.POST("", middleware.RequirePermission("purchase:write"), ctl.Purchase.CreateSupplier)
				suppliers.PUT("/:id", middleware.RequirePermission("purchase:write"), ctl.Purchase.UpdateSupplier)
				suppliers.DELETE("/:id", middleware.RequirePermission("purchase:write"), ctl.Purchase.DeleteSupplier)
			}

			purchaseOrders := authRequired.Group("/purchase-orders")
			{
				purchaseOrders.GET("", middleware.RequirePermission("purchase:read", "purchase:write"), ctl.Purchase.ListOrders)
				purchaseOrders.GET("/:id", middleware.RequirePermission("purchase:read", "purchase:write"), ctl.Purchase.GetOrder)
				purchaseOrders.POST("", middleware.RequirePermission("purchase:write"), ctl.Purchase.CreateOrder)
				purchaseOrders.POST("/:id/submit", middleware.RequirePermission("purchase:write"), ctl.Purchase.SubmitOrder)
				purchaseOrders.POST("/:id/cancel", middleware.RequirePermission("purchase:write"), ctl.Purchase.CancelOrder)
			}

			purchaseReceipts := authRequired.Group("/purchase-receipts")
			{
				purchaseReceipts.GET("", middleware.RequirePermission("purchase:read", "purchase:write"), ctl.Purchase.ListReceipts)
				purchaseReceipts.GET("/:id", middleware.RequirePermission("purchase:read", "purchase:write"), ctl.Purchase.GetReceipt)
				purchaseReceipts.POST("", middleware.RequirePermission("purchase:write"), ctl.Purchase.CreateReceipt)
				purchaseReceipts.POST("/:id/post", middleware.RequirePermission("purchase:write"), ctl.Purchase.PostReceipt)
			}

			settlements := authRequired.Group("/settlements")
			{
				settlements.POST("/daily", middleware.RequirePermission("finance:write"), ctl.Finance.DailySettle)
				settlements.GET("", middleware.RequirePermission("finance:read", "finance:write"), ctl.Finance.ListSettlements)
				settlements.GET("/:id", middleware.RequirePermission("finance:read", "finance:write"), ctl.Finance.GetSettlement)
			}

			reconciliations := authRequired.Group("/reconciliations")
			{
				reconciliations.POST("", middleware.RequirePermission("finance:write"), ctl.Finance.CreateReconciliation)
				reconciliations.GET("", middleware.RequirePermission("finance:read", "finance:write"), ctl.Finance.ListReconciliations)
				reconciliations.GET("/diffs", middleware.RequirePermission("finance:read", "finance:write"), ctl.Finance.ListDiffs)
				reconciliations.GET("/:id", middleware.RequirePermission("finance:read", "finance:write"), ctl.Finance.GetReconciliation)
			}

			employees := authRequired.Group("/employees")
			{
				employees.GET("", middleware.RequirePermission("hr:employee:read", "hr:employee:write"), ctl.HR.ListEmployees)
				employees.GET("/:id", middleware.RequirePermission("hr:employee:read", "hr:employee:write"), ctl.HR.GetEmployee)
				employees.POST("", middleware.RequirePermission("hr:employee:write"), ctl.HR.CreateEmployee)
				employees.PUT("/:id", middleware.RequirePermission("hr:employee:write"), ctl.HR.UpdateEmployee)
				employees.DELETE("/:id", middleware.RequirePermission("hr:employee:write"), ctl.HR.DeleteEmployee)
			}

			shifts := authRequired.Group("/shifts")
			{
				shifts.GET("", middleware.RequirePermission("hr:shift:write", "hr:employee:read"), ctl.HR.ListShifts)
				shifts.GET("/:id", middleware.RequirePermission("hr:shift:write"), ctl.HR.GetShift)
				shifts.POST("", middleware.RequirePermission("hr:shift:write"), ctl.HR.CreateShift)
				shifts.PUT("/:id", middleware.RequirePermission("hr:shift:write"), ctl.HR.UpdateShift)
				shifts.POST("/:id/cancel", middleware.RequirePermission("hr:shift:write"), ctl.HR.CancelShift)
				shifts.DELETE("/:id", middleware.RequirePermission("hr:shift:write"), ctl.HR.DeleteShift)
			}

			attendances := authRequired.Group("/attendances")
			{
				attendances.GET("", middleware.RequirePermission("hr:attendance:write", "hr:employee:read"), ctl.HR.ListAttendances)
				attendances.GET("/:id", middleware.RequirePermission("hr:attendance:write"), ctl.HR.GetAttendance)
				attendances.POST("/punch", middleware.RequirePermission("hr:attendance:write"), ctl.HR.Punch)
				attendances.POST("/makeup", middleware.RequirePermission("hr:attendance:write"), ctl.HR.Makeup)
			}
		}
	}
	return r
}
