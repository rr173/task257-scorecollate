package service

import (
	"errors"
	"testing"

	"task257-scorecollate/internal/model"
)

func TestSealedProjectRejectsNewSources(t *testing.T) {
	svc := newTestService(t)
	p, err := svc.CreateProject("封存后再改", "")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.TransitionProject(p.ID, model.ProjectSealed); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.CreateSource(p.ID, "X", "迟到来源", "", ""); !errors.Is(err, model.ErrSealed) {
		t.Fatalf("sealed project must reject new source, got %v", err)
	}
}
