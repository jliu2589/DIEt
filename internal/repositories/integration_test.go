package repositories

import (
	"context"
	"testing"
	"time"

	"diet/internal/models"
	"diet/internal/testutil"
)

func TestMealEventsRepository_UpdateEatenAtAndListRecent(t *testing.T) {
	pool := testutil.OpenTestDB(t)
	t.Cleanup(pool.Close)
	repos := New(pool)

	raw := "toast"
	event, err := repos.MealEvents.Insert(context.Background(), models.MealEvent{
		UserID:           "u1",
		Source:           "web",
		EventType:        "text",
		RawText:          &raw,
		LoggedAt:         time.Now().UTC(),
		EatenAt:          time.Now().UTC(),
		TimeSource:       "explicit",
		ProcessingStatus: "pending",
	})
	if err != nil {
		t.Fatalf("insert meal event: %v", err)
	}
	if err := repos.MealAnalysis.Insert(context.Background(), models.MealAnalysis{MealEventID: event.ID, UserID: "u1", CanonicalName: "toast"}); err != nil {
		t.Fatalf("insert meal analysis: %v", err)
	}

	updatedAt := time.Now().UTC().Add(-1 * time.Hour)
	updated, err := repos.MealEvents.UpdateEatenAtByIDAndUserID(context.Background(), event.ID, "u1", updatedAt)
	if err != nil {
		t.Fatalf("update eaten_at: %v", err)
	}
	if updated == nil || updated.TimeSource != "edited" {
		t.Fatalf("expected edited update, got %+v", updated)
	}

	recent, err := repos.MealEvents.ListRecentByUserID(context.Background(), "u1", 10)
	if err != nil {
		t.Fatalf("list recent: %v", err)
	}
	if len(recent) != 1 {
		t.Fatalf("expected 1 recent meal, got %d", len(recent))
	}
}

func TestIdempotencyRepository_BeginOrGetAndMarkSucceeded(t *testing.T) {
	pool := testutil.OpenTestDB(t)
	t.Cleanup(pool.Close)
	repo := NewIdempotencyKeysRepository(pool)

	rec, created, err := repo.BeginOrGet(context.Background(), "u1", "POST:/v1/meals", "abc", "hash-1")
	if err != nil {
		t.Fatalf("begin idempotency: %v", err)
	}
	if !created || rec == nil {
		t.Fatalf("expected created record")
	}

	again, created, err := repo.BeginOrGet(context.Background(), "u1", "POST:/v1/meals", "abc", "hash-1")
	if err != nil {
		t.Fatalf("get existing idempotency: %v", err)
	}
	if created {
		t.Fatalf("expected existing record")
	}
	if again.ID != rec.ID {
		t.Fatalf("expected same record id, got %d and %d", rec.ID, again.ID)
	}

	if err := repo.MarkSucceeded(context.Background(), rec.ID, 200, []byte(`{"ok":true}`)); err != nil {
		t.Fatalf("mark succeeded: %v", err)
	}
}
