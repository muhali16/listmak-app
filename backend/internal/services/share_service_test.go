package services

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/muhali16/listmak-service/internal/models"
	"gorm.io/gorm"
)

// --- fakes ---------------------------------------------------------------

// fakeShareLinkRepo stores share-links in memory.
type fakeShareLinkRepo struct {
	active map[uint]*models.ShareLink // keyed by listmak_id
	store  map[string]models.ShareLink // keyed by share_id
}

func (f *fakeShareLinkRepo) CreateShareLink(sl models.ShareLink) (models.ShareLink, error) {
	if f.store == nil {
		f.store = map[string]models.ShareLink{}
	}
	f.store[sl.ShareID] = sl
	return sl, nil
}

func (f *fakeShareLinkRepo) GetShareLinkByShareId(shareId string) (models.ShareLink, error) {
	sl, ok := f.store[shareId]
	if !ok {
		return models.ShareLink{}, gorm.ErrRecordNotFound
	}
	return sl, nil
}

func (f *fakeShareLinkRepo) DeleteShareLink(id uint) error { return nil }

func (f *fakeShareLinkRepo) GetActiveShareLinkByListmakId(listmakId uint) (*models.ShareLink, error) {
	if f.active == nil {
		return nil, nil
	}
	sl, ok := f.active[listmakId]
	if !ok {
		return nil, nil
	}
	return sl, nil
}

// fakeViewShareRepo stores view-shares in memory keyed by view_id.
type fakeViewShareRepo struct {
	store map[string]models.ViewShare
}

func (f *fakeViewShareRepo) CreateViewShare(vs models.ViewShare) (models.ViewShare, error) {
	if f.store == nil {
		f.store = map[string]models.ViewShare{}
	}
	f.store[vs.ViewID] = vs
	return vs, nil
}

func (f *fakeViewShareRepo) GetViewShareByViewId(viewId string) (models.ViewShare, error) {
	vs, ok := f.store[viewId]
	if !ok {
		return models.ViewShare{}, gorm.ErrRecordNotFound
	}
	return vs, nil
}

func (f *fakeViewShareRepo) GetViewShareByListmakId(listmakId uint) (*models.ViewShare, error) {
	for _, vs := range f.store {
		if vs.ListmakID == listmakId {
			copy := vs
			return &copy, nil
		}
	}
	return nil, nil
}

// fakeListmakRepo lets a test control what GetListmakById returns. Only
// GetListmakById is exercised by the view-share code path; the rest satisfy
// the interface.
type fakeListmakRepo struct {
	listmak    models.Listmak
	getByIDErr error
	getCalls   int
}

func (f *fakeListmakRepo) GetListmakById(id uint) (models.Listmak, error) {
	f.getCalls++
	if f.getByIDErr != nil {
		return models.Listmak{}, f.getByIDErr
	}
	return f.listmak, nil
}

func (f *fakeListmakRepo) GetAllListmaks(page, limit int, status string, startDate, endDate *time.Time) ([]models.Listmak, int64, error) {
	return nil, 0, nil
}
func (f *fakeListmakRepo) GetListmakByDate(date time.Time) ([]models.Listmak, error) {
	return nil, nil
}
func (f *fakeListmakRepo) CreateListmak(l models.Listmak) (models.Listmak, error) { return l, nil }
func (f *fakeListmakRepo) UpdateListmak(l models.Listmak) (models.Listmak, error) { return l, nil }
func (f *fakeListmakRepo) DeleteListmak(id uint) error                            { return nil }

func newServiceWithFakes(sl *fakeShareLinkRepo, vs *fakeViewShareRepo, lr *fakeListmakRepo) ShareService {
	return NewShareService(sl, vs, lr)
}

// --- tests ---------------------------------------------------------------

// Legacy links (is_live == false) must keep returning the stored snapshot
// untouched, and must NOT re-query the listmak.
func TestGetViewShare_SnapshotMode_ReturnsFrozenSnapshot(t *testing.T) {
	frozen := json.RawMessage(`{"id":1,"title":"Old Title","paid_amount":0}`)
	vsRepo := &fakeViewShareRepo{store: map[string]models.ViewShare{
		"abc": {ViewID: "abc", ListmakID: 1, IsLive: false, SnapshotData: frozen},
	}}
	// Live listmak has DIFFERENT data; snapshot mode must ignore it entirely.
	lmRepo := &fakeListmakRepo{listmak: models.Listmak{ID: 1, Title: "New Title", PaidAmount: 99000}}

	svc := newServiceWithFakes(nil, vsRepo, lmRepo)

	got, err := svc.GetViewShare("abc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(got.SnapshotData) != string(frozen) {
		t.Fatalf("snapshot was mutated.\n got: %s\nwant: %s", got.SnapshotData, frozen)
	}
	if lmRepo.getCalls != 0 {
		t.Fatalf("listmak should not be queried in snapshot mode, got %d calls", lmRepo.getCalls)
	}
}

// Live links (is_live == true) must overwrite SnapshotData in-memory with a
// fresh marshal of the current listmak, keeping the same response shape.
func TestGetViewShare_LiveMode_ReturnsFreshData(t *testing.T) {
	staleSnapshot := json.RawMessage(`{"id":1,"title":"Stale","paid_amount":0}`)
	vsRepo := &fakeViewShareRepo{store: map[string]models.ViewShare{
		"live1": {ViewID: "live1", ListmakID: 1, IsLive: true, SnapshotData: staleSnapshot},
	}}
	liveListmak := models.Listmak{
		ID:          1,
		Title:       "Current",
		TotalOrders: 3,
		PaidAmount:  50000,
		Orders: []models.Order{
			{ID: 10, Name: "Budi", IsPaid: true},
			{ID: 11, Name: "Andi", IsPaid: false},
		},
	}
	lmRepo := &fakeListmakRepo{listmak: liveListmak}

	svc := newServiceWithFakes(nil, vsRepo, lmRepo)

	got, err := svc.GetViewShare("live1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if lmRepo.getCalls != 1 {
		t.Fatalf("live mode should query listmak exactly once, got %d", lmRepo.getCalls)
	}
	if string(got.SnapshotData) == string(staleSnapshot) {
		t.Fatal("SnapshotData should have been replaced with live data")
	}

	// Response shape must equal json.Marshal(listmak) — same as the create path.
	wantBytes, _ := json.Marshal(liveListmak)
	if string(got.SnapshotData) != string(wantBytes) {
		t.Fatalf("live snapshot shape mismatch.\n got: %s\nwant: %s", got.SnapshotData, wantBytes)
	}

	// Sanity: the fresh data is actually reflected.
	var decoded map[string]interface{}
	if err := json.Unmarshal(got.SnapshotData, &decoded); err != nil {
		t.Fatalf("decode live snapshot: %v", err)
	}
	if decoded["title"] != "Current" {
		t.Fatalf("expected live title 'Current', got %v", decoded["title"])
	}
}

// A live link whose listmak was soft-deleted must yield ErrListmakUnavailable
// (mapped to 404 upstream), never a leaked GORM error.
func TestGetViewShare_LiveMode_ListmakDeleted(t *testing.T) {
	vsRepo := &fakeViewShareRepo{store: map[string]models.ViewShare{
		"live2": {ViewID: "live2", ListmakID: 7, IsLive: true, SnapshotData: json.RawMessage(`{"id":7}`)},
	}}
	// Simulate soft-deleted listmak: GetListmakById returns record-not-found.
	lmRepo := &fakeListmakRepo{getByIDErr: gorm.ErrRecordNotFound}

	svc := newServiceWithFakes(nil, vsRepo, lmRepo)

	_, err := svc.GetViewShare("live2")
	if !errors.Is(err, ErrListmakUnavailable) {
		t.Fatalf("expected ErrListmakUnavailable, got %v", err)
	}
}

// A missing view_id propagates the not-found error (controller maps to a
// generic 404), distinct from the deleted-listmak case.
func TestGetViewShare_NotFound(t *testing.T) {
	vsRepo := &fakeViewShareRepo{store: map[string]models.ViewShare{}}
	lmRepo := &fakeListmakRepo{}

	svc := newServiceWithFakes(nil, vsRepo, lmRepo)

	_, err := svc.GetViewShare("missing")
	if err == nil {
		t.Fatal("expected an error for missing view id")
	}
	if errors.Is(err, ErrListmakUnavailable) {
		t.Fatal("missing view id should not be reported as ErrListmakUnavailable")
	}
}

// New links created after this change must be flagged is_live = true so the
// read path serves live data.
func TestCreateViewShare_SetsIsLiveTrue(t *testing.T) {
	vsRepo := &fakeViewShareRepo{}
	lmRepo := &fakeListmakRepo{listmak: models.Listmak{ID: 1, Title: "X"}}

	svc := newServiceWithFakes(nil, vsRepo, lmRepo)

	created, err := svc.CreateViewShare(1, "X", 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !created.IsLive {
		t.Fatal("newly created view share must have IsLive = true")
	}
}

func TestGetActiveSharesForListmak_BothExist(t *testing.T) {
	expiry := time.Now().Add(24 * time.Hour)
	slRepo := &fakeShareLinkRepo{
		active: map[uint]*models.ShareLink{
			1: {ShareID: "abc", ListmakID: 1, ExpiresAt: expiry, IsActive: true},
		},
	}
	vsRepo := &fakeViewShareRepo{store: map[string]models.ViewShare{
		"xyz": {ViewID: "xyz", ListmakID: 1},
	}}
	lmRepo := &fakeListmakRepo{}

	svc := newServiceWithFakes(slRepo, vsRepo, lmRepo)

	sl, vs, err := svc.GetActiveSharesForListmak(1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sl == nil || sl.ShareID != "abc" {
		t.Fatalf("expected share link abc, got %v", sl)
	}
	if vs == nil || vs.ViewID != "xyz" {
		t.Fatalf("expected view share xyz, got %v", vs)
	}
}

func TestGetActiveSharesForListmak_NoneExist(t *testing.T) {
	slRepo := &fakeShareLinkRepo{}
	vsRepo := &fakeViewShareRepo{store: map[string]models.ViewShare{}}
	lmRepo := &fakeListmakRepo{}

	svc := newServiceWithFakes(slRepo, vsRepo, lmRepo)

	sl, vs, err := svc.GetActiveSharesForListmak(99)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sl != nil {
		t.Fatalf("expected nil share link, got %v", sl)
	}
	if vs != nil {
		t.Fatalf("expected nil view share, got %v", vs)
	}
}

func TestGetActiveSharesForListmak_OnlyViewShare(t *testing.T) {
	slRepo := &fakeShareLinkRepo{}
	vsRepo := &fakeViewShareRepo{store: map[string]models.ViewShare{
		"onlyview": {ViewID: "onlyview", ListmakID: 5},
	}}
	lmRepo := &fakeListmakRepo{}

	svc := newServiceWithFakes(slRepo, vsRepo, lmRepo)

	sl, vs, err := svc.GetActiveSharesForListmak(5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sl != nil {
		t.Fatalf("expected nil share link, got %v", sl)
	}
	if vs == nil || vs.ViewID != "onlyview" {
		t.Fatalf("expected view share onlyview, got %v", vs)
	}
}
