package auth

import (
	"ec.com/database"
	m "ec.com/models"
	"fmt"
	"github.com/go-oauth2/oauth2/v4/errors"
)

//123456789
//123456789

func GetUserWithPassword(username, password string) (m.User, error) {
	fmt.Println(username, password)
	var user m.User
	if err := database.DB.Find(&m.User{}, &m.User{Email: username}).Scan(&user).Error; err != nil {
		fmt.Println(err)
		return m.User{}, errors.ErrAccessDenied
	}
	fmt.Println(password, user.Password)
	if password == user.Password {
		return user, nil
	}
	return m.User{}, errors.ErrAccessDenied
}

func GetUserWithValidaCode(username, password string) (m.User, error) {
	var user m.User
	if err := database.DB.Find(&m.User{}, &m.User{Email: username}).Scan(&user).Error; err != nil {
		return m.User{}, errors.ErrAccessDenied
	}
	// TODO: Add validation code check
	return m.User{}, errors.ErrAccessDenied
}
