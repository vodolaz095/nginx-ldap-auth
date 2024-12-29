package endpoints

import (
	"context"
	"errors"
	"strings"

	"github.com/vodolaz095/ldap4gin"
	"go.opentelemetry.io/otel/trace"
)

var errAccessDenied = errors.New("access denied")

func (api *API) checkPermissions(ctx context.Context, hostname, path string, user *ldap4gin.User) error {
	span := trace.SpanFromContext(ctx)
	var uidMatched, groupMatched bool
	span.AddEvent("Checking hostname " + hostname + " with path " + path)
	for i := range api.Permissions {
		if api.Permissions[i].Host == hostname && strings.HasPrefix(path, api.Permissions[i].Prefix) {
			if len(api.Permissions[i].UIDs) == 0 {
				uidMatched = true
			}
			for j := range api.Permissions[i].UIDs {
				if user.UID == api.Permissions[i].UIDs[j] {
					uidMatched = true
				}
			}
			if len(api.Permissions[i].GIDs) == 0 {
				groupMatched = true
			}
			for k := range api.Permissions[i].GIDs {
				if user.HasGroupByGID(api.Permissions[i].GIDs[k]) {
					groupMatched = true
				}
			}
			if uidMatched && groupMatched {
				span.AddEvent("user allowed")
				return nil
			}
		}
	}
	return errAccessDenied
}
