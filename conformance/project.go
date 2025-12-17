package conformance

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/dspo/go-homework/sdk"
)

var _ = PDescribe("Projects", func() {
	Context("Project CRUD Operations", Ordered, func() {
		var teamID, projectID int
		var leaderUser, memberUser, outsiderUser *sdk.User
		var leaderPass, memberPass, outsiderPass string

		BeforeAll(func() {
			s := sdk.GetSDK().Guest()
			leaderUser, leaderPass = createAndSetupUser(helperUniqueName("project_crud_leader"), "pass")
			memberUser, memberPass = createAndSetupUser(helperUniqueName("project_crud_member"), "pass")
			outsiderUser, outsiderPass = createAndSetupUser(helperUniqueName("project_crud_outsider"), "pass")

			team, err := s.Teams().Create(&sdk.CreateTeamRequest{Name: helperUniqueName("team_project_crud")})
			Expect(err).NotTo(HaveOccurred(), "failed to Create team: %v", err)
			teamID = team.ID
			Expect(s.Teams().AddUser(teamID, leaderUser.ID)).NotTo(HaveOccurred())
			Expect(s.Teams().AddUser(teamID, memberUser.ID)).NotTo(HaveOccurred())
			Expect(s.Teams().UpdateLeader(teamID, Ptr(leaderUser.ID))).NotTo(HaveOccurred())
		})

		AfterAll(func() {
			s := loginAsAdmin(sdk.GetSDK())
			_ = s.Teams().Delete(teamID)
			_ = s.Users().Delete(leaderUser.ID)
			_ = s.Users().Delete(memberUser.ID)
			_ = s.Users().Delete(outsiderUser.ID)
		})

		It("should create project in team", func() {
			s := sdk.GetSDK().Guest()
			_, err := s.LoginWithUsername(leaderUser.Username, leaderPass)
			Expect(err).NotTo(HaveOccurred(), "failed to LoginWithUsername: %v", err)
			project, err := s.Teams().CreateProject(teamID, &sdk.CreateProjectRequest{Name: helperUniqueName("project_crud_main")})
			Expect(err).NotTo(HaveOccurred(), "failed to CreateProject: %v", err)
			Expect(project.Status).To(Equal("WAIT_FOR_SCHEDULE"))
			projectID = project.ID
		})

		It("should get project details by admin", func() {
			s := loginAsAdmin(sdk.GetSDK())
			project, err := s.Projects().Get(projectID)
			Expect(err).NotTo(HaveOccurred(), "failed to Get project: %v", err)
			Expect(project.ID).To(Equal(projectID))
		})

		It("should get project by team leader", func() {
			s := sdk.GetSDK().Guest()
			_, err := s.LoginWithUsername(leaderUser.Username, leaderPass)
			Expect(err).NotTo(HaveOccurred(), "failed to LoginWithUsername: %v", err)
			project, err := s.Projects().Get(projectID)
			Expect(err).NotTo(HaveOccurred(), "failed to Get project: %v", err)
			Expect(project.ID).To(Equal(projectID))
		})

		It("should get project by participant", func() {
			s := loginAsAdmin(sdk.GetSDK())
			Expect(s.Projects().AddUser(projectID, memberUser.ID)).NotTo(HaveOccurred())
			_, err := s.LoginWithUsername(memberUser.Username, memberPass)
			Expect(err).NotTo(HaveOccurred(), "failed to LoginWithUsername: %v", err)
			project, err := s.Projects().Get(projectID)
			Expect(err).NotTo(HaveOccurred(), "failed to Get project: %v", err)
			Expect(project.ID).To(Equal(projectID))
		})

		It("should fail to get project by non-participant", func() {
			s := sdk.GetSDK().Guest()
			_, err := s.LoginWithUsername(outsiderUser.Username, outsiderPass)
			Expect(err).NotTo(HaveOccurred(), "failed to LoginWithUsername: %v", err)
			_, err = s.Projects().Get(projectID)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("403"))
		})

		It("should update project by admin", func() {
			s := loginAsAdmin(sdk.GetSDK())
			status := "IN_PROGRESS"
			project, err := s.Projects().Update(projectID, &sdk.UpdateProjectRequest{
				Name:   helperUniqueName("project_admin_update"),
				Desc:   Ptr("updated by admin"),
				Status: &status,
			})
			Expect(err).NotTo(HaveOccurred(), "failed to Update project: %v", err)
			Expect(project.Status).To(Equal(status))
		})

		It("should update project by team leader", func() {
			s := sdk.GetSDK().Guest()
			_, err := s.LoginWithUsername(leaderUser.Username, leaderPass)
			Expect(err).NotTo(HaveOccurred(), "failed to LoginWithUsername: %v", err)
			project, err := s.Projects().Update(projectID, &sdk.UpdateProjectRequest{
				Name: helperUniqueName("project_leader_update"),
			})
			Expect(err).NotTo(HaveOccurred(), "failed to Update project: %v", err)
			Expect(project.ID).To(Equal(projectID))
		})

		It("should fail to update project by normal member", func() {
			s := sdk.GetSDK().Guest()
			_, err := s.LoginWithUsername(memberUser.Username, memberPass)
			Expect(err).NotTo(HaveOccurred(), "failed to LoginWithUsername: %v", err)
			_, err = s.Projects().Update(projectID, &sdk.UpdateProjectRequest{Name: helperUniqueName("project_member_update")})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("403"))
		})

		It("should patch project status", func() {
			s := sdk.GetSDK().Guest()
			_, err := s.LoginWithUsername(leaderUser.Username, leaderPass)
			Expect(err).NotTo(HaveOccurred(), "failed to LoginWithUsername: %v", err)
			patch := []sdk.PatchProjectRequest{{Op: "replace", Path: "/status", Value: "FINISHED"}}
			project, err := s.Projects().Patch(projectID, patch)
			Expect(err).NotTo(HaveOccurred(), "failed to Patch project: %v", err)
			Expect(project.Status).To(Equal("FINISHED"))
		})

		It("should patch project name", func() {
			s := sdk.GetSDK().Guest()
			_, err := s.LoginWithUsername(leaderUser.Username, leaderPass)
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			newName := helperUniqueName("patched_name")
			patch := []sdk.PatchProjectRequest{{Op: "replace", Path: "/name", Value: newName}}
			project, err := s.Projects().Patch(projectID, patch)
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			Expect(project.Name).To(Equal(newName))
		})

		It("should patch project desc", func() {
			s := sdk.GetSDK().Guest()
			_, err := s.LoginWithUsername(leaderUser.Username, leaderPass)
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			newDesc := "patched desc"
			patch := []sdk.PatchProjectRequest{{Op: "replace", Path: "/desc", Value: newDesc}}
			project, err := s.Projects().Patch(projectID, patch)
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			Expect(project.Desc).NotTo(BeNil())
			Expect(*project.Desc).To(Equal(newDesc))
		})

		It("should fail to patch by normal member", func() {
			s := sdk.GetSDK().Guest()
			_, err := s.LoginWithUsername(memberUser.Username, memberPass)
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			patch := []sdk.PatchProjectRequest{{Op: "replace", Path: "/name", Value: "fail"}}
			_, err = s.Projects().Patch(projectID, patch)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("403"))
		})

		It("should delete project by admin", func() {
			s := loginAsAdmin(sdk.GetSDK())
			project, err := s.Teams().CreateProject(teamID, &sdk.CreateProjectRequest{Name: helperUniqueName("project_delete_admin")})
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			Expect(s.Projects().Delete(project.ID)).NotTo(HaveOccurred())
			_, err = s.Projects().Get(project.ID)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("404"))
		})

		It("should delete project by team leader", func() {
			s := sdk.GetSDK().Guest()
			_, err := s.LoginWithUsername(leaderUser.Username, leaderPass)
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			project, err := s.Teams().CreateProject(teamID, &sdk.CreateProjectRequest{Name: helperUniqueName("project_delete_leader")})
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			Expect(s.Projects().Delete(project.ID)).NotTo(HaveOccurred())
		})

		It("should fail to delete project by normal member", func() {
			s := sdk.GetSDK().Guest()
			_, err := s.LoginWithUsername(memberUser.Username, memberPass)
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			err = s.Projects().Delete(projectID)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("403"))
		})
	})

	Context("Project Members Management", Ordered, func() {
		var teamID, projectID int
		var leaderUser, userA, userB, userC, userD, extraUser *sdk.User
		var leaderPass, extraUserPass string
		var hiddenTeamID int
		var extraAdded bool

		BeforeAll(func() {
			s := sdk.GetSDK().Guest()
			leaderUser, leaderPass = createAndSetupUser(helperUniqueName("project_member_leader"), "pass")
			userA, _ = createAndSetupUser(helperUniqueName("project_member_a"), "pass")
			userB, _ = createAndSetupUser(helperUniqueName("project_member_b"), "pass")
			userC, _ = createAndSetupUser(helperUniqueName("project_member_c"), "pass")
			userD, _ = createAndSetupUser(helperUniqueName("project_member_d"), "pass")
			extraUser, extraUserPass = createAndSetupUser(helperUniqueName("project_member_extra"), "pass")

			s = loginAsAdmin(s)
			team, err := s.Teams().Create(&sdk.CreateTeamRequest{Name: helperUniqueName("team_project_members")})
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			teamID = team.ID
			Expect(s.Teams().AddUser(teamID, leaderUser.ID)).NotTo(HaveOccurred())
			Expect(s.Teams().AddUser(teamID, userA.ID)).NotTo(HaveOccurred())
			Expect(s.Teams().AddUser(teamID, userB.ID)).NotTo(HaveOccurred())
			Expect(s.Teams().UpdateLeader(teamID, Ptr(leaderUser.ID))).NotTo(HaveOccurred())

			project, err := s.Teams().CreateProject(teamID, &sdk.CreateProjectRequest{Name: helperUniqueName("project_members")})
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			projectID = project.ID

			hiddenTeam, err := s.Teams().Create(&sdk.CreateTeamRequest{Name: helperUniqueName("team_hidden_proj")})
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			hiddenTeamID = hiddenTeam.ID
			Expect(s.Teams().AddUser(hiddenTeamID, userD.ID)).NotTo(HaveOccurred())
		})

		AfterAll(func() {
			s := loginAsAdmin(sdk.GetSDK())
			_ = s.Teams().Delete(teamID)
			_ = s.Teams().Delete(hiddenTeamID)
			_ = s.Users().Delete(leaderUser.ID)
			_ = s.Users().Delete(userA.ID)
			_ = s.Users().Delete(userB.ID)
			_ = s.Users().Delete(userC.ID)
			_ = s.Users().Delete(userD.ID)
			_ = s.Users().Delete(extraUser.ID)
		})

		It("should list project members initially empty", func() {
			s := loginAsAdmin(sdk.GetSDK())
			members, err := s.Projects().ListUsers(projectID, nil)
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			Expect(members.Total).To(Equal(0))
		})

		It("should add member by admin", func() {
			s := loginAsAdmin(sdk.GetSDK())
			Expect(s.Projects().AddUser(projectID, userA.ID)).NotTo(HaveOccurred())
			members, err := s.Projects().ListUsers(projectID, nil)
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			ids := []int{}
			for _, u := range members.List {
				ids = append(ids, u.ID)
			}
			Expect(ids).To(ContainElement(userA.ID))
		})

		It("should add member by team leader", func() {
			s := sdk.GetSDK().Guest()
			_, err := s.LoginWithUsername(leaderUser.Username, leaderPass)
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			Expect(s.Projects().AddUser(projectID, userB.ID)).NotTo(HaveOccurred())
		})

		It("should auto-add to team when adding to project", func() {
			s := loginAsAdmin(sdk.GetSDK())
			Expect(s.Projects().AddUser(projectID, userC.ID)).NotTo(HaveOccurred())
			teamMembers, err := s.Teams().ListUsers(teamID, nil)
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			ids := []int{}
			for _, u := range teamMembers.List {
				ids = append(ids, u.ID)
			}
			Expect(ids).To(ContainElement(userC.ID))
		})

		It("should fail to add invisible user by leader", func() {
			s := sdk.GetSDK().Guest()
			_, err := s.LoginWithUsername(leaderUser.Username, leaderPass)
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			err = s.Projects().AddUser(projectID, userD.ID)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Or(ContainSubstring("403"), ContainSubstring("404")))
		})

		It("should list project members with pagination", func() {
			s := loginAsAdmin(sdk.GetSDK())
			page := 1
			pageSize := 2
			params := &sdk.ListParams{Page: &page, PageSize: &pageSize}
			pageOne, err := s.Projects().ListUsers(projectID, params)
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			Expect(len(pageOne.List)).To(BeNumerically("<=", pageSize))
			page = 2
			_, err = s.Projects().ListUsers(projectID, params)
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
		})

		It("should search members by name", func() {
			s := loginAsAdmin(sdk.GetSDK())
			params := &sdk.ListParams{Name: Ptr(userA.Username)}
			members, err := s.Projects().ListUsers(projectID, params)
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			Expect(members.Total).To(Equal(1))
			Expect(members.List[0].ID).To(Equal(userA.ID))
		})

		It("should remove member by admin", func() {
			s := loginAsAdmin(sdk.GetSDK())
			Expect(s.Projects().RemoveUser(projectID, userA.ID)).NotTo(HaveOccurred())
			members, err := s.Projects().ListUsers(projectID, nil)
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			for _, u := range members.List {
				Expect(u.ID).NotTo(Equal(userA.ID))
			}
			teamMembers, err := s.Teams().ListUsers(teamID, nil)
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			ids := []int{}
			for _, u := range teamMembers.List {
				ids = append(ids, u.ID)
			}
			Expect(ids).To(ContainElement(userA.ID))
		})

		It("should remove member by team leader", func() {
			s := sdk.GetSDK().Guest()
			_, err := s.LoginWithUsername(leaderUser.Username, leaderPass)
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			Expect(s.Projects().RemoveUser(projectID, userB.ID)).NotTo(HaveOccurred())
		})

		It("should fail to remove member by normal user", func() {
			s := sdk.GetSDK().Guest()
			if !extraAdded {
				s = loginAsAdmin(s)
				Expect(s.Teams().AddUser(teamID, extraUser.ID)).NotTo(HaveOccurred())
				Expect(s.Projects().AddUser(projectID, extraUser.ID)).NotTo(HaveOccurred())
				extraAdded = true
			}
			_, err := s.LoginWithUsername(extraUser.Username, extraUserPass)
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			err = s.Projects().RemoveUser(projectID, userC.ID)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("403"))
		})
	})

	Context("Project Permissions", Ordered, func() {
		var teamID, projectID int
		var leaderUser, memberUser, outsiderUser *sdk.User
		var leaderPass, memberPass, outsiderPass string

		BeforeAll(func() {
			leaderUser, leaderPass = createAndSetupUser(helperUniqueName("project_perm_leader"), "pass")
			memberUser, memberPass = createAndSetupUser(helperUniqueName("project_perm_member"), "pass")
			outsiderUser, outsiderPass = createAndSetupUser(helperUniqueName("project_perm_out"), "pass")

			s := loginAsAdmin(sdk.GetSDK())
			team, err := s.Teams().Create(&sdk.CreateTeamRequest{Name: helperUniqueName("team_project_permissions")})
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			teamID = team.ID
			Expect(s.Teams().AddUser(teamID, leaderUser.ID)).NotTo(HaveOccurred())
			Expect(s.Teams().AddUser(teamID, memberUser.ID)).NotTo(HaveOccurred())
			Expect(s.Teams().UpdateLeader(teamID, Ptr(leaderUser.ID))).NotTo(HaveOccurred())
			project, err := s.Teams().CreateProject(teamID, &sdk.CreateProjectRequest{Name: helperUniqueName("project_perm_main")})
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			projectID = project.ID
			Expect(s.Projects().AddUser(projectID, memberUser.ID)).NotTo(HaveOccurred())
		})

		AfterAll(func() {
			s := loginAsAdmin(sdk.GetSDK())
			_ = s.Teams().Delete(teamID)
			_ = s.Users().Delete(leaderUser.ID)
			_ = s.Users().Delete(memberUser.ID)
			_ = s.Users().Delete(outsiderUser.ID)
		})

		It("should allow admin to manage any project", func() {
			s := loginAsAdmin(sdk.GetSDK())
			_, err := s.Projects().Get(projectID)
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)

			status := "IN_PROGRESS"
			_, err = s.Projects().Update(projectID, &sdk.UpdateProjectRequest{Name: helperUniqueName("proj_admin_perm"), Status: &status})
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)

			patch := []sdk.PatchProjectRequest{{Op: "replace", Path: "/desc", Value: "admin patch"}}
			_, err = s.Projects().Patch(projectID, patch)
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
		})

		It("should allow leader to manage team projects", func() {
			s := sdk.GetSDK().Guest()
			_, err := s.LoginWithUsername(leaderUser.Username, leaderPass)
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			project, err := s.Teams().CreateProject(teamID, &sdk.CreateProjectRequest{Name: helperUniqueName("project_perm_leader_manage")})
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			_, err = s.Projects().Update(project.ID, &sdk.UpdateProjectRequest{Name: helperUniqueName("project_perm_leader_put")})
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			patch := []sdk.PatchProjectRequest{{Op: "replace", Path: "/status", Value: "FINISHED"}}
			_, err = s.Projects().Patch(project.ID, patch)
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			Expect(s.Projects().Delete(project.ID)).NotTo(HaveOccurred())
		})

		It("should restrict normal member permissions", func() {
			s := sdk.GetSDK().Guest()
			_, err := s.LoginWithUsername(memberUser.Username, memberPass)
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			_, err = s.Projects().Get(projectID)
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			_, err = s.Projects().Update(projectID, &sdk.UpdateProjectRequest{Name: helperUniqueName("project_perm_member_put")})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("403"))
			patch := []sdk.PatchProjectRequest{{Op: "replace", Path: "/name", Value: "deny"}}
			_, err = s.Projects().Patch(projectID, patch)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("403"))
			err = s.Projects().Delete(projectID)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("403"))
		})

		It("should restrict non-participant access", func() {
			s := sdk.GetSDK().Guest()
			_, err := s.LoginWithUsername(outsiderUser.Username, outsiderPass)
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			_, err = s.Projects().Get(projectID)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("403"))
		})
	})

	Context("Me Project APIs", Ordered, func() {
		var teamAID, teamBID, projectA1ID, projectA2ID, projectB1ID int
		var projectA1Name string
		var userX *sdk.User
		var userPass string

		BeforeAll(func() {
			s := loginAsAdmin(sdk.GetSDK())
			teamA, err := s.Teams().Create(&sdk.CreateTeamRequest{Name: helperUniqueName("team_me_a")})
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			teamAID = teamA.ID
			teamB, err := s.Teams().Create(&sdk.CreateTeamRequest{Name: helperUniqueName("team_me_b")})
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			teamBID = teamB.ID

			projectA1, err := s.Teams().CreateProject(teamAID, &sdk.CreateProjectRequest{Name: helperUniqueName("project_A1")})
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			projectA1ID = projectA1.ID
			projectA1Name = projectA1.Name
			projectA2, err := s.Teams().CreateProject(teamAID, &sdk.CreateProjectRequest{Name: helperUniqueName("project_A2")})
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			projectA2ID = projectA2.ID
			projectB1, err := s.Teams().CreateProject(teamBID, &sdk.CreateProjectRequest{Name: helperUniqueName("project_B1")})
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			projectB1ID = projectB1.ID

			userX, userPass = createAndSetupUser(helperUniqueName("project_me_user"), "pass")
			Expect(s.Teams().AddUser(teamAID, userX.ID)).NotTo(HaveOccurred())
			Expect(s.Teams().AddUser(teamBID, userX.ID)).NotTo(HaveOccurred())
			Expect(s.Projects().AddUser(projectA1ID, userX.ID)).NotTo(HaveOccurred())
			Expect(s.Projects().AddUser(projectB1ID, userX.ID)).NotTo(HaveOccurred())
		})

		AfterAll(func() {
			s := loginAsAdmin(sdk.GetSDK())
			_ = s.Teams().Delete(teamAID)
			_ = s.Teams().Delete(teamBID)
			_ = s.Users().Delete(userX.ID)
		})

		It("should list all my projects", func() {
			s := sdk.GetSDK().Guest()
			_, err := s.LoginWithUsername(userX.Username, userPass)
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			projects, err := s.Me().ListProjects(nil)
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			ids := []int{}
			for _, p := range projects.List {
				ids = append(ids, p.ID)
			}
			Expect(ids).To(ContainElements(projectA1ID, projectB1ID))
			Expect(ids).NotTo(ContainElement(projectA2ID))
		})

		It("should filter projects by team", func() {
			s := sdk.GetSDK().Guest()
			_, err := s.LoginWithUsername(userX.Username, userPass)
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			params := &sdk.ListParams{TeamIds: []int{teamAID}}
			projects, err := s.Me().ListProjects(params)
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			Expect(projects.Total).To(Equal(1))
			Expect(projects.List[0].ID).To(Equal(projectA1ID))
		})

		It("should filter by multiple teams", func() {
			s := sdk.GetSDK().Guest()
			_, err := s.LoginWithUsername(userX.Username, userPass)
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			params := &sdk.ListParams{TeamIds: []int{teamAID, teamBID}}
			projects, err := s.Me().ListProjects(params)
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			ids := []int{}
			for _, p := range projects.List {
				ids = append(ids, p.ID)
			}
			Expect(ids).To(ContainElements(projectA1ID, projectB1ID))
		})

		It("should search projects by name", func() {
			s := sdk.GetSDK().Guest()
			_, err := s.LoginWithUsername(userX.Username, userPass)
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			params := &sdk.ListParams{Name: Ptr(projectA1Name)}
			projects, err := s.Me().ListProjects(params)
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			Expect(projects.Total).To(Equal(1))
			Expect(projects.List[0].ID).To(Equal(projectA1ID))
		})

		It("should exit from project", func() {
			s := sdk.GetSDK().Guest()
			_, err := s.LoginWithUsername(userX.Username, userPass)
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			Expect(s.Me().ExitProject(projectA1ID)).NotTo(HaveOccurred())
			projects, err := s.Me().ListProjects(nil)
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)

			var ids []int
			for _, p := range projects.List {
				ids = append(ids, p.ID)
			}
			Expect(ids).NotTo(ContainElement(projectA1ID))
			s = loginAsAdmin(s)
			teamMembers, err := s.Teams().ListUsers(teamAID, nil)
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			found := false
			for _, u := range teamMembers.List {
				if u.ID == userX.ID {
					found = true
					break
				}
			}
			Expect(found).To(BeTrue())
		})
	})

	Context("Integration: Project Lifecycle", Ordered, func() {
		It("should handle complete project lifecycle", func() {
			s := sdk.GetSDK().Guest()
			leader, leaderPass := createAndSetupUser(helperUniqueName("lifecycle_leader"), "pass")
			memberA, memberAPass := createAndSetupUser(helperUniqueName("lifecycle_member_a"), "pass")
			memberB, memberBPass := createAndSetupUser(helperUniqueName("lifecycle_member_b"), "pass")

			DeferCleanup(func() {
				s := loginAsAdmin(sdk.GetSDK())
				_ = s.Users().Delete(leader.ID)
				_ = s.Users().Delete(memberA.ID)
				_ = s.Users().Delete(memberB.ID)
			})

			s = loginAsAdmin(s)
			team, err := s.Teams().Create(&sdk.CreateTeamRequest{Name: helperUniqueName("team_lifecycle")})
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			teamID := team.ID
			DeferCleanup(func() {
				s = loginAsAdmin(s)
				_ = s.Teams().Delete(teamID)
			})

			Expect(s.Teams().AddUser(teamID, leader.ID)).NotTo(HaveOccurred())
			Expect(s.Teams().AddUser(teamID, memberA.ID)).NotTo(HaveOccurred())
			Expect(s.Teams().UpdateLeader(teamID, Ptr(leader.ID))).NotTo(HaveOccurred())

			By("Leader creates project")
			_, err = s.LoginWithUsername(leader.Username, leaderPass)
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			project, err := s.Teams().CreateProject(teamID, &sdk.CreateProjectRequest{Name: helperUniqueName("project_lifecycle")})
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			Expect(project.Status).To(Equal("WAIT_FOR_SCHEDULE"))
			projectID := project.ID

			By("Leader adds members to project")
			Expect(s.Projects().AddUser(projectID, memberA.ID)).NotTo(HaveOccurred())
			Expect(s.Projects().AddUser(projectID, memberB.ID)).NotTo(HaveOccurred())
			s = loginAsAdmin(s)
			teamMembers, err := s.Teams().ListUsers(teamID, nil)
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			memberBFound := false
			for _, u := range teamMembers.List {
				if u.ID == memberB.ID {
					memberBFound = true
					break
				}
			}
			Expect(memberBFound).To(BeTrue(), "member not previously in team should auto-join via project add")

			By("Leader updates project status to IN_PROGRESS")
			_, err = s.LoginWithUsername(leader.Username, leaderPass)
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			inProgress := "IN_PROGRESS"
			project, err = s.Projects().Update(projectID, &sdk.UpdateProjectRequest{Status: &inProgress})
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			Expect(project.Status).To(Equal(inProgress))

			By("Members can view project details")
			_, err = s.LoginWithUsername(memberA.Username, memberAPass)
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			_, err = s.Projects().Get(projectID)
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			_, err = s.LoginWithUsername(memberB.Username, memberBPass)
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			_, err = s.Projects().Get(projectID)
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)

			By("Member exits project but stays in team")
			_, err = s.LoginWithUsername(memberA.Username, memberAPass)
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			Expect(s.Me().ExitProject(projectID)).NotTo(HaveOccurred())
			projects, err := s.Me().ListProjects(nil)
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			for _, p := range projects.List {
				Expect(p.ID).NotTo(Equal(projectID))
			}
			s = loginAsAdmin(s)
			projectMembers, err := s.Projects().ListUsers(projectID, nil)
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			for _, u := range projectMembers.List {
				Expect(u.ID).NotTo(Equal(memberA.ID))
			}
			teamMembers, err = s.Teams().ListUsers(teamID, nil)
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			memberAFound := false
			for _, u := range teamMembers.List {
				if u.ID == memberA.ID {
					memberAFound = true
					break
				}
			}
			Expect(memberAFound).To(BeTrue())

			By("Leader finishes project")
			_, err = s.LoginWithUsername(leader.Username, leaderPass)
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			finished := "FINISHED"
			project, err = s.Projects().Update(projectID, &sdk.UpdateProjectRequest{Status: &finished})
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			Expect(project.Status).To(Equal(finished))

			By("Leader deletes project")
			Expect(s.Projects().Delete(projectID)).NotTo(HaveOccurred())
			s = loginAsAdmin(s)
			_, err = s.Projects().Get(projectID)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("404"))
			teamInfo, err := s.Teams().Get(teamID)
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			Expect(teamInfo.ID).To(Equal(teamID))
			teamMembers, err = s.Teams().ListUsers(teamID, nil)
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			leaderStill, memberAStill, memberBStill := false, false, false
			for _, u := range teamMembers.List {
				switch u.ID {
				case leader.ID:
					leaderStill = true
				case memberA.ID:
					memberAStill = true
				case memberB.ID:
					memberBStill = true
				}
			}
			Expect(leaderStill).To(BeTrue())
			Expect(memberAStill).To(BeTrue())
			Expect(memberBStill).To(BeTrue())

			By("Admin deletes team and keeps users")
			Expect(s.Teams().Delete(teamID)).NotTo(HaveOccurred())
			_, err = s.Teams().Get(teamID)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("404"))
			s = loginAsAdmin(s)
			_, err = s.Users().Get(memberB.ID)
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
		})
	})
})
