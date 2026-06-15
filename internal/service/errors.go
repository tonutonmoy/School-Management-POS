package service

import "errors"

var (
	ErrNotFound           = errors.New("resource not found")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrForbidden          = errors.New("forbidden")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrInactiveUser       = errors.New("user account is inactive")
	ErrEmailExists        = errors.New("email already exists")
	ErrRoleExists         = errors.New("role slug already exists")
	ErrSystemRole         = errors.New("system roles cannot be modified")
	ErrInvalidToken       = errors.New("invalid or expired token")
	ErrPasswordMismatch   = errors.New("current password is incorrect")
	ErrValidation         = errors.New("validation failed")
	ErrSessionConflict    = errors.New("only one academic session can be active")
)
