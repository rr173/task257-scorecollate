package service

import (
	"errors"
	"testing"

	"task257-scorecollate/internal/model"
)

func TestExcludedFragmentDoesNotJoinAlignment(t *testing.T) {
	svc := newTestService(t)
	p, err := svc.CreateProject("排除后对齐", "")
	if err != nil {
		t.Fatal(err)
	}
	a, err := svc.CreateSource(p.ID, "A", "A", "", "")
	if err != nil {
		t.Fatal(err)
	}
	b, err := svc.CreateSource(p.ID, "B", "B", "", "")
	if err != nil {
		t.Fatal(err)
	}
	fa, err := svc.CreateFragment(p.ID, a.ID, "A", rawGood)
	if err != nil {
		t.Fatal(err)
	}
	fb, err := svc.CreateFragment(p.ID, b.ID, "B", rawBreak)
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := svc.ParseFragment(fa.ID); err != nil {
		t.Fatal(err)
	}
	if _, _, err := svc.ParseFragment(fb.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.SetFragmentState(fb.ID, model.FragExcluded); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.AlignProject(p.ID); err == nil || !errors.Is(err, model.ErrInvalidInput) {
		t.Fatalf("excluded fragment must not satisfy alignment, err=%v", err)
	}
}
