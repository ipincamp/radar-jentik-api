package domain

type ContainerDetail struct {
	ID                 string `json:"id"`
	InspectionReportID string `json:"inspection_report_id"`
	ContainerType      string `json:"container_type"`
	InspectedCount     int    `json:"inspected_count"`
	PositiveCount      int    `json:"positive_count"`
}
