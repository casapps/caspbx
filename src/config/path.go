package config

import (
	"errors"
	"path"
	"regexp"
	"strings"
)

var (
	ErrPathTraversal = errors.New("path traversal attempt detected")
	ErrInvalidPath   = errors.New("invalid path characters")
	ErrPathTooLong   = errors.New("path exceeds maximum length")

	validPathSegmentPattern = regexp.MustCompile(`^[a-z0-9_-]+$`)
)

func normalizePath(input string) string {
	if input == "" {
		return ""
	}

	cleanedPath := path.Clean(input)
	cleanedPath = strings.Trim(cleanedPath, "/")

	if strings.Contains(cleanedPath, "..") {
		return ""
	}

	return cleanedPath
}

func validatePathSegment(segment string) error {
	if segment == "" {
		return ErrInvalidPath
	}
	if len(segment) > 64 {
		return ErrPathTooLong
	}
	if segment == "." || segment == ".." {
		return ErrPathTraversal
	}
	if !validPathSegmentPattern.MatchString(segment) {
		return ErrInvalidPath
	}
	return nil
}

func validatePath(input string) error {
	if len(input) > 2048 {
		return ErrPathTooLong
	}
	if strings.Contains(input, "..") {
		return ErrPathTraversal
	}

	segments := strings.Split(strings.Trim(input, "/"), "/")
	for _, segment := range segments {
		if segment == "" {
			continue
		}
		if validationError := validatePathSegment(segment); validationError != nil {
			return validationError
		}
	}

	return nil
}

func SafePath(input string) (string, error) {
	if validationError := validatePath(input); validationError != nil {
		return "", validationError
	}
	return normalizePath(input), nil
}

func NormalizeBaseURL(input string) (string, error) {
	if strings.TrimSpace(input) == "" || strings.TrimSpace(input) == "/" {
		return "/", nil
	}
	if strings.Contains(input, "..") {
		return "", ErrPathTraversal
	}

	cleanedPath := path.Clean("/" + strings.TrimSpace(input))
	segments := strings.Split(strings.Trim(cleanedPath, "/"), "/")
	for _, segment := range segments {
		if validationError := validatePathSegment(segment); validationError != nil {
			return "", validationError
		}
	}

	return cleanedPath, nil
}
