package conformance

import (
	. "github.com/onsi/ginkgo/v2"
)

var _ = BeforeSuite(GetSuit().BeforeSuite)

type Suite interface {
	BeforeSuite()
	AfterSuit()
	Admin() AdminClient
	User(name string) UserClient
}

func GetSuit() Suite {
	return &suit{}
}

type AdminClient interface {
}

type UserClient interface {
}

type suit struct {
}

func (s *suit) BeforeSuite() {
	By("admin login first time")

	By("admin process resource without changing password should be fail")

	By("admin change password")

	By("admin process resource without login again should fail")

	By("admin login again")

	By("admin invites user")

	By("admin set role and permission for the user")

	By("user login")

	By("user process resource without changing password should fail")

	By("user change password")

	By("user process resource without login again should fail")

	By("user login again")

	By("user process resource")
}

func (s *suit) AfterSuit() {
	// TODO implement me
	panic("implement me")
}

func (s *suit) Admin() AdminClient {
	// TODO implement me
	panic("implement me")
}

func (s *suit) User(name string) UserClient {
	// TODO implement me
	panic("implement me")
}
