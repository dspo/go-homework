package conformance

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/dspo/go-homework/sdk"
)

var _ = BeforeSuite(func() {
	By("1. Admin first login with initial password (admin/adminadmin)")
	admin := loginWithUsername(sdk.GetSDK(), "admin", "adminadmin")

	By("2. Admin tries to access protected resource without changing password")
	By("- Should get 403 error indicating password change required")
	_, err := admin.Me().Get()
	Expect(err).To(HaveOccurred())
	Expect(err.Error()).To(ContainSubstring("403"))

	By("3. Admin changes password")
	err = admin.Me().UpdatePassword("adminadmin", "admin123")
	Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)

	By("4. Admin tries to access protected resource without re-login")
	By("- Should get 401 error because session invalidated after password change")
	_, err = admin.Me().Get()
	Expect(err).To(HaveOccurred())
	Expect(err.Error()).To(ContainSubstring("401"))

	By("5. Admin re-login with new password")
	admin = loginWithUsername(sdk.GetSDK(), "admin", "admin123")

	By("6. Admin tries to access protected resource again")
	_, err = admin.Me().Get()
	Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
	By("13. All test users ready for testing")
})

var _ = AfterSuite(func() {
	By("Cleanup all test data")
	By("- Delete all created teams (will cascade delete projects)")
	By("- Delete all test users")
	By("- Delete all custom roles")

	// Note: Actual cleanup depends on specific test implementation
	// Usually cleanup happens in AfterAll blocks of each test context
})
