package domain

type ContainerDetail struct {
	ID                 string `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
	InspectionReportID string `gorm:"type:uuid;not null;index" json:"inspection_report_id"`
	ContainerType      string `gorm:"type:varchar(50);not null" json:"container_type"`
	InspectedCount     int    `gorm:"type:int;not null;default:0" json:"inspected_count"`
	PositiveCount      int    `gorm:"type:int;not null;default:0" json:"positive_count"`
}
