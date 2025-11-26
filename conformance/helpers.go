package conformance

import (
	"fmt"
	"time"

	sdk "github.com/dspo/go-homework/sdk"
)

// helperUniqueName returns a unique name with the given prefix.
func helperUniqueName(prefix string) string {
	return fmt.Sprintf("%s_%d", prefix, time.Now().UnixNano())
}

func helperStringPtr(v string) *string {
	return &v
}

func helperBoolPtr(v bool) *bool {
	return &v
}

func helperIntPtr(v int) *int {
	return &v
}

func helperInt64Ptr(v int64) *int64 {
	return &v
}

func helperRolesContain(roles []sdk.Role, roleName string) bool {
	for _, role := range roles {
		if role.Name == roleName {
			return true
		}
	}
	return false
}
