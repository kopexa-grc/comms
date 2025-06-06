// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: BUSL-1.1

package comms

import (
	"bytes"
	"errors"
	"fmt"
)

var ErrBothTemplatesFailed = errors.New("both text and html templates failed")

func Render(name string, data any) (text, html string, err error) {
	text, textErr := render(name+".txt", data)
	html, htmlErr := render(name+".html", data)

	if textErr != nil && htmlErr != nil {
		return "", "", fmt.Errorf("%w: text error: %v, html error: %v", ErrBothTemplatesFailed, textErr, htmlErr)
	}

	return text, html, nil
}

func render(name string, data any) (string, error) {
	t, ok := templates[name]
	if !ok {
		return "", fmt.Errorf("%w: %q not found", ErrTemplateNotFound, name)
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}
