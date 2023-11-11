package message

type FileUpdatedSchemaMessage struct {
	FileName string `json:"file_name"`
	IsCreate bool   `json:"is_create"`
}
