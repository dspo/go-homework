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
)

// SDK is the client interface
type SDK interface {
	// Auth returns the authentication API
	Auth() AuthAPI
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

// AuthAPI provides authentication operations
type AuthAPI interface {
	// LoginWithUsername logs in with username
	LoginWithUsername(username, password string) error
	// LoginWithEmail logs in with email
	LoginWithEmail(email, password string) error
	// Logout logs out the current user
	Logout() error
	// Healthz performs health check
	Healthz() error
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

// NewSDK creates a new SDK instance
func NewSDK(addr string) SDK {
	jar, _ := cookiejar.New(nil)
	baseURL, _ := url.Parse(strings.TrimRight(addr, "/"))
	return &sdk{
		baseURL: baseURL,
		client: &http.Client{
			Jar: jar,
		},
	}
}

type sdk struct {
	baseURL *url.URL
	client  *http.Client
}

func (s *sdk) Auth() AuthAPI {
	return &authAPI{sdk: s}
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

func (s *sdk) doRequest(method, pathStr string, body interface{}, result interface{}) error {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	// Construct full URL using url.URL
	fullURL := s.baseURL.ResolveReference(&url.URL{Path: pathStr})

	req, err := http.NewRequest(method, fullURL.String(), bodyReader)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		var errResp Error
		if err := json.Unmarshal(respBody, &errResp); err == nil {
			return fmt.Errorf("API error (%d): %s", resp.StatusCode, errResp.Error)
		}
		return fmt.Errorf("API error: status code %d, body: %s", resp.StatusCode, string(respBody))
	}

	if result != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("unmarshal response: %w", err)
		}
	}

	return nil
}

func buildQueryParams(params *ListParams) url.Values {
	query := url.Values{}
	if params == nil {
		return query
	}

	if params.OrderBy != nil {
		query.Set("order_by", *params.OrderBy)
	}
	if params.Page != nil {
		query.Set("page", strconv.Itoa(*params.Page))
	}
	if params.PageSize != nil {
		query.Set("page_size", strconv.Itoa(*params.PageSize))
	}
	if params.Keyword != nil {
		query.Set("keyword", *params.Keyword)
	}
	if params.Name != nil {
		query.Set("name", *params.Name)
	}
	if params.Leading != nil {
		query.Set("leading", strconv.FormatBool(*params.Leading))
	}
	if params.PartIn != nil {
		query.Set("part_in", strconv.FormatBool(*params.PartIn))
	}
	if params.StartAt != nil {
		query.Set("start_at", strconv.FormatInt(*params.StartAt, 10))
	}
	if params.EndAt != nil {
		query.Set("end_at", strconv.FormatInt(*params.EndAt, 10))
	}

	for _, id := range params.TeamIDs {
		query.Add("team_id", strconv.Itoa(id))
	}

	for _, name := range params.RoleNames {
		query.Add("role_name", name)
	}

	return query
}

// =============== Authentication implementations ===============

type authAPI struct {
	sdk *sdk
}

func (a *authAPI) LoginWithUsername(username, password string) error {
	req := LoginWithUsername{
		Username: username,
		Password: password,
	}
	return a.sdk.doRequest(http.MethodPost, "/api/login", req, nil)
}

func (a *authAPI) LoginWithEmail(email, password string) error {
	req := LoginWithEmail{
		Email:    email,
		Password: password,
	}
	return a.sdk.doRequest(http.MethodPost, "/api/login", req, nil)
}

func (a *authAPI) Logout() error {
	return a.sdk.doRequest(http.MethodPost, "/api/logout", nil, nil)
}

func (a *authAPI) Healthz() error {
	return a.sdk.doRequest(http.MethodGet, "/healthz", nil, nil)
}

// =============== Me implementations ===============

type meAPI struct {
	sdk *sdk
}

func (m *meAPI) Get() (*User, error) {
	var user User
	err := m.sdk.doRequest(http.MethodGet, "/api/me", nil, &user)
	return &user, err
}

func (m *meAPI) Update(req *UpdateMeRequest) (*User, error) {
	var user User
	err := m.sdk.doRequest(http.MethodPut, "/api/me", req, &user)
	return &user, err
}

func (m *meAPI) UpdatePassword(oldPassword, newPassword string) error {
	req := UpdatePasswordRequest{
		OldPassword: oldPassword,
		NewPassword: newPassword,
	}
	return m.sdk.doRequest(http.MethodPut, "/api/me/password", req, nil)
}

func (m *meAPI) ListTeams(leading *bool) (*TeamsListResponse, error) {
	params := &ListParams{Leading: leading}
	query := buildQueryParams(params)
	pathURL := &url.URL{
		Path:     "/api/me/teams",
		RawQuery: query.Encode(),
	}
	var resp TeamsListResponse
	err := m.sdk.doRequest(http.MethodGet, pathURL.String(), nil, &resp)
	return &resp, err
}

func (m *meAPI) ExitTeam(teamID int) error {
	pathStr := path.Join("/api/me/teams", strconv.Itoa(teamID))
	return m.sdk.doRequest(http.MethodDelete, pathStr, nil, nil)
}

func (m *meAPI) ListProjects(params *ListParams) (*ProjectsListResponse, error) {
	query := buildQueryParams(params)
	pathURL := &url.URL{
		Path:     "/api/me/projects",
		RawQuery: query.Encode(),
	}
	var resp ProjectsListResponse
	err := m.sdk.doRequest(http.MethodGet, pathURL.String(), nil, &resp)
	return &resp, err
}

func (m *meAPI) ExitProject(projectID int) error {
	pathStr := path.Join("/api/me/projects", strconv.Itoa(projectID))
	return m.sdk.doRequest(http.MethodDelete, pathStr, nil, nil)
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
	var user User
	err := u.sdk.doRequest(http.MethodPost, "/api/users", req, &user)
	return &user, err
}

func (u *usersAPI) List(params *ListParams) (*UsersListResponse, error) {
	query := buildQueryParams(params)
	pathURL := &url.URL{
		Path:     "/api/users",
		RawQuery: query.Encode(),
	}
	var resp UsersListResponse
	err := u.sdk.doRequest(http.MethodGet, pathURL.String(), nil, &resp)
	return &resp, err
}

func (u *usersAPI) Get(userID int) (*User, error) {
	pathStr := path.Join("/api/users", strconv.Itoa(userID))
	var user User
	err := u.sdk.doRequest(http.MethodGet, pathStr, nil, &user)
	return &user, err
}

func (u *usersAPI) Delete(userID int) error {
	pathStr := path.Join("/api/users", strconv.Itoa(userID))
	return u.sdk.doRequest(http.MethodDelete, pathStr, nil, nil)
}

func (u *usersAPI) ListTeams(userID int, params *ListParams) (*TeamsListResponse, error) {
	query := buildQueryParams(params)
	pathURL := &url.URL{
		Path:     path.Join("/api/users", strconv.Itoa(userID), "teams"),
		RawQuery: query.Encode(),
	}
	var resp TeamsListResponse
	err := u.sdk.doRequest(http.MethodGet, pathURL.String(), nil, &resp)
	return &resp, err
}

func (u *usersAPI) ListProjects(userID int) (*ProjectsListResponse, error) {
	pathStr := path.Join("/api/users", strconv.Itoa(userID), "projects")
	var resp ProjectsListResponse
	err := u.sdk.doRequest(http.MethodGet, pathStr, nil, &resp)
	return &resp, err
}

func (u *usersAPI) AddRole(userID, roleID int) error {
	pathStr := path.Join("/api/users", strconv.Itoa(userID), "roles")
	req := AddRoleToUserRequest{RoleID: roleID}
	return u.sdk.doRequest(http.MethodPost, pathStr, req, nil)
}

func (u *usersAPI) RemoveRole(userID, roleID int) error {
	pathStr := path.Join("/api/users", strconv.Itoa(userID), "roles", strconv.Itoa(roleID))
	return u.sdk.doRequest(http.MethodDelete, pathStr, nil, nil)
}

// =============== Teams implementations ===============

type teamsAPI struct {
	sdk *sdk
}

func (t *teamsAPI) List(params *ListParams) (*TeamsListResponse, error) {
	query := buildQueryParams(params)
	pathURL := &url.URL{
		Path:     "/api/teams",
		RawQuery: query.Encode(),
	}
	var resp TeamsListResponse
	err := t.sdk.doRequest(http.MethodGet, pathURL.String(), nil, &resp)
	return &resp, err
}

func (t *teamsAPI) Create(req *CreateTeamRequest) (*Team, error) {
	var team Team
	err := t.sdk.doRequest(http.MethodPost, "/api/teams", req, &team)
	return &team, err
}

func (t *teamsAPI) Get(teamID int) (*Team, error) {
	pathStr := path.Join("/api/teams", strconv.Itoa(teamID))
	var team Team
	err := t.sdk.doRequest(http.MethodGet, pathStr, nil, &team)
	return &team, err
}

func (t *teamsAPI) Update(teamID int, req *UpdateTeamRequest) (*Team, error) {
	pathStr := path.Join("/api/teams", strconv.Itoa(teamID))
	var team Team
	err := t.sdk.doRequest(http.MethodPut, pathStr, req, &team)
	return &team, err
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

	var team Team
	err := t.sdk.doRequest(http.MethodPatch, pathStr, req, &team)
	return &team, err
}

func (t *teamsAPI) Delete(teamID int) error {
	pathStr := path.Join("/api/teams", strconv.Itoa(teamID))
	return t.sdk.doRequest(http.MethodDelete, pathStr, nil, nil)
}

func (t *teamsAPI) ListUsers(teamID int, params *ListParams) (*UsersListResponse, error) {
	query := buildQueryParams(params)
	pathURL := &url.URL{
		Path:     path.Join("/api/teams", strconv.Itoa(teamID), "users"),
		RawQuery: query.Encode(),
	}
	var resp UsersListResponse
	err := t.sdk.doRequest(http.MethodGet, pathURL.String(), nil, &resp)
	return &resp, err
}

func (t *teamsAPI) AddUser(teamID, userID int) error {
	pathStr := path.Join("/api/teams", strconv.Itoa(teamID), "users")
	req := AddUserToTeamRequest{UserID: userID}
	return t.sdk.doRequest(http.MethodPost, pathStr, req, nil)
}

func (t *teamsAPI) RemoveUser(teamID, userID int) error {
	pathStr := path.Join("/api/teams", strconv.Itoa(teamID), "users", strconv.Itoa(userID))
	return t.sdk.doRequest(http.MethodDelete, pathStr, nil, nil)
}

func (t *teamsAPI) ListProjects(teamID int, params *ListParams) (*ProjectsListResponse, error) {
	query := buildQueryParams(params)
	pathURL := &url.URL{
		Path:     path.Join("/api/teams", strconv.Itoa(teamID), "projects"),
		RawQuery: query.Encode(),
	}
	var resp ProjectsListResponse
	err := t.sdk.doRequest(http.MethodGet, pathURL.String(), nil, &resp)
	return &resp, err
}

func (t *teamsAPI) CreateProject(teamID int, req *CreateProjectRequest) (*Project, error) {
	pathStr := path.Join("/api/teams", strconv.Itoa(teamID), "projects")
	var project Project
	err := t.sdk.doRequest(http.MethodPost, pathStr, req, &project)
	return &project, err
}

// =============== Projects implementations ===============

type projectsAPI struct {
	sdk *sdk
}

func (p *projectsAPI) Get(projectID int) (*Project, error) {
	pathStr := path.Join("/api/projects", strconv.Itoa(projectID))
	var project Project
	err := p.sdk.doRequest(http.MethodGet, pathStr, nil, &project)
	return &project, err
}

func (p *projectsAPI) Update(projectID int, req *UpdateProjectRequest) (*Project, error) {
	pathStr := path.Join("/api/projects", strconv.Itoa(projectID))
	var project Project
	err := p.sdk.doRequest(http.MethodPut, pathStr, req, &project)
	return &project, err
}

func (p *projectsAPI) Patch(projectID int, patches []PatchProjectRequest) (*Project, error) {
	pathStr := path.Join("/api/projects", strconv.Itoa(projectID))
	var project Project
	err := p.sdk.doRequest(http.MethodPatch, pathStr, patches, &project)
	return &project, err
}

func (p *projectsAPI) Delete(projectID int) error {
	pathStr := path.Join("/api/projects", strconv.Itoa(projectID))
	return p.sdk.doRequest(http.MethodDelete, pathStr, nil, nil)
}

func (p *projectsAPI) ListUsers(projectID int, params *ListParams) (*UsersListResponse, error) {
	query := buildQueryParams(params)
	pathURL := &url.URL{
		Path:     path.Join("/api/projects", strconv.Itoa(projectID), "users"),
		RawQuery: query.Encode(),
	}
	var resp UsersListResponse
	err := p.sdk.doRequest(http.MethodGet, pathURL.String(), nil, &resp)
	return &resp, err
}

func (p *projectsAPI) AddUser(projectID, userID int) error {
	pathStr := path.Join("/api/projects", strconv.Itoa(projectID), "users")
	req := AddUserToProjectRequest{UserID: userID}
	return p.sdk.doRequest(http.MethodPost, pathStr, req, nil)
}

func (p *projectsAPI) RemoveUser(projectID, userID int) error {
	pathStr := path.Join("/api/projects", strconv.Itoa(projectID), "users", strconv.Itoa(userID))
	return p.sdk.doRequest(http.MethodDelete, pathStr, nil, nil)
}

// =============== Roles implementations ===============

type rolesAPI struct {
	sdk *sdk
}

func (r *rolesAPI) List() (*RolesListResponse, error) {
	var resp RolesListResponse
	err := r.sdk.doRequest(http.MethodGet, "/api/roles", nil, &resp)
	return &resp, err
}

func (r *rolesAPI) Create(req *CreateRoleRequest) (*Role, error) {
	var role Role
	err := r.sdk.doRequest(http.MethodPost, "/api/roles", req, &role)
	return &role, err
}

func (r *rolesAPI) Delete(roleID int) error {
	pathStr := path.Join("/api/roles", strconv.Itoa(roleID))
	return r.sdk.doRequest(http.MethodDelete, pathStr, nil, nil)
}

// =============== Audits implementations ===============

type auditsAPI struct {
	sdk *sdk
}

func (a *auditsAPI) List(params *ListParams) (*AuditsListResponse, error) {
	query := buildQueryParams(params)
	pathURL := &url.URL{
		Path:     "/api/audits",
		RawQuery: query.Encode(),
	}
	var resp AuditsListResponse
	err := a.sdk.doRequest(http.MethodGet, pathURL.String(), nil, &resp)
	return &resp, err
}
