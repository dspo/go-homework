package sdk

import (
	"net/url"
	"strconv"
	"time"
)

// Error represents an error response
type Error struct {
	Error string `json:"error"`
}

// User represents a user model
type User struct {
	ID        int     `json:"id"`
	Username  string  `json:"username"`
	Email     *string `json:"email,omitempty"`
	Nickname  *string `json:"nickname,omitempty"`
	Logo      *string `json:"logo,omitempty"`
	Roles     []Role  `json:"roles"`
	CreatedAt int64   `json:"created_at"`
	UpdatedAt int64   `json:"updated_at"`
}

// Role represents a role model
type Role struct {
	ID   int     `json:"id"`
	Name string  `json:"name"`
	Type string  `json:"type"` // System or Custom
	Desc *string `json:"desc,omitempty"`
}

// Team represents a team model
type Team struct {
	ID        int           `json:"id"`
	Name      string        `json:"name"`
	Desc      *string       `json:"desc,omitempty"`
	Leader    *User         `json:"leader,omitempty"`
	Projects  []TeamProject `json:"projects,omitempty"`
	CreatedAt int64         `json:"created_at"`
	UpdatedAt int64         `json:"updated_at"`
}

// TeamProject represents a brief project info in team details
type TeamProject struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// Project represents a project model
type Project struct {
	ID        int     `json:"id"`
	Name      string  `json:"name"`
	Desc      *string `json:"desc,omitempty"`
	Status    string  `json:"status"` // WAIT_FOR_SCHEDULE, IN_PROGRESS, FINISHED
	CreatedAt int64   `json:"created_at"`
	UpdatedAt int64   `json:"updated_at"`
}

// AuditLog represents an audit log entry
type AuditLog struct {
	ID        int    `json:"id"`
	Content   string `json:"content"`
	CreatedAt int64  `json:"created_at"`
}

// ListResponse represents a paginated list response
type ListResponse struct {
	Total int `json:"total"`
	List  any `json:"list"`
}

// UsersListResponse represents a users list response
type UsersListResponse struct {
	Total int    `json:"total"`
	List  []User `json:"list"`
}

// TeamsListResponse represents a teams list response
type TeamsListResponse struct {
	Total int    `json:"total"`
	List  []Team `json:"list"`
}

// ProjectsListResponse represents a projects list response
type ProjectsListResponse struct {
	Total int       `json:"total"`
	List  []Project `json:"list"`
}

// RolesListResponse represents a roles list response
type RolesListResponse struct {
	Total int    `json:"total"`
	List  []Role `json:"list"`
}

// AuditsListResponse represents an audit logs list response
type AuditsListResponse struct {
	Total int        `json:"total"`
	List  []AuditLog `json:"list"`
}

// LoginWithUsername represents a login request with username
type LoginWithUsername struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginWithEmail represents a login request with email
type LoginWithEmail struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// UpdateMeRequest represents a request to update current user info
type UpdateMeRequest struct {
	Email    *string `json:"email,omitempty"`
	Nickname *string `json:"nickname,omitempty"`
	Logo     *string `json:"logo,omitempty"`
}

// UpdatePasswordRequest represents a request to change password
type UpdatePasswordRequest struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

// CreateUserRequest represents a request to create a user
type CreateUserRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// CreateTeamRequest represents a request to create a team
type CreateTeamRequest struct {
	Name string  `json:"name"`
	Desc *string `json:"desc,omitempty"`
}

// UpdateTeamRequest represents a request to update a team
type UpdateTeamRequest struct {
	Name string  `json:"name"`
	Desc *string `json:"desc,omitempty"`
}

// UpdateTeamLeaderRequest represents a request to update team leader
type UpdateTeamLeaderRequest struct {
	Op    string                 `json:"op"`   // replace
	Path  string                 `json:"path"` // /leader
	Value *UpdateTeamLeaderValue `json:"value"`
}

// UpdateTeamLeaderValue represents the value for updating team leader
type UpdateTeamLeaderValue struct {
	ID *int `json:"id"`
}

// CreateProjectRequest represents a request to create a project
type CreateProjectRequest struct {
	Name string  `json:"name"`
	Desc *string `json:"desc,omitempty"`
}

// UpdateProjectRequest represents a request to update a project
type UpdateProjectRequest struct {
	Name   string  `json:"name"`
	Desc   *string `json:"desc,omitempty"`
	Status *string `json:"status,omitempty"`
}

// PatchProjectRequest represents a request to partially update a project
type PatchProjectRequest struct {
	Op    string `json:"op"`
	Path  string `json:"path"`
	Value any    `json:"value"`
}

// CreateRoleRequest represents a request to create a role
type CreateRoleRequest struct {
	Name string  `json:"name"`
	Desc *string `json:"desc,omitempty"`
}

// AddUserToTeamRequest represents a request to add a user to a team
type AddUserToTeamRequest struct {
	UserID int `json:"user_id"`
}

// AddUserToProjectRequest represents a request to add a user to a project
type AddUserToProjectRequest struct {
	UserID int `json:"user_id"`
}

// AddRoleToUserRequest represents a request to add a role to a user
type AddRoleToUserRequest struct {
	RoleID int `json:"role_id"`
}

// ListParams represents common list query parameters
type ListParams struct {
	OrderBy   *string  `json:"order_by,omitempty"`
	Page      *int     `json:"page,omitempty"`
	PageSize  *int     `json:"page_size,omitempty"`
	Keyword   *string  `json:"keyword,omitempty"`
	Name      *string  `json:"name,omitempty"`
	TeamIds   []int    `json:"team_id,omitempty"`
	RoleNames []string `json:"role_name,omitempty"`
	Leading   *bool    `json:"leading,omitempty"`
	PartIn    *bool    `json:"part_in,omitempty"`
	StartAt   *int64   `json:"start_at,omitempty"`
	EndAt     *int64   `json:"end_at,omitempty"`
}

func (p *ListParams) ToURLValues() url.Values {
	var values = make(url.Values)
	if p == nil {
		return values
	}
	if p.OrderBy != nil {
		values.Set("order_by", *p.OrderBy)
	}
	if p.Page != nil {
		values.Set("page", strconv.Itoa(*p.Page))
	}
	if p.PageSize != nil {
		values.Set("page_size", strconv.Itoa(*p.PageSize))
	}
	if p.Keyword != nil {
		values.Set("keyword", *p.Keyword)
	}
	if p.Name != nil {
		values.Set("name", *p.Name)
	}
	for _, teamId := range p.TeamIds {
		values.Add("team_id", strconv.Itoa(teamId))
	}
	for _, name := range p.RoleNames {
		values.Add("role_name", name)
	}
	if p.Leading != nil {
		values.Set("leading", strconv.FormatBool(*p.Leading))
	}
	if p.PartIn != nil {
		values.Set("part_in", strconv.FormatBool(*p.PartIn))
	}
	if p.StartAt != nil {
		values.Set("start_at", strconv.FormatInt(*p.StartAt, 10))
	}
	if p.EndAt != nil {
		values.Set("end_at", strconv.FormatInt(*p.EndAt, 10))
	}
	return values
}

// UnixToTime converts Unix timestamp to time.Time
func UnixToTime(unix int64) time.Time {
	return time.Unix(unix, 0)
}

// TimeToUnix converts time.Time to Unix timestamp
func TimeToUnix(t time.Time) int64 {
	return t.Unix()
}
