package service

import (
	"testing"

	"task257-scorecollate/internal/model"
)

func TestCandidateCannotSkipStraightToConfirmed(t *testing.T) {
	svc := newTestService(t)
	p, err := svc.CreateProject("跳步确认", "")
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
	if _, err := svc.AlignProject(p.ID); err != nil {
		t.Fatal(err)
	}
	variants, err := svc.ListVariants(p.ID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.AdjudicateVariant(variants[0].ID, model.VarConfirmed, "skip", "e"); err == nil {
		t.Fatal("candidate must not jump straight to confirmed")
	}
	if _, err := svc.AdjudicateVariant(variants[0].ID, model.VarError, "r", "e"); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.AdjudicateVariant(variants[0].ID, model.VarConfirmed, "ok", "e"); err != nil {
		t.Fatal(err)
	}
}
