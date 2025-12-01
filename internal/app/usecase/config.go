package usecase

import (
	"github.com/example/gwm/internal/domain"
)

type ConfigInteractor struct {
	Service *domain.ConfigService
}

func (u *ConfigInteractor) Add(entry domain.ConfigEntry) error {
	return u.Service.Add(entry)
}

func (u *ConfigInteractor) List() ([]domain.ConfigEntry, error) {
	return u.Service.List()
}

func (u *ConfigInteractor) Remove(path string) error {
	return u.Service.Remove(path)
}
