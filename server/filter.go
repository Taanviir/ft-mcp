package tools

import "encoding/json"

// filterJSON unmarshals raw into T (dropping unknown fields) then re-marshals it.
// If parsing fails it returns the raw bytes unchanged so callers always get something.
func filterJSON[T any](raw []byte) []byte {
	var v T
	if err := json.Unmarshal(raw, &v); err != nil {
		return raw
	}
	out, err := json.Marshal(v)
	if err != nil {
		return raw
	}
	return out
}

// --- 42 API response types (minimal fields only) ---

type ftUser struct {
	ID               int            `json:"id"`
	Login            string         `json:"login"`
	DisplayName      string         `json:"displayname"`
	Email            string         `json:"email"`
	CorrectionPoints int            `json:"correction_points"`
	Wallet           int            `json:"wallet"`
	Location         *string        `json:"location"`
	PoolMonth        string         `json:"pool_month"`
	PoolYear         string         `json:"pool_year"`
	Active           bool           `json:"active?"`
	Campus           []ftCampus     `json:"campus"`
	CursusUsers      []ftCursusUser `json:"cursus_users"`
}

type ftUserMin struct {
	ID          int     `json:"id"`
	Login       string  `json:"login"`
	DisplayName string  `json:"displayname"`
	Location    *string `json:"location"`
}

type ftCampus struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Country string `json:"country"`
	City    string `json:"city"`
}

type ftCampusFull struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	Country    string `json:"country"`
	City       string `json:"city"`
	UsersCount int    `json:"users_count"`
	Active     bool   `json:"active"`
	TimeZone   string `json:"time_zone"`
}

type ftCursusUser struct {
	Grade        *string   `json:"grade"`
	Level        float64   `json:"level"`
	BlackholedAt *string   `json:"blackholed_at"`
	BeginAt      string    `json:"begin_at"`
	Skills       []ftSkill `json:"skills"`
	Cursus       ftCursus  `json:"cursus"`
}

type ftSkill struct {
	Name  string  `json:"name"`
	Level float64 `json:"level"`
}

type ftCursus struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type ftProjectUser struct {
	ID        int        `json:"id"`
	FinalMark *int       `json:"final_mark"`
	Status    string     `json:"status"`
	Validated bool       `json:"validated?"`
	MarkedAt  *string    `json:"marked_at"`
	Project   ftProject  `json:"project"`
	User      ftUserMin  `json:"user"`
	Team      *ftTeam    `json:"team"`
}

type ftProject struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type ftTeam struct {
	ID    int         `json:"id"`
	Name  string      `json:"name"`
	Users []ftUserMin `json:"users"`
}

type ftAchievement struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Kind        string `json:"kind"`
	Tier        string `json:"tier"`
}

type ftLocation struct {
	ID      int       `json:"id"`
	Host    string    `json:"host"`
	BeginAt string    `json:"begin_at"`
	EndAt   *string   `json:"end_at"`
	User    ftUserMin `json:"user"`
}

type ftEvent struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Location    string `json:"location"`
	Kind        string `json:"kind"`
	MaxPeople   *int   `json:"max_people"`
	BeginAt     string `json:"begin_at"`
	EndAt       string `json:"end_at"`
}
