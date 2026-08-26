package service

import "testing"

func TestRealignDoesNotDuplicateVariants(t *testing.T) {
	svc := newTestService(t)
	p, err := svc.CreateProject("重复对齐", "")
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
	first, err := svc.ListVariants(p.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(first) == 0 {
		t.Fatal("first align must create variants")
	}
	if _, err := svc.AlignProject(p.ID); err != nil {
		t.Fatal(err)
	}
	second, err := svc.ListVariants(p.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(second) != len(first) {
		t.Fatalf("realign must not duplicate variants: first=%d second=%d", len(first), len(second))
	}
}
