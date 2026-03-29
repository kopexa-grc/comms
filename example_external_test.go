// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: BUSL-1.1

package comms_test

import (
	"context"
	"fmt"

	"github.com/kopexa-grc/comms/v2"
	"github.com/kopexa-grc/comms/v2/driver/mock"
)

func ExampleRenderTemplate() {
	tmpl := comms.ExternalTemplate{
		SubjectTemplate: "Welcome to {{.OrgName}}",
		BodyTemplate:    "<h1>Hello {{.Name}}</h1><p>Welcome aboard!</p>",
		TextTemplate:    "Hello {{.Name}}, welcome aboard!",
	}

	rendered, err := comms.RenderTemplate(tmpl, nil, map[string]any{
		"Name":    "Max",
		"OrgName": "Acme Corp",
	})
	if err != nil {
		panic(err)
	}

	fmt.Println(rendered.Subject)
	fmt.Println(rendered.Text)
	// Output:
	// Welcome to Acme Corp
	// Hello Max, welcome aboard!
}

func ExampleRenderTemplate_withBranding() {
	tmpl := comms.ExternalTemplate{
		SubjectTemplate: "Hello",
		BodyTemplate:    `<a style="background-color: {{.Branding.ButtonColor}};">Click</a>`,
	}

	branding := &comms.Branding{
		BrandName:    "Acme Corp",
		PrimaryColor: "#ff6600",
		ButtonColor:  "#cc0000",
	}

	rendered, err := comms.RenderTemplate(tmpl, branding, nil)
	if err != nil {
		panic(err)
	}

	fmt.Println(rendered.Subject)
	// Output:
	// Hello
}

func ExampleComms_SendFromTemplate() {
	d := mock.NewDriver()
	c := comms.New(
		comms.WithDriver(d),
		comms.WithFrom("noreply@example.com"),
	)

	recipient := comms.NewRecipient("user@example.com", "Max", "Mustermann")

	tmpl := comms.ExternalTemplate{
		SubjectTemplate: "Welcome to {{.OrgName}}",
		BodyTemplate:    "<h1>Hello {{.Name}}</h1>",
		TextTemplate:    "Hello {{.Name}}",
	}

	err := c.SendFromTemplate(context.Background(), recipient, tmpl, nil, map[string]any{
		"Name":    "Max",
		"OrgName": "Acme Corp",
	})
	if err != nil {
		panic(err)
	}

	msg, _ := d.LastMessage()
	fmt.Println(msg.Subject)
	// Output:
	// Welcome to Acme Corp
}
