package conformance

import (
	"net/http"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/dspo/go-homework/sdk"
)

var _ = Describe("Audits", func() {
	Context("Audit Logs", Ordered, func() {
		var auditKeyword string
		var auditTeamID int
		var auditUserID int
		var timeStart, timeEnd int64

		BeforeAll(func() {
			s := loginAsAdmin(sdk.GetSDK())
			auditKeyword = helperUniqueName("audit")
			timeStart = time.Now().Add(-time.Minute).Unix()

			user, err := s.Users().Create(auditKeyword+"_user", "auditpass")
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			auditUserID = user.ID

			team, err := s.Teams().Create(&sdk.CreateTeamRequest{Name: auditKeyword + "_team"})
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			auditTeamID = team.ID

			desc := Ptr("desc " + auditKeyword)
			_, err = s.Teams().Update(team.ID, &sdk.UpdateTeamRequest{Desc: desc})
			Expect(err).NotTo(HaveOccurred(), "failed to update team description: %v", err)

			project, err := s.Teams().CreateProject(team.ID, &sdk.CreateProjectRequest{Name: auditKeyword + "_project"})
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			Expect(s.Projects().Delete(project.ID)).NotTo(HaveOccurred())

			timeEnd = time.Now().Add(time.Minute).Unix()
		})

		AfterAll(func() {
			s := loginAsAdmin(sdk.GetSDK())
			_ = s.Users().Delete(auditUserID)
			_ = s.Teams().Delete(auditTeamID)
		})

		It("should query audit logs by admin", func() {
			s := loginAsAdmin(sdk.GetSDK())
			logs, err := s.Audits().List(nil)
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			Expect(logs.Total).To(BeNumerically(">", 0))
			Expect(logs.List).NotTo(BeEmpty())
			for _, entry := range logs.List {
				Expect(entry.ID).To(BeNumerically(">", 0))
				Expect(entry.Content).NotTo(BeEmpty())
				Expect(entry.CreatedAt).To(BeNumerically(">", 0))
			}
		})

		It("should fail to query by normal user", func() {
			s := sdk.GetSDK().Guest()
			user, pass := createAndSetupUser(helperUniqueName("audit_normal"), "pass1234")
			DeferCleanup(func() {
				s = loginAsAdmin(s)
				_ = s.Users().Delete(user.ID)
			})

			s, err := s.LoginWithUsername(user.Username, pass)
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			_, err = s.Audits().List(nil)
			Expect(err).To(sdk.HaveOccurredWithStatusCode(http.StatusForbidden))
		})

		PIt("should filter by keyword", func() {
			s := loginAsAdmin(sdk.GetSDK())
			params := &sdk.ListParams{Keyword: Ptr(auditKeyword)}
			logs, err := s.Audits().List(params)
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			Expect(logs.Total).To(BeNumerically(">", 0))
			Expect(logs.List).NotTo(BeEmpty())
			for _, entry := range logs.List {
				Expect(strings.Contains(strings.ToLower(entry.Content), strings.ToLower(auditKeyword))).To(BeTrue())
			}
		})

		PIt("should filter by time range", func() {
			s := loginAsAdmin(sdk.GetSDK())
			params := &sdk.ListParams{StartAt: helperInt64Ptr(timeStart), EndAt: helperInt64Ptr(timeEnd)}
			logs, err := s.Audits().List(params)
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			Expect(logs.Total).To(BeNumerically(">", 0))
			for _, entry := range logs.List {
				Expect(entry.CreatedAt).To(BeNumerically(">=", timeStart))
				Expect(entry.CreatedAt).To(BeNumerically("<=", timeEnd))
			}
		})

		PIt("should paginate audit logs", func() {
			s := loginAsAdmin(sdk.GetSDK())
			page := 1
			pageSize := 1
			params := &sdk.ListParams{Page: &page, PageSize: &pageSize}
			firstPage, err := s.Audits().List(params)
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			Expect(firstPage.List).To(HaveLen(1))
			page = 2
			secondPage, err := s.Audits().List(params)
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			Expect(secondPage.List).To(HaveLen(1))
			Expect(firstPage.List[0].ID).NotTo(Equal(secondPage.List[0].ID))
		})

		PIt("should order audit logs", func() {
			s := loginAsAdmin(sdk.GetSDK())
			orderBy := "created_at"
			logs, err := s.Audits().List(&sdk.ListParams{OrderBy: &orderBy})
			Expect(err).NotTo(HaveOccurred(), "unexpected error: %v", err)
			prev := int64(0)
			for _, entry := range logs.List {
				Expect(entry.CreatedAt).To(BeNumerically(">=", prev))
				prev = entry.CreatedAt
			}
		})
	})
})
