package main

import (
	"flag"
	"log"

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
	"openbows-rms/internal/config"
	"openbows-rms/internal/pkg/jwtutil"
	"openbows-rms/internal/repository"
	financerepo "openbows-rms/internal/repository/finance"
	hrrepo "openbows-rms/internal/repository/hr"
	kdsrepo "openbows-rms/internal/repository/kds"
	memberrepo "openbows-rms/internal/repository/member"
	menurepo "openbows-rms/internal/repository/menu"
	onlinerepo "openbows-rms/internal/repository/online"
	orderrepo "openbows-rms/internal/repository/order"
	paymentrepo "openbows-rms/internal/repository/payment"
	purchaserepo "openbows-rms/internal/repository/purchase"
	queuerepo "openbows-rms/internal/repository/queue"
	resrepo "openbows-rms/internal/repository/reservation"
	rolerepo "openbows-rms/internal/repository/role"
	stockrepo "openbows-rms/internal/repository/stock"
	storerepo "openbows-rms/internal/repository/store"
	tablerepo "openbows-rms/internal/repository/table"
	userrepo "openbows-rms/internal/repository/user"
	"openbows-rms/internal/router"
	authsvc "openbows-rms/internal/service/auth"
	financesvc "openbows-rms/internal/service/finance"
	hrsvc "openbows-rms/internal/service/hr"
	kdssvc "openbows-rms/internal/service/kds"
	membersvc "openbows-rms/internal/service/member"
	menusvc "openbows-rms/internal/service/menu"
	onlinesvc "openbows-rms/internal/service/online"
	ordersvc "openbows-rms/internal/service/order"
	paymentsvc "openbows-rms/internal/service/payment"
	purchasesvc "openbows-rms/internal/service/purchase"
	queuesvc "openbows-rms/internal/service/queue"
	ressvc "openbows-rms/internal/service/reservation"
	rolesvc "openbows-rms/internal/service/role"
	stocksvc "openbows-rms/internal/service/stock"
	storesvc "openbows-rms/internal/service/store"
	tablesvc "openbows-rms/internal/service/table"
	usersvc "openbows-rms/internal/service/user"
)

func main() {
	cfgPath := flag.String("config", "configs/config.yaml", "config file path")
	flag.Parse()

	cfg, err := config.Load(*cfgPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	db, err := repository.NewDB(cfg.Database.DSN)
	if err != nil {
		log.Fatalf("db: %v", err)
	}

	jm := jwtutil.NewManager(cfg.JWT.Secret, cfg.JWT.ExpireHours)

	storeRepo := storerepo.New(db)
	userRepo := userrepo.New(db)
	roleRepo := rolerepo.New(db)
	stationRepo := menurepo.NewStationRepo(db)
	categoryRepo := menurepo.NewCategoryRepo(db)
	dishRepo := menurepo.NewDishRepo(db)
	areaRepo := tablerepo.NewAreaRepo(db)
	diningTableRepo := tablerepo.NewTableRepo(db)
	orderRepo := orderrepo.New(db)
	paymentRepo := paymentrepo.New(db)
	kdsRepo := kdsrepo.New(db)
	queueRepo := queuerepo.New(db)
	reservationRepo := resrepo.New(db)
	memberRepo := memberrepo.New(db)
	onlineRepo := onlinerepo.New(db)
	stockRepo := stockrepo.New(db)
	purchaseRepo := purchaserepo.New(db)
	financeRepo := financerepo.New(db)
	hrRepo := hrrepo.New(db)

	authService := authsvc.New(userRepo, jm)
	storeService := storesvc.New(storeRepo)
	userService := usersvc.New(userRepo, roleRepo)
	roleService := rolesvc.New(roleRepo)
	stationService := menusvc.NewStationService(stationRepo)
	categoryService := menusvc.NewCategoryService(categoryRepo)
	dishService := menusvc.NewDishService(dishRepo, categoryRepo, stationRepo, stockRepo)
	areaService := tablesvc.NewAreaService(areaRepo)
	diningTableService := tablesvc.NewTableService(diningTableRepo, areaRepo)
	orderService := ordersvc.New(orderRepo, diningTableRepo, dishRepo)
	stockService := stocksvc.New(stockRepo)
	memberService := membersvc.New(memberRepo)
	paymentService := paymentsvc.New(paymentRepo, orderRepo, diningTableRepo, dishRepo, stockService, memberService)
	kdsService := kdssvc.New(kdsRepo, dishRepo)
	orderService.SetKitchenSyncer(kdsService)
	queueService := queuesvc.New(queueRepo, diningTableRepo, orderRepo)
	reservationService := ressvc.New(reservationRepo, diningTableRepo, orderRepo)
	onlineService := onlinesvc.New(orderService, onlineRepo, memberRepo)
	purchaseService := purchasesvc.New(purchaseRepo, stockService, stockRepo)
	financeService := financesvc.New(financeRepo, paymentRepo)
	hrService := hrsvc.New(hrRepo)

	engine := router.Setup(jm, router.Controllers{
		Auth:        authctl.New(authService),
		Store:       storectl.New(storeService),
		User:        userctl.New(userService),
		Role:        rolectl.New(roleService),
		Station:     menuctl.NewStationController(stationService),
		Category:    menuctl.NewCategoryController(categoryService),
		Dish:        menuctl.NewDishController(dishService),
		Area:        tablectl.NewAreaController(areaService),
		Table:       tablectl.NewTableController(diningTableService),
		Order:       orderctl.New(orderService),
		Payment:     paymentctl.New(paymentService),
		Kds:         kdsctl.New(kdsService),
		Queue:       queuectl.New(queueService),
		Reservation: reservationctl.New(reservationService),
		Member:      memberctl.New(memberService),
		Online:      onlinectl.New(onlineService),
		Stock:       stockctl.New(stockService),
		Purchase:    purchasectl.New(purchaseService),
		Finance:     financectl.New(financeService),
		HR:          hrctl.New(hrService),
	})
	// 同源托管管理后台 / 收银 / 厨显（需先 npm run build）
	router.MountFrontend(engine, "web/apps")

	log.Printf("server listening on %s", cfg.Server.Addr)
	if err := engine.Run(cfg.Server.Addr); err != nil {
		log.Fatalf("run: %v", err)
	}
}
