package service

import (
	"testing"

	"task257-scorecollate/internal/model"
)

func TestVoiceCountMismatchStaysUnreadable(t *testing.T) {
	svc := newTestService(t)
	p, err := svc.CreateProject("声部不符", "")
	if err != nil {
		t.Fatal(err)
	}
	src, err := svc.CreateSource(p.ID, "A", "A", "", "")
	if err != nil {
		t.Fatal(err)
	}
	f, err := svc.CreateFragment(p.ID, src.ID, "A", "M1 b4 v2 : [C4 E4]\n")
	if err != nil {
		t.Fatal(err)
	}
	got, _, err := svc.ParseFragment(f.ID)
	if err == nil {
		t.Fatal("voice count mismatch must fail parse")
	}
	if got == nil || got.State != model.FragUnreadable {
		t.Fatalf("mismatch must stay unreadable, got %#v", got)
	}
}
