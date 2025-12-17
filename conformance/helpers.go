package conformance

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
	. "github.com/onsi/gomega"

	"github.com/dspo/go-homework/sdk"
)

// helperUniqueName returns a unique name with the given prefix.
func helperUniqueName(prefix string) string {
	suffix := uuid.New().String()
	suffix = strings.ReplaceAll(suffix, "-", "_")
	name := fmt.Sprintf("%s_%s", prefix, suffix)
	if len(name) > 20 {
		name = name[:20]
	}
	return name
}

func Ptr[T any](v T) *T {
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

// loginWithUsername 返回带登录态的客户端，便于链式访问。
func loginWithUsername(sdk sdk.SDK, username, password string) sdk.UserClient {
	client, err := sdk.LoginWithUsername(username, password)
	Expect(err).NotTo(HaveOccurred(), "failed to LoginWithUsername: %v", err)
	return client
}

// loginWithEmail 返回带登录态的客户端。
func loginWithEmail(email, password string) sdk.UserClient {
	client, err := sdk.GetSDK().LoginWithEmail(email, password)
	Expect(err).NotTo(HaveOccurred(), "failed to LoginWithEmail: %v", err)
	return client
}
