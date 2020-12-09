package user

import "time"

// Info represents someone with access to the system.
type Info struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	Role         string    `json:"role"`
	PasswordHash string    `json:"password_hash"`
	DateCreated  time.Time `json:"date_created"`
	DateUpdated  time.Time `json:"date_updated"`
}

// NewUser contains information needed to create a new User.
type NewUser struct {
	Name            string `json:"name"`
	Email           string `json:"email"`
	Role            string `json:"role"`
	Password        string `json:"password"`
	PasswordConfirm string `json:"password_confirm"`
}

// =============================================================================

type id struct {
	Resp struct {
		Entities []struct {
			ID string `json:"id"`
		} `json:"entities"`
	} `json:"resp"`
}

func (id) document() string {
	return `{
		entities: user {
			id
		}
	}`
}

type result struct {
	Resp struct {
		Msg     string
		NumUids int
	} `json:"resp"`
}

func (result) document() string {
	return `{
		msg,
		numUids,
	}`
}
