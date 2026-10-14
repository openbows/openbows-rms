package model

import (
	"time"

	"gorm.io/gorm"
)

const (
	ShiftStatusScheduled = "scheduled"
	ShiftStatusCancelled = "cancelled"

	AttendanceSourcePunch  = "punch"
	AttendanceSourceMakeup = "makeup"

	AttendanceStatusNormal = "normal"
	AttendanceStatusMakeup = "makeup"
)

type Employee struct {
	ID        int64          `gorm:"column:id;primaryKey;autoIncrement;comment:主键" json:"id"`
	StoreID   int64          `gorm:"column:store_id;index;not null;comment:门店ID" json:"store_id"`
	UserID    *int64         `gorm:"column:user_id;index;comment:关联登录用户ID" json:"user_id"`
	EmpNo     string         `gorm:"column:emp_no;type:varchar(64);not null;comment:工号" json:"emp_no"`
	Name      string         `gorm:"column:name;type:varchar(128);not null;comment:姓名" json:"name"`
	Phone     string         `gorm:"column:phone;type:varchar(32);not null;default:'';comment:手机号" json:"phone"`
	JobTitle  string         `gorm:"column:job_title;type:varchar(64);not null;default:'';comment:岗位" json:"job_title"`
	Status    bool           `gorm:"column:status;type:boolean;not null;default:true;comment:是否启用" json:"status"`
	HiredAt   *time.Time     `gorm:"column:hired_at;type:date;comment:入职日期" json:"hired_at"`
	Remark    string         `gorm:"column:remark;type:varchar(512);not null;default:'';comment:备注" json:"remark"`
	CreatedAt time.Time      `gorm:"column:created_at;autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at;autoUpdateTime;comment:更新时间" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index;comment:删除时间" json:"-"`
}

func (Employee) TableName() string { return "employees" }

type Shift struct {
	ID         int64          `gorm:"column:id;primaryKey;autoIncrement;comment:主键" json:"id"`
	StoreID    int64          `gorm:"column:store_id;index;not null;comment:门店ID" json:"store_id"`
	EmployeeID int64          `gorm:"column:employee_id;index;not null;comment:员工ID" json:"employee_id"`
	ShiftDate  time.Time      `gorm:"column:shift_date;type:date;index;not null;comment:排班日期" json:"shift_date"`
	StartTime  string         `gorm:"column:start_time;type:time;not null;comment:开始时间" json:"start_time"`
	EndTime    string         `gorm:"column:end_time;type:time;not null;comment:结束时间" json:"end_time"`
	ShiftName  string         `gorm:"column:shift_name;type:varchar(64);not null;default:'';comment:班次名称" json:"shift_name"`
	Status     string         `gorm:"column:status;type:varchar(32);not null;default:scheduled;comment:状态" json:"status"`
	Remark     string         `gorm:"column:remark;type:varchar(255);not null;default:'';comment:备注" json:"remark"`
	CreatedBy  int64          `gorm:"column:created_by;not null;default:0;comment:创建人" json:"created_by"`
	CreatedAt  time.Time      `gorm:"column:created_at;autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt  time.Time      `gorm:"column:updated_at;autoUpdateTime;comment:更新时间" json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"column:deleted_at;index;comment:删除时间" json:"-"`
}

func (Shift) TableName() string { return "shifts" }

type Attendance struct {
	ID         int64          `gorm:"column:id;primaryKey;autoIncrement;comment:主键" json:"id"`
	StoreID    int64          `gorm:"column:store_id;index;not null;comment:门店ID" json:"store_id"`
	EmployeeID int64          `gorm:"column:employee_id;index;not null;comment:员工ID" json:"employee_id"`
	WorkDate   time.Time      `gorm:"column:work_date;type:date;index;not null;comment:出勤日" json:"work_date"`
	ClockInAt  *time.Time     `gorm:"column:clock_in_at;comment:上班打卡时间" json:"clock_in_at"`
	ClockOutAt *time.Time     `gorm:"column:clock_out_at;comment:下班打卡时间" json:"clock_out_at"`
	Source     string         `gorm:"column:source;type:varchar(32);not null;default:punch;comment:来源" json:"source"`
	Status     string         `gorm:"column:status;type:varchar(32);not null;default:normal;comment:状态" json:"status"`
	Remark     string         `gorm:"column:remark;type:varchar(255);not null;default:'';comment:备注" json:"remark"`
	CreatedBy  int64          `gorm:"column:created_by;not null;default:0;comment:创建人" json:"created_by"`
	UpdatedBy  int64          `gorm:"column:updated_by;not null;default:0;comment:更新人" json:"updated_by"`
	CreatedAt  time.Time      `gorm:"column:created_at;autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt  time.Time      `gorm:"column:updated_at;autoUpdateTime;comment:更新时间" json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"column:deleted_at;index;comment:删除时间" json:"-"`
}

func (Attendance) TableName() string { return "attendances" }
