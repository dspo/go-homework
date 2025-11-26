package conformance

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	sdk "github.com/dspo/go-homework/sdk"
)

var _ = Describe("Teams", func() {
	Context("Team CRUD Operations", Ordered, func() {
		var teamID int
		var memberUser, leaderUser, outsiderUser *sdk.User
		var memberPass, leaderPass, outsiderPass string

		BeforeAll(func() {
			s := sdk.GetSDK()
			memberUser, memberPass = createAndSetupUser(helperUniqueName("team_member"), "pass")
			leaderUser, leaderPass = createAndSetupUser(helperUniqueName("team_leader"), "pass")
			outsiderUser, outsiderPass = createAndSetupUser(helperUniqueName("team_outsider"), "pass")

			loginAsAdmin()
			team, err := s.Teams().Create(&sdk.CreateTeamRequest{Name: helperUniqueName("team_base"), Desc: helperStringPtr("base team")})
			Expect(err).NotTo(HaveOccurred())
			teamID = team.ID

			Expect(s.Teams().AddUser(teamID, memberUser.ID)).NotTo(HaveOccurred())
			Expect(s.Teams().AddUser(teamID, leaderUser.ID)).NotTo(HaveOccurred())
			Expect(s.Teams().UpdateLeader(teamID, helperIntPtr(leaderUser.ID))).NotTo(HaveOccurred())
		})

		AfterAll(func() {
			s := sdk.GetSDK()
			loginAsAdmin()
			_ = s.Teams().Delete(teamID)
			_ = s.Users().Delete(memberUser.ID)
			_ = s.Users().Delete(leaderUser.ID)
			_ = s.Users().Delete(outsiderUser.ID)
		})

		It("should create team by admin", func() {
			s := sdk.GetSDK()
			loginAsAdmin()
			team, err := s.Teams().Create(&sdk.CreateTeamRequest{Name: helperUniqueName("team_create"), Desc: helperStringPtr("created by admin")})
			Expect(err).NotTo(HaveOccurred())
			Expect(team.ID).To(BeNumerically(">", 0))
			DeferCleanup(func() {
				loginAsAdmin()
				_ = s.Teams().Delete(team.ID)
			})
		})

		It("should fail to create team by normal user", func() {
			s := sdk.GetSDK()
			Expect(s.Auth().LoginWithUsername(memberUser.Username, memberPass)).NotTo(HaveOccurred())
			_, err := s.Teams().Create(&sdk.CreateTeamRequest{Name: helperUniqueName("team_fail")})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("403"))
		})

		It("should list all teams by admin", func() {
			s := sdk.GetSDK()
			loginAsAdmin()
			teams, err := s.Teams().List(nil)
			Expect(err).NotTo(HaveOccurred())
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
			s := sdk.GetSDK()
			Expect(s.Auth().LoginWithUsername(memberUser.Username, memberPass)).NotTo(HaveOccurred())
			teams, err := s.Teams().List(nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(teams.Total).To(Equal(1))
			Expect(teams.List[0].ID).To(Equal(teamID))
		})

		It("should get team details by admin", func() {
			s := sdk.GetSDK()
			loginAsAdmin()
			team, err := s.Teams().Get(teamID)
			Expect(err).NotTo(HaveOccurred())
			Expect(team.ID).To(Equal(teamID))
			Expect(team.Leader).NotTo(BeNil())
			Expect(team.Leader.ID).To(Equal(leaderUser.ID))
		})

		It("should get team details by member", func() {
			s := sdk.GetSDK()
			Expect(s.Auth().LoginWithUsername(memberUser.Username, memberPass)).NotTo(HaveOccurred())
			team, err := s.Teams().Get(teamID)
			Expect(err).NotTo(HaveOccurred())
			Expect(team.ID).To(Equal(teamID))
		})

		It("should fail to get team details by non-member", func() {
			s := sdk.GetSDK()
			Expect(s.Auth().LoginWithUsername(outsiderUser.Username, outsiderPass)).NotTo(HaveOccurred())
			_, err := s.Teams().Get(teamID)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("403"))
		})

		It("should update team by admin", func() {
			s := sdk.GetSDK()
			loginAsAdmin()
			newName := helperUniqueName("team_admin_update")
			newDesc := helperStringPtr("updated by admin")
			team, err := s.Teams().Update(teamID, &sdk.UpdateTeamRequest{Name: newName, Desc: newDesc})
			Expect(err).NotTo(HaveOccurred())
			Expect(team.Name).To(Equal(newName))
			Expect(team.Desc).NotTo(BeNil())
			Expect(*team.Desc).To(Equal(*newDesc))
		})

		It("should update team by leader", func() {
			s := sdk.GetSDK()
			Expect(s.Auth().LoginWithUsername(leaderUser.Username, leaderPass)).NotTo(HaveOccurred())
			newDesc := helperStringPtr("leader updated desc")
			team, err := s.Teams().Update(teamID, &sdk.UpdateTeamRequest{Name: helperUniqueName("team_leader_update"), Desc: newDesc})
			Expect(err).NotTo(HaveOccurred())
			Expect(team.Leader).NotTo(BeNil())
			Expect(team.Leader.ID).To(Equal(leaderUser.ID))
		})

		It("should fail to update team by normal member", func() {
			s := sdk.GetSDK()
			Expect(s.Auth().LoginWithUsername(memberUser.Username, memberPass)).NotTo(HaveOccurred())
			_, err := s.Teams().Update(teamID, &sdk.UpdateTeamRequest{Name: helperUniqueName("team_member_update")})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("403"))
		})

		It("should delete team by admin", func() {
			s := sdk.GetSDK()
			loginAsAdmin()
			team, err := s.Teams().Create(&sdk.CreateTeamRequest{Name: helperUniqueName("team_delete_admin")})
			Expect(err).NotTo(HaveOccurred())
			proj, err := s.Teams().CreateProject(team.ID, &sdk.CreateProjectRequest{Name: helperUniqueName("team_delete_proj")})
			Expect(err).NotTo(HaveOccurred())
			Expect(s.Teams().Delete(team.ID)).NotTo(HaveOccurred())
			_, err = s.Teams().Get(team.ID)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("404"))
			_, err = s.Projects().Get(proj.ID)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("404"))
		})

		It("should delete team by leader", func() {
			s := sdk.GetSDK()
			loginAsAdmin()
			team, err := s.Teams().Create(&sdk.CreateTeamRequest{Name: helperUniqueName("team_delete_leader")})
			Expect(err).NotTo(HaveOccurred())
			Expect(s.Teams().AddUser(team.ID, leaderUser.ID)).NotTo(HaveOccurred())
			Expect(s.Teams().UpdateLeader(team.ID, helperIntPtr(leaderUser.ID))).NotTo(HaveOccurred())

			Expect(s.Auth().LoginWithUsername(leaderUser.Username, leaderPass)).NotTo(HaveOccurred())
			Expect(s.Teams().Delete(team.ID)).NotTo(HaveOccurred())
			loginAsAdmin()
			_, err = s.Teams().Get(team.ID)
			Expect(err).To(HaveOccurred())
		})
	})

	Context("Team Members Management", Ordered, func() {
		var teamID, otherTeamID int
		var leaderUser, userA, userB, invisibleUser, extraUser *sdk.User
		var leaderPass, userAPass, extraUserPass string
		var extraAdded bool

		BeforeAll(func() {
			s := sdk.GetSDK()
			leaderUser, leaderPass = createAndSetupUser(helperUniqueName("team_member_leader"), "pass")
			userA, userAPass = createAndSetupUser(helperUniqueName("team_member_a"), "pass")
			userB, _ = createAndSetupUser(helperUniqueName("team_member_b"), "pass")
			invisibleUser, _ = createAndSetupUser(helperUniqueName("team_member_invisible"), "pass")
			extraUser, extraUserPass = createAndSetupUser(helperUniqueName("team_member_extra"), "pass")

			loginAsAdmin()
			team, err := s.Teams().Create(&sdk.CreateTeamRequest{Name: helperUniqueName("team_members"), Desc: helperStringPtr("member tests")})
			Expect(err).NotTo(HaveOccurred())
			teamID = team.ID

			otherTeam, err := s.Teams().Create(&sdk.CreateTeamRequest{Name: helperUniqueName("team_hidden")})
			Expect(err).NotTo(HaveOccurred())
			otherTeamID = otherTeam.ID
			Expect(s.Teams().AddUser(otherTeamID, invisibleUser.ID)).NotTo(HaveOccurred())
		})

		AfterAll(func() {
			s := sdk.GetSDK()
			loginAsAdmin()
			_ = s.Teams().Delete(teamID)
			_ = s.Teams().Delete(otherTeamID)
			_ = s.Users().Delete(leaderUser.ID)
			_ = s.Users().Delete(userA.ID)
			_ = s.Users().Delete(userB.ID)
			_ = s.Users().Delete(invisibleUser.ID)
			_ = s.Users().Delete(extraUser.ID)
		})

		It("should list team members initially empty", func() {
			s := sdk.GetSDK()
			loginAsAdmin()
			members, err := s.Teams().ListUsers(teamID, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(members.Total).To(Equal(0))

			Expect(s.Teams().AddUser(teamID, leaderUser.ID)).NotTo(HaveOccurred())
			Expect(s.Teams().UpdateLeader(teamID, helperIntPtr(leaderUser.ID))).NotTo(HaveOccurred())
		})

		It("should add member by admin", func() {
			s := sdk.GetSDK()
			loginAsAdmin()
			Expect(s.Teams().AddUser(teamID, userA.ID)).NotTo(HaveOccurred())
			members, err := s.Teams().ListUsers(teamID, nil)
			Expect(err).NotTo(HaveOccurred())
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
			s := sdk.GetSDK()
			Expect(s.Auth().LoginWithUsername(leaderUser.Username, leaderPass)).NotTo(HaveOccurred())
			Expect(s.Teams().AddUser(teamID, userB.ID)).NotTo(HaveOccurred())
		})

		It("should fail to add invisible user by leader", func() {
			s := sdk.GetSDK()
			Expect(s.Auth().LoginWithUsername(leaderUser.Username, leaderPass)).NotTo(HaveOccurred())
			err := s.Teams().AddUser(teamID, invisibleUser.ID)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Or(ContainSubstring("403"), ContainSubstring("404")))
		})

		It("should fail to add member by normal user", func() {
			s := sdk.GetSDK()
			Expect(s.Auth().LoginWithUsername(userA.Username, userAPass)).NotTo(HaveOccurred())
			err := s.Teams().AddUser(teamID, extraUser.ID)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("403"))
		})

		It("should search members by name", func() {
			s := sdk.GetSDK()
			loginAsAdmin()
			params := &sdk.ListParams{Name: helperStringPtr(userA.Username)}
			members, err := s.Teams().ListUsers(teamID, params)
			Expect(err).NotTo(HaveOccurred())
			Expect(members.Total).To(Equal(1))
			Expect(members.List[0].ID).To(Equal(userA.ID))
		})

		It("should paginate team members", func() {
			s := sdk.GetSDK()
			loginAsAdmin()
			if !extraAdded {
				Expect(s.Teams().AddUser(teamID, extraUser.ID)).NotTo(HaveOccurred())
				extraAdded = true
			}

			page := 1
			pageSize := 2
			params := &sdk.ListParams{Page: &page, PageSize: &pageSize}
			pageOne, err := s.Teams().ListUsers(teamID, params)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(pageOne.List)).To(BeNumerically("<=", pageSize))
			page = 2
			pageTwo, err := s.Teams().ListUsers(teamID, params)
			Expect(err).NotTo(HaveOccurred())
			if len(pageTwo.List) > 0 {
				Expect(pageOne.List[0].ID).NotTo(Equal(pageTwo.List[0].ID))
			}
		})

		It("should remove member by admin", func() {
			s := sdk.GetSDK()
			loginAsAdmin()
			Expect(s.Teams().RemoveUser(teamID, userA.ID)).NotTo(HaveOccurred())
			members, err := s.Teams().ListUsers(teamID, nil)
			Expect(err).NotTo(HaveOccurred())
			for _, u := range members.List {
				Expect(u.ID).NotTo(Equal(userA.ID))
			}
		})

		It("should remove member by leader", func() {
			s := sdk.GetSDK()
			Expect(s.Auth().LoginWithUsername(leaderUser.Username, leaderPass)).NotTo(HaveOccurred())
			Expect(s.Teams().RemoveUser(teamID, userB.ID)).NotTo(HaveOccurred())
		})

		It("should fail to remove member by normal user", func() {
			s := sdk.GetSDK()
			Expect(s.Auth().LoginWithUsername(extraUser.Username, extraUserPass)).NotTo(HaveOccurred())
			err := s.Teams().RemoveUser(teamID, leaderUser.ID)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("403"))
		})
	})

	Context("Team Leader Management", Ordered, func() {
		var teamID int
		var userLeader, userMember *sdk.User
		var leaderPass, memberPass string

		BeforeAll(func() {
			s := sdk.GetSDK()
			userLeader, leaderPass = createAndSetupUser(helperUniqueName("leader_manage_leader"), "pass")
			userMember, memberPass = createAndSetupUser(helperUniqueName("leader_manage_member"), "pass")

			loginAsAdmin()
			team, err := s.Teams().Create(&sdk.CreateTeamRequest{Name: helperUniqueName("team_leader_mgmt")})
			Expect(err).NotTo(HaveOccurred())
			teamID = team.ID
			Expect(s.Teams().AddUser(teamID, userLeader.ID)).NotTo(HaveOccurred())
			Expect(s.Teams().AddUser(teamID, userMember.ID)).NotTo(HaveOccurred())
		})

		AfterAll(func() {
			s := sdk.GetSDK()
			loginAsAdmin()
			_ = s.Teams().Delete(teamID)
			_ = s.Users().Delete(userLeader.ID)
			_ = s.Users().Delete(userMember.ID)
		})

		It("should have no leader initially", func() {
			s := sdk.GetSDK()
			loginAsAdmin()
			team, err := s.Teams().Get(teamID)
			Expect(err).NotTo(HaveOccurred())
			Expect(team.Leader).To(BeNil())
		})

		It("should set leader by admin", func() {
			s := sdk.GetSDK()
			loginAsAdmin()
			team, err := s.Teams().UpdateLeader(teamID, helperIntPtr(userLeader.ID))
			Expect(err).NotTo(HaveOccurred())
			Expect(team.Leader).NotTo(BeNil())
			Expect(team.Leader.ID).To(Equal(userLeader.ID))
		})

		It("should show team leader role assigned", func() {
			s := sdk.GetSDK()
			Expect(s.Auth().LoginWithUsername(userLeader.Username, leaderPass)).NotTo(HaveOccurred())
			me, err := s.Me().Get()
			Expect(err).NotTo(HaveOccurred())
			Expect(helperRolesContain(me.Roles, "team leader")).To(BeTrue())
		})

		It("should change leader by admin", func() {
			s := sdk.GetSDK()
			loginAsAdmin()
			team, err := s.Teams().UpdateLeader(teamID, helperIntPtr(userMember.ID))
			Expect(err).NotTo(HaveOccurred())
			Expect(team.Leader.ID).To(Equal(userMember.ID))

			Expect(s.Auth().LoginWithUsername(userLeader.Username, leaderPass)).NotTo(HaveOccurred())
			leaderMe, err := s.Me().Get()
			Expect(err).NotTo(HaveOccurred())
			Expect(helperRolesContain(leaderMe.Roles, "team leader")).To(BeFalse())

			Expect(s.Auth().LoginWithUsername(userMember.Username, memberPass)).NotTo(HaveOccurred())
			memberMe, err := s.Me().Get()
			Expect(err).NotTo(HaveOccurred())
			Expect(helperRolesContain(memberMe.Roles, "team leader")).To(BeTrue())
		})

		It("should change leader by current leader", func() {
			s := sdk.GetSDK()
			Expect(s.Auth().LoginWithUsername(userMember.Username, memberPass)).NotTo(HaveOccurred())
			team, err := s.Teams().UpdateLeader(teamID, helperIntPtr(userLeader.ID))
			Expect(err).NotTo(HaveOccurred())
			Expect(team.Leader.ID).To(Equal(userLeader.ID))
		})

		It("should fail to change leader by normal member", func() {
			s := sdk.GetSDK()
			Expect(s.Auth().LoginWithUsername(userMember.Username, memberPass)).NotTo(HaveOccurred())
			_, err := s.Teams().UpdateLeader(teamID, helperIntPtr(userMember.ID))
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("403"))
		})

		It("should clear leader", func() {
			s := sdk.GetSDK()
			loginAsAdmin()
			team, err := s.Teams().UpdateLeader(teamID, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(team.Leader).To(BeNil())
		})

		It("should clear leader when leader exits team", func() {
			s := sdk.GetSDK()
			loginAsAdmin()
			Expect(s.Teams().AddUser(teamID, userLeader.ID)).NotTo(HaveOccurred())
			Expect(s.Teams().UpdateLeader(teamID, helperIntPtr(userLeader.ID))).NotTo(HaveOccurred())

			Expect(s.Auth().LoginWithUsername(userLeader.Username, leaderPass)).NotTo(HaveOccurred())
			Expect(s.Me().ExitTeam(teamID)).NotTo(HaveOccurred())

			loginAsAdmin()
			team, err := s.Teams().Get(teamID)
			Expect(err).NotTo(HaveOccurred())
			Expect(team.Leader).To(BeNil())
		})

		It("should clear leader when leader is removed", func() {
			s := sdk.GetSDK()
			loginAsAdmin()
			Expect(s.Teams().AddUser(teamID, userLeader.ID)).NotTo(HaveOccurred())
			Expect(s.Teams().UpdateLeader(teamID, helperIntPtr(userLeader.ID))).NotTo(HaveOccurred())
			Expect(s.Teams().RemoveUser(teamID, userLeader.ID)).NotTo(HaveOccurred())
			team, err := s.Teams().Get(teamID)
			Expect(err).NotTo(HaveOccurred())
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
			s := sdk.GetSDK()
			leaderUser, leaderPass = createAndSetupUser(helperUniqueName("team_proj_leader"), "pass")
			memberUser, memberPass = createAndSetupUser(helperUniqueName("team_proj_member"), "pass")

			loginAsAdmin()
			team, err := s.Teams().Create(&sdk.CreateTeamRequest{Name: helperUniqueName("team_projects_suite")})
			Expect(err).NotTo(HaveOccurred())
			teamID = team.ID
			Expect(s.Teams().AddUser(teamID, leaderUser.ID)).NotTo(HaveOccurred())
			Expect(s.Teams().AddUser(teamID, memberUser.ID)).NotTo(HaveOccurred())
			Expect(s.Teams().UpdateLeader(teamID, helperIntPtr(leaderUser.ID))).NotTo(HaveOccurred())
		})

		AfterAll(func() {
			s := sdk.GetSDK()
			loginAsAdmin()
			_ = s.Teams().Delete(teamID)
			_ = s.Users().Delete(leaderUser.ID)
			_ = s.Users().Delete(memberUser.ID)
		})

		It("should list team projects initially empty", func() {
			s := sdk.GetSDK()
			loginAsAdmin()
			projects, err := s.Teams().ListProjects(teamID, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(projects.Total).To(Equal(0))
		})

		It("should create project in team by admin", func() {
			s := sdk.GetSDK()
			loginAsAdmin()
			projectAdminName = helperUniqueName("team_proj_admin")
			project, err := s.Teams().CreateProject(teamID, &sdk.CreateProjectRequest{Name: projectAdminName})
			Expect(err).NotTo(HaveOccurred())
			Expect(project.Status).To(Equal("WAIT_FOR_SCHEDULE"))
			projectAdminID = project.ID
		})

		It("should create project by team leader", func() {
			s := sdk.GetSDK()
			Expect(s.Auth().LoginWithUsername(leaderUser.Username, leaderPass)).NotTo(HaveOccurred())
			projectLeaderName = helperUniqueName("team_proj_leader")
			project, err := s.Teams().CreateProject(teamID, &sdk.CreateProjectRequest{Name: projectLeaderName})
			Expect(err).NotTo(HaveOccurred())
			projectLeaderID = project.ID

			loginAsAdmin()
			Expect(s.Projects().AddUser(projectAdminID, memberUser.ID)).NotTo(HaveOccurred())
		})

		It("should fail to create project by normal member", func() {
			s := sdk.GetSDK()
			Expect(s.Auth().LoginWithUsername(memberUser.Username, memberPass)).NotTo(HaveOccurred())
			_, err := s.Teams().CreateProject(teamID, &sdk.CreateProjectRequest{Name: helperUniqueName("team_proj_member_fail")})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("403"))
		})

		It("should list team projects", func() {
			s := sdk.GetSDK()
			loginAsAdmin()
			projects, err := s.Teams().ListProjects(teamID, nil)
			Expect(err).NotTo(HaveOccurred())
			ids := []int{}
			for _, p := range projects.List {
				ids = append(ids, p.ID)
			}
			Expect(ids).To(ContainElements(projectAdminID, projectLeaderID))
		})

		It("should filter projects by participation", func() {
			s := sdk.GetSDK()
			Expect(s.Auth().LoginWithUsername(memberUser.Username, memberPass)).NotTo(HaveOccurred())
			partIn := true
			params := &sdk.ListParams{PartIn: helperBoolPtr(partIn)}
			projectsIn, err := s.Teams().ListProjects(teamID, params)
			Expect(err).NotTo(HaveOccurred())
			Expect(projectsIn.Total).To(Equal(1))
			Expect(projectsIn.List[0].ID).To(Equal(projectAdminID))

			partIn = false
			params = &sdk.ListParams{PartIn: helperBoolPtr(partIn)}
			projectsOut, err := s.Teams().ListProjects(teamID, params)
			Expect(err).NotTo(HaveOccurred())
			Expect(projectsOut.Total).To(BeNumerically(">=", 1))
			ids := []int{}
			for _, p := range projectsOut.List {
				ids = append(ids, p.ID)
			}
			Expect(ids).To(ContainElement(projectLeaderID))
			Expect(ids).NotTo(ContainElement(projectAdminID))
		})

		It("should search projects by name", func() {
			s := sdk.GetSDK()
			loginAsAdmin()
			params := &sdk.ListParams{Name: helperStringPtr(projectAdminName)}
			projects, err := s.Teams().ListProjects(teamID, params)
			Expect(err).NotTo(HaveOccurred())
			Expect(projects.Total).To(Equal(1))
			Expect(projects.List[0].ID).To(Equal(projectAdminID))
		})

		It("should cascade delete projects when team deleted", func() {
			s := sdk.GetSDK()
			loginAsAdmin()
			tempTeam, err := s.Teams().Create(&sdk.CreateTeamRequest{Name: helperUniqueName("team_proj_delete")})
			Expect(err).NotTo(HaveOccurred())
			project, err := s.Teams().CreateProject(tempTeam.ID, &sdk.CreateProjectRequest{Name: helperUniqueName("team_proj_delete_proj")})
			Expect(err).NotTo(HaveOccurred())
			Expect(s.Teams().Delete(tempTeam.ID)).NotTo(HaveOccurred())
			_, err = s.Projects().Get(project.ID)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("404"))
		})
	})
})
