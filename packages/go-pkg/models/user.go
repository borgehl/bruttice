package models

type User struct {
	BaseModel

	Email        string
	PasswordHash []byte
}
