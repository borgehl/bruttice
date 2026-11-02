package controlctx

import (
	"context"
	"github.com/borgehl/bruttice/packages/go-pkg/models"
	"net/http"
)

type CtxKey string

const CtxUser = CtxKey("user")

// return a shallow copy of r with the user added to the context
// user is accessible cia the CtxUser key
func RequestWithUser(r *http.Request, user *models.User) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), CtxUser, user))
}

// check the request contex for a user
// return user, or nil if no user found
func GetUserFromRequestContext(r *http.Request) *models.User {
	user := r.Context().Value(CtxUser)
	if user == nil {
		return nil
	}
	return user.(*models.User)
}
