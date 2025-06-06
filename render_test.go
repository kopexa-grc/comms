// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: BUSL-1.1

package comms

import (
	"strings"
	"testing"
)

func TestRender(t *testing.T) {
	// Test data
	data := struct {
		Name string
	}{
		Name: "World",
	}

	// Test both text and HTML rendering
	text, html, err := Render("hello", data)
	if err != nil {
		t.Errorf("Render failed: %v", err)
	}

	// Verify text output
	if text == "" {
		t.Error("Expected non-empty text output")
	}

	if !strings.Contains(text, "Hello World!") {
		t.Errorf("Expected text to contain 'Hello World!', got: %s", text)
	}

	// Verify HTML output
	if html == "" {
		t.Error("Expected non-empty HTML output")
	}

	if !strings.Contains(html, "<h1>Hello World!</h1>") {
		t.Errorf("Expected HTML to contain '<h1>Hello World!</h1>', got: %s", html)
	}
}

func TestRenderTemplateNotFound(t *testing.T) {
	// Test with non-existent template
	_, _, err := Render("nonexistent", nil)
	if err == nil {
		t.Error("Expected error for non-existent template")
	}

	if !strings.Contains(err.Error(), "template not found") {
		t.Errorf("Expected error to contain 'template not found', got: %v", err)
	}
}

func TestRenderSingleTemplate(t *testing.T) {
	// Test data
	data := struct {
		Name string
	}{
		Name: "World",
	}

	// Create a temporary template with only text
	templates["single.txt"] = parseTemplate("single.txt")
	defer delete(templates, "single.txt")

	// Test rendering with only text template
	text, html, err := Render("single", data)
	if err != nil {
		t.Errorf("Render failed: %v", err)
	}

	// Verify text output exists
	if text == "" {
		t.Error("Expected non-empty text output")
	}

	// Verify HTML output is empty
	if html != "" {
		t.Error("Expected empty HTML output")
	}
}
