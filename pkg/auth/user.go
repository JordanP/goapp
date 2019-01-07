package auth

type Who interface {
	Who() string
}

type User struct {
	Login string
	Email string
	Role  string
}

func NewUser(login, email, role string) User {
	return User{
		Login: login,
		Email: email,
		Role:  role,
	}
}

func (u User) Who() string { return u.Login }

type AdminUser struct {
	Login string
}

func NewAdminUser(login string) AdminUser {
	return AdminUser{Login: login}
}

func (u AdminUser) Who() string { return u.Login }
