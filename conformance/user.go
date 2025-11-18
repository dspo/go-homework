package conformance

import (
	. "github.com/onsi/ginkgo/v2"
)

var _ = Describe("Users", func() {
	Context("/api/me", func() {
		It("login user access", func() {

		})

		It("visitor access", func() {
			By("visitor access should get 401")
		})
	})

	Context("/api/users[/:user_id]", func() {

	})
})
