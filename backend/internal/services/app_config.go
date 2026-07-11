package services

import (
	"strconv"
	"sync/atomic"

	"github.com/muhali16/listmak-service/internal/repository"
)

const settingTestingMode = "testing_mode"

// AppConfig exposes runtime, admin-togglable app settings. TestingMode() is read
// on every listmak create/read and payment checkout, so it is cached in memory.
type AppConfig interface {
	TestingMode() bool
	SetTestingMode(v bool) error
}

type appConfig struct {
	repo    repository.AppSettingRepository
	testing atomic.Bool
}

// NewAppConfig loads testing_mode from the DB, falling back to defaultTesting
// (seeded from env) when unset.
func NewAppConfig(repo repository.AppSettingRepository, defaultTesting bool) AppConfig {
	c := &appConfig{repo: repo}
	v, _ := repo.Get(settingTestingMode)
	if v == "" {
		c.testing.Store(defaultTesting)
	} else {
		c.testing.Store(v == "true")
	}
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
