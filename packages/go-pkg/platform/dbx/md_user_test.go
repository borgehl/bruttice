package dbx

import (
	"testing"

	"github.com/borgehl/bruttice/packages/go-pkg/models"
	"github.com/google/uuid"
)

// Needs a database, see md_inbox_message_test.go.
func TestUserNameOnARowFromBeforeTheColumn(t *testing.T) {
	d := testInboxDB(t)
	if err := d.DB.AutoMigrate(&models.User{}); err != nil {
		t.Fatal(err)
	}
	email := uuid.NewString() + "@example.com"
	// a user as it was saved before names existed: the new column is NULL
	if err := d.DB.Exec("INSERT INTO users (email, name) VALUES (?, NULL)", email).Error; err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { d.DB.Unscoped().Where("email = ?", email).Delete(&models.User{}) })

	user, err := d.GetUser(models.User{Email: email}, d.DB)
	if err != nil || user.ID == 0 {
		t.Fatalf("user = %+v, err = %v", user, err)
	}
	if user.Name != "" || user.DisplayName() != email {
		t.Errorf("a missing name is empty and the email is shown: %q, %q", user.Name, user.DisplayName())
	}

	if err := d.SetUserName(user.ID, "Sam"); err != nil {
		t.Fatal(err)
	}
	if got, _ := d.GetUser(models.User{Email: email}, d.DB); got.Name != "Sam" || got.DisplayName() != "Sam" {
		t.Errorf("name = %q", got.Name)
	}
	if err := d.SetUserName(user.ID, ""); err != nil {
		t.Fatal(err)
	}
	if got, _ := d.GetUser(models.User{Email: email}, d.DB); got.Name != "" {
		t.Errorf("clearing the name: %q", got.Name)
	}
}
