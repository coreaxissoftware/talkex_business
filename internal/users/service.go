package users

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	ErrEmailTaken    = errors.New("email already registered")
	ErrUserNotFound  = errors.New("user not found")
	ErrInactiveUser  = errors.New("account is inactive")
	ErrBadCredentials = errors.New("incorrect email or password")
)

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func CheckPassword(password, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

func CreateUser(db *gorm.DB, email, password, fullName string) (*User, error) {
	var existing User
	if err := db.Where("email = ?", email).First(&existing).Error; err == nil {
		return nil, ErrEmailTaken
	}

	hashed, err := HashPassword(password)
	if err != nil {
		return nil, err
	}

	user := &User{
		Email:          email,
		HashedPassword: hashed,
		FullName:       fullName,
		Role:           RoleOwner,
		IsActive:       true,
	}
	if err := db.Create(user).Error; err != nil {
		return nil, err
	}
	return user, nil
}

func GetByID(db *gorm.DB, id string) (*User, error) {
	var user User
	if err := db.First(&user, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

func GetByEmail(db *gorm.DB, email string) (*User, error) {
	var user User
	if err := db.First(&user, "email = ?", email).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

func Authenticate(db *gorm.DB, email, password string) (*User, error) {
	user, err := GetByEmail(db, email)
	if err != nil {
		return nil, ErrBadCredentials
	}
	if !CheckPassword(password, user.HashedPassword) {
		return nil, ErrBadCredentials
	}
	if !user.IsActive {
		return nil, ErrInactiveUser
	}
	return user, nil
}
