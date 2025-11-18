package conformance

import (
	. "github.com/onsi/ginkgo/v2"
)

var _ = Describe("Team", Ordered, func() {
	AfterAll(func() {
		By("remove the team")
	})

	Context("/api/teams[/:team_id]", func() {

	})

	Context("/api/teams/:team_id/users", func() {
		It("add users to a team", func() {
			By("list users from the team")

			By("list teams for the user")

			By("add users to the team")

			By("list users from the team again")

			By("list teams for the user again")
		})

		It("remove users from a team", func() {
			By("list users from the team should get total 0")

			By("remove users from the team")

			By("list users from the team again")
		})

		It("mark a user as team leader", func() {

			By("get team leader should get 404")

			By("mark a user as team leader")

			By("get team leader again")

			By("list all my teams /api/me/teams")

			By("list my leading teams /api/me/teams?leading=true")

			By("list my non-leading teams /api/me/teams?leading=false")
		})
	})

})
