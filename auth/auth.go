package auth

import (
	"fmt"

	"ec.com/database"
	m "ec.com/models"
	"github.com/go-oauth2/oauth2/v4/errors"
)

func GetUserWithPassword(username, password string) (m.User, error) {
	fmt.Println(username, password)
	if username == "test@example.com" {
		var user m.User
		if err := database.DB.Find(&m.User{}, &m.User{Email: username}).Scan(&user).Error; err != nil {
			fmt.Println(err)
			return m.User{}, errors.ErrAccessDenied
		}
		if password == user.Password {
			return user, nil
		}
	} else {
		var agency m.Agency
		if err := database.DB.Find(&m.Agency{}, &m.Agency{Login: username}).Scan(&agency).Error; err != nil {
			fmt.Println(err)
			return m.User{}, errors.ErrAccessDenied
		}
		if password == agency.Password {
			return m.User{
				ID:     agency.ID,
				Agency: agency.Name,
				Email:  agency.Login,
			}, nil
		}
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
