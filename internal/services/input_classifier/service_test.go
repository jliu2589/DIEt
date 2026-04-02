package input_classifier

import "testing"

func TestServiceClassify_MealLogWithSupplementAndPortions(t *testing.T) {
	svc := NewService()

	intent := svc.Classify("1 scoop ON Whey protein, 250ml 2% milk, 5mg of creatine")

	if intent != IntentMealLog {
		t.Fatalf("expected %q, got %q", IntentMealLog, intent)
	}
}

func TestServiceClassify_GeneralChatStillClassified(t *testing.T) {
	svc := NewService()

	intent := svc.Classify("hello there")

	if intent != IntentGeneralChat {
		t.Fatalf("expected %q, got %q", IntentGeneralChat, intent)
	}
}
