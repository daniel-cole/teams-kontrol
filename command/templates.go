package command

import (
	"github.com/google/uuid"
	"reflect"
	"text/template"
)

var templateFns = template.FuncMap{
	"last": func(x int, a interface{}) bool {
		return x == reflect.ValueOf(a).Len()-1
	},
	"uuid": func() string {
		return uuid.New().String()
	},
}

const teamsAdaptiveCardPodListTmpl = `{
  "$schema": "http://adaptivecards.io/schemas/adaptive-card.json",
  "type": "AdaptiveCard",
  "version": "1.0",
  "body": [
    {
      "type": "TextBlock",
      "id": "{{ uuid }}",
      "text": "Pod Detail",
      "wrap": true,
      "size": "Large",
      "weight": "Bolder",
      "color": "Accent",
      "horizontalAlignment": "Center"
    },
    {{ range $i, $pod := . }}
    {
      "type": "Container",
      "id": " {{ uuid }}",
      "padding": "None",
      "items": [
        {
          "type": "FactSet",
          "id": "{{ uuid }}",
          "facts": [
            {
              "title": "Name",
              "value": "{{ $pod.Name }}"
            },
            {
              "title": "Age",
              "value": "{{ $pod.CreationTimestamp }}"
            },
            {
              "title": "Status",
              "value": "{{ $pod.Status.Phase }}"
            },
            {
              "title": "Namespace",
              "value": "{{ $pod.Namespace }}"
            }
          ]
        }
      ],
      "style": "emphasis"
    }{{ if not (last $i $) }},{{ end }}{{ end }}
  ],
  "padding": "None"
}
`
