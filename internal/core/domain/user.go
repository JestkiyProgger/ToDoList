package domain

import (
	"fmt"
	"regexp"

	core_errors "github.com/JestkiyProgger/ToDoList/internal/core/errors"
)

type User struct {
	ID          int
	Version     int
	Name        string
	PhoneNumber *string
}

func NewUser(id int, version int, name string, phoneNumber *string) User {
	return User{
		ID:          id,
		Version:     version,
		Name:        name,
		PhoneNumber: phoneNumber,
	}
}

func NewUserUninitialized(name string, phoneNumber *string) User {
	return NewUser(UninitializedID, UninitializedVersion, name, phoneNumber)
}

func (u *User) Validate() error {
	lenName := len([]rune(u.Name))
	if lenName < 3 || lenName > 100 {
		return fmt.Errorf("invalid \"Name\" len: %d, %w", lenName, core_errors.ErrInvalidArgument)
	}

	if u.PhoneNumber != nil {
		lenPhoneNumber := len([]rune(*u.PhoneNumber))
		if lenPhoneNumber < 10 || lenPhoneNumber > 15 {
			return fmt.Errorf("invalid \"Phone number\" len: %d, %w", lenPhoneNumber, core_errors.ErrInvalidArgument)
		}

		re := regexp.MustCompile(`^\+[0-9]+$`)

		if !re.MatchString(*u.PhoneNumber) {
			return fmt.Errorf("invalid \"Phone number\" format: %s, %w", *u.PhoneNumber, core_errors.ErrInvalidArgument)
		}
	}

	return nil
}

type UserPatch struct {
	Name        Nullable[string]
	PhoneNumber Nullable[string]
}

func (up *UserPatch) Validate() error {
	if up.Name.Set && up.Name.Value == nil {
		return fmt.Errorf("'name' can't be patched to NULL: %w", core_errors.ErrInvalidArgument)
	}

	return nil
}

func (u *User) ApplyPatch(us UserPatch) error {
	if err := us.Validate(); err != nil {
		return fmt.Errorf("validate user patch: %w", err)
	}

	tmp := *u

	if us.Name.Set {
		tmp.Name = *us.Name.Value
	}
	if us.PhoneNumber.Set {
		tmp.PhoneNumber = us.PhoneNumber.Value
	}

	if err := tmp.Validate(); err != nil {
		return fmt.Errorf("validate patched user: %w", err)
	}

	*u = tmp

	return nil
}
