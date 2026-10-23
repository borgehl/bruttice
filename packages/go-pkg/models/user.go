package models

type User struct {
	BaseModel

	Email        string
	PasswordHash []byte
	// how the admin is shown to other admins, such as on their replies in the inbox. Empty uses
	// the email.
	Name string
}

// the name to show for an admin
func (u User) DisplayName() string {
	if u.Name != "" {
		return u.Name
	}
	return u.Email
}
