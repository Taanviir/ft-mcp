# 42 API Reference

> **Status:** The API is approaching End of Life (EOL) and will eventually be deprecated. It still works fully as of May 2026.
>
> **Full live docs:** https://api.intra.42.fr/apidoc
> **Full OpenAPI spec (JSON):** https://api.intra.42.fr/apidoc/2.0.json
> **Full OpenAPI spec (HTML):** https://api.intra.42.fr/apidoc/2.0.html
> **Guides:** https://api.intra.42.fr/apidoc/guides

---

## Overview

- **Base URL:** `https://api.intra.42.fr/v2`
- **Protocol:** HTTPS only (connection refused on HTTP)
- **Format:** JSON — all responses are JSON, blank fields are `null` (not omitted)
- **Timestamps:** ISO 8601
- **Total endpoints:** 739 across ~60 resource types
- **Auth:** OAuth 2.0

---

## Authentication

Register your app at https://profile.intra.42.fr/oauth/applications/new to get a **Client UID** and **Client Secret**.

### Flow 1 — Client Credentials (server-to-server, no user)

Use this when you only need public data and don't act on behalf of a user.

```bash
curl -X POST https://api.intra.42.fr/oauth/token \
  -d "grant_type=client_credentials&client_id=UID&client_secret=SECRET"
```

Response:
```json
{
  "access_token": "42804d1f...",
  "token_type": "bearer",
  "expires_in": 7200,
  "scope": "public",
  "created_at": 1443451918
}
```

Token expires in **7200 seconds (2 hours)**.

### Flow 2 — Authorization Code (web app, acts as a user)

Redirect the user to:
```
GET https://api.intra.42.fr/oauth/authorize
  ?client_id=UID
  &redirect_uri=https://your.app/callback
  &response_type=code
  &scope=public
  &state=RANDOM_STRING
```

After the user approves, they're redirected to your `redirect_uri?code=ABC`. Exchange that code:

```bash
curl -X POST https://api.intra.42.fr/oauth/token \
  -F grant_type=authorization_code \
  -F client_id=UID \
  -F client_secret=SECRET \
  -F code=ABC \
  -F redirect_uri=https://your.app/callback
```

### Using a token

Pass in a header (preferred):
```
Authorization: Bearer YOUR_ACCESS_TOKEN
```

Or as a query param (fallback):
```
GET /v2/users?access_token=YOUR_ACCESS_TOKEN
```

### Inspecting your token

```bash
curl -H "Authorization: Bearer TOKEN" https://api.intra.42.fr/oauth/token/info
# {"resource_owner_id":74,"scopes":["public"],"expires_in_seconds":7174,...}
```

---

## Scopes

Request only the scopes you actually need. Insufficient scope returns `403` with a `WWW-Authenticate` header explaining what's missing.

| Scope | Access granted |
|-------|---------------|
| `public` | Public data: users, campus, cursus, projects, events, achievements, etc. |
| `profile` | User's own profile details |
| `projects` | Subscribe to / interact with projects as a user |
| `forum` | Create/update topics and messages |
| `elearning` | Access e-learning content |
| `tig` | Community services |

---

## Errors

| HTTP Code | Meaning |
|-----------|---------|
| 400 | Malformed request |
| 401 | Unauthorized (missing or invalid token) |
| 403 | Forbidden (insufficient scope or role) |
| 404 | Resource not found |
| 422 | Unprocessable entity (validation error) |
| 429 | Rate limit exceeded |
| 500 | Server error |

---

## Rate Limits

Default limits (no special role):
- **2 requests/second**
- **1200 requests/hour**

Limits are raised by application roles:
- **Official App** — higher rate limit (apply by emailing intrateam@42.fr)
- **Certified App** — highest limits + access to restricted endpoints

Current role is returned in response headers as `X-Application-Roles`.

---

## Pagination

All index endpoints are paginated. Default page size: **30**. Maximum: **100** (some endpoints cap lower).

Two equivalent parameter styles:

```
GET /v2/users?page=2&per_page=50
GET /v2/users?page[number]=2&page[size]=50
```

**Response headers:**

| Header | Value |
|--------|-------|
| `X-Page` | Current page number |
| `X-Per-Page` | Items per page |
| `X-Total` | Total number of items |
| `Link` | `first`, `prev`, `next`, `last` page URLs |

---

## Filtering

Use `filter[field]=value` to filter results. Multiple values are comma-separated.

```
GET /v2/users?filter[pool_year]=2013&filter[pool_month]=september,july
```

---

## Sorting

Use `sort=field` for ascending, `sort=-field` for descending. Comma-separate multiple fields:

```
GET /v2/users?sort=kind,-login
```

---

## Resource Reference

Endpoints marked **[restricted]** require a privileged application role (Official App, Certified App, or staff). Endpoints marked **[user]** require user-level OAuth (Authorization Code flow, not Client Credentials).

---

### Users

A 42 student, staff member, or any entity with a 42 account.

| Method | Endpoint | Notes |
|--------|----------|-------|
| GET | `/v2/users` | List all users |
| GET | `/v2/users/:id` | Get user by ID or login |
| GET | `/v2/me` | Get authenticated user's own profile |
| GET | `/v2/staff` | List staff members [restricted] |
| GET | `/v2/campus/:campus_id/users` | Users at a campus |
| GET | `/v2/cursus/:cursus_id/users` | Users in a cursus |
| GET | `/v2/coalitions/:coalition_id/users` | Users in a coalition |
| GET | `/v2/events/:event_id/users` | Users registered to an event |
| GET | `/v2/projects/:project_id/users` | Users on a project |
| GET | `/v2/groups/:group_id/users` | Users in a group |
| GET | `/v2/users/:id/locations_stats` | Login time stats for a user |

Useful filter params on `GET /v2/users`:
- `filter[login]=somelogin` — exact login match
- `filter[campus_id]=1` — filter by campus
- `filter[pool_year]=2022` — filter by piscine year
- `filter[pool_month]=october` — filter by piscine month
- `filter[kind]=student` or `staff`

---

### Campus

Physical locations where 42 students work.

| Method | Endpoint | Notes |
|--------|----------|-------|
| GET | `/v2/campus` | List all campuses |
| GET | `/v2/campus/:id` | Get a campus |
| GET | `/v2/campus/:campus_id/stats` | Campus statistics |
| GET | `/v2/campus/:campus_id/products` | Shop products at campus |

---

### Cursus

An educational cycle / curriculum (e.g. "42", "Piscine C").

| Method | Endpoint | Notes |
|--------|----------|-------|
| GET | `/v2/cursus` | List all cursus |
| GET | `/v2/cursus/:id` | Get a cursus |

---

### Cursus Users

Tracks which users are enrolled in which cursus, including level and grade.

| Method | Endpoint | Notes |
|--------|----------|-------|
| GET | `/v2/cursus_users` | All cursus memberships |
| GET | `/v2/users/:user_id/cursus_users` | A user's cursus info (level, grade, skills) |
| GET | `/v2/cursus/:cursus_id/cursus_users` | All users in a cursus |

The `cursus_users` object contains: `level`, `grade`, `skills` (array with score per skill), `begin_at`, `end_at`, `blackholed_at`.

---

### Projects

Pedagogic projects belonging to a cursus.

| Method | Endpoint | Notes |
|--------|----------|-------|
| GET | `/v2/projects` | List all projects |
| GET | `/v2/projects/:id` | Get a project |
| GET | `/v2/cursus/:cursus_id/projects` | Projects in a cursus |
| GET | `/v2/me/projects` | Authenticated user's projects [user] |
| GET | `/v2/project_sessions` | Project sessions (configuration per campus/cursus) |
| GET | `/v2/project_sessions/:id` | Get a project session |
| GET | `/v2/projects/:project_id/project_sessions` | Sessions for a project |

---

### Projects Users

Tracks user submissions on projects (the `projects_users` join table).

| Method | Endpoint | Notes |
|--------|----------|-------|
| GET | `/v2/projects_users` | All project submissions |
| GET | `/v2/users/:user_id/projects_users` | A user's project submissions |
| GET | `/v2/projects/:project_id/projects_users` | All submissions for a project |

Key fields: `status` (`in_progress`, `finished`, `searching_a_group`, etc.), `final_mark`, `validated?`, `occurrence`, `marked_at`.

---

### Locations

Tracks which computer a user is logged into at a campus. `end_at: null` means currently active.

| Method | Endpoint | Notes |
|--------|----------|-------|
| GET | `/v2/locations` | All locations (current and historical) |
| GET | `/v2/campus/:campus_id/locations` | Locations at a campus |
| GET | `/v2/users/:user_id/locations` | A user's location history |

Filter for currently active: `filter[active]=true` or `filter[end_at]=null`.
Key fields: `host` (machine name), `begin_at`, `end_at`.

---

### Events

Events hosted at a campus or within a cursus.

| Method | Endpoint | Notes |
|--------|----------|-------|
| GET | `/v2/events` | All events |
| GET | `/v2/campus/:campus_id/events` | Events at a campus |
| GET | `/v2/cursus/:cursus_id/events` | Events in a cursus |
| GET | `/v2/campus/:campus_id/cursus/:cursus_id/events` | Scoped to both |
| GET | `/v2/users/:user_id/events` | Events a user is registered to |
| GET | `/v2/events/:id` | Get an event |

---

### Events Users

Users registered to an event.

| Method | Endpoint | Notes |
|--------|----------|-------|
| GET | `/v2/events_users` | All registrations |
| GET | `/v2/events/:event_id/events_users` | Registrations for an event |
| GET | `/v2/users/:user_id/events_users` | Events a user is registered to |

---

### Achievements

Meta-goals/badges earned by users along their progression.

| Method | Endpoint | Notes |
|--------|----------|-------|
| GET | `/v2/achievements` | All achievements |
| GET | `/v2/achievements/:id` | Get an achievement |
| GET | `/v2/cursus/:cursus_id/achievements` | Achievements in a cursus |
| GET | `/v2/campus/:campus_id/achievements` | Achievements at a campus |
| GET | `/v2/users/:user_id/achievements` | A user's earned achievements |
| GET | `/v2/achievements_users` | All achievement-user records |

---

### Coalitions

Groups of users competing inside a bloc (the gamification system).

| Method | Endpoint | Notes |
|--------|----------|-------|
| GET | `/v2/coalitions` | All coalitions |
| GET | `/v2/coalitions/:id` | Get a coalition |
| GET | `/v2/blocs/:bloc_id/coalitions` | Coalitions in a bloc |
| GET | `/v2/users/:user_id/coalitions` | A user's coalitions |
| GET | `/v2/coalitions_users` | All coalition memberships |
| GET | `/v2/coalitions/:coalition_id/coalitions_users` | Members of a coalition |
| GET | `/v2/blocs` | List blocs |

---

### Scale Teams

Peer-evaluations (defences) of project teams.

| Method | Endpoint | Notes |
|--------|----------|-------|
| GET | `/v2/scale_teams` | All scale teams |
| GET | `/v2/scale_teams/:id` | Get a scale team |
| GET | `/v2/users/:user_id/scale_teams` | Scale teams involving a user |
| GET | `/v2/users/:user_id/scale_teams/as_corrector` | As evaluator |
| GET | `/v2/users/:user_id/scale_teams/as_corrected` | As evaluated |
| GET | `/v2/projects/:project_id/scale_teams` | Scale teams for a project |
| GET | `/v2/me/scale_teams` | Current user's scale teams [user] |

---

### Teams

One or more users working on a project together.

| Method | Endpoint | Notes |
|--------|----------|-------|
| GET | `/v2/teams` | All teams |
| GET | `/v2/teams/:id` | Get a team |
| GET | `/v2/users/:user_id/teams` | Teams a user is in |
| GET | `/v2/projects/:project_id/teams` | Teams for a project |
| GET | `/v2/me/teams` | Current user's teams [user] |
| GET | `/v2/teams_uploads` | Moulinette/bot evaluation uploads |

---

### Titles

Titles a user can obtain, displayed on profile and forum.

| Method | Endpoint | Notes |
|--------|----------|-------|
| GET | `/v2/titles` | All titles |
| GET | `/v2/titles/:id` | Get a title |
| GET | `/v2/users/:user_id/titles` | A user's titles |
| GET | `/v2/titles_users` | All title-user records |

---

### Skills & Expertises

| Method | Endpoint | Notes |
|--------|----------|-------|
| GET | `/v2/skills` | All skills |
| GET | `/v2/cursus/:cursus_id/skills` | Skills in a cursus |
| GET | `/v2/expertises` | All pedagogic expertises |
| GET | `/v2/expertises_users` | Users with an expertise |
| GET | `/v2/users/:user_id/expertises_users` | A user's expertises |

---

### Groups

Named groups users can belong to (displayed as labels on profile).

| Method | Endpoint | Notes |
|--------|----------|-------|
| GET | `/v2/groups` | All groups |
| GET | `/v2/groups/:id` | Get a group |
| GET | `/v2/users/:user_id/groups` | A user's groups |
| GET | `/v2/groups_users` | All group memberships |

---

### Partnerships

Pedagogic partnerships between users.

| Method | Endpoint | Notes |
|--------|----------|-------|
| GET | `/v2/partnerships` | All partnerships |
| GET | `/v2/partnerships_users` | All partnership memberships |
| GET | `/v2/users/:user_id/partnerships_users` | A user's partnerships (via `patroning`/`patroned` in user object) |

---

### Notions & Subnotions

E-learning content within a cursus.

| Method | Endpoint | Notes |
|--------|----------|-------|
| GET | `/v2/notions` | All notions |
| GET | `/v2/cursus/:cursus_id/notions` | Notions in a cursus |
| GET | `/v2/subnotions` | All subnotions |
| GET | `/v2/notions/:notion_id/subnotions` | Subnotions of a notion |

---

### Tags

Non-hierarchical keywords attached to entities.

| Method | Endpoint | Notes |
|--------|----------|-------|
| GET | `/v2/tags` | All tags |
| GET | `/v2/projects/:project_id/tags` | Tags on a project |
| GET | `/v2/tags_users` | Tag-user associations |

---

### Slots

Time slots available for booking peer evaluations.

| Method | Endpoint | Notes |
|--------|----------|-------|
| GET | `/v2/slots` | All slots |
| GET | `/v2/users/:user_id/slots` | A user's available slots |
| GET | `/v2/me/slots` | Current user's slots [user] |

---

### Feedbacks

Feedback on scale teams (evaluations) or events.

| Method | Endpoint | Notes |
|--------|----------|-------|
| GET | `/v2/feedbacks` | All feedbacks |
| GET | `/v2/scale_teams/:scale_team_id/feedbacks` | Feedback on a scale team |
| GET | `/v2/events/:event_id/feedbacks` | Feedback on an event |

---

### Correction Point Historics

| Method | Endpoint | Notes |
|--------|----------|-------|
| GET | `/v2/users/:user_id/correction_point_historics` | A user's correction point history |

---

### Exams

Exams at a campus or in a cursus.

| Method | Endpoint | Notes |
|--------|----------|-------|
| GET | `/v2/exams/:id` | Get an exam |
| GET | `/v2/campus/:campus_id/exams` | Exams at a campus [user] |
| GET | `/v2/cursus/:cursus_id/exams` | Exams in a cursus [user] |
| GET | `/v2/users/:user_id/exams` | Exams a user is registered to [user] |

---

### Community Services

Tasks a user must do for the community (often linked to a close/penalty).

| Method | Endpoint | Notes |
|--------|----------|-------|
| GET | `/v2/community_services` | All community services |
| GET | `/v2/community_services/:id` | Get one |

---

### Languages

| Method | Endpoint | Notes |
|--------|----------|-------|
| GET | `/v2/languages` | All languages |
| GET | `/v2/languages_users` | Language preferences per user |
| GET | `/v2/users/:user_id/languages_users` | A user's languages |

---

### Products & Shop

Items sold on the 42 intranet shop.

| Method | Endpoint | Notes |
|--------|----------|-------|
| GET | `/v2/products` | All products |
| GET | `/v2/campus/:campus_id/products` | Products at a campus |

---

### Roles

Grants particular privileges to users or applications.

| Method | Endpoint | Notes |
|--------|----------|-------|
| GET | `/v2/roles` | All roles |
| GET | `/v2/users/:user_id/roles` | A user's roles |

---

### Apps

OAuth2 applications registered with the 42 API.

| Method | Endpoint | Notes |
|--------|----------|-------|
| GET | `/v2/apps` | All apps |
| GET | `/v2/apps/:id` | Get an app |
| GET | `/v2/users/:user_id/apps` | Apps owned by a user |

---

### Restricted-only Resources

These require elevated application roles or staff access:

| Resource | What it is |
|----------|-----------|
| `anti_grav_units` / `anti_grav_units_users` | Anti-gravity unit system |
| `balances` | Pool evaluation point balances |
| `bloc_deadlines` | Tournament deadlines |
| `certificates` / `certificates_users` | User certificates |
| `closes` | Account closure records |
| `clusters` | Physical cluster (room) definitions |
| `companies` | Company records from jobs site |
| `dashes` / `dashes_users` | Short-time projects |
| `endpoints` | Campus network endpoints |
| `evaluations` | Raw evaluation records |
| `experiences` | Skill XP records |
| `flash_users` / `flashes` | Flash challenge system |
| `gitlab_users` | Vogsphere/GitLab user records |
| `internships` | Student internship records |
| `journals` | Campus journals |
| `levels` | Level definitions per cursus |
| `mailings` | Internal mailing records |
| `offers` / `offers_users` | Job offer subscriptions |
| `patronages` / `patronages_reports` | Mentoring records |
| `pools` | Evaluation point pools |
| `quests` / `quests_users` | Quest system (restricted) |
| `scales` | Evaluation grids |
| `scores` | Coalition score records |
| `squads` / `squads_users` | Coalition sub-groups |
| `transactions` | Altarian Dollar transactions |
| `user_candidatures` | Admission candidatures |
| `waitlists` | Event/exam waitlists |

---

## Common Patterns

### Get a user by login (not ID)

```bash
GET /v2/users/somelogin
# Works with both login string and numeric ID
```

### Get who's currently on campus

```bash
GET /v2/campus/1/locations?filter[active]=true
```

### Get a user's current level in the main 42 cursus

```bash
GET /v2/users/somelogin/cursus_users
# Find the entry where cursus.slug == "42"
# The "level" field is a float, e.g. 7.42
```

### Search users by name (partial)

The API doesn't support `LIKE` queries directly, but you can filter by login:

```bash
GET /v2/users?filter[login]=somelogin
# Exact match only. For fuzzy search, use range filter or sort then paginate.
```

### Get all projects a user completed (validated)

```bash
GET /v2/users/somelogin/projects_users?filter[status]=finished
```

---

## Response Headers

Every API response includes useful metadata:

| Header | Meaning |
|--------|---------|
| `X-Application-Id` | Your app's numeric ID |
| `X-Application-Name` | Your app's name |
| `X-Application-Roles` | Roles granted to your app |
| `X-Page` | Current page |
| `X-Per-Page` | Items per page |
| `X-Total` | Total item count |
| `Link` | Pagination links (first/prev/next/last) |

---

## Getting More Information

- **Live interactive docs:** https://api.intra.42.fr/apidoc/2.0.html — lists every endpoint with full request/response schema, filterable by resource
- **Full OpenAPI JSON spec:** https://api.intra.42.fr/apidoc/2.0.json — machine-readable schema for all 739 endpoints (also in `a81659f7-.../019e4ed5-.../api.intra.42.fr_apidoc_2.0.json.md` locally, ~1.6MB)
- **Getting started guide:** https://api.intra.42.fr/apidoc/guides/getting_started
- **Web Application Flow (OAuth for users):** https://api.intra.42.fr/apidoc/guides/web_application_flow
- **Full specification:** https://api.intra.42.fr/apidoc/guides/specification
- **Register an app:** https://profile.intra.42.fr/oauth/applications/new
- **Request Official App role** (higher rate limits): email intrateam@42.fr
