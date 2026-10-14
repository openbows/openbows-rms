package hr

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"openbows-rms/internal/model"
)

type Repository struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) DB() *gorm.DB { return r.db }

func IsNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}

// ---------- employees ----------

func (r *Repository) CreateEmployee(ctx context.Context, m *model.Employee) error {
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *Repository) SaveEmployee(ctx context.Context, m *model.Employee) error {
	return r.db.WithContext(ctx).Save(m).Error
}

func (r *Repository) SoftDeleteEmployee(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&model.Employee{}, id).Error
}

func (r *Repository) GetEmployeeByID(ctx context.Context, id int64) (*model.Employee, error) {
	var m model.Employee
	err := r.db.WithContext(ctx).First(&m, id).Error
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *Repository) GetEmployeeByEmpNo(ctx context.Context, storeID int64, empNo string) (*model.Employee, error) {
	var m model.Employee
	err := r.db.WithContext(ctx).Where("store_id = ? AND emp_no = ?", storeID, empNo).First(&m).Error
	if err != nil {
		return nil, err
	}
	return &m, nil
}

type EmployeeListFilter struct {
	StoreID  int64
	Keyword  string
	Status   *bool
	Page     int
	PageSize int
}

func (r *Repository) ListEmployees(ctx context.Context, f EmployeeListFilter) ([]model.Employee, int64, error) {
	q := r.db.WithContext(ctx).Model(&model.Employee{}).Where("store_id = ?", f.StoreID)
	if f.Keyword != "" {
		like := "%" + f.Keyword + "%"
		q = q.Where("emp_no ILIKE ? OR name ILIKE ? OR phone ILIKE ?", like, like, like)
	}
	if f.Status != nil {
		q = q.Where("status = ?", *f.Status)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if f.Page < 1 {
		f.Page = 1
	}
	if f.PageSize < 1 {
		f.PageSize = 10
	}
	var list []model.Employee
	err := q.Order("id DESC").Offset((f.Page - 1) * f.PageSize).Limit(f.PageSize).Find(&list).Error
	return list, total, err
}

// ---------- shifts ----------

func (r *Repository) CreateShift(ctx context.Context, m *model.Shift) error {
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *Repository) SaveShift(ctx context.Context, m *model.Shift) error {
	return r.db.WithContext(ctx).Save(m).Error
}

func (r *Repository) SoftDeleteShift(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&model.Shift{}, id).Error
}

func (r *Repository) GetShiftByID(ctx context.Context, id int64) (*model.Shift, error) {
	var m model.Shift
	err := r.db.WithContext(ctx).First(&m, id).Error
	if err != nil {
		return nil, err
	}
	return &m, nil
}

type ShiftListFilter struct {
	StoreID    int64
	EmployeeID int64
	ShiftDate  string // YYYY-MM-DD
	Status     string
	Page       int
	PageSize   int
}

func (r *Repository) ListShifts(ctx context.Context, f ShiftListFilter) ([]model.Shift, int64, error) {
	q := r.db.WithContext(ctx).Model(&model.Shift{}).Where("store_id = ?", f.StoreID)
	if f.EmployeeID > 0 {
		q = q.Where("employee_id = ?", f.EmployeeID)
	}
	if f.ShiftDate != "" {
		q = q.Where("shift_date = ?", f.ShiftDate)
	}
	if f.Status != "" {
		q = q.Where("status = ?", f.Status)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if f.Page < 1 {
		f.Page = 1
	}
	if f.PageSize < 1 {
		f.PageSize = 10
	}
	var list []model.Shift
	err := q.Order("shift_date DESC, id DESC").Offset((f.Page - 1) * f.PageSize).Limit(f.PageSize).Find(&list).Error
	return list, total, err
}

// ---------- attendances ----------

func (r *Repository) CreateAttendance(ctx context.Context, m *model.Attendance) error {
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *Repository) SaveAttendance(ctx context.Context, m *model.Attendance) error {
	return r.db.WithContext(ctx).Save(m).Error
}

func (r *Repository) GetAttendanceByID(ctx context.Context, id int64) (*model.Attendance, error) {
	var m model.Attendance
	err := r.db.WithContext(ctx).First(&m, id).Error
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *Repository) GetAttendanceByEmpDate(ctx context.Context, storeID, employeeID int64, workDate time.Time) (*model.Attendance, error) {
	var m model.Attendance
	err := r.db.WithContext(ctx).
		Where("store_id = ? AND employee_id = ? AND work_date = ?", storeID, employeeID, workDate.Format("2006-01-02")).
		First(&m).Error
	if err != nil {
		return nil, err
	}
	return &m, nil
}

type AttendanceListFilter struct {
	StoreID    int64
	EmployeeID int64
	WorkDate   string // YYYY-MM-DD
	Page       int
	PageSize   int
}

func (r *Repository) ListAttendances(ctx context.Context, f AttendanceListFilter) ([]model.Attendance, int64, error) {
	q := r.db.WithContext(ctx).Model(&model.Attendance{}).Where("store_id = ?", f.StoreID)
	if f.EmployeeID > 0 {
		q = q.Where("employee_id = ?", f.EmployeeID)
	}
	if f.WorkDate != "" {
		q = q.Where("work_date = ?", f.WorkDate)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if f.Page < 1 {
		f.Page = 1
	}
	if f.PageSize < 1 {
		f.PageSize = 10
	}
	var list []model.Attendance
	err := q.Order("work_date DESC, id DESC").Offset((f.Page - 1) * f.PageSize).Limit(f.PageSize).Find(&list).Error
	return list, total, err
}
