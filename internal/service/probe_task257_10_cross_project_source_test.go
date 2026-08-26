package service

import (
	"errors"
	"testing"

	"task257-scorecollate/internal/model"
)

func TestFragmentSourceMustBelongToProject(t *testing.T) {
	svc := newTestService(t)
	p1, err := svc.CreateProject("甲", "")
	if err != nil {
		t.Fatal(err)
	}
	p2, err := svc.CreateProject("乙", "")
	if err != nil {
		t.Fatal(err)
	}
	foreign, err := svc.CreateSource(p2.ID, "Z", "乙的来源", "", "")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.CreateFragment(p1.ID, foreign.ID, "错挂", rawGood); !errors.Is(err, model.ErrInvalidInput) {
		t.Fatalf("foreign source must be rejected, got %v", err)
	}
	own, err := svc.CreateSource(p1.ID, "A", "甲的来源", "", "")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.CreateFragment(p1.ID, own.ID, "正挂", rawGood); err != nil {
		t.Fatal(err)
	}
}
