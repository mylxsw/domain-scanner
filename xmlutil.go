package main

import (
	"fmt"
	"strings"

	"github.com/beevik/etree"
)

// ParseXML parses XML text and returns the root element
func ParseXML(xmlText string) (*etree.Element, error) {
	doc := etree.NewDocument()
	if err := doc.ReadFromString(xmlText); err != nil {
		return nil, fmt.Errorf("XML parse failed: %w", err)
	}
	return doc.Root(), nil
}

// ErrorInfo represents an API error
type ErrorInfo struct {
	Number string `json:"number"`
	Text   string `json:"text"`
}

// GetErrors extracts error information from XML response
func GetErrors(root *etree.Element) []ErrorInfo {
	ns := "http://api.namecheap.com/xml.response"

	var errors []ErrorInfo
	errorsElem := root.FindElement(".//{" + ns + "}Errors")
	if errorsElem == nil {
		return errors
	}

	for _, err := range errorsElem.ChildElements() {
		if err.Tag != "Error" {
			continue
		}
		number := err.SelectAttrValue("Number", err.SelectAttrValue("number", ""))
		text := strings.TrimSpace(err.Text())
		errors = append(errors, ErrorInfo{
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
