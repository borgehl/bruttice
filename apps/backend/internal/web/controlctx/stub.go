package controlctx

import (
	"github.com/borgehl/bruttice/packages/go-pkg/models"
	"context"
	"net/http"
)

type CtxKey string

const CtxUser = CtxKey("user")

func RequestWithUser(r *http.Request, user *models.User) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), CtxUser, user))
}

func GetUserFromRequestContext(r *http.Request) *models.User {
	user := r.Context().Value(CtxUser)
	if user == nil {
		return nil
	}
	return user.(*models.User)
}
