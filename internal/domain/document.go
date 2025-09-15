package domain

type Document struct {
	ID            string `json:"id"`
	FileName      string `json:"fileName" `
	FileExtension string `json:"fileExtension" `
	DriveLink     string `json:"driveLink"`
}
