# Go Homework - 项目管理系统 API

本项目是一个基于 RESTful API 的项目管理系统，支持用户、团队、项目的完整生命周期管理，包含简单的权限控制、角色管理和审计日志等功能。

## 模型设计

### 核心模型

系统包含以下 5 个主要模型。
本文列出这些模型的主要字段，开发中是否需求更多字段由开发者根据 OpenAPI 文档决定。
但要求开发者设计的模型能满足 OpenAPI 所述的需求。

本文描述模型有缺漏之处、或与 OpenAPI 不一致的，以 OpenAPI 文档为准。
开发者遇到这种情况请及时告知文档作者。

#### 1. User (用户)

```
User {
  id: integer
  username: string（系统内唯一）
  email: string (可选，系统内唯一)
  nickname: string (可选，默认同 username)
  logo: string
}
```

**特殊用户：**
- `admin` 用户：系统初始化时创建，用户名 `admin`，初始密码 `admin`
- admin 用户拥有超然权限。但不能删除自身，不能解绑 admin Role

#### 2. Role (角色)
```
Role {
  id: integer
  name: string (全局唯一)
  type: enum ["System", "Custom"]
  desc: string
}
```

**系统角色 (System Roles)：**
系统初始化时必须创建以下 3 个系统角色：
- `admin` - 管理员角色，与 admin 用户绑定
- `team leader` - 团队领导角色，用户成为 Team Leader 时自动绑定
- `normal user` - 普通用户角色，新用户创建时自动绑定

**注意：** Role 在本系统中仅作为维表资源，无实际权限判定作用。

#### 3. Team (团队)
```
Team {
  id: integer
  name: string (系统内唯一)
  desc: string
}
```

Team 可以有一个 Leader，可以有多个 Projects。

#### 4. Project (项目)
```
Project {
  id: integer
  name: string (Team 内唯一)
  desc: string
  status: enum ["WAIT_FOR_SCHEDULE", "IN_PROGRESS", "FINISHED"]
}
```

**项目状态：**
- `WAIT_FOR_SCHEDULE` - 默认状态，待调度
- `IN_PROGRESS` - 进行中
- `FINISHED` - 已完成

Project 任意时刻都应当处于一个状态。

#### 5. Audit (审计日志)
```
Audit {
  id: integer
  content: string (日志内容，格式自定义)
}
```

### 模型关系

```
┌─────────────────────────────────────────────────────────────┐
│                        关系图谱                              │
└─────────────────────────────────────────────────────────────┘

User ←───────┐
  │          │ N:M
  │          │ (用户-角色)
  ↓          │
Role ←───────┘

本项目设计的 Role 比较简单，并没有关联权限。
一个 User 可以绑定多个 Roles。
一个 Role 可以被多个 User 绑定。
User 上绑定的 Role 可以被解绑。
User 或 Role 被删除都不会令对方被级联删除。
更多业务逻辑见 OpenAPI 文档。

User ←───────┐
  │          │ N:M
  │          │ (团队成员)
  ↓          │
Team ←───────┘
  │
  │ 1:1 (可选)
  │ (Team Leader)
  ↓
User

一个 User 可以加入多个 Teams。
一个 Team 可以容纳 多个 Users。
Team 可以没有 Team Leader，最多可以有一个 Team Leader。
User 可以不担任任何 Team Leader，也可以担任任意多个 Team 的 Leader。

Team ─────→ Project
  │           (1:N)
  └─────────→ (级联删除)

Project belongs to a Team.
一个 Team 名下可以有多个 Project。
一个 Project 属于且只能属于一个 Team。

User ←───────┐
  │          │ N:M
  │          │ (项目成员)
  ↓          │
Project ←────┘

User 可以参与其 Team 下的 Project。
一个 User 可以参与多个 Projects。
一个 Project 可以有多个 User 共同参与。

Team ─────→ User
  │          |
  └───────→ User
User 之间的可见性：
只有两个 User 处于同一个 Team 时，才能互相可见对方。
否则双方不能通过任何 API 查看到对方。
```

#### 关系说明

更多关系说明见 OpenAPI 文档。

---

## 业务逻辑

### 用户管理

#### 用户创建与首次登录
1. **创建用户** (仅 admin)
   - 创建时指定用户名和初始密码
   - 自动绑定 `normal user` Role
   
2. **首次登录限制**
   - 使用初始密码首次登录后，必须先修改密码
   - 在修改密码前访问其他接口返回 `403 Forbidden`
   - 修改密码后原会话失效，需重新登录
   - 此限制适用于所有用户（包括 admin）

3. **用户可见性**
   ```
   示例：
   - UserA in [TeamX, TeamY]
   - UserB in [TeamY, TeamZ]
   - UserC in [TeamZ, TeamW]
   
   结果：
   - UserA 和 UserB 可见（同在 TeamY）
   - UserA 和 UserC 不可见（无共同 Team）
   - UserB 和 UserC 可见（同在 TeamZ）
   ```

#### 用户操作
- **登录：** 支持用户名或邮箱登录
- **登出：** 会话失效
- **更新个人信息：** 可修改 email, nickname, logo
- **修改密码：** 会使当前会话失效
- **退出团队：** `/api/me/teams/{team_id}` (DELETE)
- **退出项目：** `/api/me/projects/{project_id}` (DELETE)

### 团队管理

#### 团队生命周期
1. **创建团队** (仅 admin)
   - 创建时可指定名称和描述
   - 初始无 Leader，无成员

2. **添加/移除成员**
   - admin 和 Team Leader 可以操作
   - Leader 只能添加/移除对自己可见的用户

3. **设置 Team Leader**
   - admin 和 当前 Leader 可以操作
   - 更多操作见 OpenAPI 文档

4. **删除团队** (admin 或 Leader)
   - 更多操作见 OpenAPI 文档

### 项目管理

#### 项目生命周期
1. **创建项目** (admin 或 Team Leader)
   - 在指定 Team 下创建
   - 初始状态为 `WAIT_FOR_SCHEDULE`
   - 项目名称在同一 Team 内唯一

2. **添加项目成员**
   - admin 和 Team Leader 可以操作
   - **自动添加到 Team：** 如果用户不在 Project 所属的 Team 中，会自动加入该 Team
   
3. **更新项目**
   - admin 和 Team Leader 可以更新项目信息
   - 支持 PUT（完整更新）和 PATCH（部分更新）
   - 可更新的字段：name, desc, status

4. **状态转换**
   ```
   WAIT_FOR_SCHEDULE → IN_PROGRESS → FINISHED
   ```

5. **删除项目** (admin 或 Team Leader)
   - 项目成员不会被删除
   - 项目成员仍保留在 Team 中

### 角色管理

#### 系统角色 (System Roles)
- 不能删除
- 不能手动绑定/解绑
- 由系统自动管理

#### 自定义角色 (Custom Roles)
- **创建：** 仅 admin
- **删除：** 仅 admin
- **绑定/解绑：** 仅 admin

更多操作见 OpenAPI 文档。

### 审计日志

#### 记录范围
应当记录以下操作：
- ✅ 用户登录/登出
- ✅ 修改密码
- ✅ 创建/删除用户
- ✅ 创建/更新/删除团队
- ✅ 添加/移除团队成员
- ✅ 设置/更换 Team Leader
- ✅ 创建/更新/删除项目
- ✅ 添加/移除项目成员
- ✅ 创建/删除角色
- ✅ 绑定/解绑角色

开发者可以依据实际情况决定是否记录更多日志。

不建议记录：
- ❌ 查询操作（GET 请求）
- ❌ 健康检查

#### 日志格式
格式由实现者自定义，建议包含：
```
{谁} 在 {什么时间} {做了什么操作} {结果如何}

示例：
- "admin (ID:1) 于 2025-01-15 10:30:45 创建了用户 testuser (ID:10) - 成功"
- "user123 (ID:5) 于 2025-01-15 10:35:20 更新了团队 DevTeam (ID:3) 的描述 - 成功"
- "teamleader (ID:8) 于 2025-01-15 10:40:00 将项目 ProjectX (ID:15) 状态更新为 IN_PROGRESS - 成功"
```

#### 隐私

注意某些字段可能不能被明文记录。

---

## 权限系统

### 权限角色对照表

| 操作 | admin | Team Leader | 普通成员 | 非成员 |
|------|-------|-------------|----------|--------|
| **用户管理** |
| 创建用户 | ✅ | ❌ | ❌ | ❌ |
| 删除用户 | ✅ (除自己) | ❌ | ❌ | ❌ |
| 查看用户列表 | ✅ (全部) | ✅ (可见) | ✅ (可见) | ✅ (可见) |
| 查看用户详情 | ✅ | ✅ (可见) | ✅ (可见) | ❌ |
| **团队管理** |
| 创建团队 | ✅ | ❌ | ❌ | ❌ |
| 删除团队 | ✅ | ✅ (自己的) | ❌ | ❌ |
| 更新团队 | ✅ | ✅ (自己的) | ❌ | ❌ |
| 设置 Leader | ✅ | ✅ (自己的) | ❌ | ❌ |
| 添加成员 | ✅ | ✅ (自己的) | ❌ | ❌ |
| 移除成员 | ✅ | ✅ (自己的) | ❌ | ❌ |
| 查看团队 | ✅ (全部) | ✅ (自己的) | ✅ (自己的) | ❌ |
| **项目管理** |
| 创建项目 | ✅ | ✅ (自己团队) | ❌ | ❌ |
| 删除项目 | ✅ | ✅ (自己团队) | ❌ | ❌ |
| 更新项目 | ✅ | ✅ (自己团队) | ❌ | ❌ |
| 添加成员 | ✅ | ✅ (自己团队) | ❌ | ❌ |
| 移除成员 | ✅ | ✅ (自己团队) | ❌ | ❌ |
| 查看项目 | ✅ | ✅ (自己团队) | ✅ (参与的) | ❌ |
| **角色管理** |
| 创建角色 | ✅ | ❌ | ❌ | ❌ |
| 删除角色 | ✅ | ❌ | ❌ | ❌ |
| 绑定角色 | ✅ | ❌ | ❌ | ❌ |
| 查看角色 | ✅ | ✅ | ✅ | ✅ |
| **审计日志** |
| 查看审计日志 | ✅ | ❌ | ❌ | ❌ |

### 特殊权限规则

#### Admin 权限
- ✅ 可以查询和操作所有资源
- ✅ 可以查看所有用户（不受可见性限制）
- ❌ 不能删除自身
- ❌ 不能解绑自身的 admin Role

#### Team Leader 权限
- ✅ 可以管理自己负责的团队及其项目
- ✅ 自动绑定 `team leader` Role
- ✅ 可以更换自己团队的 Leader（包括卸任）
- ⚠️ 只能操作对自己可见的用户
- ⚠️ 卸任或被移除时自动解绑 team leader Role

#### 密码相关
- 🔐 首次登录必须修改密码
- 🔐 修改密码会使当前会话失效
- 🔐 密码要求：8-30个字符，仅字母数字下划线连字符

#### 用户可见性
- 👥 同一 Team 的用户互相可见
- 🔍 列表接口只返回可见用户
- ⛔ 不能查看不可见用户的详情
- 🌐 admin 不受可见性限制

---

## 一致性测试

一致性测试旨在保障同一服务约定的不同实现者在关键功能上表现的一致性。

开发者实现的应用程序应当通过一致性测试的所有用例。

本项目提供了完整的 BDD 风格一致性测试套件，使用 Ginkgo v2 + Gomega 编写。

### 测试文件结构

```
conformance/
├── suit.go              # 测试套件初始化
├── user.go              # 用户、认证、角色、审计测试
├── team.go              # 团队管理测试
└── project.go           # 项目管理测试
```

### 如何编写测试

在你的项目中添加

```
conformance/
└── e2e_test.go           # 组织一致性测试入口代码
e2e/
└── e2e_test.go           # 组织 e2e 测试入口代码
```

```go
package conformance

import (
	"testing"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	_ "github.com/dspo/go-homework/conformance"
	"github.com/dspo/go-homework/sdk"
)

func TestConformance(t *testing.T) {
	// do your init jobs, e.g. deploy service, database, prepare data

	var _ = sdk.NewSDK("http://my_app_server:8080") // init sdk with your app server address

	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "Conformance Suite")
}

```

### 如何运行测试

```bash
go test ./conformance -v
```

## e2e 测试

一致性测试是本项目的设计者为保障所有开发者在基础功能实现上的完整性、一致性而编写的。

实际开发和业务中，还有很多边缘用例、与实现相关的个性化用例没能在一致性测试中覆盖到，对于这些用例，需要开发者自行编写 e2e 测试进行覆盖。

---

## 作业要求

### 技术栈

- ✅ 主要语言：Go
- ✅ 主要 Web 框架：Gin
- ✅ 必须使用 GORM 作为 Database Access 的主要库
- ✅ 必须使用 Ginkgo + Gomega 库进行测试

### 功能完整性

- ✅ 必须实现 `openapi.yaml` 中约定的所有 APIs 和功能。
- ✅ 必须实现服务配置、日志初始化、审计日志记录等一般应用程序所需的组件。
- ✅ 应当实现中间件以复用处理 API 的基本能力。
- ✅ 应当正确处理模型关系和业务逻辑：层级、级联、相关、权限，以及其他在服务端开发中**不言自明**的业务逻辑。

### 代码质量

- ✅ Go 版本应当 go1.24+
- ✅ 项目应当用 `go.mod` 组织，不要用其他方式组织依赖。
- ✅ 注意错误处理：一般情况下应当处理 err，不处理时应当有充分的理由。
- ✅ 代码应当整洁、格式化、健壮、注意性能（虽然本项目足够简单）。
- ✅ 代码应表现出对 Go 语言特性的运用。

### 测试

- ✅ 不要求进行单元测试，但提倡进行单元测试。
- ✅ **项目应当通过全部一致性测试用例**。注意：这条很重要。
- ✅ 项目在一致性测试之外，应当补充充分的 e2e 测试。

### 脚本

项目可以用一些脚本语言进行必要的组织，如使用 shell、python 等脚本来组织项目构建、测试服务编排等。

项目要求在根目录提供 `Makefile`，至少提供两个命令用以验收项目：

```cmake
conformance-test:
    # 执行该命令运行一致性测试

e2e-test:
    # 执行该命令运行开发者自己编写的 e2e 测试
```

```shell
make conformance-test
make e2e
```

祝开发顺利！🚀

---

## 常见问题

### Q: Role 有什么实际作用？
A: Role 在本系统中仅作为维表资源，无实际权限判定作用。权限判定基于用户身份（admin / Team Leader / 普通用户）和资源关系（是否在同一 Team）。

### Q: 删除团队时会发生什么？
A: 
- ✅ 团队下的所有项目会被级联删除
- ✅ 如有 Leader，会解绑其 team leader Role
- ❌ 团队成员不会被删除
- ❌ 团队成员不会从其他团队移除

### Q: 将用户添加到项目时会自动加入团队吗？
A: 是的。如果用户不在项目所属的团队中，会自动将用户添加到该团队。

### Q: 从项目移除用户会将其从团队移除吗？
A: 不会。从项目移除用户只会解除项目成员关系，不影响团队成员关系。

### Q: 用户可以同时是多个团队的 Leader 吗？
A: 可以。一个用户可以是多个团队的 Leader。

### Q: Team Leader 可以操作不可见的用户吗？
A: 不可以。Team Leader 只能操作对自己可见的用户（即至少与自己在同一个团队的用户）。

### Q: 首次登录的密码修改是如何实现的？
A: 
1. 用户使用初始密码登录成功
2. 尝试访问其他接口时返回 403（提示需要修改密码）
3. 调用 `/api/me/password` 修改密码
4. 修改成功后原会话失效
5. 使用新密码重新登录
