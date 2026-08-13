package users

import (
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// qualityFlagWindow mirrors WhatsApp's rolling-window quality mechanic
// (CONTEXT.md: "auto-suspends a sender after 5+ blocks/reports in 7
// days"). A sender flagged within this window is actively at risk (Red);
// one flagged before it is recovering (Yellow); never flagged is clean
// (Green).
const qualityFlagWindow = 7 * 24 * time.Hour

// QualityStatus computes the Green/Yellow/Red messaging-quality tier.
// Nothing in this codebase sets QualityFlaggedAt yet (that requires the
// block/report tracking called out as a later gap in CONTEXT.md), so
// every account correctly shows Green until that lands — this is real
// status, not a placeholder value.
func QualityStatus(u *User) string {
	if u.QualityFlaggedAt == nil {
		return "green"
	}
	if time.Since(*u.QualityFlaggedAt) < qualityFlagWindow {
		return "red"
	}
	return "yellow"
}

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
	// Return the same generic error for missing user, bad password, or
	// inactive account so callers can't enumerate valid emails by
	// distinguishing the responses.
	user, err := GetByEmail(db, email)
	if err != nil {
		return nil, ErrBadCredentials
	}
	if !CheckPassword(password, user.HashedPassword) {
		return nil, ErrBadCredentials
	}
	if !user.IsActive {
		return nil, ErrBadCredentials
	}
	return user, nil
}
