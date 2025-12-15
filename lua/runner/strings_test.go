package runner

import "testing"

func TestDedent(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "SQL query with tabs",
			input: `
		INSERT INTO deployments (team_slug, repository, environment_name)
			VALUES ('slug-2', CONCAT('org/repo-', $1::text), 'dev')
		RETURNING id::text
	`,
			expected: `INSERT INTO deployments (team_slug, repository, environment_name)
	VALUES ('slug-2', CONCAT('org/repo-', $1::text), 'dev')
RETURNING id::text`,
		},
		{
			name: "GraphQL query",
			input: `
		query {
			users {
				id
				name
				email
			}
		}
	`,
			expected: `query {
	users {
		id
		name
		email
	}
}`,
		},
		{
			name:     "no indentation",
			input:    "SELECT * FROM users",
			expected: "SELECT * FROM users",
		},
		{
			name: "empty lines in middle",
			input: `
		SELECT *
		FROM users

		WHERE id = 1
	`,
			expected: `SELECT *
FROM users

WHERE id = 1`,
		},
		{
			name: "mixed content with empty lines",
			input: `
		line1

		line2
			indented
		line3
	`,
			expected: `line1

line2
	indented
line3`,
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "only whitespace",
			input:    "   \n   \n   ",
			expected: "",
		},
		{
			name: "tabs and spaces mixed",
			input: `
	first line
	second line
		third indented
	`,
			expected: `first line
second line
	third indented`,
		},
		{
			name: "single line with leading whitespace",
			input: `
		single line
	`,
			expected: "single line",
		},
		{
			name: "whitespace only lines should be preserved as empty",
			input: `
		line1

		line2
	`,
			expected: `line1

line2`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Dedent(tt.input)
			if result != tt.expected {
				t.Errorf("Dedent() mismatch\ngot:\n%q\n\nexpected:\n%q", result, tt.expected)
			}
		})
	}
}
