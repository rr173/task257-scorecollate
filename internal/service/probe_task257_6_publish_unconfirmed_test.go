package service

import (
	"testing"

	"task257-scorecollate/internal/model"
)

func TestPublishEditionKeepsOnlyConfirmedVariants(t *testing.T) {
	svc := newTestService(t)
	p, err := svc.CreateProject("快照污染", "")
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
	twoDiff := "M1 b4 v2 : [C4 E4][G4 XX]\nM2 b3 v2 : [D4 F4][B4]\n"
	fa, err := svc.CreateFragment(p.ID, a.ID, "A", rawGood)
	if err != nil {
		t.Fatal(err)
	}
	fb, err := svc.CreateFragment(p.ID, b.ID, "B", twoDiff)
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
	if len(variants) < 2 {
		t.Fatalf("need at least two variants, got %d", len(variants))
	}
	if _, err := svc.AdjudicateVariant(variants[0].ID, variants[0].DetectedKind, "r", "e"); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.AdjudicateVariant(variants[0].ID, model.VarConfirmed, "c", "e"); err != nil {
		t.Fatal(err)
	}
	ed, err := svc.CreateEdition(p.ID, "v1")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.PublishEdition(ed.ID); err != nil {
		t.Fatal(err)
	}
	links, err := svc.store.ListEditionVariants(ed.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(links) != 1 || links[0].VariantID != variants[0].ID {
		t.Fatalf("frozen edition must keep only the confirmed variant, got %#v", links)
	}
}
