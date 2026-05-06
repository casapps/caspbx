package service

import (
	"fmt"

	"github.com/casapps/caspbx/src/server/model"
)

type NamespaceRegistry struct {
	usedNames map[string]string
}

func NewNamespaceRegistry(existingUsers []string, existingOrgs []string, extraReserved []string) NamespaceRegistry {
	registry := NamespaceRegistry{
		usedNames: map[string]string{},
	}

	for _, reservedName := range model.ReservedNames {
		registry.usedNames[model.NormalizeSharedName(reservedName)] = "reserved"
	}
	for _, reservedName := range extraReserved {
		registry.usedNames[model.NormalizeSharedName(reservedName)] = "reserved"
	}
	for _, username := range existingUsers {
		registry.usedNames[model.NormalizeSharedName(username)] = "user"
	}
	for _, orgSlug := range existingOrgs {
		registry.usedNames[model.NormalizeSharedName(orgSlug)] = "org"
	}

	return registry
}

func (registry NamespaceRegistry) CheckNameAvailable(name string) error {
	normalizedName := model.NormalizeSharedName(name)
	if ownerType, found := registry.usedNames[normalizedName]; found {
		return fmt.Errorf("name %q is unavailable (%s)", normalizedName, ownerType)
	}
	return nil
}

func (registry *NamespaceRegistry) ReserveUser(name string) error {
	if validationError := registry.CheckNameAvailable(name); validationError != nil {
		return validationError
	}
	registry.usedNames[model.NormalizeSharedName(name)] = "user"
	return nil
}

func (registry *NamespaceRegistry) ReserveOrg(name string) error {
	if validationError := registry.CheckNameAvailable(name); validationError != nil {
		return validationError
	}
	registry.usedNames[model.NormalizeSharedName(name)] = "org"
	return nil
}
