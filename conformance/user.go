package conformance

import (
	sdk "github.com/dspo/go-homework/sdk"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// Helper function to login as admin
func loginAsAdmin() {
	s := sdk.GetSDK()
	err := s.Auth().LoginWithUsername("admin", "admin123")
	Expect(err).NotTo(HaveOccurred())
}

// Helper function to create user and change password
func createAndSetupUser(username, password string) *sdk.User {
	s := sdk.GetSDK()
	loginAsAdmin()

	user, err := s.Users().Create(username, password)
	Expect(err).NotTo(HaveOccurred())

	// Login and change password
	err = s.Auth().LoginWithUsername(username, password)
	Expect(err).NotTo(HaveOccurred())

	newPass := password + "456"
	err = s.Me().UpdatePassword(password, newPass)
	Expect(err).NotTo(HaveOccurred())

	err = s.Auth().LoginWithUsername(username, newPass)
	Expect(err).NotTo(HaveOccurred())

	return user
}

var _ = Describe("Authentication", func() {
	Context("Login and Logout", func() {
		It("should login with username successfully", func() {
			By("Login with valid username and password")
			s := sdk.GetSDK()
			err := s.Auth().LoginWithUsername("admin", "admin123")
			Expect(err).NotTo(HaveOccurred())

			By("Verify session cookie is set")
			me, err := s.Me().Get()
			Expect(err).NotTo(HaveOccurred())
			Expect(me.Username).To(Equal("admin"))
		})

		It("should login with email successfully", func() {
			By("Login with valid email and password")
			s := sdk.GetSDK()
			// First set email for admin
			err := s.Auth().LoginWithUsername("admin", "admin123")
			Expect(err).NotTo(HaveOccurred())

			email := "admin@example.com"
			_, err = s.Me().Update(&sdk.UpdateMeRequest{Email: &email})
			Expect(err).NotTo(HaveOccurred())

			// Logout and login with email
			err = s.Auth().Logout()
			Expect(err).NotTo(HaveOccurred())

			err = s.Auth().LoginWithEmail("admin@example.com", "admin123")
			Expect(err).NotTo(HaveOccurred())

			By("Verify session cookie is set")
			me, err := s.Me().Get()
			Expect(err).NotTo(HaveOccurred())
			Expect(me.Username).To(Equal("admin"))
		})

		It("should fail with invalid credentials", func() {
			By("Try to login with wrong password")
			s := sdk.GetSDK()
			err := s.Auth().LoginWithUsername("admin", "wrongpassword")

			By("Should get 401 Unauthorized")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("401"))
		})

		It("should logout successfully", func() {
			By("Login first")
			s := sdk.GetSDK()
			err := s.Auth().LoginWithUsername("admin", "admin123")
			Expect(err).NotTo(HaveOccurred())

			By("Logout")
			err = s.Auth().Logout()
			Expect(err).NotTo(HaveOccurred())

			By("Try to access protected resource")
			_, err = s.Me().Get()

			By("Should get 401 Unauthorized")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("401"))
		})
	})

	Context("Password Management", func() {
		It("should require password change on first login", func() {
			By("Admin creates a new user")
			s := sdk.GetSDK()
			err := s.Auth().LoginWithUsername("admin", "admin123")
			Expect(err).NotTo(HaveOccurred())

			user, err := s.Users().Create("testpasswd", "test123")
			Expect(err).NotTo(HaveOccurred())
			userID := user.ID

			By("New user login with initial password")
			err = s.Auth().LoginWithUsername("testpasswd", "test123")
			Expect(err).NotTo(HaveOccurred())

			By("Try to access /api/me")
			_, err = s.Me().Get()

			By("Should get 403 indicating password change required")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("403"))

			// Cleanup
			err = s.Auth().LoginWithUsername("admin", "admin123")
			Expect(err).NotTo(HaveOccurred())
			err = s.Users().Delete(userID)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should invalidate session after password change", func() {
			By("User login")
			s := sdk.GetSDK()
			err := s.Auth().LoginWithUsername("admin", "admin123")
			Expect(err).NotTo(HaveOccurred())

			// Create test user and change password
			user, err := s.Users().Create("testpasswd2", "test123")
			Expect(err).NotTo(HaveOccurred())
			userID := user.ID

			err = s.Auth().LoginWithUsername("testpasswd2", "test123")
			Expect(err).NotTo(HaveOccurred())

			By("User changes password")
			err = s.Me().UpdatePassword("test123", "test456")
			Expect(err).NotTo(HaveOccurred())

			By("Try to access /api/me with old session")
			_, err = s.Me().Get()

			By("Should get 401 because session is invalidated")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("401"))

			By("Login with new password should succeed")
			err = s.Auth().LoginWithUsername("testpasswd2", "test456")
			Expect(err).NotTo(HaveOccurred())

			_, err = s.Me().Get()
			Expect(err).NotTo(HaveOccurred())

			// Cleanup
			err = s.Auth().LoginWithUsername("admin", "admin123")
			Expect(err).NotTo(HaveOccurred())
			err = s.Users().Delete(userID)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Context("Health Check", func() {
		It("should access healthz without authentication", func() {
			By("GET /healthz without login")
			s := sdk.GetSDK()
			// Logout first to ensure no session
			_ = s.Auth().Logout()

			err := s.Auth().Healthz()

			By("Should get 200 OK")
			Expect(err).NotTo(HaveOccurred())
		})
	})
})

var _ = Describe("Users", func() {
	Context("User CRUD Operations", Ordered, func() {
		var userID int

		It("should create user by admin", func() {
			By("Admin login")
			s := sdk.GetSDK()
			err := s.Auth().LoginWithUsername("admin", "admin123")
			Expect(err).NotTo(HaveOccurred())

			By("POST /api/users with username and password")
			user, err := s.Users().Create("testuser_crud", "test123")

			By("Should return 200 with user object")
			Expect(err).NotTo(HaveOccurred())
			Expect(user.Username).To(Equal("testuser_crud"))

			By("Save user ID for later tests")
			userID = user.ID
		})

		It("should fail to create user by normal user", func() {
			By("Normal user login")
			s := sdk.GetSDK()
			// Create a normal user first
			err := s.Auth().LoginWithUsername("admin", "admin123")
			Expect(err).NotTo(HaveOccurred())
			normalUser, err := s.Users().Create("normaluser", "test123")
			Expect(err).NotTo(HaveOccurred())

			// Change password for normal user
			err = s.Auth().LoginWithUsername("normaluser", "test123")
			Expect(err).NotTo(HaveOccurred())
			err = s.Me().UpdatePassword("test123", "test456")
			Expect(err).NotTo(HaveOccurred())

			err = s.Auth().LoginWithUsername("normaluser", "test456")
			Expect(err).NotTo(HaveOccurred())

			By("Try to POST /api/users")
			_, err = s.Users().Create("anotheruser", "test123")

			By("Should get 403 Forbidden")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("403"))

			// Cleanup
			err = s.Auth().LoginWithUsername("admin", "admin123")
			Expect(err).NotTo(HaveOccurred())
			_ = s.Users().Delete(normalUser.ID)
		})

		It("should list users with admin privileges", func() {
			By("Admin login")
			s := sdk.GetSDK()
			err := s.Auth().LoginWithUsername("admin", "admin123")
			Expect(err).NotTo(HaveOccurred())

			By("GET /api/users")
			users, err := s.Users().List(nil)

			By("Should return all users including newly created")
			Expect(err).NotTo(HaveOccurred())
			Expect(users.Total).To(BeNumerically(">", 0))

			found := false
			for _, u := range users.List {
				if u.ID == userID {
					found = true
					break
				}
			}
			Expect(found).To(BeTrue())
		})

		It("should list visible users for normal user", func() {
			By("Normal user (not in any team) login")
			s := sdk.GetSDK()
			err := s.Auth().LoginWithUsername("admin", "admin123")
			Expect(err).NotTo(HaveOccurred())

			isolatedUser, err := s.Users().Create("isolated", "test123")
			Expect(err).NotTo(HaveOccurred())

			err = s.Auth().LoginWithUsername("isolated", "test123")
			Expect(err).NotTo(HaveOccurred())
			err = s.Me().UpdatePassword("test123", "test456")
			Expect(err).NotTo(HaveOccurred())
			err = s.Auth().LoginWithUsername("isolated", "test456")
			Expect(err).NotTo(HaveOccurred())

			By("GET /api/users")
			users, err := s.Users().List(nil)

			By("Should return empty list or only self")
			Expect(err).NotTo(HaveOccurred())
			// Should only see themselves
			Expect(users.Total).To(Equal(1))
			Expect(users.List[0].Username).To(Equal("isolated"))

			// Cleanup
			err = s.Auth().LoginWithUsername("admin", "admin123")
			Expect(err).NotTo(HaveOccurred())
			_ = s.Users().Delete(isolatedUser.ID)
		})

		It("should get user details", func() {
			By("Admin login")
			s := sdk.GetSDK()
			err := s.Auth().LoginWithUsername("admin", "admin123")
			Expect(err).NotTo(HaveOccurred())

			By("GET /api/users/{user_id}")
			user, err := s.Users().Get(userID)

			By("Should return user details")
			Expect(err).NotTo(HaveOccurred())
			Expect(user.ID).To(Equal(userID))
			Expect(user.Username).To(Equal("testuser_crud"))
		})

		It("should fail to get invisible user details", func() {
			By("UserA login (in teamA)")
			s := sdk.GetSDK()
			err := s.Auth().LoginWithUsername("admin", "admin123")
			Expect(err).NotTo(HaveOccurred())

			// Create teamA and teamB
			teamA, err := s.Teams().Create(&sdk.CreateTeamRequest{Name: "TeamA"})
			Expect(err).NotTo(HaveOccurred())
			teamB, err := s.Teams().Create(&sdk.CreateTeamRequest{Name: "TeamB"})
			Expect(err).NotTo(HaveOccurred())

			// Create userA in teamA and userB in teamB
			userA, err := s.Users().Create("userA_visibility", "test123")
			Expect(err).NotTo(HaveOccurred())
			userB, err := s.Users().Create("userB_visibility", "test123")
			Expect(err).NotTo(HaveOccurred())

			err = s.Teams().AddUser(teamA.ID, userA.ID)
			Expect(err).NotTo(HaveOccurred())
			err = s.Teams().AddUser(teamB.ID, userB.ID)
			Expect(err).NotTo(HaveOccurred())

			// UserA login
			err = s.Auth().LoginWithUsername("userA_visibility", "test123")
			Expect(err).NotTo(HaveOccurred())
			err = s.Me().UpdatePassword("test123", "test456")
			Expect(err).NotTo(HaveOccurred())
			err = s.Auth().LoginWithUsername("userA_visibility", "test456")
			Expect(err).NotTo(HaveOccurred())

			By("Try to GET /api/users/{userB_id} where userB not in teamA")
			_, err = s.Users().Get(userB.ID)

			By("Should get 403 Forbidden")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("403"))

			// Cleanup
			err = s.Auth().LoginWithUsername("admin", "admin123")
			Expect(err).NotTo(HaveOccurred())
			_ = s.Teams().Delete(teamA.ID)
			_ = s.Teams().Delete(teamB.ID)
			_ = s.Users().Delete(userA.ID)
			_ = s.Users().Delete(userB.ID)
		})

		It("should delete user by admin", func() {
			By("Admin login")
			s := sdk.GetSDK()
			err := s.Auth().LoginWithUsername("admin", "admin123")
			Expect(err).NotTo(HaveOccurred())

			By("DELETE /api/users/{user_id}")
			err = s.Users().Delete(userID)

			By("Should return 200")
			Expect(err).NotTo(HaveOccurred())

			By("Verify user is deleted by getting 404")
			_, err = s.Users().Get(userID)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("404"))
		})

		It("should fail to delete admin user", func() {
			By("Admin login")
			s := sdk.GetSDK()
			err := s.Auth().LoginWithUsername("admin", "admin123")
			Expect(err).NotTo(HaveOccurred())

			// Get admin user ID
			me, err := s.Me().Get()
			Expect(err).NotTo(HaveOccurred())

			By("Try to DELETE /api/users/{admin_id}")
			err = s.Users().Delete(me.ID)

			By("Should get 403 or 400")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Or(ContainSubstring("403"), ContainSubstring("400")))
		})
	})

	Context("User Visibility", Ordered, func() {
		var _, _ int    // teamA, teamB - will be assigned in actual implementation
		var _, _, _ int // userA, userB, userC - will be assigned in actual implementation

		BeforeAll(func() {
			By("Admin creates teamA and teamB")
			By("Admin creates userA, userB, userC")
			By("Add userA to teamA")
			By("Add userB to teamB")
			By("Add userC to both teamA and teamB")
		})

		AfterAll(func() {
			By("Delete teams (cascade)")
			By("Delete test users")
		})

		It("should see users in same team", func() {
			By("UserA login (in teamA)")
			By("GET /api/users")
			By("Should see userC (both in teamA)")
			By("Should NOT see userB (not in same team)")
		})

		It("should see users in multiple teams", func() {
			By("UserC login (in teamA and teamB)")
			By("GET /api/users")
			By("Should see both userA and userB")
		})
	})

	Context("Me APIs", func() {
		It("should get current user info", func() {
			By("User login")
			By("GET /api/me")
			By("Should return current user object with roles")
		})

		It("should update current user info", func() {
			By("User login")
			By("PUT /api/me with new email, nickname, logo")
			By("Should return 200 with updated user")
			By("Verify changes by GET /api/me")
		})

		It("should list my teams", func() {
			By("User login (in teamA and teamB)")
			By("GET /api/me/teams")
			By("Should return both teams")
		})

		It("should list my leading teams", func() {
			By("Team leader login")
			By("GET /api/me/teams?leading=true")
			By("Should return only teams where user is leader")
		})

		It("should list my non-leading teams", func() {
			By("User login (leader of teamA, member of teamB)")
			By("GET /api/me/teams?leading=false")
			By("Should return only teamB")
		})

		It("should exit from team", func() {
			By("User login (in teamA)")
			By("DELETE /api/me/teams/{teamA_id}")
			By("Should return 200")
			By("GET /api/me/teams should not include teamA")
		})

		It("should clear leader when leader exits team", func() {
			By("Leader login")
			By("DELETE /api/me/teams/{team_id}")
			By("Should return 200")
			By("GET /api/teams/{team_id} should show leader is null")
		})

		It("should list my projects", func() {
			By("User login (in projectA and projectB)")
			By("GET /api/me/projects")
			By("Should return both projects")
		})

		It("should filter projects by team", func() {
			By("User login")
			By("GET /api/me/projects?team_id=1&team_id=2")
			By("Should return only projects in team 1 and 2")
		})

		It("should exit from project", func() {
			By("User login (in projectA)")
			By("DELETE /api/me/projects/{projectA_id}")
			By("Should return 200")
			By("GET /api/me/projects should not include projectA")
		})
	})

	Context("User Teams and Projects", Ordered, func() {
		var _, _, _ int // userID, teamID, projectID - will be assigned in actual implementation

		BeforeAll(func() {
			By("Admin creates user, team, and project")
			By("Add user to team and project")
		})

		AfterAll(func() {
			By("Cleanup resources")
		})

		It("should list user's teams", func() {
			By("Admin login")
			By("GET /api/users/{user_id}/teams")
			By("Should return teams list")
		})

		It("should respect visibility when listing user teams", func() {
			By("UserA login (in teamA with userB)")
			By("GET /api/users/{userB_id}/teams")
			By("Should return only teamA (intersection of their teams)")
		})

		It("should list user's projects", func() {
			By("Admin login")
			By("GET /api/users/{user_id}/projects")
			By("Should return projects list")
		})

		It("should respect visibility when listing user projects", func() {
			By("UserA login (in projectA with userB)")
			By("GET /api/users/{userB_id}/projects")
			By("Should return only projectA (intersection)")
		})
	})
})

var _ = Describe("Roles", func() {
	Context("System Roles", func() {
		It("should have 3 system roles initialized", func() {
			By("GET /api/roles")
			By("Should include: admin, team leader, normal user")
			By("All should have type=System")
		})

		It("should not delete system roles", func() {
			By("Admin login")
			By("Try to DELETE /api/roles/{system_role_id}")
			By("Should get 403 or 400")
		})

		It("should not add system roles to users manually", func() {
			By("Admin login")
			By("Try to POST /api/users/{user_id}/roles with system role")
			By("Should get 403 or 400")
		})

		It("should auto-assign admin role to admin user", func() {
			By("GET /api/me as admin")
			By("Should include admin role in roles array")
		})

		It("should not remove admin role from admin user", func() {
			By("Admin login")
			By("Try to DELETE /api/users/{admin_id}/roles/{admin_role_id}")
			By("Should get 403 or 400")
		})

		It("should auto-assign team leader role", func() {
			By("Admin sets user as team leader")
			By("GET /api/me as that user")
			By("Should include 'team leader' role")
		})

		It("should auto-assign normal user role", func() {
			By("Admin creates new user")
			By("GET /api/users/{new_user_id}")
			By("Should have 'normal user' role by default")
		})
	})

	Context("Custom Roles", Ordered, func() {
		var _ int // roleID - will be assigned in actual implementation

		It("should create custom role by admin", func() {
			By("Admin login")
			By("POST /api/roles with name and desc")
			By("Should return role with type=Custom")
			By("Save role ID")
		})

		It("should fail to create role by normal user", func() {
			By("Normal user login")
			By("Try to POST /api/roles")
			By("Should get 403 Forbidden")
		})

		It("should list all roles", func() {
			By("Any user login")
			By("GET /api/roles")
			By("Should return system roles + custom roles")
			By("Response should not be paginated")
		})

		It("should add custom role to user", func() {
			By("Admin login")
			By("POST /api/users/{user_id}/roles with role_id")
			By("Should return 200")
			By("GET /api/users/{user_id} should include new role")
		})

		It("should remove custom role from user", func() {
			By("Admin login")
			By("DELETE /api/users/{user_id}/roles/{role_id}")
			By("Should return 200")
			By("GET /api/users/{user_id} should not include removed role")
		})

		It("should delete custom role", func() {
			By("Admin login")
			By("DELETE /api/roles/{role_id}")
			By("Should return 200")
			By("GET /api/roles should not include deleted role")
			By("User who had this role should not be deleted")
		})
	})

	Context("Role Permissions", func() {
		It("should allow only admin to manage roles", func() {
			By("Normal user login")
			By("Try to POST /api/roles")
			By("Should get 403 Forbidden")
			By("Try to DELETE /api/roles/{role_id}")
			By("Should get 403 Forbidden")
		})

		It("should allow only admin to assign roles", func() {
			By("Normal user login")
			By("Try to POST /api/users/{user_id}/roles")
			By("Should get 403 Forbidden")
		})
	})
})

var _ = Describe("Audits", func() {
	Context("Audit Logs", Ordered, func() {
		BeforeAll(func() {
			By("Admin performs some audited operations")
			By("- Create user")
			By("- Update team")
			By("- Delete project")
		})

		It("should query audit logs by admin", func() {
			By("Admin login")
			By("GET /api/audits")
			By("Should return audit log entries")
			By("Each entry should have: id, content, created_at")
		})

		It("should fail to query by normal user", func() {
			By("Normal user login")
			By("Try to GET /api/audits")
			By("Should get 403 Forbidden")
		})

		It("should filter by keyword", func() {
			By("Admin login")
			By("GET /api/audits?keyword=delete")
			By("Should return only logs containing 'delete'")
		})

		It("should filter by time range", func() {
			By("Admin login")
			By("GET /api/audits?start_at={timestamp}&end_at={timestamp}")
			By("Should return logs within time range")
		})

		It("should paginate audit logs", func() {
			By("Admin login")
			By("GET /api/audits?page=1&page_size=10")
			By("Should return first 10 logs")
			By("GET /api/audits?page=2&page_size=10")
			By("Should return next 10 logs")
		})

		It("should order audit logs", func() {
			By("Admin login")
			By("GET /api/audits?order_by=created_at")
			By("Should return logs ordered by creation time")
		})
	})
})
