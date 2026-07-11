package services

import "testing"

type fakeSettingRepo struct{ store map[string]string }

func (r *fakeSettingRepo) Get(key string) (string, error) { return r.store[key], nil }
func (r *fakeSettingRepo) Set(key, value string) error {
	if r.store == nil {
		r.store = map[string]string{}
	}
	r.store[key] = value
	return nil
}

func TestAppConfig(t *testing.T) {
	// Unset -> falls back to the env default.
	repo := &fakeSettingRepo{store: map[string]string{}}
	if NewAppConfig(repo, true).TestingMode() != true {
		t.Fatal("empty store should use default=true")
	}
	if NewAppConfig(repo, false).TestingMode() != false {
		t.Fatal("empty store should use default=false")
	}

	// Stored value wins over the default.
	repo2 := &fakeSettingRepo{store: map[string]string{"testing_mode": "true"}}
	if NewAppConfig(repo2, false).TestingMode() != true {
		t.Fatal("stored true should override default=false")
	}

	// SetTestingMode updates both the live cache and the store.
	c := NewAppConfig(&fakeSettingRepo{store: map[string]string{}}, false)
	if err := c.SetTestingMode(true); err != nil {
		t.Fatalf("set: %v", err)
	}
	if !c.TestingMode() {
		t.Fatal("cache not updated after SetTestingMode(true)")
	}
	if err := c.SetTestingMode(false); err != nil {
		t.Fatalf("set: %v", err)
	}
	if c.TestingMode() {
		t.Fatal("cache not updated after SetTestingMode(false)")
	}
}
