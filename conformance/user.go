package conformance

import (
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/dspo/go-homework/sdk"
)

// Helper function to login as admin
func loginAsAdmin() {
	s := sdk.GetSDK()
	err := s.Auth().LoginWithUsername("admin", "admin123")
	Expect(err).NotTo(HaveOccurred())
}

// Helper function to create user and change password
func createAndSetupUser(username, password string) (*sdk.User, string) {
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

	return user, newPass
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
		var teamAID, teamBID int
		var userA, userB, userC *sdk.User
		var passA, passC string

		BeforeAll(func() {
			s := sdk.GetSDK()
			loginAsAdmin()

			tA, err := s.Teams().Create(&sdk.CreateTeamRequest{Name: helperUniqueName("teamA")})
			Expect(err).NotTo(HaveOccurred())
			teamAID = tA.ID

			tB, err := s.Teams().Create(&sdk.CreateTeamRequest{Name: helperUniqueName("teamB")})
			Expect(err).NotTo(HaveOccurred())
			teamBID = tB.ID

			userA, passA = createAndSetupUser(helperUniqueName("userA"), "passA")
			userB, _ = createAndSetupUser(helperUniqueName("userB"), "passB")
			userC, passC = createAndSetupUser(helperUniqueName("userC"), "passC")

			loginAsAdmin()
			Expect(s.Teams().AddUser(teamAID, userA.ID)).NotTo(HaveOccurred())
			Expect(s.Teams().AddUser(teamBID, userB.ID)).NotTo(HaveOccurred())
			Expect(s.Teams().AddUser(teamAID, userC.ID)).NotTo(HaveOccurred())
			Expect(s.Teams().AddUser(teamBID, userC.ID)).NotTo(HaveOccurred())
		})

		AfterAll(func() {
			s := sdk.GetSDK()
			loginAsAdmin()
			_ = s.Teams().Delete(teamAID)
			_ = s.Teams().Delete(teamBID)
			_ = s.Users().Delete(userA.ID)
			_ = s.Users().Delete(userB.ID)
			_ = s.Users().Delete(userC.ID)
		})

		It("should see users in same team", func() {
			s := sdk.GetSDK()
			Expect(s.Auth().LoginWithUsername(userA.Username, passA)).NotTo(HaveOccurred())
			users, err := s.Users().List(nil)
			Expect(err).NotTo(HaveOccurred())

			visibleIDs := make(map[int]bool)
			for _, u := range users.List {
				visibleIDs[u.ID] = true
			}

			Expect(visibleIDs[userA.ID]).To(BeTrue())
			Expect(visibleIDs[userC.ID]).To(BeTrue())
			Expect(visibleIDs[userB.ID]).To(BeFalse())
		})

		It("should see users in multiple teams", func() {
			s := sdk.GetSDK()
			Expect(s.Auth().LoginWithUsername(userC.Username, passC)).NotTo(HaveOccurred())
			users, err := s.Users().List(nil)
			Expect(err).NotTo(HaveOccurred())

			visibleIDs := make(map[int]bool)
			for _, u := range users.List {
				visibleIDs[u.ID] = true
			}

			Expect(visibleIDs[userA.ID]).To(BeTrue())
			Expect(visibleIDs[userB.ID]).To(BeTrue())
			Expect(visibleIDs[userC.ID]).To(BeTrue())
		})
	})

	Context("Me APIs", func() {
		It("should get current user info", func() {
			s := sdk.GetSDK()
			user, pass := createAndSetupUser(helperUniqueName("me_user"), "pass")
			DeferCleanup(func() {
				loginAsAdmin()
				_ = s.Users().Delete(user.ID)
			})

			Expect(s.Auth().LoginWithUsername(user.Username, pass)).NotTo(HaveOccurred())
			me, err := s.Me().Get()
			Expect(err).NotTo(HaveOccurred())
			Expect(me.ID).To(Equal(user.ID))
			Expect(me.Username).To(Equal(user.Username))
			Expect(me.Roles).NotTo(BeEmpty())
		})

		It("should update current user info", func() {
			s := sdk.GetSDK()
			user, pass := createAndSetupUser(helperUniqueName("me_update"), "pass")
			DeferCleanup(func() {
				loginAsAdmin()
				_ = s.Users().Delete(user.ID)
			})

			Expect(s.Auth().LoginWithUsername(user.Username, pass)).NotTo(HaveOccurred())
			email := helperUniqueName("me_email") + "@example.com"
			nickname := helperUniqueName("nick")
			logo := "https://logo.example.com/" + helperUniqueName("logo")
			updated, err := s.Me().Update(&sdk.UpdateMeRequest{
				Email:    helperStringPtr(email),
				Nickname: helperStringPtr(nickname),
				Logo:     helperStringPtr(logo),
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(updated.Email).NotTo(BeNil())
			Expect(*updated.Email).To(Equal(email))
			Expect(updated.Nickname).NotTo(BeNil())
			Expect(*updated.Nickname).To(Equal(nickname))
			Expect(updated.Logo).NotTo(BeNil())
			Expect(*updated.Logo).To(Equal(logo))

			me, err := s.Me().Get()
			Expect(err).NotTo(HaveOccurred())
			Expect(me.Email).NotTo(BeNil())
			Expect(*me.Email).To(Equal(email))
			Expect(me.Nickname).NotTo(BeNil())
			Expect(*me.Nickname).To(Equal(nickname))
		})

		It("should list my teams", func() {
			s := sdk.GetSDK()
			user, pass := createAndSetupUser(helperUniqueName("me_teams"), "pass")
			DeferCleanup(func() {
				loginAsAdmin()
				_ = s.Users().Delete(user.ID)
			})

			loginAsAdmin()
			teamA, err := s.Teams().Create(&sdk.CreateTeamRequest{Name: helperUniqueName("teamA")})
			Expect(err).NotTo(HaveOccurred())
			teamB, err := s.Teams().Create(&sdk.CreateTeamRequest{Name: helperUniqueName("teamB")})
			Expect(err).NotTo(HaveOccurred())
			DeferCleanup(func() {
				loginAsAdmin()
				_ = s.Teams().Delete(teamA.ID)
				_ = s.Teams().Delete(teamB.ID)
			})

			Expect(s.Teams().AddUser(teamA.ID, user.ID)).NotTo(HaveOccurred())
			Expect(s.Teams().AddUser(teamB.ID, user.ID)).NotTo(HaveOccurred())

			Expect(s.Auth().LoginWithUsername(user.Username, pass)).NotTo(HaveOccurred())
			teams, err := s.Me().ListTeams(nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(teams.Total).To(Equal(2))
			ids := []int{}
			for _, t := range teams.List {
				ids = append(ids, t.ID)
			}
			Expect(ids).To(ContainElements(teamA.ID, teamB.ID))
		})

		It("should list my leading teams", func() {
			s := sdk.GetSDK()
			user, pass := createAndSetupUser(helperUniqueName("me_lead"), "pass")
			DeferCleanup(func() {
				loginAsAdmin()
				_ = s.Users().Delete(user.ID)
			})

			loginAsAdmin()
			teamLead, err := s.Teams().Create(&sdk.CreateTeamRequest{Name: helperUniqueName("team_lead")})
			Expect(err).NotTo(HaveOccurred())
			teamMember, err := s.Teams().Create(&sdk.CreateTeamRequest{Name: helperUniqueName("team_member")})
			Expect(err).NotTo(HaveOccurred())
			DeferCleanup(func() {
				loginAsAdmin()
				_ = s.Teams().Delete(teamLead.ID)
				_ = s.Teams().Delete(teamMember.ID)
			})

			Expect(s.Teams().AddUser(teamLead.ID, user.ID)).NotTo(HaveOccurred())
			Expect(s.Teams().AddUser(teamMember.ID, user.ID)).NotTo(HaveOccurred())
			Expect(s.Teams().UpdateLeader(teamLead.ID, helperIntPtr(user.ID))).NotTo(HaveOccurred())

			Expect(s.Auth().LoginWithUsername(user.Username, pass)).NotTo(HaveOccurred())
			leading := true
			teams, err := s.Me().ListTeams(&leading)
			Expect(err).NotTo(HaveOccurred())
			Expect(teams.Total).To(Equal(1))
			Expect(teams.List[0].ID).To(Equal(teamLead.ID))
		})

		It("should list my non-leading teams", func() {
			s := sdk.GetSDK()
			user, pass := createAndSetupUser(helperUniqueName("me_nonlead"), "pass")
			DeferCleanup(func() {
				loginAsAdmin()
				_ = s.Users().Delete(user.ID)
			})

			loginAsAdmin()
			teamLeader, err := s.Teams().Create(&sdk.CreateTeamRequest{Name: helperUniqueName("team_leader")})
			Expect(err).NotTo(HaveOccurred())
			teamMember, err := s.Teams().Create(&sdk.CreateTeamRequest{Name: helperUniqueName("team_member")})
			Expect(err).NotTo(HaveOccurred())
			DeferCleanup(func() {
				loginAsAdmin()
				_ = s.Teams().Delete(teamLeader.ID)
				_ = s.Teams().Delete(teamMember.ID)
			})

			Expect(s.Teams().AddUser(teamLeader.ID, user.ID)).NotTo(HaveOccurred())
			Expect(s.Teams().AddUser(teamMember.ID, user.ID)).NotTo(HaveOccurred())
			Expect(s.Teams().UpdateLeader(teamLeader.ID, helperIntPtr(user.ID))).NotTo(HaveOccurred())

			Expect(s.Auth().LoginWithUsername(user.Username, pass)).NotTo(HaveOccurred())
			leading := false
			teams, err := s.Me().ListTeams(&leading)
			Expect(err).NotTo(HaveOccurred())
			Expect(teams.Total).To(Equal(1))
			Expect(teams.List[0].ID).To(Equal(teamMember.ID))
		})

		It("should exit from team", func() {
			s := sdk.GetSDK()
			user, pass := createAndSetupUser(helperUniqueName("me_exit_team"), "pass")
			DeferCleanup(func() {
				loginAsAdmin()
				_ = s.Users().Delete(user.ID)
			})

			loginAsAdmin()
			team, err := s.Teams().Create(&sdk.CreateTeamRequest{Name: helperUniqueName("team_exit")})
			Expect(err).NotTo(HaveOccurred())
			DeferCleanup(func() {
				loginAsAdmin()
				_ = s.Teams().Delete(team.ID)
			})

			Expect(s.Teams().AddUser(team.ID, user.ID)).NotTo(HaveOccurred())

			Expect(s.Auth().LoginWithUsername(user.Username, pass)).NotTo(HaveOccurred())
			Expect(s.Me().ExitTeam(team.ID)).NotTo(HaveOccurred())
			teams, err := s.Me().ListTeams(nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(teams.Total).To(Equal(0))
		})

		It("should clear leader when leader exits team", func() {
			s := sdk.GetSDK()
			user, pass := createAndSetupUser(helperUniqueName("me_exit_leader"), "pass")
			DeferCleanup(func() {
				loginAsAdmin()
				_ = s.Users().Delete(user.ID)
			})

			loginAsAdmin()
			team, err := s.Teams().Create(&sdk.CreateTeamRequest{Name: helperUniqueName("team_leader_exit")})
			Expect(err).NotTo(HaveOccurred())
			DeferCleanup(func() {
				loginAsAdmin()
				_ = s.Teams().Delete(team.ID)
			})

			Expect(s.Teams().AddUser(team.ID, user.ID)).NotTo(HaveOccurred())
			Expect(s.Teams().UpdateLeader(team.ID, helperIntPtr(user.ID))).NotTo(HaveOccurred())

			Expect(s.Auth().LoginWithUsername(user.Username, pass)).NotTo(HaveOccurred())
			Expect(s.Me().ExitTeam(team.ID)).NotTo(HaveOccurred())

			loginAsAdmin()
			teamInfo, err := s.Teams().Get(team.ID)
			Expect(err).NotTo(HaveOccurred())
			Expect(teamInfo.Leader).To(BeNil())
		})

		It("should list my projects", func() {
			s := sdk.GetSDK()
			user, pass := createAndSetupUser(helperUniqueName("me_projects"), "pass")
			DeferCleanup(func() {
				loginAsAdmin()
				_ = s.Users().Delete(user.ID)
			})

			loginAsAdmin()
			team, err := s.Teams().Create(&sdk.CreateTeamRequest{Name: helperUniqueName("team_projects")})
			Expect(err).NotTo(HaveOccurred())
			DeferCleanup(func() {
				loginAsAdmin()
				_ = s.Teams().Delete(team.ID)
			})

			Expect(s.Teams().AddUser(team.ID, user.ID)).NotTo(HaveOccurred())
			projA, err := s.Teams().CreateProject(team.ID, &sdk.CreateProjectRequest{Name: helperUniqueName("projectA")})
			Expect(err).NotTo(HaveOccurred())
			projB, err := s.Teams().CreateProject(team.ID, &sdk.CreateProjectRequest{Name: helperUniqueName("projectB")})
			Expect(err).NotTo(HaveOccurred())
			Expect(s.Projects().AddUser(projA.ID, user.ID)).NotTo(HaveOccurred())
			Expect(s.Projects().AddUser(projB.ID, user.ID)).NotTo(HaveOccurred())

			Expect(s.Auth().LoginWithUsername(user.Username, pass)).NotTo(HaveOccurred())
			projects, err := s.Me().ListProjects(nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(projects.Total).To(Equal(2))
			ids := []int{}
			for _, p := range projects.List {
				ids = append(ids, p.ID)
			}
			Expect(ids).To(ContainElements(projA.ID, projB.ID))
		})

		It("should filter projects by team", func() {
			s := sdk.GetSDK()
			user, pass := createAndSetupUser(helperUniqueName("me_projects_filter"), "pass")
			DeferCleanup(func() {
				loginAsAdmin()
				_ = s.Users().Delete(user.ID)
			})

			loginAsAdmin()
			teamA, err := s.Teams().Create(&sdk.CreateTeamRequest{Name: helperUniqueName("team_filter_a")})
			Expect(err).NotTo(HaveOccurred())
			teamB, err := s.Teams().Create(&sdk.CreateTeamRequest{Name: helperUniqueName("team_filter_b")})
			Expect(err).NotTo(HaveOccurred())
			DeferCleanup(func() {
				loginAsAdmin()
				_ = s.Teams().Delete(teamA.ID)
				_ = s.Teams().Delete(teamB.ID)
			})

			Expect(s.Teams().AddUser(teamA.ID, user.ID)).NotTo(HaveOccurred())
			Expect(s.Teams().AddUser(teamB.ID, user.ID)).NotTo(HaveOccurred())
			projA, err := s.Teams().CreateProject(teamA.ID, &sdk.CreateProjectRequest{Name: helperUniqueName("projA")})
			Expect(err).NotTo(HaveOccurred())
			projB, err := s.Teams().CreateProject(teamB.ID, &sdk.CreateProjectRequest{Name: helperUniqueName("projB")})
			Expect(err).NotTo(HaveOccurred())
			Expect(s.Projects().AddUser(projA.ID, user.ID)).NotTo(HaveOccurred())
			Expect(s.Projects().AddUser(projB.ID, user.ID)).NotTo(HaveOccurred())

			Expect(s.Auth().LoginWithUsername(user.Username, pass)).NotTo(HaveOccurred())
			paramsA := &sdk.ListParams{TeamIDs: []int{teamA.ID}}
			projectsA, err := s.Me().ListProjects(paramsA)
			Expect(err).NotTo(HaveOccurred())
			Expect(projectsA.Total).To(Equal(1))
			Expect(projectsA.List[0].ID).To(Equal(projA.ID))

			paramsBoth := &sdk.ListParams{TeamIDs: []int{teamA.ID, teamB.ID}}
			projectsBoth, err := s.Me().ListProjects(paramsBoth)
			Expect(err).NotTo(HaveOccurred())
			Expect(projectsBoth.Total).To(Equal(2))
			ids := []int{}
			for _, p := range projectsBoth.List {
				ids = append(ids, p.ID)
			}
			Expect(ids).To(ContainElements(projA.ID, projB.ID))
		})

		It("should exit from project", func() {
			s := sdk.GetSDK()
			user, pass := createAndSetupUser(helperUniqueName("me_exit_project"), "pass")
			DeferCleanup(func() {
				loginAsAdmin()
				_ = s.Users().Delete(user.ID)
			})

			loginAsAdmin()
			team, err := s.Teams().Create(&sdk.CreateTeamRequest{Name: helperUniqueName("team_project_exit")})
			Expect(err).NotTo(HaveOccurred())
			DeferCleanup(func() {
				loginAsAdmin()
				_ = s.Teams().Delete(team.ID)
			})

			Expect(s.Teams().AddUser(team.ID, user.ID)).NotTo(HaveOccurred())
			project, err := s.Teams().CreateProject(team.ID, &sdk.CreateProjectRequest{Name: helperUniqueName("proj_exit")})
			Expect(err).NotTo(HaveOccurred())
			Expect(s.Projects().AddUser(project.ID, user.ID)).NotTo(HaveOccurred())

			Expect(s.Auth().LoginWithUsername(user.Username, pass)).NotTo(HaveOccurred())
			Expect(s.Me().ExitProject(project.ID)).NotTo(HaveOccurred())
			projects, err := s.Me().ListProjects(nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(projects.Total).To(Equal(0))

			loginAsAdmin()
			teamUsers, err := s.Teams().ListUsers(team.ID, nil)
			Expect(err).NotTo(HaveOccurred())
			found := false
			for _, u := range teamUsers.List {
				if u.ID == user.ID {
					found = true
					break
				}
			}
			Expect(found).To(BeTrue(), "exiting project should not remove user from team")
		})
	})

	Context("User Teams and Projects", Ordered, func() {
		var targetUser, viewerUser *sdk.User
		var viewerPass string
		var teamCommonID, teamPrivateID int
		var projectCommonID, projectPrivateID int

		BeforeAll(func() {
			s := sdk.GetSDK()
			loginAsAdmin()

			var err error
			targetUser, _ = createAndSetupUser(helperUniqueName("user_target"), "pass")
			viewerUser, viewerPass = createAndSetupUser(helperUniqueName("user_viewer"), "pass")

			loginAsAdmin()
			teamCommon, err := s.Teams().Create(&sdk.CreateTeamRequest{Name: helperUniqueName("team_common")})
			Expect(err).NotTo(HaveOccurred())
			teamPrivate, err := s.Teams().Create(&sdk.CreateTeamRequest{Name: helperUniqueName("team_private")})
			Expect(err).NotTo(HaveOccurred())
			teamCommonID = teamCommon.ID
			teamPrivateID = teamPrivate.ID

			Expect(s.Teams().AddUser(teamCommonID, targetUser.ID)).NotTo(HaveOccurred())
			Expect(s.Teams().AddUser(teamPrivateID, targetUser.ID)).NotTo(HaveOccurred())
			Expect(s.Teams().AddUser(teamCommonID, viewerUser.ID)).NotTo(HaveOccurred())

			projectCommon, err := s.Teams().CreateProject(teamCommonID, &sdk.CreateProjectRequest{Name: helperUniqueName("project_common")})
			Expect(err).NotTo(HaveOccurred())
			projectPrivate, err := s.Teams().CreateProject(teamPrivateID, &sdk.CreateProjectRequest{Name: helperUniqueName("project_private")})
			Expect(err).NotTo(HaveOccurred())
			projectCommonID = projectCommon.ID
			projectPrivateID = projectPrivate.ID

			Expect(s.Projects().AddUser(projectCommonID, targetUser.ID)).NotTo(HaveOccurred())
			Expect(s.Projects().AddUser(projectPrivateID, targetUser.ID)).NotTo(HaveOccurred())
			Expect(s.Projects().AddUser(projectCommonID, viewerUser.ID)).NotTo(HaveOccurred())
		})

		AfterAll(func() {
			s := sdk.GetSDK()
			loginAsAdmin()
			_ = s.Teams().Delete(teamCommonID)
			_ = s.Teams().Delete(teamPrivateID)
			_ = s.Users().Delete(targetUser.ID)
			_ = s.Users().Delete(viewerUser.ID)
		})

		It("should list user's teams", func() {
			s := sdk.GetSDK()
			loginAsAdmin()
			teams, err := s.Users().ListTeams(targetUser.ID, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(teams.Total).To(Equal(2))
			ids := []int{}
			for _, team := range teams.List {
				ids = append(ids, team.ID)
			}
			Expect(ids).To(ContainElements(teamCommonID, teamPrivateID))
		})

		It("should respect visibility when listing user teams", func() {
			s := sdk.GetSDK()
			Expect(s.Auth().LoginWithUsername(viewerUser.Username, viewerPass)).NotTo(HaveOccurred())
			teams, err := s.Users().ListTeams(targetUser.ID, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(teams.Total).To(Equal(1))
			Expect(teams.List[0].ID).To(Equal(teamCommonID))
		})

		It("should list user's projects", func() {
			s := sdk.GetSDK()
			loginAsAdmin()
			projects, err := s.Users().ListProjects(targetUser.ID)
			Expect(err).NotTo(HaveOccurred())
			Expect(projects.Total).To(Equal(2))
			ids := []int{}
			for _, project := range projects.List {
				ids = append(ids, project.ID)
			}
			Expect(ids).To(ContainElements(projectCommonID, projectPrivateID))
		})

		It("should respect visibility when listing user projects", func() {
			s := sdk.GetSDK()
			Expect(s.Auth().LoginWithUsername(viewerUser.Username, viewerPass)).NotTo(HaveOccurred())
			projects, err := s.Users().ListProjects(targetUser.ID)
			Expect(err).NotTo(HaveOccurred())
			Expect(projects.Total).To(Equal(1))
			Expect(projects.List[0].ID).To(Equal(projectCommonID))
		})
	})
})

var _ = Describe("Roles", func() {
	Context("System Roles", func() {
		It("should have 3 system roles initialized", func() {
			s := sdk.GetSDK()
			loginAsAdmin()
			roles, err := s.Roles().List()
			Expect(err).NotTo(HaveOccurred())

			systemNames := map[string]bool{}
			for _, role := range roles.List {
				if role.Type == "System" {
					systemNames[role.Name] = true
				}
			}
			Expect(systemNames["admin"]).To(BeTrue())
			Expect(systemNames["team leader"]).To(BeTrue())
			Expect(systemNames["normal user"]).To(BeTrue())
		})

		It("should not delete system roles", func() {
			s := sdk.GetSDK()
			loginAsAdmin()
			roles, err := s.Roles().List()
			Expect(err).NotTo(HaveOccurred())

			var adminRoleID int
			for _, role := range roles.List {
				if role.Name == "admin" {
					adminRoleID = role.ID
					break
				}
			}
			Expect(adminRoleID).NotTo(Equal(0))

			err = s.Roles().Delete(adminRoleID)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Or(ContainSubstring("403"), ContainSubstring("400")))
		})

		It("should not add system roles to users manually", func() {
			s := sdk.GetSDK()
			user, _ := createAndSetupUser(helperUniqueName("role_sys"), "pass")
			DeferCleanup(func() {
				loginAsAdmin()
				_ = s.Users().Delete(user.ID)
			})

			loginAsAdmin()
			roles, err := s.Roles().List()
			Expect(err).NotTo(HaveOccurred())
			var adminRoleID int
			for _, role := range roles.List {
				if role.Name == "admin" {
					adminRoleID = role.ID
					break
				}
			}
			Expect(adminRoleID).NotTo(Equal(0))

			err = s.Users().AddRole(user.ID, adminRoleID)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Or(ContainSubstring("403"), ContainSubstring("400")))
		})

		It("should auto-assign admin role to admin user", func() {
			s := sdk.GetSDK()
			loginAsAdmin()
			me, err := s.Me().Get()
			Expect(err).NotTo(HaveOccurred())
			Expect(helperRolesContain(me.Roles, "admin")).To(BeTrue())
		})

		It("should not remove admin role from admin user", func() {
			s := sdk.GetSDK()
			loginAsAdmin()
			me, err := s.Me().Get()
			Expect(err).NotTo(HaveOccurred())

			roles, err := s.Roles().List()
			Expect(err).NotTo(HaveOccurred())
			var adminRoleID int
			for _, role := range roles.List {
				if role.Name == "admin" {
					adminRoleID = role.ID
					break
				}
			}
			Expect(adminRoleID).NotTo(Equal(0))

			err = s.Users().RemoveRole(me.ID, adminRoleID)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Or(ContainSubstring("403"), ContainSubstring("400")))
		})

		It("should auto-assign team leader role", func() {
			s := sdk.GetSDK()
			user, pass := createAndSetupUser(helperUniqueName("role_leader"), "pass")
			DeferCleanup(func() {
				loginAsAdmin()
				_ = s.Users().Delete(user.ID)
			})

			loginAsAdmin()
			team, err := s.Teams().Create(&sdk.CreateTeamRequest{Name: helperUniqueName("team_leader_role")})
			Expect(err).NotTo(HaveOccurred())
			DeferCleanup(func() {
				loginAsAdmin()
				_ = s.Teams().Delete(team.ID)
			})

			Expect(s.Teams().AddUser(team.ID, user.ID)).NotTo(HaveOccurred())
			Expect(s.Teams().UpdateLeader(team.ID, helperIntPtr(user.ID))).NotTo(HaveOccurred())

			Expect(s.Auth().LoginWithUsername(user.Username, pass)).NotTo(HaveOccurred())
			me, err := s.Me().Get()
			Expect(err).NotTo(HaveOccurred())
			Expect(helperRolesContain(me.Roles, "team leader")).To(BeTrue())
		})

		It("should auto-assign normal user role", func() {
			s := sdk.GetSDK()
			user, _ := createAndSetupUser(helperUniqueName("role_normal"), "pass")
			DeferCleanup(func() {
				loginAsAdmin()
				_ = s.Users().Delete(user.ID)
			})

			loginAsAdmin()
			stored, err := s.Users().Get(user.ID)
			Expect(err).NotTo(HaveOccurred())
			Expect(helperRolesContain(stored.Roles, "normal user")).To(BeTrue())
		})
	})

	Context("Custom Roles", Ordered, func() {
		var roleID int
		var roleName string
		var testUser *sdk.User
		var testUserPass string

		BeforeAll(func() {
			testUser, testUserPass = createAndSetupUser(helperUniqueName("role_custom_user"), "pass")
			DeferCleanup(func() {
				s := sdk.GetSDK()
				loginAsAdmin()
				_ = s.Users().Delete(testUser.ID)
			})
		})

		AfterAll(func() {
			if roleID != 0 {
				s := sdk.GetSDK()
				loginAsAdmin()
				_ = s.Roles().Delete(roleID)
			}
		})

		It("should create custom role by admin", func() {
			s := sdk.GetSDK()
			loginAsAdmin()
			roleName = helperUniqueName("custom_role")
			role, err := s.Roles().Create(&sdk.CreateRoleRequest{Name: roleName, Desc: helperStringPtr("custom role for tests")})
			Expect(err).NotTo(HaveOccurred())
			Expect(role.Type).To(Equal("Custom"))
			roleID = role.ID
		})

		It("should fail to create role by normal user", func() {
			s := sdk.GetSDK()
			Expect(s.Auth().LoginWithUsername(testUser.Username, testUserPass)).NotTo(HaveOccurred())
			_, err := s.Roles().Create(&sdk.CreateRoleRequest{Name: helperUniqueName("custom_role_fail")})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("403"))
		})

		It("should list all roles", func() {
			s := sdk.GetSDK()
			Expect(s.Auth().LoginWithUsername(testUser.Username, testUserPass)).NotTo(HaveOccurred())
			roles, err := s.Roles().List()
			Expect(err).NotTo(HaveOccurred())
			Expect(roles.Total).To(BeNumerically(">=", 4))
			Expect(roles.List).NotTo(BeEmpty())
			found := false
			for _, role := range roles.List {
				if role.Name == roleName {
					found = true
					break
				}
			}
			Expect(found).To(BeTrue(), "custom role should be listed")
		})

		It("should add custom role to user", func() {
			s := sdk.GetSDK()
			loginAsAdmin()
			Expect(roleID).NotTo(Equal(0))
			Expect(s.Users().AddRole(testUser.ID, roleID)).NotTo(HaveOccurred())
			stored, err := s.Users().Get(testUser.ID)
			Expect(err).NotTo(HaveOccurred())
			Expect(helperRolesContain(stored.Roles, roleName)).To(BeTrue())
		})

		It("should remove custom role from user", func() {
			s := sdk.GetSDK()
			loginAsAdmin()
			Expect(s.Users().RemoveRole(testUser.ID, roleID)).NotTo(HaveOccurred())
			stored, err := s.Users().Get(testUser.ID)
			Expect(err).NotTo(HaveOccurred())
			Expect(helperRolesContain(stored.Roles, roleName)).To(BeFalse())
		})

		It("should delete custom role", func() {
			s := sdk.GetSDK()
			loginAsAdmin()
			Expect(s.Roles().Delete(roleID)).NotTo(HaveOccurred())
			roles, err := s.Roles().List()
			Expect(err).NotTo(HaveOccurred())
			for _, role := range roles.List {
				Expect(role.Name).NotTo(Equal(roleName))
			}
			stored, err := s.Users().Get(testUser.ID)
			Expect(err).NotTo(HaveOccurred())
			Expect(stored.ID).To(Equal(testUser.ID))
			roleID = 0
		})
	})

	Context("Role Permissions", func() {
		It("should allow only admin to manage roles", func() {
			s := sdk.GetSDK()
			user, pass := createAndSetupUser(helperUniqueName("role_perm_user"), "pass")
			DeferCleanup(func() {
				loginAsAdmin()
				_ = s.Users().Delete(user.ID)
			})

			loginAsAdmin()
			role, err := s.Roles().Create(&sdk.CreateRoleRequest{Name: helperUniqueName("role_perm"), Desc: helperStringPtr("perm test")})
			Expect(err).NotTo(HaveOccurred())
			DeferCleanup(func() {
				loginAsAdmin()
				_ = s.Roles().Delete(role.ID)
			})

			Expect(s.Auth().LoginWithUsername(user.Username, pass)).NotTo(HaveOccurred())
			_, err = s.Roles().Create(&sdk.CreateRoleRequest{Name: helperUniqueName("role_perm_fail")})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("403"))

			err = s.Roles().Delete(role.ID)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("403"))
		})

		It("should allow only admin to assign roles", func() {
			s := sdk.GetSDK()
			actor, pass := createAndSetupUser(helperUniqueName("role_assign_actor"), "pass")
			target, _ := createAndSetupUser(helperUniqueName("role_assign_target"), "pass")
			DeferCleanup(func() {
				loginAsAdmin()
				_ = s.Users().Delete(actor.ID)
				_ = s.Users().Delete(target.ID)
			})

			loginAsAdmin()
			role, err := s.Roles().Create(&sdk.CreateRoleRequest{Name: helperUniqueName("role_assign"), Desc: helperStringPtr("assign test")})
			Expect(err).NotTo(HaveOccurred())
			DeferCleanup(func() {
				loginAsAdmin()
				_ = s.Roles().Delete(role.ID)
			})

			Expect(s.Auth().LoginWithUsername(actor.Username, pass)).NotTo(HaveOccurred())
			err = s.Users().AddRole(target.ID, role.ID)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("403"))
		})
	})
})

var _ = Describe("Audits", func() {
	Context("Audit Logs", Ordered, func() {
		var auditKeyword string
		var auditTeamID int
		var auditUserID int
		var timeStart, timeEnd int64

		BeforeAll(func() {
			s := sdk.GetSDK()
			loginAsAdmin()
			auditKeyword = helperUniqueName("audit")
			timeStart = time.Now().Add(-time.Minute).Unix()

			user, err := s.Users().Create(auditKeyword+"_user", "auditpass")
			Expect(err).NotTo(HaveOccurred())
			auditUserID = user.ID

			team, err := s.Teams().Create(&sdk.CreateTeamRequest{Name: auditKeyword + "_team"})
			Expect(err).NotTo(HaveOccurred())
			auditTeamID = team.ID

			desc := helperStringPtr("desc " + auditKeyword)
			_, err = s.Teams().Update(team.ID, &sdk.UpdateTeamRequest{Name: team.Name, Desc: desc})
			Expect(err).NotTo(HaveOccurred())

			project, err := s.Teams().CreateProject(team.ID, &sdk.CreateProjectRequest{Name: auditKeyword + "_project"})
			Expect(err).NotTo(HaveOccurred())
			Expect(s.Projects().Delete(project.ID)).NotTo(HaveOccurred())

			timeEnd = time.Now().Add(time.Minute).Unix()
		})

		AfterAll(func() {
			s := sdk.GetSDK()
			loginAsAdmin()
			_ = s.Users().Delete(auditUserID)
			_ = s.Teams().Delete(auditTeamID)
		})

		It("should query audit logs by admin", func() {
			s := sdk.GetSDK()
			loginAsAdmin()
			logs, err := s.Audits().List(nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(logs.Total).To(BeNumerically(">", 0))
			Expect(logs.List).NotTo(BeEmpty())
			for _, entry := range logs.List {
				Expect(entry.ID).To(BeNumerically(">", 0))
				Expect(entry.Content).NotTo(BeEmpty())
				Expect(entry.CreatedAt).To(BeNumerically(">", 0))
			}
		})

		It("should fail to query by normal user", func() {
			s := sdk.GetSDK()
			user, pass := createAndSetupUser(helperUniqueName("audit_normal"), "pass")
			DeferCleanup(func() {
				loginAsAdmin()
				_ = s.Users().Delete(user.ID)
			})

			Expect(s.Auth().LoginWithUsername(user.Username, pass)).NotTo(HaveOccurred())
			_, err := s.Audits().List(nil)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("403"))
		})

		It("should filter by keyword", func() {
			s := sdk.GetSDK()
			loginAsAdmin()
			params := &sdk.ListParams{Keyword: helperStringPtr(auditKeyword)}
			logs, err := s.Audits().List(params)
			Expect(err).NotTo(HaveOccurred())
			Expect(logs.Total).To(BeNumerically(">", 0))
			Expect(logs.List).NotTo(BeEmpty())
			for _, entry := range logs.List {
				Expect(strings.Contains(strings.ToLower(entry.Content), strings.ToLower(auditKeyword))).To(BeTrue())
			}
		})

		It("should filter by time range", func() {
			s := sdk.GetSDK()
			loginAsAdmin()
			params := &sdk.ListParams{StartAt: helperInt64Ptr(timeStart), EndAt: helperInt64Ptr(timeEnd)}
			logs, err := s.Audits().List(params)
			Expect(err).NotTo(HaveOccurred())
			Expect(logs.Total).To(BeNumerically(">", 0))
			for _, entry := range logs.List {
				Expect(entry.CreatedAt).To(BeNumerically(">=", timeStart))
				Expect(entry.CreatedAt).To(BeNumerically("<=", timeEnd))
			}
		})

		It("should paginate audit logs", func() {
			s := sdk.GetSDK()
			loginAsAdmin()
			page := 1
			pageSize := 1
			params := &sdk.ListParams{Page: &page, PageSize: &pageSize}
			firstPage, err := s.Audits().List(params)
			Expect(err).NotTo(HaveOccurred())
			Expect(firstPage.List).To(HaveLen(1))
			page = 2
			secondPage, err := s.Audits().List(params)
			Expect(err).NotTo(HaveOccurred())
			Expect(secondPage.List).To(HaveLen(1))
			Expect(firstPage.List[0].ID).NotTo(Equal(secondPage.List[0].ID))
		})

		It("should order audit logs", func() {
			s := sdk.GetSDK()
			loginAsAdmin()
			orderBy := "created_at"
			logs, err := s.Audits().List(&sdk.ListParams{OrderBy: &orderBy})
			Expect(err).NotTo(HaveOccurred())
			prev := int64(0)
			for _, entry := range logs.List {
				Expect(entry.CreatedAt).To(BeNumerically(">=", prev))
				prev = entry.CreatedAt
			}
		})
	})
})
