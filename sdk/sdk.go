package sdk

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"path"
	"strconv"
	"strings"
	"sync"

	"github.com/pkg/errors"
)

// SDK is the client interface
type SDK interface {
	// LoginWithUsername logs in with username and returns a user-scoped client
	LoginWithUsername(username, password string) (UserClient, error)
	// LoginWithEmail logs in with email and returns a user-scoped client
	LoginWithEmail(email, password string) (UserClient, error)
	Guest() UserClient
	// Healthz performs health check
	Healthz() error
}

// UserClient 表示携带指定用户登录态的客户端。
// 使用该客户端发出的请求会携带对应用户的 Cookie，同时支持业务 API 调用。
type UserClient interface {
	SDK
	// Logout logs out the current user
	Logout() error
	// Me returns the current user API
	Me() MeAPI
	// Users returns the users API
	Users() UsersAPI
	// Teams returns the teams API
	Teams() TeamsAPI
	// Projects returns the projects API
	Projects() ProjectsAPI
	// Roles returns the roles API
	Roles() RolesAPI
	// Audits returns the audits API
	Audits() AuditsAPI
}

// MeAPI provides current user operations
type MeAPI interface {
	// Get gets current user info
	Get() (*User, error)
	// Update updates current user info
	Update(req *UpdateMeRequest) (*User, error)
	// UpdatePassword changes current user password
	UpdatePassword(oldPassword, newPassword string) error
	// ListTeams gets current user's teams list
	ListTeams(leading *bool) (*TeamsListResponse, error)
	// ExitTeam exits from a team
	ExitTeam(teamID int) error
	// ListProjects gets current user's projects list
	ListProjects(params *ListParams) (*ProjectsListResponse, error)
	// ExitProject exits from a project
	ExitProject(projectID int) error
}

// UsersAPI provides user management operations
type UsersAPI interface {
	// Create creates a user (admin only)
	Create(username, password string) (*User, error)
	// List lists users
	List(params *ListParams) (*UsersListResponse, error)
	// Get gets user details
	Get(userID int) (*User, error)
	// Delete deletes a user (admin only)
	Delete(userID int) error
	// ListTeams gets user's teams list
	ListTeams(userID int, params *ListParams) (*TeamsListResponse, error)
	// ListProjects gets user's projects list
	ListProjects(userID int) (*ProjectsListResponse, error)
	// AddRole adds a role to user
	AddRole(userID, roleID int) error
	// RemoveRole removes a role from user
	RemoveRole(userID, roleID int) error
}

// TeamsAPI provides team management operations
type TeamsAPI interface {
	// List lists teams
	List(params *ListParams) (*TeamsListResponse, error)
	// Create creates a team
	Create(req *CreateTeamRequest) (*Team, error)
	// Get gets team details
	Get(teamID int) (*Team, error)
	// Update updates team info
	Update(teamID int, req *UpdateTeamRequest) (*Team, error)
	// UpdateLeader updates team leader
	UpdateLeader(teamID int, leaderID *int) (*Team, error)
	// Delete deletes a team
	Delete(teamID int) error
	// ListUsers gets team members list
	ListUsers(teamID int, params *ListParams) (*UsersListResponse, error)
	// AddUser adds a user to team
	AddUser(teamID, userID int) error
	// RemoveUser removes a user from team
	RemoveUser(teamID, userID int) error
	// ListProjects gets team's projects list
	ListProjects(teamID int, params *ListParams) (*ProjectsListResponse, error)
	// CreateProject creates a project for team
	CreateProject(teamID int, req *CreateProjectRequest) (*Project, error)
}

// ProjectsAPI provides project management operations
type ProjectsAPI interface {
	// Get gets project details
	Get(projectID int) (*Project, error)
	// Update updates project info
	Update(projectID int, req *UpdateProjectRequest) (*Project, error)
	// Patch partially updates project info
	Patch(projectID int, patches []PatchProjectRequest) (*Project, error)
	// Delete deletes a project
	Delete(projectID int) error
	// ListUsers gets project members list
	ListUsers(projectID int, params *ListParams) (*UsersListResponse, error)
	// AddUser adds a user to project
	AddUser(projectID, userID int) error
	// RemoveUser removes a user from project
	RemoveUser(projectID, userID int) error
}

// RolesAPI provides role management operations
type RolesAPI interface {
	// List lists all roles
	List() (*RolesListResponse, error)
	// Create creates a role
	Create(req *CreateRoleRequest) (*Role, error)
	// Delete deletes a role
	Delete(roleID int) error
}

// AuditsAPI provides audit log operations
type AuditsAPI interface {
	// List gets audit logs
	List(params *ListParams) (*AuditsListResponse, error)
}

var once sync.Once
var globalSDK SDK

// NewSDK creates a new SDK instance and sets it as global singleton.
// It can be only init onece.
func NewSDK(addr string) SDK {
	once.Do(func() {
		initSDK(addr)
	})
	return globalSDK
}

func initSDK(addr string) {
	jar, _ := cookiejar.New(nil)
	baseURL, _ := url.Parse(strings.TrimRight(addr, "/"))
	globalSDK = &sdk{
		baseURL: baseURL,
		client: &http.Client{
			Jar: jar,
		},
	}
}

// GetSDK returns the global SDK singleton
func GetSDK() SDK {
	if globalSDK == nil {
		panic("SDK not initialized, call NewSDK first")
	}
	return globalSDK
}

type sdk struct {
	baseURL *url.URL
	client  *http.Client
	// cookieScope 记录最近一次更新 Cookie 时使用的 URL，用于复制登录态。
	cookieScope *url.URL
}

func (s *sdk) Me() MeAPI {
	return &meAPI{sdk: s}
}

func (s *sdk) Users() UsersAPI {
	return &usersAPI{sdk: s}
}

func (s *sdk) Teams() TeamsAPI {
	return &teamsAPI{sdk: s}
}

func (s *sdk) Projects() ProjectsAPI {
	return &projectsAPI{sdk: s}
}

func (s *sdk) Roles() RolesAPI {
	return &rolesAPI{sdk: s}
}

func (s *sdk) Audits() AuditsAPI {
	return &auditsAPI{sdk: s}
}

// =============== Internal utility methods ===============

func (s *sdk) cookieURLForJar(base *url.URL) *url.URL {
	if base == nil {
		return nil
	}
	u := *base
	if u.Path == "" {
		u.Path = "/"
	}
	u.RawQuery = ""
	u.Fragment = ""
	return &u
}

func (s *sdk) cloneWithCookies() UserClient {
	var baseCopy *url.URL
	if s.baseURL != nil {
		tmp := *s.baseURL
		baseCopy = &tmp
	}

	newJar, _ := cookiejar.New(nil)
	cookieURL := s.cookieScope
	if cookieURL == nil {
		cookieURL = s.cookieURLForJar(s.baseURL)
	}
	if s.client != nil && s.client.Jar != nil && cookieURL != nil {
		newJar.SetCookies(cookieURL, s.client.Jar.Cookies(cookieURL))
	}

	newClient := http.Client{}
	if s.client != nil {
		newClient = *s.client
	}
	newClient.Jar = newJar

	return &sdk{
		baseURL: baseCopy,
		client:  &newClient,
	}
}

func (s *sdk) resetCookies(targetURL *url.URL, cookies []*http.Cookie) {
	if len(cookies) == 0 {
		return
	}

	newJar, _ := cookiejar.New(nil)
	cookieURL := s.cookieURLForJar(targetURL)
	if cookieURL == nil {
		return
	}

	newJar.SetCookies(cookieURL, cookies)
	s.cookieScope = cookieURL
	if s.client == nil {
		s.client = &http.Client{}
	}
	s.client.Jar = newJar
}

func doRequest[T any](s *sdk, method, pathStr string, body any) (*T, error) {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	// Construct full URL using url.URL
	// Parse pathStr to properly handle query parameters
	relativeURL, err := url.Parse(pathStr)
	if err != nil {
		return nil, fmt.Errorf("parse path: %w", err)
	}
	fullURL := s.baseURL.ResolveReference(relativeURL)

	req, err := http.NewRequest(method, fullURL.String(), bodyReader)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	s.resetCookies(fullURL, resp.Cookies())

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		var e = &Error{StatusCode: resp.StatusCode}
		if err = json.Unmarshal(respBody, e); err != nil {
			e.Error_ = errors.Wrapf(err, "response body: %s", string(respBody)).Error()
		}
		return nil, e
	}

	var out T
	if len(respBody) > 0 {
		if err := json.Unmarshal(respBody, &out); err != nil {
			return nil, fmt.Errorf("unmarshal response: %w", err)
		}
	}

	return &out, nil
}

// =============== Authentication implementations ===============

func (s *sdk) LoginWithUsername(username, password string) (UserClient, error) {
	req := LoginWithUsername{
		Username: username,
		Password: password,
	}
	if _, err := doRequest[struct{}](s, http.MethodPost, "/api/login", req); err != nil {
		return nil, err
	}
	return s.cloneWithCookies(), nil
}

func (s *sdk) LoginWithEmail(email, password string) (UserClient, error) {
	req := LoginWithEmail{
		Email:    email,
		Password: password,
	}
	if _, err := doRequest[struct{}](s, http.MethodPost, "/api/login", req); err != nil {
		return nil, err
	}
	return s.cloneWithCookies(), nil
}

func (s *sdk) Guest() UserClient {
	return &sdk{
		baseURL:     s.baseURL,
		client:      new(http.Client),
		cookieScope: s.cookieScope,
	}
}

func (s *sdk) Logout() error {
	_, err := doRequest[struct{}](s, http.MethodPost, "/api/logout", nil)
	return err
}

func (s *sdk) Healthz() error {
	_, err := doRequest[struct{}](s, http.MethodGet, "/healthz", nil)
	return err
}

// =============== Me implementations ===============

type meAPI struct {
	sdk *sdk
}

func (m *meAPI) Get() (*User, error) {
	user, err := doRequest[User](m.sdk, http.MethodGet, "/api/me", nil)
	return user, err
}

func (m *meAPI) Update(req *UpdateMeRequest) (*User, error) {
	user, err := doRequest[User](m.sdk, http.MethodPut, "/api/me", req)
	return user, err
}

func (m *meAPI) UpdatePassword(oldPassword, newPassword string) error {
	req := UpdatePasswordRequest{
		OldPassword: oldPassword,
		NewPassword: newPassword,
	}
	_, err := doRequest[struct{}](m.sdk, http.MethodPut, "/api/me/password", req)
	return err
}

func (m *meAPI) ListTeams(leading *bool) (*TeamsListResponse, error) {
	params := &ListParams{Leading: leading}
	pathURL := &url.URL{
		Path:     "/api/me/teams",
		RawQuery: params.ToURLValues().Encode(),
	}
	resp, err := doRequest[TeamsListResponse](m.sdk, http.MethodGet, pathURL.String(), nil)
	return resp, err
}

func (m *meAPI) ExitTeam(teamID int) error {
	pathStr := path.Join("/api/me/teams", strconv.Itoa(teamID))
	_, err := doRequest[struct{}](m.sdk, http.MethodDelete, pathStr, nil)
	return err
}

func (m *meAPI) ListProjects(params *ListParams) (*ProjectsListResponse, error) {
	pathURL := &url.URL{
		Path:     "/api/me/projects",
		RawQuery: params.ToURLValues().Encode(),
	}
	resp, err := doRequest[ProjectsListResponse](m.sdk, http.MethodGet, pathURL.String(), nil)
	return resp, err
}

func (m *meAPI) ExitProject(projectID int) error {
	pathStr := path.Join("/api/me/projects", strconv.Itoa(projectID))
	_, err := doRequest[struct{}](m.sdk, http.MethodDelete, pathStr, nil)
	return err
}

// =============== Users implementations ===============

type usersAPI struct {
	sdk *sdk
}

func (u *usersAPI) Create(username, password string) (*User, error) {
	req := CreateUserRequest{
		Username: username,
		Password: password,
	}
	user, err := doRequest[User](u.sdk, http.MethodPost, "/api/users", req)
	return user, err
}

func (u *usersAPI) List(params *ListParams) (*UsersListResponse, error) {
	query := params.ToURLValues()
	pathURL := &url.URL{
		Path:     "/api/users",
		RawQuery: query.Encode(),
	}
	resp, err := doRequest[UsersListResponse](u.sdk, http.MethodGet, pathURL.String(), nil)
	return resp, err
}

func (u *usersAPI) Get(userID int) (*User, error) {
	pathStr := path.Join("/api/users", strconv.Itoa(userID))
	user, err := doRequest[User](u.sdk, http.MethodGet, pathStr, nil)
	return user, err
}

func (u *usersAPI) Delete(userID int) error {
	pathStr := path.Join("/api/users", strconv.Itoa(userID))
	_, err := doRequest[struct{}](u.sdk, http.MethodDelete, pathStr, nil)
	return err
}

func (u *usersAPI) ListTeams(userID int, params *ListParams) (*TeamsListResponse, error) {
	query := params.ToURLValues()
	pathURL := &url.URL{
		Path:     path.Join("/api/users", strconv.Itoa(userID), "teams"),
		RawQuery: query.Encode(),
	}
	resp, err := doRequest[TeamsListResponse](u.sdk, http.MethodGet, pathURL.String(), nil)
	return resp, err
}

func (u *usersAPI) ListProjects(userID int) (*ProjectsListResponse, error) {
	pathStr := path.Join("/api/users", strconv.Itoa(userID), "projects")
	resp, err := doRequest[ProjectsListResponse](u.sdk, http.MethodGet, pathStr, nil)
	return resp, err
}

func (u *usersAPI) AddRole(userID, roleID int) error {
	pathStr := path.Join("/api/users", strconv.Itoa(userID), "roles")
	req := AddRoleToUserRequest{RoleID: roleID}
	_, err := doRequest[struct{}](u.sdk, http.MethodPost, pathStr, req)
	return err
}

func (u *usersAPI) RemoveRole(userID, roleID int) error {
	pathStr := path.Join("/api/users", strconv.Itoa(userID), "roles", strconv.Itoa(roleID))
	_, err := doRequest[struct{}](u.sdk, http.MethodDelete, pathStr, nil)
	return err
}

// =============== Teams implementations ===============

type teamsAPI struct {
	sdk *sdk
}

func (t *teamsAPI) List(params *ListParams) (*TeamsListResponse, error) {
	query := params.ToURLValues()
	pathURL := &url.URL{
		Path:     "/api/teams",
		RawQuery: query.Encode(),
	}
	resp, err := doRequest[TeamsListResponse](t.sdk, http.MethodGet, pathURL.String(), nil)
	return resp, err
}

func (t *teamsAPI) Create(req *CreateTeamRequest) (*Team, error) {
	team, err := doRequest[Team](t.sdk, http.MethodPost, "/api/teams", req)
	return team, err
}

func (t *teamsAPI) Get(teamID int) (*Team, error) {
	pathStr := path.Join("/api/teams", strconv.Itoa(teamID))
	team, err := doRequest[Team](t.sdk, http.MethodGet, pathStr, nil)
	return team, err
}

func (t *teamsAPI) Update(teamID int, req *UpdateTeamRequest) (*Team, error) {
	pathStr := path.Join("/api/teams", strconv.Itoa(teamID))
	team, err := doRequest[Team](t.sdk, http.MethodPut, pathStr, req)
	return team, err
}

func (t *teamsAPI) UpdateLeader(teamID int, leaderID *int) (*Team, error) {
	pathStr := path.Join("/api/teams", strconv.Itoa(teamID))

	var value *UpdateTeamLeaderValue
	if leaderID != nil {
		value = &UpdateTeamLeaderValue{ID: leaderID}
	}

	req := []UpdateTeamLeaderRequest{
		{
			Op:    "replace",
			Path:  "/leader",
			Value: value,
		},
	}

	team, err := doRequest[Team](t.sdk, http.MethodPatch, pathStr, req)
	return team, err
}

func (t *teamsAPI) Delete(teamID int) error {
	pathStr := path.Join("/api/teams", strconv.Itoa(teamID))
	_, err := doRequest[struct{}](t.sdk, http.MethodDelete, pathStr, nil)
	return err
}

func (t *teamsAPI) ListUsers(teamID int, params *ListParams) (*UsersListResponse, error) {
	query := params.ToURLValues()
	pathURL := &url.URL{
		Path:     path.Join("/api/teams", strconv.Itoa(teamID), "users"),
		RawQuery: query.Encode(),
	}
	resp, err := doRequest[UsersListResponse](t.sdk, http.MethodGet, pathURL.String(), nil)
	return resp, err
}

func (t *teamsAPI) AddUser(teamID, userID int) error {
	pathStr := path.Join("/api/teams", strconv.Itoa(teamID), "users")
	req := AddUserToTeamRequest{UserID: userID}
	_, err := doRequest[struct{}](t.sdk, http.MethodPost, pathStr, req)
	return err
}

func (t *teamsAPI) RemoveUser(teamID, userID int) error {
	pathStr := path.Join("/api/teams", strconv.Itoa(teamID), "users", strconv.Itoa(userID))
	_, err := doRequest[struct{}](t.sdk, http.MethodDelete, pathStr, nil)
	return err
}

func (t *teamsAPI) ListProjects(teamID int, params *ListParams) (*ProjectsListResponse, error) {
	query := params.ToURLValues()
	pathURL := &url.URL{
		Path:     path.Join("/api/teams", strconv.Itoa(teamID), "projects"),
		RawQuery: query.Encode(),
	}
	resp, err := doRequest[ProjectsListResponse](t.sdk, http.MethodGet, pathURL.String(), nil)
	return resp, err
}

func (t *teamsAPI) CreateProject(teamID int, req *CreateProjectRequest) (*Project, error) {
	pathStr := path.Join("/api/teams", strconv.Itoa(teamID), "projects")
	project, err := doRequest[Project](t.sdk, http.MethodPost, pathStr, req)
	return project, err
}

// =============== Projects implementations ===============

type projectsAPI struct {
	sdk *sdk
}

func (p *projectsAPI) Get(projectID int) (*Project, error) {
	pathStr := path.Join("/api/projects", strconv.Itoa(projectID))
	project, err := doRequest[Project](p.sdk, http.MethodGet, pathStr, nil)
	return project, err
}

func (p *projectsAPI) Update(projectID int, req *UpdateProjectRequest) (*Project, error) {
	pathStr := path.Join("/api/projects", strconv.Itoa(projectID))
	project, err := doRequest[Project](p.sdk, http.MethodPut, pathStr, req)
	return project, err
}

func (p *projectsAPI) Patch(projectID int, patches []PatchProjectRequest) (*Project, error) {
	pathStr := path.Join("/api/projects", strconv.Itoa(projectID))
	project, err := doRequest[Project](p.sdk, http.MethodPatch, pathStr, patches)
	return project, err
}

func (p *projectsAPI) Delete(projectID int) error {
	pathStr := path.Join("/api/projects", strconv.Itoa(projectID))
	_, err := doRequest[struct{}](p.sdk, http.MethodDelete, pathStr, nil)
	return err
}

func (p *projectsAPI) ListUsers(projectID int, params *ListParams) (*UsersListResponse, error) {
	query := params.ToURLValues()
	pathURL := &url.URL{
		Path:     path.Join("/api/projects", strconv.Itoa(projectID), "users"),
		RawQuery: query.Encode(),
	}
	resp, err := doRequest[UsersListResponse](p.sdk, http.MethodGet, pathURL.String(), nil)
	return resp, err
}

func (p *projectsAPI) AddUser(projectID, userID int) error {
	pathStr := path.Join("/api/projects", strconv.Itoa(projectID), "users")
	req := AddUserToProjectRequest{UserID: userID}
	_, err := doRequest[struct{}](p.sdk, http.MethodPost, pathStr, req)
	return err
}

func (p *projectsAPI) RemoveUser(projectID, userID int) error {
	pathStr := path.Join("/api/projects", strconv.Itoa(projectID), "users", strconv.Itoa(userID))
	_, err := doRequest[struct{}](p.sdk, http.MethodDelete, pathStr, nil)
	return err
}

// =============== Roles implementations ===============

type rolesAPI struct {
	sdk *sdk
}

func (r *rolesAPI) List() (*RolesListResponse, error) {
	resp, err := doRequest[RolesListResponse](r.sdk, http.MethodGet, "/api/roles", nil)
	return resp, err
}

func (r *rolesAPI) Create(req *CreateRoleRequest) (*Role, error) {
	role, err := doRequest[Role](r.sdk, http.MethodPost, "/api/roles", req)
	return role, err
}

func (r *rolesAPI) Delete(roleID int) error {
	pathStr := path.Join("/api/roles", strconv.Itoa(roleID))
	_, err := doRequest[struct{}](r.sdk, http.MethodDelete, pathStr, nil)
	return err
}

// =============== Audits implementations ===============

type auditsAPI struct {
	sdk *sdk
}

func (a *auditsAPI) List(params *ListParams) (*AuditsListResponse, error) {
	query := params.ToURLValues()
	pathURL := &url.URL{
		Path:     "/api/audits",
		RawQuery: query.Encode(),
	}
	resp, err := doRequest[AuditsListResponse](a.sdk, http.MethodGet, pathURL.String(), nil)
	return resp, err
}
