package service

import (
	"fmt"
	"strings"

	"github.com/beevik/etree"
	"github.com/mylxsw/namecheap-domain-probe/backend/internal/model"
)

// ParseXML parses XML text and returns the root element
func ParseXML(xmlText string) (*etree.Element, error) {
	doc := etree.NewDocument()
	if err := doc.ReadFromString(xmlText); err != nil {
		return nil, fmt.Errorf("XML parse failed: %w", err)
	}
	return doc.Root(), nil
}

// GetErrors extracts error information from XML response
func GetErrors(root *etree.Element) []model.ErrorInfo {
	var errors []model.ErrorInfo
	errorsElem := root.FindElement(".//Errors")
	if errorsElem == nil {
		return errors
	}

	for _, err := range errorsElem.FindElements("Error") {
		number := err.SelectAttrValue("Number", err.SelectAttrValue("number", ""))
		text := strings.TrimSpace(err.Text())
		errors = append(errors, model.ErrorInfo{
			Number: number,
			Text:   text,
		})
	}
	return errors
}

// GetStatus returns the Status attribute from the root element
func GetStatus(root *etree.Element) string {
	return strings.ToUpper(root.SelectAttrValue("Status", ""))
}

// IsRateLimitError checks if the error message indicates rate limiting
func IsRateLimitError(msg string) bool {
	lower := strings.ToLower(msg)
	return strings.Contains(lower, "too many requests") || strings.Contains(lower, "rate limit")
}
