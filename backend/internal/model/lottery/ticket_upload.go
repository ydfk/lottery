package lottery

import "go-fiber-starter/internal/model/base"

type TicketUpload struct {
	base.BaseModel
	LotteryCode           string  `gorm:"index;size:32" json:"lotteryCode"`
	Status                string  `gorm:"size:32" json:"status"`
	OriginalFilename      string  `gorm:"size:255" json:"originalFilename"`
	ImagePath             string  `gorm:"size:255" json:"imagePath"`
	RecognizedText        string  `gorm:"type:text" json:"recognizedText"`
	RecognitionIssue      string  `gorm:"size:32" json:"recognitionIssue"`
	RecognitionConfidence float64 `json:"recognitionConfidence"`
	RecognitionPayload    string  `gorm:"type:text" json:"recognitionPayload"`
	ErrorMessage          string  `gorm:"type:text" json:"errorMessage"`
}

func (TicketUpload) TableName() string {
	return "ticket_uploads"
}
