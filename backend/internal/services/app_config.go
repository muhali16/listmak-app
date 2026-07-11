package services

import (
	"strconv"
	"sync/atomic"

	"github.com/muhali16/listmak-service/internal/repository"
)

const (
	settingTestingMode    = "testing_mode"
	settingFireworksModel = "fireworks_model"
)

// AppConfig exposes runtime, admin-togglable app settings. Values are cached in
// memory (read on hot paths) and persisted to the DB on write.
type AppConfig interface {
	TestingMode() bool
	SetTestingMode(v bool) error
	FireworksModel() string
	SetFireworksModel(m string) error
}

type appConfig struct {
	repo    repository.AppSettingRepository
	testing atomic.Bool
	model   atomic.Pointer[string]
}

// NewAppConfig loads settings from the DB, falling back to the given defaults
// (seeded from env) when unset.
func NewAppConfig(repo repository.AppSettingRepository, defaultTesting bool, defaultModel string) AppConfig {
	c := &appConfig{repo: repo}

	if v, _ := repo.Get(settingTestingMode); v == "" {
		c.testing.Store(defaultTesting)
	} else {
		c.testing.Store(v == "true")
	}

	m, _ := repo.Get(settingFireworksModel)
	if m == "" {
		m = defaultModel
	}
	c.model.Store(&m)

	return c
}

func (c *appConfig) TestingMode() bool {
	return c.testing.Load()
}

func (c *appConfig) SetTestingMode(v bool) error {
	if err := c.repo.Set(settingTestingMode, strconv.FormatBool(v)); err != nil {
		return err
	}
	c.testing.Store(v)
	return nil
}

func (c *appConfig) FireworksModel() string {
	if p := c.model.Load(); p != nil {
		return *p
	}
	return ""
}

func (c *appConfig) SetFireworksModel(m string) error {
	if err := c.repo.Set(settingFireworksModel, m); err != nil {
		return err
	}
	c.model.Store(&m)
	return nil
}
