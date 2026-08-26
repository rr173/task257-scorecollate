package service

import (
	"testing"

	"task257-scorecollate/internal/model"
)

func TestOmittedMeasureIsScribalError(t *testing.T) {
	svc := newTestService(t)
	p, err := svc.CreateProject("缺抄小节", "")
	if err != nil {
		t.Fatal(err)
	}
	a, err := svc.CreateSource(p.ID, "A", "祖本", "", "")
	if err != nil {
		t.Fatal(err)
	}
	b, err := svc.CreateSource(p.ID, "B", "传抄", a.ID, "")
	if err != nil {
		t.Fatal(err)
	}
	full := "M1 b4 v1 : [C4]\nM2 b4 v1 : [D4]\n"
	short := "M1 b4 v1 : [C4]\n"
	fa, err := svc.CreateFragment(p.ID, a.ID, "A", full)
	if err != nil {
		t.Fatal(err)
	}
	fb, err := svc.CreateFragment(p.ID, b.ID, "B", short)
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
	found := false
	for _, v := range variants {
		if v.MeasureNumber == 2 {
			found = true
			if v.DetectedKind != model.VarError {
				t.Fatalf("missing measure must be scribal error, got %s", v.DetectedKind)
			}
		}
	}
	if !found {
		t.Fatal("omitted terminal measure must produce a variant")
	}
}
