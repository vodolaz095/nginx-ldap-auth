package endpoints

import (
	"context"
	"errors"
	"html/template"
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
			span.AddEvent("Permission found for hostname=" + hostname + " with path=" + path)
			if len(api.Permissions[i].UIDs) == 0 {
				span.AddEvent("all UIDs allowed")
				uidMatched = true
			}
			for j := range api.Permissions[i].UIDs {
				if user.UID == api.Permissions[i].UIDs[j] {
					span.AddEvent("user has uid " + api.Permissions[i].UIDs[j])
					uidMatched = true
				}
			}
			if len(api.Permissions[i].GIDs) == 0 {
				span.AddEvent("all GIDs allowed")
				groupMatched = true
			}
			for k := range api.Permissions[i].GIDs {
				if user.HasGroupByName(api.Permissions[i].GIDs[k]) {
					span.AddEvent("user has group " + api.Permissions[i].GIDs[k])
					groupMatched = true
				}
			}
			if uidMatched && groupMatched {
				span.AddEvent("User " + user.String() + " is allowed for hostname=" + hostname + " with path=" + path)
				return nil
			}
		}
	}
	span.AddEvent("User " + user.String() + " resticted hostname=" + hostname + " with path=" + path)
	return errAccessDenied
}

func (api *API) listAllowed(hostname string, user *ldap4gin.User) (ret []template.HTMLAttr) {
	var uidMatched, groupMatched bool
	for i := range api.Permissions {
		if api.Permissions[i].Host != hostname {
			continue
		}
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
			if user.HasGroupByName(api.Permissions[i].GIDs[k]) {
				groupMatched = true
			}
		}
		if uidMatched && groupMatched {
			// https://stackoverflow.com/questions/38037615/prevent-escaping-forward-slashes-in-templates
			ret = append(ret, template.HTMLAttr(strings.TrimPrefix(api.Permissions[i].Prefix, "/")))
		}
	}
	return ret
}
