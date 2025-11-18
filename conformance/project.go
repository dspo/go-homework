package conformance

import (
	. "github.com/onsi/ginkgo/v2"
)

var _ = Describe("Project", func() {
	BeforeAll(func() {
		By("create a team")
		By("add a user to the team")
		By("mark the user as team leader")
	})

	AfterAll(func() {
		By("delete all projects")
		By("delete the team")
	})

	Context("/api/teams/:team_id/projects[:project_id]", func() {
		It("check the permission for creating a project", func() {
			By("add a user to the team")

			By("the user try to create a project in the team should fail")
		})

		It("project apis", func() {
			By("list projects in the team should get total 0")

			By("create a project POST /api/teams/:team_id/projects")

			By("list project in the team again")

			By("get the project by a normal user")

			By("update the project by the leader")

			By("update the project by a normal user should get 403")

			By("delete the project by a normal user should get 401")
		})
	})
})
