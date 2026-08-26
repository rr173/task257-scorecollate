package service

import (
	"testing"

	"task257-scorecollate/internal/model"
)

const lineageRawA = "M1 b4 v1 : [C4]\n"
const lineageRawB = "M1 b4 v1 : [D4]\n"

func variantForPair(t *testing.T, variants []*model.Variant, otherID string) *model.Variant {
	t.Helper()
	for _, v := range variants {
		if v.FragmentBID == otherID {
			return v
		}
	}
	t.Fatalf("variant for fragment %s not found", otherID)
	return nil
}

func TestLineageCopiesDoNotBecomeIndependentSupport(t *testing.T) {
	t.Run("same lineage is not independent support", func(t *testing.T) {
		svc := newTestService(t)
		p, err := svc.CreateProject("同源副本", "同一祖本的重复传抄不应增加独立支持")
		if err != nil {
			t.Fatal(err)
		}
		a, err := svc.CreateSource(p.ID, "A", "祖本 A", "", "")
		if err != nil {
			t.Fatal(err)
		}
		b, err := svc.CreateSource(p.ID, "B", "A 的传抄本", a.ID, "")
		if err != nil {
			t.Fatal(err)
		}
		c, err := svc.CreateSource(p.ID, "C", "A 的另一传抄本", a.ID, "")
		if err != nil {
			t.Fatal(err)
		}
		fa, err := svc.CreateFragment(p.ID, a.ID, "A", lineageRawA)
		if err != nil {
			t.Fatal(err)
		}
		fb, err := svc.CreateFragment(p.ID, b.ID, "B", lineageRawB)
		if err != nil {
			t.Fatal(err)
		}
		fc, err := svc.CreateFragment(p.ID, c.ID, "C", lineageRawB)
		if err != nil {
			t.Fatal(err)
		}
		for _, f := range []*model.Fragment{fa, fb, fc} {
			if _, _, err := svc.ParseFragment(f.ID); err != nil {
				t.Fatal(err)
			}
		}
		if _, err := svc.AlignProject(p.ID); err != nil {
			t.Fatal(err)
		}
		variants, err := svc.ListVariants(p.ID)
		if err != nil {
			t.Fatal(err)
		}
		v := variantForPair(t, variants, fb.ID)
		if v.SupportCount != 0 {
			t.Fatalf("same lineage copies must provide zero independent support, got %d", v.SupportCount)
		}
		if v.DetectedKind != model.VarInsufficient {
			t.Fatalf("same lineage copies must remain insufficient, got %s", v.DetectedKind)
		}
	})

	t.Run("different lineage remains independent support", func(t *testing.T) {
		svc := newTestService(t)
		p, err := svc.CreateProject("异祖本支持", "不同祖本的读法应保留独立支持")
		if err != nil {
			t.Fatal(err)
		}
		a, err := svc.CreateSource(p.ID, "A", "祖本 A", "", "")
		if err != nil {
			t.Fatal(err)
		}
		b, err := svc.CreateSource(p.ID, "B", "A 的传抄本", a.ID, "")
		if err != nil {
			t.Fatal(err)
		}
		d, err := svc.CreateSource(p.ID, "D", "祖本 D", "", "")
		if err != nil {
			t.Fatal(err)
		}
		fa, err := svc.CreateFragment(p.ID, a.ID, "A", lineageRawA)
		if err != nil {
			t.Fatal(err)
		}
		fb, err := svc.CreateFragment(p.ID, b.ID, "B", lineageRawB)
		if err != nil {
			t.Fatal(err)
		}
		fd, err := svc.CreateFragment(p.ID, d.ID, "D", lineageRawB)
		if err != nil {
			t.Fatal(err)
		}
		for _, f := range []*model.Fragment{fa, fb, fd} {
			if _, _, err := svc.ParseFragment(f.ID); err != nil {
				t.Fatal(err)
			}
		}
		if _, err := svc.AlignProject(p.ID); err != nil {
			t.Fatal(err)
		}
		variants, err := svc.ListVariants(p.ID)
		if err != nil {
			t.Fatal(err)
		}
		v := variantForPair(t, variants, fb.ID)
		if v.SupportCount != 1 {
			t.Fatalf("different lineage must provide one independent support, got %d", v.SupportCount)
		}
		if v.DetectedKind != model.VarValidVariant {
			t.Fatalf("different lineage support must classify as variant, got %s", v.DetectedKind)
		}
	})
}
