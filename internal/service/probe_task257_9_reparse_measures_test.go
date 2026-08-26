package service

import (
	"testing"

	"task257-scorecollate/internal/model"
)

func TestReparseReplacesMeasures(t *testing.T) {
	svc := newTestService(t)
	p, err := svc.CreateProject("重解析", "")
	if err != nil {
		t.Fatal(err)
	}
	src, err := svc.CreateSource(p.ID, "A", "A", "", "")
	if err != nil {
		t.Fatal(err)
	}
	f, err := svc.CreateFragment(p.ID, src.ID, "A", rawGood)
	if err != nil {
		t.Fatal(err)
	}
	if _, ms, err := svc.ParseFragment(f.ID); err != nil {
		t.Fatal(err)
	} else if len(ms) != 2 {
		t.Fatalf("first parse want 2 measures, got %d", len(ms))
	}
	if _, err := svc.SetFragmentState(f.ID, model.FragUnreadable); err != nil {
		t.Fatal(err)
	}
	if _, _, err := svc.ParseFragment(f.ID); err != nil {
		t.Fatal(err)
	}
	got, err := svc.ListMeasures(f.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("reparse must replace measures, got %d", len(got))
	}
}
