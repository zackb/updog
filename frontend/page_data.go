package frontend

import "github.com/zackb/updog/user"

type PageData struct {
	Title       string
	Description string
	Keywords    []string
	User        *user.User
	Message     string
}
