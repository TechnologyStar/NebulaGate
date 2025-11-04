package common

import (
	"errors"
	"html"
	"regexp"
	"strings"
)

// ValidateTicketTitle validates ticket title
func ValidateTicketTitle(title string) error {
	if title == "" {
		return errors.New("标题不能为空")
	}
	
	// Remove leading and trailing whitespace
	title = strings.TrimSpace(title)
	
	// Check length
	if len(title) < 1 || len(title) > 200 {
		return errors.New("标题长度必须在1-200字符之间")
	}
	
	return nil
}

// ValidateTicketContent validates ticket content
func ValidateTicketContent(content string) error {
	if content == "" {
		return errors.New("内容不能为空")
	}
	
	// Remove leading and trailing whitespace
	content = strings.TrimSpace(content)
	
	// Check length
	if len(content) < 1 || len(content) > 5000 {
		return errors.New("内容长度必须在1-5000字符之间")
	}
	
	return nil
}

// SanitizeInput sanitizes user input to prevent XSS
func SanitizeInput(input string) string {
	// HTML escape to prevent XSS
	input = html.EscapeString(input)
	
	// Remove potentially dangerous patterns
	// Remove script tags (case insensitive)
	scriptPattern := regexp.MustCompile(`(?i)<script[^>]*>.*?</script>`)
	input = scriptPattern.ReplaceAllString(input, "")
	
	// Remove event handlers (onclick, onerror, etc.)
	eventPattern := regexp.MustCompile(`(?i)on\w+\s*=`)
	input = eventPattern.ReplaceAllString(input, "")
	
	// Remove javascript: protocol
	jsPattern := regexp.MustCompile(`(?i)javascript:`)
	input = jsPattern.ReplaceAllString(input, "")
	
	return input
}

// ValidateTicketStatus validates ticket status
func ValidateTicketStatus(status string) bool {
	validStatuses := []string{"pending", "processing", "resolved", "closed"}
	for _, s := range validStatuses {
		if s == status {
			return true
		}
	}
	return false
}

// ValidateTicketPriority validates ticket priority
func ValidateTicketPriority(priority string) bool {
	validPriorities := []string{"low", "medium", "high", "urgent"}
	for _, p := range validPriorities {
		if p == priority {
			return true
		}
	}
	return false
}

// ValidateTicketCategory validates ticket category
func ValidateTicketCategory(category string) bool {
	validCategories := []string{"technical", "account", "feature", "other"}
	for _, c := range validCategories {
		if c == category {
			return true
		}
	}
	return false
}
