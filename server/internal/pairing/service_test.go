package pairing

import "testing"

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

func TestCannotApproveClosedRequestTwice(t *testing.T) {
	service := NewService()
	request, err := service.Create(CreateRequest{})
	if err != nil {
		t.Fatalf("create pairing: %v", err)
	}
	if _, err := service.Deny(request.ID, "admin"); err != nil {
		t.Fatalf("deny pairing: %v", err)
	}
	if _, err := service.Approve(request.ID, "admin"); err != ErrClosed {
		t.Fatalf("expected closed error, got %v", err)
	}
}
