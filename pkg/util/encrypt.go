package util

import "golang.org/x/crypto/bcrypt"

func HashPassword(pwd string) (string, error) {
	ihash, err := bcrypt.GenerateFromPassword([]byte(pwd), 12)
	return string(ihash), err
}

func CheckPwd(pwd string, hashedPwd string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPwd), []byte(pwd))
	return err == nil
}
