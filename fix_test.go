package main

import (
	"testing"

	"github.com/grafana/mimir/pkg/alertmanager/alertspb"
	"github.com/stretchr/testify/require"
)

func TestE2E(t *testing.T) {
	tests := []struct {
		name        string
		input       alertspb.AlertConfigDesc
		expected    alertspb.AlertConfigDesc
		expectedErr string
	}{{
		// This test shows how even a simple configuration can create large diffs
		// due to re-ordering of fields when marshalling the YAML.
		name: "simple case",
		input: alertspb.AlertConfigDesc{
			User: "1",
			RawConfig: `receivers:
  - name: default
route:
  receiver: default
  routes:
    - matchers:
      - foo=
      - bar=baz
    - routes:
      - matchers:
        - baz=qux
        - qux="corge"`,
		},
		expected: alertspb.AlertConfigDesc{
			User: "1",
			RawConfig: `route:
  receiver: default
  continue: false
  routes:
  - matchers:
    - bar="baz"
    - foo=""
    continue: false
  - continue: false
    routes:
    - matchers:
      - baz="qux"
      - qux="corge"
      continue: false
receivers:
- name: default
templates: []
`,
		},
	}, {
		name: "advanced case with secrets",
		input: alertspb.AlertConfigDesc{
			User: "2",
			RawConfig: `global:
  smtp_from: test@example.com
  smtp_smarthost: smtp.example.org:587
  smtp_auth_username: admin
  smtp_auth_password: password
  slack_api_url: https://example.com/1/
  victorops_api_key: foo
  pagerduty_url: https://example.com/2/
  opsgenie_api_key: bar
  opsgenie_api_url: https://example.com/3/
  wechat_api_url: https://example.com/4/
  wechat_api_secret: baz
  wechat_api_corp_id: qux
  telegram_api_url: https://example.com/5/
  webex_api_url: https://example.com/6/
  http_config:
    basic_auth:
      username: admin
      password: password
route:
  receiver: default
  group_by:
    - foo
  group_wait: 1m
  group_interval: 5m
  repeat_interval: 4h
  routes:
    - matchers:
      - foo=
      - bar=!baz
      active_time_intervals:
       - weekdays
      routes:
      - matchers:
        - baz=[\w+]
        - qux="[corge]"
        continue: true
        mute_time_intervals:
         - weekends
inhibit_rules:
  - target_matchers:
    - foo=
    source_matchers:
    - bar=!baz
  - target_matchers:
    - baz=[\w+]
    source_matchers:
    - qux="[corge]"
receivers:
  - name: default
    email_configs:
      - to: test@example.com
  - name: webhook
    webhook_configs:
      - url: https://example.com/
        http_config:
          authorization:
            type: Bearer
            credentials: token
templates:
  - tmpl1.tmpl
  - tmpl2.tmpl
time_intervals:
  - name: weekdays
    time_intervals:
      - times:
          - start_time: 09:00
            end_time: 17:00
        weekdays:
        - monday:friday
  - name: weekends
    time_intervals:
      - weekdays:
        - saturday
        - sunday`,
			Templates: []*alertspb.TemplateDesc{{
				Filename: "tmpl1.yml",
			}, {
				Filename: "tmpl2.yml",
			}},
		},
		expected: alertspb.AlertConfigDesc{
			User: "2",
			RawConfig: `global:
  smtp_from: test@example.com
  smtp_smarthost: smtp.example.org:587
  smtp_auth_username: admin
  smtp_auth_password: password
  slack_api_url: https://example.com/1/
  victorops_api_key: foo
  pagerduty_url: https://example.com/2/
  opsgenie_api_key: bar
  opsgenie_api_url: https://example.com/3/
  wechat_api_url: https://example.com/4/
  wechat_api_secret: baz
  wechat_api_corp_id: qux
  telegram_api_url: https://example.com/5/
  webex_api_url: https://example.com/6/
  http_config:
    basic_auth:
      username: admin
      password: password
route:
  receiver: default
  group_by:
  - foo
  continue: false
  routes:
  - matchers:
    - bar="!baz"
    - foo=""
    active_time_intervals:
    - weekdays
    continue: false
    routes:
    - matchers:
      - baz="[\\w+]"
      - qux="[corge]"
      mute_time_intervals:
      - weekends
      continue: true
  group_wait: 1m
  group_interval: 5m
  repeat_interval: 4h
inhibit_rules:
- source_matchers:
  - bar="!baz"
  target_matchers:
  - foo=""
- source_matchers:
  - qux="[corge]"
  target_matchers:
  - baz="[\\w+]"
receivers:
- name: default
  email_configs:
  - to: test@example.com
- name: webhook
  webhook_configs:
  - url: https://example.com/
    http_config:
      authorization:
        type: Bearer
        credentials: token
templates:
- tmpl1.tmpl
- tmpl2.tmpl
time_intervals:
- name: weekdays
  time_intervals:
  - times:
    - start_time: "09:00"
      end_time: "17:00"
    weekdays:
    - monday:friday
- name: weekends
  time_intervals:
  - weekdays:
    - saturday
    - sunday
`,
			Templates: []*alertspb.TemplateDesc{{
				Filename: "tmpl1.yml",
			}, {
				Filename: "tmpl2.yml",
			}},
		},
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual, err := fix(test.input)
			if test.expectedErr != "" {
				require.Equal(t, err, test.expectedErr)
				require.Nil(t, actual)
			} else {
				require.NoError(t, err)
				require.Equal(t, test.expected, *actual)
				ok, diffs, equalErr := isEqual(test.expected, *actual)
				require.NoError(t, equalErr)
				if !ok {
					t.Logf("expected and actual protos are not the same: %s", diffs)
				}
				require.True(t, ok)
			}
		})
	}
}
