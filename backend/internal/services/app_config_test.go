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
	// Unset -> falls back to the env defaults.
	repo := &fakeSettingRepo{store: map[string]string{}}
	if NewAppConfig(repo, true, "m1").TestingMode() != true {
		t.Fatal("empty store should use default=true")
	}
	if NewAppConfig(repo, false, "m1").FireworksModel() != "m1" {
		t.Fatal("empty store should use default model")
	}

	// Stored values win over defaults.
	repo2 := &fakeSettingRepo{store: map[string]string{"testing_mode": "true", "fireworks_model": "m2"}}
	c2 := NewAppConfig(repo2, false, "m1")
	if !c2.TestingMode() || c2.FireworksModel() != "m2" {
		t.Fatal("stored values should override defaults")
	}

	// Setters update both the live cache and the store.
	c := NewAppConfig(&fakeSettingRepo{store: map[string]string{}}, false, "m1")
	if err := c.SetTestingMode(true); err != nil || !c.TestingMode() {
		t.Fatalf("SetTestingMode: %v / %v", err, c.TestingMode())
	}
	if err := c.SetFireworksModel("m3"); err != nil || c.FireworksModel() != "m3" {
		t.Fatalf("SetFireworksModel: %v / %v", err, c.FireworksModel())
	}
}
