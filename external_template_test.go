// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: BUSL-1.1

package comms

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExternalTemplate_Validate(t *testing.T) {
	tests := []struct {
		name    string
		tmpl    ExternalTemplate
		wantErr bool
	}{
		{
			name:    "both empty returns error",
			tmpl:    ExternalTemplate{},
			wantErr: true,
		},
		{
			name: "body only is valid",
			tmpl: ExternalTemplate{
				BodyTemplate: "<h1>Hello {{.Name}}</h1>",
			},
			wantErr: false,
		},
		{
			name: "text only is valid",
			tmpl: ExternalTemplate{
				TextTemplate: "Hello {{.Name}}",
			},
			wantErr: false,
		},
		{
			name: "both set is valid",
			tmpl: ExternalTemplate{
				BodyTemplate: "<h1>Hello {{.Name}}</h1>",
				TextTemplate: "Hello {{.Name}}",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.tmpl.Validate()
			if tt.wantErr {
				require.Error(t, err)
				require.ErrorIs(t, err, ErrTemplateContentRequired)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
