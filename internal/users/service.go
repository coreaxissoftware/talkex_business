package users

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	ErrEmailTaken           = errors.New("email already registered")
	ErrUserNotFound         = errors.New("user not found")
	ErrInactiveUser         = errors.New("account is inactive")
	ErrBadCredentials       = errors.New("incorrect email or password")
	ErrWrongCurrentPassword = errors.New("current password is incorrect")
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

// UpdateProfileInput uses pointers so the handler can distinguish
// "field omitted" from "field cleared" the same way contacts/templates do.
type UpdateProfileInput struct {
	FullName         *string
	BusinessCategory *string
}

func UpdateProfile(db *gorm.DB, user *User, in *UpdateProfileInput) (*User, error) {
	if in.FullName != nil {
		user.FullName = *in.FullName
	}
	if in.BusinessCategory != nil {
		user.BusinessCategory = in.BusinessCategory
	}
	if err := db.Save(user).Error; err != nil {
		return nil, err
	}
	return user, nil
}

// ChangePassword verifies the current password before setting a new one —
// standard defense against a stolen session token being used to lock the
// real owner out.
func ChangePassword(db *gorm.DB, user *User, currentPassword, newPassword string) error {
	if !CheckPassword(currentPassword, user.HashedPassword) {
		return ErrWrongCurrentPassword
	}
	hashed, err := HashPassword(newPassword)
	if err != nil {
		return err
	}
	user.HashedPassword = hashed
	return db.Save(user).Error
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
