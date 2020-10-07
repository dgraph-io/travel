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

type addResult struct {
	AddUser struct {
		User []struct {
			ID string `json:"id"`
		} `json:"user"`
	} `json:"addUser"`
}

func (addResult) document() string {
	return `{
		user {
			id
		}
	}`
}

type updateResult struct {
	UpdateUser struct {
		Msg     string
		NumUids int
	} `json:"updateUser"`
}

func (updateResult) document() string {
	return `{
		msg,
		numUids,
	}`
}

type deleteResult struct {
	DeleteUser struct {
		Msg     string
		NumUids int
	} `json:"deleteUser"`
}

func (deleteResult) document() string {
	return `{
		msg,
		numUids,
	}`
}
