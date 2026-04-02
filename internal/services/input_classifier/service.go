package input_classifier

import (
	"regexp"
	"strings"
)

const (
	IntentMealLog               = "meal_log"
	IntentWeightLog             = "weight_log"
	IntentRecommendationRequest = "recommendation_request"
	IntentGeneralChat           = "general_chat"
	IntentUnknown               = "unknown"
)

type Service struct{}

func NewService() *Service {
	return &Service{}
}

var (
	weightLogPattern  = regexp.MustCompile(`(?i)\b\d{2,3}(?:\.\d+)?\s*(kg|kgs|kilograms?|lb|lbs|pounds?)\b`)
	weightHintPattern = regexp.MustCompile(`(?i)\b(weight|weigh|scale)\b`)
	mealPortionPattern = regexp.MustCompile(`(?i)\b\d+(?:\.\d+)?\s*(g|gram|grams|kg|ml|l|oz|cup|cups|tbsp|tablespoons?|tsp|teaspoons?|scoop|scoops|serving|servings|slice|slices|piece|pieces)\b`)
)

func (s *Service) Classify(rawText string) string {
	text := strings.ToLower(strings.TrimSpace(rawText))
	if text == "" {
		return IntentUnknown
	}

	if isGeneralChat(text) {
		return IntentGeneralChat
	}
	if isWeightLog(text) {
		return IntentWeightLog
	}
	if isRecommendationRequest(text) {
		return IntentRecommendationRequest
	}
	if isMealLog(text) {
		return IntentMealLog
	}

	return IntentUnknown
}

func isWeightLog(text string) bool {
	if weightLogPattern.MatchString(text) {
		return true
	}
	if weightHintPattern.MatchString(text) && strings.Contains(text, "kg") {
		return true
	}
	if weightHintPattern.MatchString(text) && strings.Contains(text, "lb") {
		return true
	}
	return false
}

func isRecommendationRequest(text string) bool {
	keywords := []string{
		"what should i eat", "what can i eat", "recommend", "suggest", "meal ideas", "meal idea", "tonight", "for dinner",
	}
	for _, kw := range keywords {
		if strings.Contains(text, kw) {
			return true
		}
	}
	return false
}

func isGeneralChat(text string) bool {
	phrases := []string{"hi", "hello", "hey", "how are you", "good morning", "good evening", "thanks", "thank you"}
	for _, p := range phrases {
		if text == p || strings.HasPrefix(text, p+" ") {
			return true
		}
	}
	return false
}

func isMealLog(text string) bool {
	foodHints := []string{
		"egg", "toast", "chicken", "rice", "salad", "sandwich", "oatmeal", "yogurt", "apple", "banana", "beef", "fish", "avocado", "pasta", "soup",
		"milk", "protein", "whey", "creatine", "ate ", "had ", "breakfast", "lunch", "dinner", "snack",
	}
	for _, hint := range foodHints {
		if strings.Contains(text, hint) {
			return true
		}
	}

	if mealPortionPattern.MatchString(text) {
		return true
	}

	return false
}
