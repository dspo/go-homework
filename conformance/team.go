package conformance

import (
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/dspo/go-homework/sdk"
)

var _ = Describe("Teams", Label("Team"), func() {
	Context("Team CRUD Operations", Ordered, func() {
		var teamID int
		var memberUser, leaderUser, outsiderUser *sdk.User
		var memberPass, leaderPass, outsiderPass string

		BeforeAll(func() {
			s := sdk.GetSDK().Guest()
			memberUser, memberPass = createAndSetupUser(helperUniqueName("team_member"), "pass1234")
			leaderUser, leaderPass = createAndSetupUser(helperUniqueName("team_leader"), "pass1234")
			outsiderUser, outsiderPass = createAndSetupUser(helperUniqueName("team_outsider"), "pass1234")

			s = loginAsAdmin(s)
			team, err := s.Teams().Create(&sdk.CreateTeamRequest{Name: helperUniqueName("team_base"), Desc: Ptr("base team")})
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			teamID = team.ID

			Expect(s.Teams().AddUser(teamID, memberUser.ID)).NotTo(HaveOccurred())
			Expect(s.Teams().AddUser(teamID, leaderUser.ID)).NotTo(HaveOccurred())
			_, err = s.Teams().UpdateLeader(teamID, Ptr(leaderUser.ID))
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
		})

		AfterAll(func() {
			s := loginAsAdmin(sdk.GetSDK())
			_ = s.Teams().Delete(teamID)
			_ = s.Users().Delete(memberUser.ID)
			_ = s.Users().Delete(leaderUser.ID)
			_ = s.Users().Delete(outsiderUser.ID)
		})

		It("should create team by admin", func() {
			s := loginAsAdmin(sdk.GetSDK())
			team, err := s.Teams().Create(&sdk.CreateTeamRequest{Name: helperUniqueName("team_create"), Desc: Ptr("created by admin")})
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			Expect(team.ID).To(BeNumerically(">", 0))
			DeferCleanup(func() {
				s := loginAsAdmin(s)
				_ = s.Teams().Delete(team.ID)
			})
		})

		It("should fail to create team by normal user", func() {
			s, err := sdk.GetSDK().Guest().LoginWithUsername(memberUser.Username, memberPass)
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			_, err = s.Teams().Create(&sdk.CreateTeamRequest{Name: helperUniqueName("team_fail")})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("403"))
		})

		It("should list all teams by admin", func() {
			s := loginAsAdmin(sdk.GetSDK())
			teams, err := s.Teams().List(nil)
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			found := false
			for _, team := range teams.List {
				if team.ID == teamID {
					found = true
					break
				}
			}
			Expect(found).To(BeTrue())
		})

		It("should list only my teams for normal user", func() {
			s, err := sdk.GetSDK().Guest().LoginWithUsername(memberUser.Username, memberPass)
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			teams, err := s.Teams().List(nil)
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			Expect(teams.Total).To(Equal(1))
			Expect(teams.List[0].ID).To(Equal(teamID))
		})

		It("should get team details by admin", func() {
			s := loginAsAdmin(sdk.GetSDK())
			team, err := s.Teams().Get(teamID)
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			Expect(team.ID).To(Equal(teamID))
			Expect(team.Leader).NotTo(BeNil())
			Expect(team.Leader.ID).To(Equal(leaderUser.ID))
		})

		It("should get team details by member", func() {
			s, err := sdk.GetSDK().Guest().LoginWithUsername(memberUser.Username, memberPass)
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			team, err := s.Teams().Get(teamID)
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			Expect(team.ID).To(Equal(teamID))
		})

		It("should fail to get team details by non-member", func() {
			s, err := sdk.GetSDK().Guest().LoginWithUsername(outsiderUser.Username, outsiderPass)
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			_, err = s.Teams().Get(teamID)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("403"))
		})

		It("should update team by admin", func() {
			s := loginAsAdmin(sdk.GetSDK())
			newName := helperUniqueName("team_admin_update")
			newDesc := Ptr("updated by admin")
			team, err := s.Teams().Update(teamID, &sdk.UpdateTeamRequest{Name: &newName, Desc: newDesc})
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			Expect(team.Name).To(Equal(newName))
			Expect(team.Desc).NotTo(BeNil())
			Expect(*team.Desc).To(Equal(*newDesc))
		})

		It("should update team by leader", func() {
			s, err := sdk.GetSDK().Guest().LoginWithUsername(leaderUser.Username, leaderPass)
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			newName := helperUniqueName("team_leader_update")
			newDesc := Ptr("leader updated desc")
			team, err := s.Teams().Update(teamID, &sdk.UpdateTeamRequest{Name: &newName, Desc: newDesc})
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			Expect(team.Leader).NotTo(BeNil())
			Expect(team.Leader.ID).To(Equal(leaderUser.ID))
		})

		It("should fail to update team by normal member", func() {
			s, err := sdk.GetSDK().Guest().LoginWithUsername(memberUser.Username, memberPass)
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			newName := helperUniqueName("team_member_update")
			_, err = s.Teams().Update(teamID, &sdk.UpdateTeamRequest{Name: &newName})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("403"))
		})

		It("should delete team by admin", func() {
			s := loginAsAdmin(sdk.GetSDK())
			team, err := s.Teams().Create(&sdk.CreateTeamRequest{Name: helperUniqueName("team_delete_admin")})
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			proj, err := s.Teams().CreateProject(team.ID, &sdk.CreateProjectRequest{Name: helperUniqueName("team_delete_proj")})
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			Expect(s.Teams().Delete(team.ID)).NotTo(HaveOccurred())
			_, err = s.Teams().Get(team.ID)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("404"))
			_, err = s.Projects().Get(proj.ID)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("404"))
		})

		It("should delete team by leader", func() {
			s := loginAsAdmin(sdk.GetSDK())
			team, err := s.Teams().Create(&sdk.CreateTeamRequest{Name: helperUniqueName("team_delete_leader")})
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			Expect(s.Teams().AddUser(team.ID, leaderUser.ID)).NotTo(HaveOccurred())
			_, err = s.Teams().UpdateLeader(team.ID, Ptr(leaderUser.ID))
			Expect(err).NotTo(HaveOccurred())

			s, err = s.LoginWithUsername(leaderUser.Username, leaderPass)
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			Expect(s.Teams().Delete(team.ID)).NotTo(HaveOccurred())
			s = loginAsAdmin(s)
			_, err = s.Teams().Get(team.ID)
			Expect(err).To(HaveOccurred())
		})
	})

	Context("Team Members Management", Ordered, func() {
		var teamID, otherTeamID, visibleTeamID int
		var leaderUser, userA, userB, invisibleUser, extraUser *sdk.User
		var leaderPass, userAPass, extraUserPass string
		var extraAdded bool

		BeforeAll(func() {
			s := sdk.GetSDK().Guest()
			leaderUser, leaderPass = createAndSetupUser(helperUniqueName("team_member_leader"), "pass1234")
			userA, userAPass = createAndSetupUser(helperUniqueName("team_member_a"), "pass1234")
			userB, _ = createAndSetupUser(helperUniqueName("team_member_b"), "pass1234")
			invisibleUser, _ = createAndSetupUser(helperUniqueName("team_member_invisible"), "pass1234")
			extraUser, extraUserPass = createAndSetupUser(helperUniqueName("team_member_extra"), "pass1234")

			s = loginAsAdmin(s)
			team, err := s.Teams().Create(&sdk.CreateTeamRequest{Name: helperUniqueName("team_members"), Desc: Ptr("member tests")})
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			teamID = team.ID

			// 为 leader 和 userB 创建一个共同 Team，满足可见性再由 leader 添加到目标 Team。
			visibleTeam, err := s.Teams().Create(&sdk.CreateTeamRequest{Name: helperUniqueName("team_visible_bridge")})
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			visibleTeamID = visibleTeam.ID
			Expect(s.Teams().AddUser(visibleTeamID, leaderUser.ID)).NotTo(HaveOccurred())
			Expect(s.Teams().AddUser(visibleTeamID, userB.ID)).NotTo(HaveOccurred())

			otherTeam, err := s.Teams().Create(&sdk.CreateTeamRequest{Name: helperUniqueName("team_hidden")})
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			otherTeamID = otherTeam.ID
			Expect(s.Teams().AddUser(otherTeamID, invisibleUser.ID)).NotTo(HaveOccurred())
		})

		AfterAll(func() {
			s := loginAsAdmin(sdk.GetSDK())
			_ = s.Teams().Delete(visibleTeamID)
			_ = s.Teams().Delete(teamID)
			_ = s.Teams().Delete(otherTeamID)
			_ = s.Users().Delete(leaderUser.ID)
			_ = s.Users().Delete(userA.ID)
			_ = s.Users().Delete(userB.ID)
			_ = s.Users().Delete(invisibleUser.ID)
			_ = s.Users().Delete(extraUser.ID)
		})

		It("should list team members initially empty", func() {
			s := loginAsAdmin(sdk.GetSDK())
			members, err := s.Teams().ListUsers(teamID, nil)
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			Expect(members.Total).To(Equal(0))

			Expect(s.Teams().AddUser(teamID, leaderUser.ID)).NotTo(HaveOccurred())
			_, err = s.Teams().UpdateLeader(teamID, Ptr(leaderUser.ID))
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
		})

		It("should add member by admin", func() {
			s := loginAsAdmin(sdk.GetSDK())
			Expect(s.Teams().AddUser(teamID, userA.ID)).NotTo(HaveOccurred())
			members, err := s.Teams().ListUsers(teamID, nil)
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			found := false
			for _, u := range members.List {
				if u.ID == userA.ID {
					found = true
					break
				}
			}
			Expect(found).To(BeTrue())
		})

		It("should add member by leader", func() {
			s, err := sdk.GetSDK().Guest().LoginWithUsername(leaderUser.Username, leaderPass)
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			Expect(s.Teams().AddUser(teamID, userB.ID)).NotTo(HaveOccurred())
		})

		It("should fail to add invisible user by leader", func() {
			s, err := sdk.GetSDK().Guest().LoginWithUsername(leaderUser.Username, leaderPass)
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			err = s.Teams().AddUser(teamID, invisibleUser.ID)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Or(ContainSubstring("403"), ContainSubstring("404")))
		})

		It("should fail to add member by normal user", func() {
			s, err := sdk.GetSDK().Guest().LoginWithUsername(userA.Username, userAPass)
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			err = s.Teams().AddUser(teamID, extraUser.ID)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("403"))
		})

		It("should search members by name", func() {
			s := loginAsAdmin(sdk.GetSDK())
			params := &sdk.ListParams{Name: Ptr(userA.Username)}
			members, err := s.Teams().ListUsers(teamID, params)
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			Expect(members.Total).To(Equal(1))
			Expect(members.List[0].ID).To(Equal(userA.ID))
		})

		It("should paginate team members", func() {
			s := loginAsAdmin(sdk.GetSDK())
			if !extraAdded {
				Expect(s.Teams().AddUser(teamID, extraUser.ID)).NotTo(HaveOccurred())
				extraAdded = true
			}

			page := 1
			pageSize := 2
			params := &sdk.ListParams{Page: &page, PageSize: &pageSize}
			pageOne, err := s.Teams().ListUsers(teamID, params)
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			Expect(len(pageOne.List)).To(BeNumerically("<=", pageSize))
			page = 2
			pageTwo, err := s.Teams().ListUsers(teamID, params)
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			if len(pageTwo.List) > 0 {
				Expect(pageOne.List[0].ID).NotTo(Equal(pageTwo.List[0].ID))
			}
		})

		It("should remove member by admin", func() {
			s := loginAsAdmin(sdk.GetSDK())
			Expect(s.Teams().RemoveUser(teamID, userA.ID)).NotTo(HaveOccurred())
			members, err := s.Teams().ListUsers(teamID, nil)
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			for _, u := range members.List {
				Expect(u.ID).NotTo(Equal(userA.ID))
			}
		})

		It("should remove member by leader", func() {
			s, err := sdk.GetSDK().Guest().LoginWithUsername(leaderUser.Username, leaderPass)
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			Expect(s.Teams().RemoveUser(teamID, userB.ID)).NotTo(HaveOccurred())
		})

		It("should fail to remove member by normal user", func() {
			s, err := sdk.GetSDK().Guest().LoginWithUsername(extraUser.Username, extraUserPass)
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			err = s.Teams().RemoveUser(teamID, leaderUser.ID)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("403"))
		})
	})

	Context("Team Leader Management", Ordered, func() {
		var teamID int
		var userLeader, userMember *sdk.User
		var leaderPass, memberPass string

		BeforeAll(func() {
			s := sdk.GetSDK().Guest()
			userLeader, leaderPass = createAndSetupUser(helperUniqueName("leader_manage_leader"), "pass1234")
			userMember, memberPass = createAndSetupUser(helperUniqueName("leader_manage_member"), "pass1234")

			s = loginAsAdmin(s)
			team, err := s.Teams().Create(&sdk.CreateTeamRequest{Name: helperUniqueName("team_leader_mgmt")})
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			teamID = team.ID
			Expect(s.Teams().AddUser(teamID, userLeader.ID)).NotTo(HaveOccurred())
			Expect(s.Teams().AddUser(teamID, userMember.ID)).NotTo(HaveOccurred())
		})

		AfterAll(func() {
			s := sdk.GetSDK().Guest()
			s = loginAsAdmin(s)
			_ = s.Teams().Delete(teamID)
			_ = s.Users().Delete(userLeader.ID)
			_ = s.Users().Delete(userMember.ID)
		})

		It("should have no leader initially", func() {
			s := loginAsAdmin(sdk.GetSDK())
			team, err := s.Teams().Get(teamID)
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			Expect(team.Leader).To(BeNil())
		})

		It("should set leader by admin", func() {
			s := loginAsAdmin(sdk.GetSDK())
			team, err := s.Teams().UpdateLeader(teamID, Ptr(userLeader.ID))
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			Expect(team.Leader).NotTo(BeNil())
			Expect(team.Leader.ID).To(Equal(userLeader.ID))
		})

		It("should show team leader role assigned", func() {
			s, err := sdk.GetSDK().Guest().LoginWithUsername(userLeader.Username, leaderPass)
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			me, err := s.Me().Get()
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			Expect(helperRolesContain(me.Roles, "team leader")).To(BeTrue())
		})

		It("should change leader by admin", func() {
			s := loginAsAdmin(sdk.GetSDK())
			team, err := s.Teams().UpdateLeader(teamID, Ptr(userMember.ID))
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			Expect(team.Leader.ID).To(Equal(userMember.ID))

			s, err = s.LoginWithUsername(userLeader.Username, leaderPass)
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			leaderMe, err := s.Me().Get()
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			Expect(helperRolesContain(leaderMe.Roles, "team leader")).To(BeFalse())

			s, err = s.LoginWithUsername(userMember.Username, memberPass)
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			memberMe, err := s.Me().Get()
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			Expect(helperRolesContain(memberMe.Roles, "team leader")).To(BeTrue())
		})

		It("should change leader by current leader", func() {
			s, err := sdk.GetSDK().Guest().LoginWithUsername(userMember.Username, memberPass)
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			_, err = s.Teams().UpdateLeader(teamID, Ptr(userLeader.ID))
			Expect(err).NotTo(HaveOccurred())
		})

		It("should fail to change leader by normal member", func() {
			s, err := sdk.GetSDK().Guest().LoginWithUsername(userMember.Username, memberPass)
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			_, err = s.Teams().UpdateLeader(teamID, Ptr(userMember.ID))
			Expect(err).To(sdk.HaveOccurredWithStatusCode(http.StatusForbidden))
		})

		It("should clear leader", func() {
			s := loginAsAdmin(sdk.GetSDK())
			team, err := s.Teams().UpdateLeader(teamID, nil)
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			Expect(team.Leader).To(BeNil())
		})

		It("should clear leader when leader exits team", func() {
			s := loginAsAdmin(sdk.GetSDK())
			Expect(s.Teams().AddUser(teamID, userLeader.ID)).NotTo(HaveOccurred())
			_, err := s.Teams().UpdateLeader(teamID, Ptr(userLeader.ID))
			Expect(err).NotTo(HaveOccurred())

			s, err = s.LoginWithUsername(userLeader.Username, leaderPass)
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			Expect(s.Me().ExitTeam(teamID)).NotTo(HaveOccurred())

			s = loginAsAdmin(s)
			team, err := s.Teams().Get(teamID)
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			Expect(team.Leader).To(BeNil())
		})

		It("should clear leader when leader is removed", func() {
			s := loginAsAdmin(sdk.GetSDK())
			Expect(s.Teams().AddUser(teamID, userLeader.ID)).NotTo(HaveOccurred())
			_, err := s.Teams().UpdateLeader(teamID, Ptr(userLeader.ID))
			Expect(err).NotTo(HaveOccurred())
			Expect(s.Teams().RemoveUser(teamID, userLeader.ID)).NotTo(HaveOccurred())
			team, err := s.Teams().Get(teamID)
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			Expect(team.Leader).To(BeNil())
		})
	})

	Context("Team-Project Relationship", Ordered, func() {
		var teamID int
		var leaderUser, memberUser *sdk.User
		var leaderPass, memberPass string
		var projectAdminID, projectLeaderID int
		var projectAdminName, projectLeaderName string

		BeforeAll(func() {
			s := sdk.GetSDK().Guest()
			leaderUser, leaderPass = createAndSetupUser(helperUniqueName("team_proj_leader"), "pass1234")
			memberUser, memberPass = createAndSetupUser(helperUniqueName("team_proj_member"), "pass1234")

			s = loginAsAdmin(s)
			team, err := s.Teams().Create(&sdk.CreateTeamRequest{Name: helperUniqueName("team_projects_suite")})
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			teamID = team.ID
			Expect(s.Teams().AddUser(teamID, leaderUser.ID)).NotTo(HaveOccurred())
			Expect(s.Teams().AddUser(teamID, memberUser.ID)).NotTo(HaveOccurred())
			_, err = s.Teams().UpdateLeader(teamID, Ptr(leaderUser.ID))
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
		})

		AfterAll(func() {
			s := loginAsAdmin(sdk.GetSDK())
			_ = s.Teams().Delete(teamID)
			_ = s.Users().Delete(leaderUser.ID)
			_ = s.Users().Delete(memberUser.ID)
		})

		It("should list team projects initially empty", func() {
			s := loginAsAdmin(sdk.GetSDK())
			projects, err := s.Teams().ListProjects(teamID, nil)
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			Expect(projects.Total).To(Equal(0))
		})

		It("should create project in team by admin", func() {
			s := loginAsAdmin(sdk.GetSDK())
			projectAdminName = helperUniqueName("team_proj_admin")
			project, err := s.Teams().CreateProject(teamID, &sdk.CreateProjectRequest{Name: projectAdminName})
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			Expect(project.Status).To(Equal("WAIT_FOR_SCHEDULE"))
			projectAdminID = project.ID
		})

		It("should create project by team leader", func() {
			s, err := sdk.GetSDK().Guest().LoginWithUsername(leaderUser.Username, leaderPass)
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			projectLeaderName = helperUniqueName("team_proj_leader")
			project, err := s.Teams().CreateProject(teamID, &sdk.CreateProjectRequest{Name: projectLeaderName})
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			projectLeaderID = project.ID

			s = loginAsAdmin(s)
			Expect(s.Projects().AddUser(projectAdminID, memberUser.ID)).NotTo(HaveOccurred())
		})

		It("should fail to create project by normal member", func() {
			s, err := sdk.GetSDK().Guest().LoginWithUsername(memberUser.Username, memberPass)
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			_, err = s.Teams().CreateProject(teamID, &sdk.CreateProjectRequest{Name: helperUniqueName("team_proj_member_fail")})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("403"))
		})

		It("should list team projects", func() {
			s := loginAsAdmin(sdk.GetSDK())
			projects, err := s.Teams().ListProjects(teamID, nil)
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			ids := []int{}
			for _, p := range projects.List {
				ids = append(ids, p.ID)
			}
			Expect(ids).To(ContainElements(projectAdminID, projectLeaderID))
		})

		It("should filter projects by participation", func() {
			s, err := sdk.GetSDK().Guest().LoginWithUsername(memberUser.Username, memberPass)
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			partIn := true
			params := &sdk.ListParams{PartIn: Ptr(partIn)}
			projectsIn, err := s.Teams().ListProjects(teamID, params)
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			Expect(projectsIn.Total).To(Equal(1))
			Expect(projectsIn.List[0].ID).To(Equal(projectAdminID))

			partIn = false
			params = &sdk.ListParams{PartIn: Ptr(partIn)}
			projectsOut, err := s.Teams().ListProjects(teamID, params)
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			Expect(projectsOut.Total).To(BeNumerically(">=", 1))
			ids := []int{}
			for _, p := range projectsOut.List {
				ids = append(ids, p.ID)
			}
			Expect(ids).To(ContainElement(projectLeaderID))
			Expect(ids).NotTo(ContainElement(projectAdminID))
		})

		It("should search projects by name", func() {
			s := loginAsAdmin(sdk.GetSDK())
			params := &sdk.ListParams{Name: Ptr(projectAdminName)}
			projects, err := s.Teams().ListProjects(teamID, params)
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			Expect(projects.Total).To(Equal(1))
			Expect(projects.List[0].ID).To(Equal(projectAdminID))
		})

		It("should cascade delete projects when team deleted", func() {
			s := loginAsAdmin(sdk.GetSDK())
			tempTeam, err := s.Teams().Create(&sdk.CreateTeamRequest{Name: helperUniqueName("team_proj_delete")})
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			project, err := s.Teams().CreateProject(tempTeam.ID, &sdk.CreateProjectRequest{Name: helperUniqueName("team_proj_delete_proj")})
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			Expect(s.Teams().Delete(tempTeam.ID)).NotTo(HaveOccurred())
			_, err = s.Projects().Get(project.ID)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("404"))
		})
	})
})
