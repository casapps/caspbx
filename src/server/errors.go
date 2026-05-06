package server

import "errors"

var (
	ErrInvalidAPIVersion  = errors.New("invalid API version")
	ErrInvalidAdminPath   = errors.New("invalid admin path")
	ErrInvalidAdminRoute  = errors.New("invalid admin route")
	ErrUnknownTenantHost  = errors.New("unknown tenant host")
	ErrInactiveDomainHost = errors.New("inactive tenant domain")
)
