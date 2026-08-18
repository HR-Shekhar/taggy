package sqlc

func (u User) IsAdmin() bool {
	return u.Role == UserRoleADMIN
}
