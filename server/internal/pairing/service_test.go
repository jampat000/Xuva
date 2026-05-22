package pairing

import (
	"context"
	"testing"

	"github.com/jampat000/Xuva/server/internal/database"
)

func TestCreateApproveAndHideCodeAfterApproval(t *testing.T) {
	service := NewService()
	request, err := service.Create(CreateRequest{DeviceName: "Living Room", ClientProfile: "apple-tv"})
	if err != nil {
		t.Fatalf("create pairing: %v", err)
	}
	if request.ID == "" || len(request.Code) != 6 || request.Status != StatusPending {
		t.Fatalf("expected pending request with six digit code, got %#v", request)
	}

	approved, err := service.Approve(request.ID, "admin")
	if err != nil {
		t.Fatalf("approve pairing: %v", err)
	}
	if approved.Status != StatusApproved || approved.DeviceID == "" || approved.Code != "" {
		t.Fatalf("expected approved request with device id and hidden code, got %#v", approved)
	}
}

func TestCannotApproveAlreadyApprovedRequest(t *testing.T) {
	service := NewService()
	request, err := service.Create(CreateRequest{})
	if err != nil {
		t.Fatalf("create pairing: %v", err)
	}
	if _, err := service.Approve(request.ID, "admin"); err != nil {
		t.Fatalf("first approve: %v", err)
	}
	if _, err := service.Approve(request.ID, "admin"); err != ErrClosed {
		t.Fatalf("expected ErrClosed on double-approve, got %v", err)
	}
}

func TestDenyDeletesRequestImmediately(t *testing.T) {
	service := NewService()
	request, err := service.Create(CreateRequest{})
	if err != nil {
		t.Fatalf("create pairing: %v", err)
	}
	if _, err := service.Deny(request.ID, "admin"); err != nil {
		t.Fatalf("deny pairing: %v", err)
	}
	if _, ok := service.Get(request.ID); ok {
		t.Fatal("expected denied request to be gone, but Get still returned it")
	}
	if _, err := service.Approve(request.ID, "admin"); err != ErrNotFound {
		t.Fatalf("expected ErrNotFound after deny, got %v", err)
	}
}

func TestPersistentServiceKeepsPendingRequestAcrossServiceRestart(t *testing.T) {
	db, err := database.Open(context.Background(), t.TempDir())
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Fatalf("close database: %v", err)
		}
	}()

	created, err := NewPersistentService(db).Create(CreateRequest{
		DeviceName:    "Apple TV",
		ClientProfile: "apple-tv",
		DeviceID:      "device-1",
	})
	if err != nil {
		t.Fatalf("create request: %v", err)
	}

	got, ok := NewPersistentService(db).Get(created.ID)
	if !ok {
		t.Fatalf("expected request %s after service restart", created.ID)
	}
	if got.ID != created.ID || got.Status != StatusPending || got.Code == "" {
		t.Fatalf("unexpected restored request: %#v", got)
	}
}
