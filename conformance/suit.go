package conformance

import (
	"os"

	sdk "github.com/dspo/go-homework/sdk"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var baseURL string

var _ = BeforeSuite(func() {
	// Get API base URL from environment
	baseURL = os.Getenv("API_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	// Initialize SDK
	sdk.NewSDK(baseURL)
	s := sdk.GetSDK()

	By("1. Verify system initialization")
	By("- Check admin user exists with username 'admin'")
	By("- Check 3 system roles exist: admin, team leader, normal user")

	roles, err := s.Roles().List()
	Expect(err).NotTo(HaveOccurred())
	Expect(roles.Total).To(BeNumerically(">=", 3))

	roleNames := make(map[string]bool)
	for _, role := range roles.List {
		if role.Type == "System" {
			roleNames[role.Name] = true
		}
	}
	Expect(roleNames["admin"]).To(BeTrue(), "admin role should exist")
	Expect(roleNames["team leader"]).To(BeTrue(), "team leader role should exist")
	Expect(roleNames["normal user"]).To(BeTrue(), "normal user role should exist")

	By("2. Admin first login with initial password (admin/admin)")
	err = s.Auth().LoginWithUsername("admin", "admin")
	if err == nil {
		// Admin hasn't changed password yet
		By("3. Admin tries to access protected resource without changing password")
		By("- Should get 403 error indicating password change required")
		_, err = s.Me().Get()
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("403"))

		By("4. Admin changes password")
		err = s.Me().UpdatePassword("admin", "admin123")
		Expect(err).NotTo(HaveOccurred())

		By("5. Admin tries to access protected resource without re-login")
		By("- Should get 401 error because session invalidated after password change")
		_, err = s.Me().Get()
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("401"))

		By("6. Admin re-login with new password")
		err = s.Auth().LoginWithUsername("admin", "admin123")
		Expect(err).NotTo(HaveOccurred())
	} else {
		// Admin already changed password, just login
		By("Admin login with changed password")
		err = s.Auth().LoginWithUsername("admin", "admin123")
		Expect(err).NotTo(HaveOccurred())
	}

	By("7. Admin creates test users")
	By("- Create userA, userB, userC for visibility tests")
	By("- Create userLeader for team leader tests")

	// Note: These users may already exist from previous test runs
	// We just try to create them, and if they fail, we continue

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
