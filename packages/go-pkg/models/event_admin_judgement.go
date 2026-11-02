package models

type EventAdminJudgement struct {
	BaseModel

	EventID    int64 `gorm:"index:,unique,composite:idx_event_user_jdgmt"`
	UserID     int64 `gorm:"index:,unique,composite:idx_event_user_jdgmt"`
	IsApproval bool
}
