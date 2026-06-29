package domain

type ContainerDetail struct {
	ID                 string  `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
	InspectionReportID string  `gorm:"type:uuid;not null;index" json:"inspection_report_id"`
	ContainerTypeID    string  `gorm:"type:uuid;not null;index" json:"container_type_id"`
	CustomName         *string `gorm:"type:varchar(255)" json:"custom_name"`
	InspectedCount     int     `gorm:"type:int;not null;default:0" json:"inspected_count"`
	PositiveCount      int     `gorm:"type:int;not null;default:0" json:"positive_count"`

	// Relasi Normalisasi
	ContainerType *ContainerType `gorm:"foreignKey:ContainerTypeID" json:"container_type,omitempty"`
}
