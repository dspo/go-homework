package conformance

import (
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/dspo/go-homework/sdk"
)

// Helper function to login as admin
func loginAsAdmin(sdk sdk.SDK) sdk.UserClient {
	return loginWithUsername(sdk, "admin", "admin123")
}

// Helper function to create user and change password
func createAndSetupUser(username, password string) (*sdk.User, string) {
	admin := loginWithUsername(sdk.GetSDK(), "admin", "admin123")

	user, err := admin.Users().Create(username, password)
	Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)

	// Login and change password
	userClient := loginWithUsername(sdk.GetSDK(), username, password)

	newPass := password + "456"
	err = userClient.Me().UpdatePassword(password, newPass)
	Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)

	userClient = loginWithUsername(sdk.GetSDK(), username, newPass)

	return user, newPass
}

var _ = Describe("Authentication", Label("Auth"), func() {
	Context("Login and Logout", func() {
		It("should login with username successfully", func() {
			By("Login with valid username and password")
			client := loginWithUsername(sdk.GetSDK(), "admin", "admin123")

			By("Verify session cookie is set")
			me, err := client.Me().Get()
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			Expect(me.Username).To(Equal("admin"))
		})

		It("should login with email successfully", func() {
			By("Login with valid email and password")
			// First set email for admin
			admin := loginWithUsername(sdk.GetSDK(), "admin", "admin123")

			email := "admin@example.com"
			_, err := admin.Me().Update(&sdk.UpdateMeRequest{Email: &email})
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)

			// Logout and login with email
			Expect(admin.Logout()).NotTo(HaveOccurred())

			client := loginWithEmail("admin@example.com", "admin123")

			By("Verify session cookie is set")
			me, err := client.Me().Get()
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			Expect(me.Username).To(Equal("admin"))
		})

		It("should fail with invalid credentials", func() {
			By("Try to login with wrong password")
			_, err := sdk.GetSDK().LoginWithUsername("admin", "wrongpassword")

			By("Should get 401 Unauthorized")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("401"))
		})

		It("should logout successfully", func() {
			By("Login first")
			client := loginWithUsername(sdk.GetSDK(), "admin", "admin123")

			By("Logout")
			err := client.Logout()
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)

			By("Try to access protected resource")
			_, err = client.Me().Get()

			By("Should get 401 Unauthorized")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("401"))
		})
	})

	Context("Password Management", func() {
		It("should require password change on first login", func() {
			By("Admin creates a new user")
			admin := loginWithUsername(sdk.GetSDK(), "admin", "admin123")

			user, err := admin.Users().Create("testpasswd", "test1234")
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			userID := user.ID

			By("New user login with initial password")
			testUser := loginWithUsername(sdk.GetSDK(), "testpasswd", "test1234")

			By("Try to access /api/me")
			_, err = testUser.Me().Get()

			By("Should get 403 indicating password change required")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("403"))

			// Cleanup
			admin = loginWithUsername(sdk.GetSDK(), "admin", "admin123")
			err = admin.Users().Delete(userID)
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
		})

		It("should invalidate session after password change", func() {
			By("User login")
			admin := loginWithUsername(sdk.GetSDK(), "admin", "admin123")

			// Create test user and change password
			user, err := admin.Users().Create("testpasswd2", "test1234")
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			userID := user.ID

			testClient := loginWithUsername(sdk.GetSDK(), "testpasswd2", "test1234")

			By("User changes password")
			err = testClient.Me().UpdatePassword("test1234", "test4567")
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)

			By("Try to access /api/me with old session")
			_, err = testClient.Me().Get()

			By("Should get 401 because session is invalidated")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("401"))

			By("Login with new password should succeed")
			testClient = loginWithUsername(sdk.GetSDK(), "testpasswd2", "test4567")

			_, err = testClient.Me().Get()
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)

			// Cleanup
			admin = loginWithUsername(sdk.GetSDK(), "admin", "admin123")
			err = admin.Users().Delete(userID)
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
		})
	})

	Context("Health Check", func() {
		It("should access healthz without authentication", func() {
			By("GET /healthz without login")
			err := sdk.GetSDK().Healthz()

			By("Should get 200 OK")
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
		})
	})
})

var _ = Describe("Users", Label("User"), func() {
	Context("User CRUD Operations", Ordered, func() {
		var userID int

		It("should create user by admin", func() {
			By("Admin login")
			admin := loginWithUsername(sdk.GetSDK(), "admin", "admin123")

			By("POST /api/users with username and password")
			user, err := admin.Users().Create("testuser_crud", "test1234")

			By("Should return 200 with user object")
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			Expect(user.Username).To(Equal("testuser_crud"))

			By("Save user ID for later tests")
			userID = user.ID
		})

		It("should fail to create user by normal user", func() {
			By("Normal user login")
			// Create a normal user first
			admin := loginWithUsername(sdk.GetSDK(), "admin", "admin123")
			normalUser, err := admin.Users().Create("normaluser", "test1234")
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)

			// Change password for normal user
			normalClient := loginWithUsername(sdk.GetSDK(), "normaluser", "test1234")
			err = normalClient.Me().UpdatePassword("test1234", "test4567")
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)

			normalClient = loginWithUsername(sdk.GetSDK(), "normaluser", "test4567")

			By("Try to POST /api/users")
			_, err = normalClient.Users().Create("anotheruser", "test1234")

			By("Should get 403 Forbidden")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("403"))

			// Cleanup
			admin = loginWithUsername(sdk.GetSDK(), "admin", "admin123")
			_ = admin.Users().Delete(normalUser.ID)
		})

		It("should list users with admin privileges", func() {
			By("Admin login")
			admin := loginWithUsername(sdk.GetSDK(), "admin", "admin123")

			By("GET /api/users")
			users, err := admin.Users().List(nil)

			By("Should return all users including newly created")
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
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
			admin := loginWithUsername(sdk.GetSDK(), "admin", "admin123")

			isolatedUser, err := admin.Users().Create("isolated", "test1234")
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)

			isolatedClient := loginWithUsername(sdk.GetSDK(), "isolated", "test1234")
			err = isolatedClient.Me().UpdatePassword("test1234", "test4567")
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			isolatedClient = loginWithUsername(sdk.GetSDK(), "isolated", "test4567")

			By("GET /api/users")
			users, err := isolatedClient.Users().List(nil)

			By("Should return empty list or only self")
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			// Should only see themselves
			Expect(users.Total).To(Equal(1))
			Expect(users.List[0].Username).To(Equal("isolated"))

			// Cleanup
			admin = loginWithUsername(sdk.GetSDK(), "admin", "admin123")
			_ = admin.Users().Delete(isolatedUser.ID)
		})

		It("should get user details", func() {
			By("Admin login")
			admin := loginWithUsername(sdk.GetSDK(), "admin", "admin123")

			By("GET /api/users/{user_id}")
			user, err := admin.Users().Get(userID)

			By("Should return user details")
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			Expect(user.ID).To(Equal(userID))
			Expect(user.Username).To(Equal("testuser_crud"))
		})

		It("should fail to get invisible user details", func() {
			By("UserA login (in teamA)")
			admin := loginWithUsername(sdk.GetSDK(), "admin", "admin123")

			// Create teamA and teamB
			teamA, err := admin.Teams().Create(&sdk.CreateTeamRequest{Name: "TeamA"})
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			teamB, err := admin.Teams().Create(&sdk.CreateTeamRequest{Name: "TeamB"})
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)

			// Create userA in teamA and userB in teamB
			userA, err := admin.Users().Create("userA_visibility", "test1234")
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			userB, err := admin.Users().Create("userB_visibility", "test1234")
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)

			err = admin.Teams().AddUser(teamA.ID, userA.ID)
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			err = admin.Teams().AddUser(teamB.ID, userB.ID)
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)

			// UserA login
			userAClient := loginWithUsername(sdk.GetSDK(), "userA_visibility", "test1234")
			err = userAClient.Me().UpdatePassword("test1234", "test4567")
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			userAClient = loginWithUsername(sdk.GetSDK(), "userA_visibility", "test4567")

			By("Try to GET /api/users/{userB_id} where userB not in teamA")
			_, err = userAClient.Users().Get(userB.ID)

			By("Should get 403 Forbidden")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("403"))

			// Cleanup
			admin = loginWithUsername(sdk.GetSDK(), "admin", "admin123")
			_ = admin.Teams().Delete(teamA.ID)
			_ = admin.Teams().Delete(teamB.ID)
			_ = admin.Users().Delete(userA.ID)
			_ = admin.Users().Delete(userB.ID)
		})

		It("should delete user by admin", func() {
			By("Admin login")
			s, err := sdk.GetSDK().Guest().LoginWithUsername("admin", "admin123")
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)

			By("DELETE /api/users/{user_id}")
			err = s.Users().Delete(userID)

			By("Should return 200")
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)

			By("Verify user is deleted by getting 404")
			_, err = s.Users().Get(userID)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("404"))
		})

		It("should fail to delete admin user", func() {
			By("Admin login")
			admin := loginWithUsername(sdk.GetSDK(), "admin", "admin123")

			// Get admin user ID
			me, err := admin.Me().Get()
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)

			By("Try to DELETE /api/users/{admin_id}")
			err = admin.Users().Delete(me.ID)

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
			admin := loginWithUsername(sdk.GetSDK(), "admin", "admin123")

			tA, err := admin.Teams().Create(&sdk.CreateTeamRequest{Name: helperUniqueName("teamA")})
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			teamAID = tA.ID

			tB, err := admin.Teams().Create(&sdk.CreateTeamRequest{Name: helperUniqueName("teamB")})
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			teamBID = tB.ID

			userA, passA = createAndSetupUser(helperUniqueName("userA"), "passA1234")
			userB, _ = createAndSetupUser(helperUniqueName("userB"), "passB1234")
			userC, passC = createAndSetupUser(helperUniqueName("userC"), "passC1234")

			admin = loginWithUsername(sdk.GetSDK(), "admin", "admin123")
			Expect(admin.Teams().AddUser(teamAID, userA.ID)).NotTo(HaveOccurred())
			Expect(admin.Teams().AddUser(teamBID, userB.ID)).NotTo(HaveOccurred())
			Expect(admin.Teams().AddUser(teamAID, userC.ID)).NotTo(HaveOccurred())
			Expect(admin.Teams().AddUser(teamBID, userC.ID)).NotTo(HaveOccurred())
		})

		AfterAll(func() {
			admin := loginWithUsername(sdk.GetSDK(), "admin", "admin123")
			_ = admin.Teams().Delete(teamAID)
			_ = admin.Teams().Delete(teamBID)
			_ = admin.Users().Delete(userA.ID)
			_ = admin.Users().Delete(userB.ID)
			_ = admin.Users().Delete(userC.ID)
		})

		It("should see users in same team", func() {
			client := loginWithUsername(sdk.GetSDK(), userA.Username, passA)
			users, err := client.Users().List(nil)
			Expect(err).NotTo(HaveOccurred(), "failed to list users: %v", err)

			visibleIDs := make(map[int]bool)
			for _, u := range users.List {
				visibleIDs[u.ID] = true
			}

			Expect(visibleIDs[userA.ID]).To(BeTrue())
			Expect(visibleIDs[userC.ID]).To(BeTrue())
			Expect(visibleIDs[userB.ID]).To(BeFalse())
		})

		It("should see users in multiple teams", func() {
			client := loginWithUsername(sdk.GetSDK(), userC.Username, passC)
			users, err := client.Users().List(nil)
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)

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
			user, pass := createAndSetupUser(helperUniqueName("me_user"), "pass1234")
			DeferCleanup(func() {
				_ = loginWithUsername(sdk.GetSDK(), "admin", "admin123").Users().Delete(user.ID)
			})

			client := loginWithUsername(sdk.GetSDK(), user.Username, pass)
			me, err := client.Me().Get()
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			Expect(me.ID).To(Equal(user.ID))
			Expect(me.Username).To(Equal(user.Username))
			Expect(me.Roles).NotTo(BeEmpty())
		})

		It("should update current user info", func() {
			user, pass := createAndSetupUser(helperUniqueName("me_update"), "pass1234")
			DeferCleanup(func() {
				_ = loginWithUsername(sdk.GetSDK(), "admin", "admin123").Users().Delete(user.ID)
			})

			client := loginWithUsername(sdk.GetSDK(), user.Username, pass)
			email := helperUniqueName("me_email") + "@example.com"
			nickname := helperUniqueName("nick")
			logo := "https://logo.example.com/" + helperUniqueName("logo")
			updated, err := client.Me().Update(&sdk.UpdateMeRequest{
				Email:    Ptr(email),
				Nickname: Ptr(nickname),
				Logo:     Ptr(logo),
			})
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			Expect(updated.Email).NotTo(BeNil())
			Expect(*updated.Email).To(Equal(email))
			Expect(updated.Nickname).NotTo(BeNil())
			Expect(*updated.Nickname).To(Equal(nickname))
			Expect(updated.Logo).NotTo(BeNil())
			Expect(*updated.Logo).To(Equal(logo))

			me, err := client.Me().Get()
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			Expect(me.Email).NotTo(BeNil())
			Expect(*me.Email).To(Equal(email))
			Expect(me.Nickname).NotTo(BeNil())
			Expect(*me.Nickname).To(Equal(nickname))
		})

		It("should list my teams", func() {
			s := sdk.GetSDK().Guest()
			user, pass := createAndSetupUser(helperUniqueName("me_teams"), "pass1234")
			DeferCleanup(func() {
				s = loginAsAdmin(s)
				_ = s.Users().Delete(user.ID)
			})

			s = loginAsAdmin(s)
			teamA, err := s.Teams().Create(&sdk.CreateTeamRequest{Name: helperUniqueName("teamA")})
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			teamB, err := s.Teams().Create(&sdk.CreateTeamRequest{Name: helperUniqueName("teamB")})
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			DeferCleanup(func() {
				s = loginAsAdmin(s)
				_ = s.Teams().Delete(teamA.ID)
				_ = s.Teams().Delete(teamB.ID)
			})

			Expect(s.Teams().AddUser(teamA.ID, user.ID)).NotTo(HaveOccurred())
			Expect(s.Teams().AddUser(teamB.ID, user.ID)).NotTo(HaveOccurred())

			s, err = s.LoginWithUsername(user.Username, pass)
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			teams, err := s.Me().ListTeams(nil)
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			Expect(teams.Total).To(Equal(2))
			ids := []int{}
			for _, t := range teams.List {
				ids = append(ids, t.ID)
			}
			Expect(ids).To(ContainElements(teamA.ID, teamB.ID))
		})

		It("should list my leading teams", func() {
			s := sdk.GetSDK().Guest()
			user, pass := createAndSetupUser(helperUniqueName("me_lead"), "pass1234")
			DeferCleanup(func() {
				s = loginAsAdmin(s)
				_ = s.Users().Delete(user.ID)
			})

			s = loginAsAdmin(s)
			myLeadTeam, err := s.Teams().Create(&sdk.CreateTeamRequest{Name: helperUniqueName("team_lead")})
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			myNormalTeam, err := s.Teams().Create(&sdk.CreateTeamRequest{Name: helperUniqueName("team_member")})
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			DeferCleanup(func() {
				s = loginAsAdmin(s)
				_ = s.Teams().Delete(myLeadTeam.ID)
				_ = s.Teams().Delete(myNormalTeam.ID)
			})

			Expect(s.Teams().AddUser(myLeadTeam.ID, user.ID)).NotTo(HaveOccurred())
			Expect(s.Teams().AddUser(myNormalTeam.ID, user.ID)).NotTo(HaveOccurred())

			team, err := s.Teams().UpdateLeader(myLeadTeam.ID, Ptr(user.ID))
			Expect(err).NotTo(HaveOccurred())
			Expect(team.Leader).NotTo(BeNil())
			Expect(team.Leader.ID).To(Equal(user.ID))

			s, err = s.LoginWithUsername(user.Username, pass)
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			leading := true
			teams, err := s.Me().ListTeams(&leading)
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			Expect(teams.Total).To(Equal(1))
			Expect(teams.List[0].ID).To(Equal(myLeadTeam.ID))
		})

		It("should list my non-leading teams", func() {
			s := sdk.GetSDK().Guest()
			user, pass := createAndSetupUser(helperUniqueName("me_nonlead"), "pass1234")
			DeferCleanup(func() {
				s = loginAsAdmin(s)
				_ = s.Users().Delete(user.ID)
			})

			s = loginAsAdmin(s)
			teamLeader, err := s.Teams().Create(&sdk.CreateTeamRequest{Name: helperUniqueName("team_leader")})
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			teamMember, err := s.Teams().Create(&sdk.CreateTeamRequest{Name: helperUniqueName("team_member")})
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			DeferCleanup(func() {
				s = loginAsAdmin(s)
				_ = s.Teams().Delete(teamLeader.ID)
				_ = s.Teams().Delete(teamMember.ID)
			})

			Expect(s.Teams().AddUser(teamLeader.ID, user.ID)).NotTo(HaveOccurred())
			Expect(s.Teams().AddUser(teamMember.ID, user.ID)).NotTo(HaveOccurred())
			team, err := s.Teams().UpdateLeader(teamLeader.ID, Ptr(user.ID))
			Expect(err).NotTo(HaveOccurred())
			Expect(team.Leader).NotTo(BeNil())
			Expect(team.Leader.ID).To(Equal(user.ID))

			s, err = s.LoginWithUsername(user.Username, pass)
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			leading := false
			teams, err := s.Me().ListTeams(&leading)
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			Expect(teams.Total).To(Equal(1))
			Expect(teams.List[0].ID).To(Equal(teamMember.ID))
		})

		It("should exit from team", func() {
			s := sdk.GetSDK().Guest()
			user, pass := createAndSetupUser(helperUniqueName("me_exit_team"), "pass1234")
			DeferCleanup(func() {
				s = loginAsAdmin(s)
				_ = s.Users().Delete(user.ID)
			})

			s = loginAsAdmin(s)
			team, err := s.Teams().Create(&sdk.CreateTeamRequest{Name: helperUniqueName("team_exit")})
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			DeferCleanup(func() {
				s = loginAsAdmin(s)
				_ = s.Teams().Delete(team.ID)
			})

			Expect(s.Teams().AddUser(team.ID, user.ID)).NotTo(HaveOccurred())

			s, err = s.LoginWithUsername(user.Username, pass)
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			Expect(s.Me().ExitTeam(team.ID)).NotTo(HaveOccurred())
			teams, err := s.Me().ListTeams(nil)
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			Expect(teams.Total).To(Equal(0))
		})

		It("should clear leader when leader exits team", func() {
			s := sdk.GetSDK().Guest()
			user, pass := createAndSetupUser(helperUniqueName("me_exit_leader"), "pass1234")
			DeferCleanup(func() {
				s = loginAsAdmin(s)
				_ = s.Users().Delete(user.ID)
			})

			s = loginAsAdmin(s)
			team, err := s.Teams().Create(&sdk.CreateTeamRequest{Name: helperUniqueName("team_leader_exit")})
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			DeferCleanup(func() {
				s = loginAsAdmin(s)
				_ = s.Teams().Delete(team.ID)
			})

			Expect(s.Teams().AddUser(team.ID, user.ID)).NotTo(HaveOccurred())
			team, err = s.Teams().UpdateLeader(team.ID, Ptr(user.ID))
			Expect(err).NotTo(HaveOccurred())
			Expect(team.Leader).NotTo(BeNil())
			Expect(team.Leader.ID).To(Equal(user.ID))

			s, err = s.LoginWithUsername(user.Username, pass)
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			Expect(s.Me().ExitTeam(team.ID)).NotTo(HaveOccurred())

			s = loginAsAdmin(s)
			teamInfo, err := s.Teams().Get(team.ID)
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			Expect(teamInfo.Leader).To(BeNil())
		})

		It("should list my projects", func() {
			s := sdk.GetSDK().Guest()
			user, pass := createAndSetupUser(helperUniqueName("me_projects"), "pass1234")
			DeferCleanup(func() {
				s = loginAsAdmin(s)
				_ = s.Users().Delete(user.ID)
			})

			s = loginAsAdmin(s)
			team, err := s.Teams().Create(&sdk.CreateTeamRequest{Name: helperUniqueName("team_projects")})
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			DeferCleanup(func() {
				s = loginAsAdmin(s)
				_ = s.Teams().Delete(team.ID)
			})

			Expect(s.Teams().AddUser(team.ID, user.ID)).NotTo(HaveOccurred())
			projA, err := s.Teams().CreateProject(team.ID, &sdk.CreateProjectRequest{Name: helperUniqueName("projectA")})
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			projB, err := s.Teams().CreateProject(team.ID, &sdk.CreateProjectRequest{Name: helperUniqueName("projectB")})
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			Expect(s.Projects().AddUser(projA.ID, user.ID)).NotTo(HaveOccurred())
			Expect(s.Projects().AddUser(projB.ID, user.ID)).NotTo(HaveOccurred())

			s, err = s.LoginWithUsername(user.Username, pass)
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			projects, err := s.Me().ListProjects(nil)
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			Expect(projects.Total).To(Equal(2))
			var ids []int
			for _, p := range projects.List {
				ids = append(ids, p.ID)
			}
			Expect(ids).To(ContainElements(projA.ID, projB.ID))
		})

		It("should filter projects by team", func() {
			s := sdk.GetSDK().Guest()
			user, pass := createAndSetupUser(helperUniqueName("me_projects_filter"), "pass1234")
			DeferCleanup(func() {
				s = loginAsAdmin(s)
				_ = s.Users().Delete(user.ID)
			})

			s = loginAsAdmin(s)
			teamA, err := s.Teams().Create(&sdk.CreateTeamRequest{Name: helperUniqueName("team_filter_a")})
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			teamB, err := s.Teams().Create(&sdk.CreateTeamRequest{Name: helperUniqueName("team_filter_b")})
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			DeferCleanup(func() {
				s = loginAsAdmin(s)
				_ = s.Teams().Delete(teamA.ID)
				_ = s.Teams().Delete(teamB.ID)
			})

			Expect(s.Teams().AddUser(teamA.ID, user.ID)).NotTo(HaveOccurred())
			Expect(s.Teams().AddUser(teamB.ID, user.ID)).NotTo(HaveOccurred())
			projA, err := s.Teams().CreateProject(teamA.ID, &sdk.CreateProjectRequest{Name: helperUniqueName("projA")})
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			projB, err := s.Teams().CreateProject(teamB.ID, &sdk.CreateProjectRequest{Name: helperUniqueName("projB")})
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			Expect(s.Projects().AddUser(projA.ID, user.ID)).NotTo(HaveOccurred())
			Expect(s.Projects().AddUser(projB.ID, user.ID)).NotTo(HaveOccurred())

			s, err = s.LoginWithUsername(user.Username, pass)
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			paramsA := &sdk.ListParams{TeamIds: []int{teamA.ID}}
			projectsA, err := s.Me().ListProjects(paramsA)
			Expect(err).NotTo(HaveOccurred(), "failed to ListProjects: %v", err)
			Expect(projectsA.Total).To(Equal(1))
			Expect(projectsA.List[0].ID).To(Equal(projA.ID))

			paramsBoth := &sdk.ListParams{TeamIds: []int{teamA.ID, teamB.ID}}
			projectsBoth, err := s.Me().ListProjects(paramsBoth)
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			Expect(projectsBoth.Total).To(Equal(2))
			ids := []int{}
			for _, p := range projectsBoth.List {
				ids = append(ids, p.ID)
			}
			Expect(ids).To(ContainElements(projA.ID, projB.ID))
		})

		It("should exit from project", func() {
			s := sdk.GetSDK().Guest()
			user, pass := createAndSetupUser(helperUniqueName("me_exit_project"), "pass1234")
			DeferCleanup(func() {
				s = loginAsAdmin(s)
				_ = s.Users().Delete(user.ID)
			})

			s = loginAsAdmin(s)
			team, err := s.Teams().Create(&sdk.CreateTeamRequest{Name: helperUniqueName("team_project_exit")})
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			DeferCleanup(func() {
				s = loginAsAdmin(s)
				_ = s.Teams().Delete(team.ID)
			})

			Expect(s.Teams().AddUser(team.ID, user.ID)).NotTo(HaveOccurred())
			project, err := s.Teams().CreateProject(team.ID, &sdk.CreateProjectRequest{Name: helperUniqueName("proj_exit")})
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			Expect(s.Projects().AddUser(project.ID, user.ID)).NotTo(HaveOccurred())

			s, err = s.LoginWithUsername(user.Username, pass)
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			Expect(s.Me().ExitProject(project.ID)).NotTo(HaveOccurred())
			projects, err := s.Me().ListProjects(nil)
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			Expect(projects.Total).To(Equal(0))

			s = loginAsAdmin(s)
			teamUsers, err := s.Teams().ListUsers(team.ID, nil)
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
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
			s := sdk.GetSDK().Guest()
			s = loginAsAdmin(s)

			var err error
			targetUser, _ = createAndSetupUser(helperUniqueName("user_target"), "pass1234")
			viewerUser, viewerPass = createAndSetupUser(helperUniqueName("user_viewer"), "pass1234")

			s = loginAsAdmin(s)
			teamCommon, err := s.Teams().Create(&sdk.CreateTeamRequest{Name: helperUniqueName("team_common")})
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			teamPrivate, err := s.Teams().Create(&sdk.CreateTeamRequest{Name: helperUniqueName("team_private")})
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			teamCommonID = teamCommon.ID
			teamPrivateID = teamPrivate.ID

			Expect(s.Teams().AddUser(teamCommonID, targetUser.ID)).NotTo(HaveOccurred())
			Expect(s.Teams().AddUser(teamPrivateID, targetUser.ID)).NotTo(HaveOccurred())
			Expect(s.Teams().AddUser(teamCommonID, viewerUser.ID)).NotTo(HaveOccurred())

			projectCommon, err := s.Teams().CreateProject(teamCommonID, &sdk.CreateProjectRequest{Name: helperUniqueName("project_common")})
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			projectPrivate, err := s.Teams().CreateProject(teamPrivateID, &sdk.CreateProjectRequest{Name: helperUniqueName("project_private")})
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			projectCommonID = projectCommon.ID
			projectPrivateID = projectPrivate.ID

			Expect(s.Projects().AddUser(projectCommonID, targetUser.ID)).NotTo(HaveOccurred())
			Expect(s.Projects().AddUser(projectPrivateID, targetUser.ID)).NotTo(HaveOccurred())
			Expect(s.Projects().AddUser(projectCommonID, viewerUser.ID)).NotTo(HaveOccurred())
		})

		AfterAll(func() {
			s := loginAsAdmin(sdk.GetSDK())
			_ = s.Teams().Delete(teamCommonID)
			_ = s.Teams().Delete(teamPrivateID)
			_ = s.Users().Delete(targetUser.ID)
			_ = s.Users().Delete(viewerUser.ID)
		})

		It("should list user's teams", func() {
			s := loginAsAdmin(sdk.GetSDK())
			teams, err := s.Users().ListTeams(targetUser.ID, nil)
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			Expect(teams.Total).To(Equal(2))
			ids := []int{}
			for _, team := range teams.List {
				ids = append(ids, team.ID)
			}
			Expect(ids).To(ContainElements(teamCommonID, teamPrivateID))
		})

		It("should respect visibility when listing user teams", func() {
			s, err := sdk.GetSDK().Guest().LoginWithUsername(viewerUser.Username, viewerPass)
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			teams, err := s.Users().ListTeams(targetUser.ID, nil)
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			Expect(teams.Total).To(Equal(1))
			Expect(teams.List[0].ID).To(Equal(teamCommonID))
		})

		It("should list user's projects", func() {
			s := loginAsAdmin(sdk.GetSDK())
			projects, err := s.Users().ListProjects(targetUser.ID)
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			Expect(projects.Total).To(Equal(2))
			ids := []int{}
			for _, project := range projects.List {
				ids = append(ids, project.ID)
			}
			Expect(ids).To(ContainElements(projectCommonID, projectPrivateID))
		})

		It("should respect visibility when listing user projects", func() {
			s, err := sdk.GetSDK().LoginWithUsername(viewerUser.Username, viewerPass)
			Expect(err).NotTo(HaveOccurred(), "failed to LoginWithUsername: %v", err)
			projects, err := s.Users().ListProjects(targetUser.ID)
			Expect(err).NotTo(HaveOccurred(), "failed to ListProjects: %v", err)
			Expect(projects.Total).To(Equal(1))
			Expect(projects.List[0].ID).To(Equal(projectCommonID))
		})
	})
})

var _ = Describe("Roles", Label("Role"), func() {
	Context("System Roles", func() {
		It("should have 3 system roles initialized", func() {
			s := loginAsAdmin(sdk.GetSDK())
			roles, err := s.Roles().List()
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)

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
			s := loginAsAdmin(sdk.GetSDK())
			roles, err := s.Roles().List()
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)

			var adminRoleID int
			for _, role := range roles.List {
				if role.Name == "admin" {
					adminRoleID = role.ID
					break
				}
			}
			Expect(adminRoleID).NotTo(Equal(0))

			err = s.Roles().Delete(adminRoleID)
			Expect(err).To(Or(
				sdk.HaveOccurredWithStatusCode(http.StatusBadRequest),
				sdk.HaveOccurredWithStatusCode(http.StatusForbidden),
				sdk.HaveOccurredWithStatusCode(http.StatusNotAcceptable),
				sdk.HaveOccurredWithStatusCode(http.StatusConflict),
			))
		})

		It("should not add system roles to users manually", func() {
			s := sdk.GetSDK().Guest()
			user, _ := createAndSetupUser(helperUniqueName("role_sys"), "pass1234")
			DeferCleanup(func() {
				s = loginAsAdmin(s)
				_ = s.Users().Delete(user.ID)
			})

			s = loginAsAdmin(s)
			roles, err := s.Roles().List()
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			var adminRoleID int
			for _, role := range roles.List {
				if role.Name == "admin" {
					adminRoleID = role.ID
					break
				}
			}
			Expect(adminRoleID).NotTo(Equal(0))

			err = s.Users().AddRole(user.ID, adminRoleID)
			Expect(err).To(Or(
				sdk.HaveOccurredWithStatusCode(http.StatusBadRequest),
				sdk.HaveOccurredWithStatusCode(http.StatusForbidden),
				sdk.HaveOccurredWithStatusCode(http.StatusNotAcceptable),
				sdk.HaveOccurredWithStatusCode(http.StatusConflict),
			))
		})

		It("should auto-assign admin role to admin user", func() {
			s := loginAsAdmin(sdk.GetSDK())
			me, err := s.Me().Get()
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			Expect(helperRolesContain(me.Roles, "admin")).To(BeTrue())
		})

		It("should not remove admin role from admin user", func() {
			s := loginAsAdmin(sdk.GetSDK())
			me, err := s.Me().Get()
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)

			roles, err := s.Roles().List()
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			var adminRoleID int
			for _, role := range roles.List {
				if role.Name == "admin" {
					adminRoleID = role.ID
					break
				}
			}
			Expect(adminRoleID).NotTo(Equal(0))

			err = s.Users().RemoveRole(me.ID, adminRoleID)
			Expect(err).To(Or(
				sdk.HaveOccurredWithStatusCode(http.StatusBadRequest),
				sdk.HaveOccurredWithStatusCode(http.StatusForbidden),
				sdk.HaveOccurredWithStatusCode(http.StatusNotAcceptable),
				sdk.HaveOccurredWithStatusCode(http.StatusConflict),
			))
		})

		It("should auto-assign team leader role", func() {
			s := sdk.GetSDK().Guest()
			user, pass := createAndSetupUser(helperUniqueName("role_leader"), "pass1234")
			DeferCleanup(func() {
				s = loginAsAdmin(s)
				_ = s.Users().Delete(user.ID)
			})

			s = loginAsAdmin(s)
			team, err := s.Teams().Create(&sdk.CreateTeamRequest{Name: helperUniqueName("team_leader_role")})
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			DeferCleanup(func() {
				s = loginAsAdmin(s)
				_ = s.Teams().Delete(team.ID)
			})

			Expect(s.Teams().AddUser(team.ID, user.ID)).NotTo(HaveOccurred())
			team, err = s.Teams().UpdateLeader(team.ID, Ptr(user.ID))
			Expect(err).NotTo(HaveOccurred())
			Expect(team.Leader).NotTo(BeNil())
			Expect(team.Leader.ID).To(Equal(user.ID))

			s, err = s.LoginWithUsername(user.Username, pass)
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			me, err := s.Me().Get()
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			Expect(helperRolesContain(me.Roles, "team leader")).To(BeTrue())
		})

		It("should auto-assign normal user role", func() {
			s := sdk.GetSDK().Guest()
			user, _ := createAndSetupUser(helperUniqueName("role_normal"), "pass1234")
			DeferCleanup(func() {
				s = loginAsAdmin(s)
				_ = s.Users().Delete(user.ID)
			})

			s = loginAsAdmin(s)
			stored, err := s.Users().Get(user.ID)
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			Expect(helperRolesContain(stored.Roles, "normal user")).To(BeTrue())
		})
	})

	Context("Custom Roles", Ordered, func() {
		var roleID int
		var roleName string
		var testUser *sdk.User
		var testUserPass string

		BeforeAll(func() {
			testUser, testUserPass = createAndSetupUser(helperUniqueName("role_custom_user"), "pass1234")
			DeferCleanup(func() {
				s := loginAsAdmin(sdk.GetSDK())
				_ = s.Users().Delete(testUser.ID)
			})
		})

		AfterAll(func() {
			if roleID != 0 {
				s := loginAsAdmin(sdk.GetSDK())
				_ = s.Roles().Delete(roleID)
			}
		})

		It("should create custom role by admin", func() {
			s := loginAsAdmin(sdk.GetSDK())
			roleName = helperUniqueName("custom_role")
			role, err := s.Roles().Create(&sdk.CreateRoleRequest{Name: roleName, Desc: Ptr("custom role for tests")})
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			Expect(role.Type).To(Equal("Custom"))
			roleID = role.ID
		})

		It("should fail to create role by normal user", func() {
			s, err := sdk.GetSDK().Guest().LoginWithUsername(testUser.Username, testUserPass)
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			_, err = s.Roles().Create(&sdk.CreateRoleRequest{Name: helperUniqueName("custom_role_fail")})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("403"))
		})

		It("should list all roles", func() {
			s, err := sdk.GetSDK().Guest().LoginWithUsername(testUser.Username, testUserPass)
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			roles, err := s.Roles().List()
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
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
			s := loginAsAdmin(sdk.GetSDK())
			Expect(roleID).NotTo(Equal(0))
			Expect(s.Users().AddRole(testUser.ID, roleID)).NotTo(HaveOccurred())
			stored, err := s.Users().Get(testUser.ID)
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			Expect(helperRolesContain(stored.Roles, roleName)).To(BeTrue())
		})

		It("should remove custom role from user", func() {
			s := loginAsAdmin(sdk.GetSDK())
			Expect(s.Users().RemoveRole(testUser.ID, roleID)).NotTo(HaveOccurred())
			stored, err := s.Users().Get(testUser.ID)
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			Expect(helperRolesContain(stored.Roles, roleName)).To(BeFalse())
		})

		It("should delete custom role", func() {
			s := loginAsAdmin(sdk.GetSDK())
			Expect(s.Roles().Delete(roleID)).NotTo(HaveOccurred())
			roles, err := s.Roles().List()
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			for _, role := range roles.List {
				Expect(role.Name).NotTo(Equal(roleName))
			}
			stored, err := s.Users().Get(testUser.ID)
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			Expect(stored.ID).To(Equal(testUser.ID))
			roleID = 0
		})
	})

	Context("Role Permissions", func() {
		It("should allow only admin to manage roles", func() {
			s := sdk.GetSDK().Guest()
			user, pass := createAndSetupUser(helperUniqueName("role_perm_user"), "pass1234")
			DeferCleanup(func() {
				s = loginAsAdmin(s)
				_ = s.Users().Delete(user.ID)
			})

			s = loginAsAdmin(s)
			role, err := s.Roles().Create(&sdk.CreateRoleRequest{Name: helperUniqueName("role_perm"), Desc: Ptr("perm test")})
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			DeferCleanup(func() {
				s = loginAsAdmin(s)
				_ = s.Roles().Delete(role.ID)
			})

			s, err = s.LoginWithUsername(user.Username, pass)
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			_, err = s.Roles().Create(&sdk.CreateRoleRequest{Name: helperUniqueName("role_perm_fail")})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("403"))

			err = s.Roles().Delete(role.ID)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("403"))
		})

		It("should allow only admin to assign roles", func() {
			s := sdk.GetSDK().Guest()
			actor, pass := createAndSetupUser(helperUniqueName("role_assign_actor"), "pass1234")
			target, _ := createAndSetupUser(helperUniqueName("role_assign_target"), "pass1234")
			DeferCleanup(func() {
				s = loginAsAdmin(s)
				_ = s.Users().Delete(actor.ID)
				_ = s.Users().Delete(target.ID)
			})

			s = loginAsAdmin(s)
			role, err := s.Roles().Create(&sdk.CreateRoleRequest{Name: helperUniqueName("role_assign"), Desc: Ptr("assign test")})
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			DeferCleanup(func() {
				s = loginAsAdmin(s)
				_ = s.Roles().Delete(role.ID)
			})
			s, err = s.LoginWithUsername(actor.Username, pass)
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			err = s.Users().AddRole(target.ID, role.ID)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("403"))
		})
	})
})
