package hr

import (
	"context"
	"fmt"
	"strings"
	"time"

	"openbows-rms/internal/model"
	hrrepo "openbows-rms/internal/repository/hr"
)

type Service struct {
	repo *hrrepo.Repository
}

func New(repo *hrrepo.Repository) *Service {
	return &Service{repo: repo}
}

type Operator struct {
	UserID int64
}

// ---------- employees ----------

type EmployeeCreateInput struct {
	UserID   *int64 `json:"user_id"`
	EmpNo    string `json:"emp_no" binding:"required,max=64"`
	Name     string `json:"name" binding:"required,max=128"`
	Phone    string `json:"phone" binding:"max=32"`
	JobTitle string `json:"job_title" binding:"max=64"`
	Status   *bool  `json:"status"`
	HiredAt  string `json:"hired_at"` // YYYY-MM-DD
	Remark   string `json:"remark" binding:"max=512"`
}

type EmployeeUpdateInput struct {
	UserID   *int64 `json:"user_id"`
	EmpNo    string `json:"emp_no" binding:"required,max=64"`
	Name     string `json:"name" binding:"required,max=128"`
	Phone    string `json:"phone" binding:"max=32"`
	JobTitle string `json:"job_title" binding:"max=64"`
	Status   *bool  `json:"status"`
	HiredAt  string `json:"hired_at"`
	Remark   string `json:"remark" binding:"max=512"`
}

func (s *Service) CreateEmployee(ctx context.Context, storeID int64, in EmployeeCreateInput) (*model.Employee, error) {
	empNo := strings.TrimSpace(in.EmpNo)
	name := strings.TrimSpace(in.Name)
	if empNo == "" || name == "" {
		return nil, fmt.Errorf("工号和姓名不能为空")
	}
	if _, err := s.repo.GetEmployeeByEmpNo(ctx, storeID, empNo); err == nil {
		return nil, fmt.Errorf("工号已存在")
	} else if !hrrepo.IsNotFound(err) {
		return nil, err
	}
	hiredAt, err := parseDateOptional(in.HiredAt)
	if err != nil {
		return nil, err
	}
	status := true
	if in.Status != nil {
		status = *in.Status
	}
	m := &model.Employee{
		StoreID: storeID, UserID: in.UserID, EmpNo: empNo, Name: name,
		Phone: strings.TrimSpace(in.Phone), JobTitle: strings.TrimSpace(in.JobTitle),
		Status: status, HiredAt: hiredAt, Remark: in.Remark,
	}
	if err := s.repo.CreateEmployee(ctx, m); err != nil {
		return nil, fmt.Errorf("创建员工失败: %w", err)
	}
	return m, nil
}

func (s *Service) UpdateEmployee(ctx context.Context, storeID, id int64, in EmployeeUpdateInput) (*model.Employee, error) {
	m, err := s.getOwnedEmployee(ctx, storeID, id)
	if err != nil {
		return nil, err
	}
	empNo := strings.TrimSpace(in.EmpNo)
	name := strings.TrimSpace(in.Name)
	if empNo == "" || name == "" {
		return nil, fmt.Errorf("工号和姓名不能为空")
	}
	if empNo != m.EmpNo {
		if other, err := s.repo.GetEmployeeByEmpNo(ctx, storeID, empNo); err == nil && other.ID != id {
			return nil, fmt.Errorf("工号已存在")
		} else if err != nil && !hrrepo.IsNotFound(err) {
			return nil, err
		}
	}
	hiredAt, err := parseDateOptional(in.HiredAt)
	if err != nil {
		return nil, err
	}
	m.UserID = in.UserID
	m.EmpNo = empNo
	m.Name = name
	m.Phone = strings.TrimSpace(in.Phone)
	m.JobTitle = strings.TrimSpace(in.JobTitle)
	if in.Status != nil {
		m.Status = *in.Status
	}
	if in.HiredAt != "" {
		m.HiredAt = hiredAt
	}
	m.Remark = in.Remark
	if err := s.repo.SaveEmployee(ctx, m); err != nil {
		return nil, fmt.Errorf("更新员工失败: %w", err)
	}
	return m, nil
}

func (s *Service) DeleteEmployee(ctx context.Context, storeID, id int64) error {
	if _, err := s.getOwnedEmployee(ctx, storeID, id); err != nil {
		return err
	}
	return s.repo.SoftDeleteEmployee(ctx, id)
}

func (s *Service) GetEmployee(ctx context.Context, storeID, id int64) (*model.Employee, error) {
	return s.getOwnedEmployee(ctx, storeID, id)
}

func (s *Service) ListEmployees(ctx context.Context, storeID int64, keyword string, status *bool, page, pageSize int) ([]model.Employee, int64, error) {
	return s.repo.ListEmployees(ctx, hrrepo.EmployeeListFilter{
		StoreID: storeID, Keyword: keyword, Status: status, Page: page, PageSize: pageSize,
	})
}

// ---------- shifts ----------

type ShiftCreateInput struct {
	EmployeeID int64  `json:"employee_id" binding:"required"`
	ShiftDate  string `json:"shift_date" binding:"required"` // YYYY-MM-DD
	StartTime  string `json:"start_time" binding:"required"` // HH:MM or HH:MM:SS
	EndTime    string `json:"end_time" binding:"required"`
	ShiftName  string `json:"shift_name" binding:"max=64"`
	Remark     string `json:"remark" binding:"max=255"`
}

type ShiftUpdateInput struct {
	EmployeeID int64  `json:"employee_id" binding:"required"`
	ShiftDate  string `json:"shift_date" binding:"required"`
	StartTime  string `json:"start_time" binding:"required"`
	EndTime    string `json:"end_time" binding:"required"`
	ShiftName  string `json:"shift_name" binding:"max=64"`
	Remark     string `json:"remark" binding:"max=255"`
}

func (s *Service) CreateShift(ctx context.Context, storeID int64, op Operator, in ShiftCreateInput) (*model.Shift, error) {
	if _, err := s.getOwnedEmployee(ctx, storeID, in.EmployeeID); err != nil {
		return nil, err
	}
	day, err := parseDateRequired(in.ShiftDate, "排班日期")
	if err != nil {
		return nil, err
	}
	start, err := normalizeTime(in.StartTime)
	if err != nil {
		return nil, fmt.Errorf("开始时间格式错误")
	}
	end, err := normalizeTime(in.EndTime)
	if err != nil {
		return nil, fmt.Errorf("结束时间格式错误")
	}
	m := &model.Shift{
		StoreID: storeID, EmployeeID: in.EmployeeID, ShiftDate: day,
		StartTime: start, EndTime: end, ShiftName: strings.TrimSpace(in.ShiftName),
		Status: model.ShiftStatusScheduled, Remark: in.Remark, CreatedBy: op.UserID,
	}
	if err := s.repo.CreateShift(ctx, m); err != nil {
		return nil, fmt.Errorf("创建排班失败: %w", err)
	}
	return m, nil
}

func (s *Service) UpdateShift(ctx context.Context, storeID, id int64, in ShiftUpdateInput) (*model.Shift, error) {
	m, err := s.getOwnedShift(ctx, storeID, id)
	if err != nil {
		return nil, err
	}
	if m.Status == model.ShiftStatusCancelled {
		return nil, fmt.Errorf("已取消排班不可修改")
	}
	if _, err := s.getOwnedEmployee(ctx, storeID, in.EmployeeID); err != nil {
		return nil, err
	}
	day, err := parseDateRequired(in.ShiftDate, "排班日期")
	if err != nil {
		return nil, err
	}
	start, err := normalizeTime(in.StartTime)
	if err != nil {
		return nil, fmt.Errorf("开始时间格式错误")
	}
	end, err := normalizeTime(in.EndTime)
	if err != nil {
		return nil, fmt.Errorf("结束时间格式错误")
	}
	m.EmployeeID = in.EmployeeID
	m.ShiftDate = day
	m.StartTime = start
	m.EndTime = end
	m.ShiftName = strings.TrimSpace(in.ShiftName)
	m.Remark = in.Remark
	if err := s.repo.SaveShift(ctx, m); err != nil {
		return nil, fmt.Errorf("更新排班失败: %w", err)
	}
	return m, nil
}

func (s *Service) CancelShift(ctx context.Context, storeID, id int64) (*model.Shift, error) {
	m, err := s.getOwnedShift(ctx, storeID, id)
	if err != nil {
		return nil, err
	}
	if m.Status == model.ShiftStatusCancelled {
		return m, nil
	}
	m.Status = model.ShiftStatusCancelled
	if err := s.repo.SaveShift(ctx, m); err != nil {
		return nil, fmt.Errorf("取消排班失败: %w", err)
	}
	return m, nil
}

func (s *Service) DeleteShift(ctx context.Context, storeID, id int64) error {
	if _, err := s.getOwnedShift(ctx, storeID, id); err != nil {
		return err
	}
	return s.repo.SoftDeleteShift(ctx, id)
}

func (s *Service) GetShift(ctx context.Context, storeID, id int64) (*model.Shift, error) {
	return s.getOwnedShift(ctx, storeID, id)
}

func (s *Service) ListShifts(ctx context.Context, storeID, employeeID int64, shiftDate, status string, page, pageSize int) ([]model.Shift, int64, error) {
	return s.repo.ListShifts(ctx, hrrepo.ShiftListFilter{
		StoreID: storeID, EmployeeID: employeeID, ShiftDate: shiftDate, Status: status,
		Page: page, PageSize: pageSize,
	})
}

// ---------- attendances ----------

type PunchInput struct {
	EmployeeID int64  `json:"employee_id" binding:"required"`
	WorkDate   string `json:"work_date"` // 可选，默认当天
	Type       string `json:"type" binding:"required"` // in / out
	Remark     string `json:"remark" binding:"max=255"`
}

type MakeupInput struct {
	EmployeeID int64  `json:"employee_id" binding:"required"`
	WorkDate   string `json:"work_date" binding:"required"`
	ClockInAt  string `json:"clock_in_at"`  // RFC3339 或 2006-01-02 15:04:05
	ClockOutAt string `json:"clock_out_at"`
	Remark     string `json:"remark" binding:"max=255"`
}

func (s *Service) Punch(ctx context.Context, storeID int64, op Operator, in PunchInput) (*model.Attendance, error) {
	emp, err := s.getOwnedEmployee(ctx, storeID, in.EmployeeID)
	if err != nil {
		return nil, err
	}
	if !emp.Status {
		return nil, fmt.Errorf("员工已停用")
	}
	punchType := strings.ToLower(strings.TrimSpace(in.Type))
	if punchType != "in" && punchType != "out" {
		return nil, fmt.Errorf("打卡类型须为 in 或 out")
	}
	workDate, err := parseDateOptional(in.WorkDate)
	if err != nil {
		return nil, err
	}
	if workDate == nil {
		now := time.Now()
		d := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
		workDate = &d
	}
	now := time.Now()
	att, err := s.repo.GetAttendanceByEmpDate(ctx, storeID, in.EmployeeID, *workDate)
	if err != nil && !hrrepo.IsNotFound(err) {
		return nil, err
	}
	if att == nil {
		att = &model.Attendance{
			StoreID: storeID, EmployeeID: in.EmployeeID, WorkDate: *workDate,
			Source: model.AttendanceSourcePunch, Status: model.AttendanceStatusNormal,
			Remark: in.Remark, CreatedBy: op.UserID, UpdatedBy: op.UserID,
		}
		if punchType == "in" {
			att.ClockInAt = &now
		} else {
			att.ClockOutAt = &now
		}
		if err := s.repo.CreateAttendance(ctx, att); err != nil {
			return nil, fmt.Errorf("打卡失败: %w", err)
		}
		return att, nil
	}
	if punchType == "in" {
		if att.ClockInAt != nil {
			return nil, fmt.Errorf("今日已上班打卡")
		}
		att.ClockInAt = &now
	} else {
		if att.ClockOutAt != nil {
			return nil, fmt.Errorf("今日已下班打卡")
		}
		att.ClockOutAt = &now
	}
	if in.Remark != "" {
		att.Remark = in.Remark
	}
	att.UpdatedBy = op.UserID
	if err := s.repo.SaveAttendance(ctx, att); err != nil {
		return nil, fmt.Errorf("打卡失败: %w", err)
	}
	return att, nil
}

func (s *Service) Makeup(ctx context.Context, storeID int64, op Operator, in MakeupInput) (*model.Attendance, error) {
	if _, err := s.getOwnedEmployee(ctx, storeID, in.EmployeeID); err != nil {
		return nil, err
	}
	day, err := parseDateRequired(in.WorkDate, "出勤日")
	if err != nil {
		return nil, err
	}
	if in.ClockInAt == "" && in.ClockOutAt == "" {
		return nil, fmt.Errorf("补卡须提供上班或下班时间")
	}
	var clockIn, clockOut *time.Time
	if in.ClockInAt != "" {
		t, err := parseDateTime(in.ClockInAt)
		if err != nil {
			return nil, fmt.Errorf("上班时间格式错误")
		}
		clockIn = &t
	}
	if in.ClockOutAt != "" {
		t, err := parseDateTime(in.ClockOutAt)
		if err != nil {
			return nil, fmt.Errorf("下班时间格式错误")
		}
		clockOut = &t
	}
	att, err := s.repo.GetAttendanceByEmpDate(ctx, storeID, in.EmployeeID, day)
	if err != nil && !hrrepo.IsNotFound(err) {
		return nil, err
	}
	if att == nil {
		att = &model.Attendance{
			StoreID: storeID, EmployeeID: in.EmployeeID, WorkDate: day,
			ClockInAt: clockIn, ClockOutAt: clockOut,
			Source: model.AttendanceSourceMakeup, Status: model.AttendanceStatusMakeup,
			Remark: in.Remark, CreatedBy: op.UserID, UpdatedBy: op.UserID,
		}
		if err := s.repo.CreateAttendance(ctx, att); err != nil {
			return nil, fmt.Errorf("补卡失败: %w", err)
		}
		return att, nil
	}
	if clockIn != nil {
		att.ClockInAt = clockIn
	}
	if clockOut != nil {
		att.ClockOutAt = clockOut
	}
	att.Source = model.AttendanceSourceMakeup
	att.Status = model.AttendanceStatusMakeup
	if in.Remark != "" {
		att.Remark = in.Remark
	}
	att.UpdatedBy = op.UserID
	if err := s.repo.SaveAttendance(ctx, att); err != nil {
		return nil, fmt.Errorf("补卡失败: %w", err)
	}
	return att, nil
}

func (s *Service) ListAttendances(ctx context.Context, storeID, employeeID int64, workDate string, page, pageSize int) ([]model.Attendance, int64, error) {
	return s.repo.ListAttendances(ctx, hrrepo.AttendanceListFilter{
		StoreID: storeID, EmployeeID: employeeID, WorkDate: workDate, Page: page, PageSize: pageSize,
	})
}

func (s *Service) GetAttendance(ctx context.Context, storeID, id int64) (*model.Attendance, error) {
	m, err := s.repo.GetAttendanceByID(ctx, id)
	if err != nil {
		if hrrepo.IsNotFound(err) {
			return nil, fmt.Errorf("考勤不存在")
		}
		return nil, err
	}
	if m.StoreID != storeID {
		return nil, fmt.Errorf("无权操作其他门店考勤")
	}
	return m, nil
}

// ---------- helpers ----------

func (s *Service) getOwnedEmployee(ctx context.Context, storeID, id int64) (*model.Employee, error) {
	m, err := s.repo.GetEmployeeByID(ctx, id)
	if err != nil {
		if hrrepo.IsNotFound(err) {
			return nil, fmt.Errorf("员工不存在")
		}
		return nil, err
	}
	if m.StoreID != storeID {
		return nil, fmt.Errorf("无权操作其他门店员工")
	}
	return m, nil
}

func (s *Service) getOwnedShift(ctx context.Context, storeID, id int64) (*model.Shift, error) {
	m, err := s.repo.GetShiftByID(ctx, id)
	if err != nil {
		if hrrepo.IsNotFound(err) {
			return nil, fmt.Errorf("排班不存在")
		}
		return nil, err
	}
	if m.StoreID != storeID {
		return nil, fmt.Errorf("无权操作其他门店排班")
	}
	return m, nil
}

func parseDateOptional(s string) (*time.Time, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, nil
	}
	t, err := time.ParseInLocation("2006-01-02", s, time.Local)
	if err != nil {
		return nil, fmt.Errorf("日期格式须为 YYYY-MM-DD")
	}
	return &t, nil
}

func parseDateRequired(s, label string) (time.Time, error) {
	t, err := parseDateOptional(s)
	if err != nil {
		return time.Time{}, err
	}
	if t == nil {
		return time.Time{}, fmt.Errorf("%s不能为空", label)
	}
	return *t, nil
}

func normalizeTime(s string) (string, error) {
	s = strings.TrimSpace(s)
	if t, err := time.Parse("15:04:05", s); err == nil {
		return t.Format("15:04:05"), nil
	}
	if t, err := time.Parse("15:04", s); err == nil {
		return t.Format("15:04:05"), nil
	}
	return "", fmt.Errorf("invalid time")
}

func parseDateTime(s string) (time.Time, error) {
	s = strings.TrimSpace(s)
	layouts := []string{
		time.RFC3339,
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
		"2006-01-02 15:04",
	}
	for _, layout := range layouts {
		if t, err := time.ParseInLocation(layout, s, time.Local); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("invalid datetime")
}
