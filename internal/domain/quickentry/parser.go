// Package quickentry parses Quick Entry input strings of the form
// "buy milk #tag @today @project !2026-06-01" into a structured Parsed value.
// It does NOT resolve project names to IDs — that is a service-layer concern.
package quickentry

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jtprogru/todushka/internal/domain/task"
)

type Parsed struct {
	Title      string
	Tags       []string
	StartDate  *task.Date
	Deadline   *task.Date
	ProjectRef *string
}

type ParseError struct {
	Position int
	Token    string
	Reason   string
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("quickentry: token %q at position %d: %s", e.Token, e.Position, e.Reason)
}

var (
	ErrEmptyInput   = errors.New("quickentry: empty input")
	ErrInvalidDate  = errors.New("quickentry: invalid date")
	ErrEmptyTag     = errors.New("quickentry: empty tag name")
	ErrEmptyMention = errors.New("quickentry: empty mention")
)

// IsEmpty reports whether input is empty or whitespace-only.
func IsEmpty(input string) bool { return strings.TrimSpace(input) == "" }

// Parse breaks input into tokens and classifies them.
//
// Recognized tokens:
//   - `#name` — tag (name normalized at service layer)
//   - `@today` — sets start_date to today (clock injected by service)
//   - `@<name>` — project reference (resolution is service's job)
//   - `!YYYY-MM-DD` — deadline
//
// Anything else is treated as part of the title, preserved in input order.
func Parse(input string) (Parsed, error) {
	if IsEmpty(input) {
		return Parsed{}, ErrEmptyInput
	}
	var p Parsed
	titleParts := make([]string, 0, 8)
	tokens := tokenize(input)
	for _, tok := range tokens {
		switch {
		case strings.HasPrefix(tok.text, "#"):
			name := strings.TrimPrefix(tok.text, "#")
			if name == "" {
				return Parsed{}, &ParseError{Position: tok.pos, Token: tok.text, Reason: "empty tag name"}
			}
			p.Tags = append(p.Tags, name)
		case strings.HasPrefix(tok.text, "@"):
			body := strings.TrimPrefix(tok.text, "@")
			if body == "" {
				return Parsed{}, &ParseError{Position: tok.pos, Token: tok.text, Reason: "empty mention"}
			}
			if strings.EqualFold(body, "today") {
				// service-layer resolves to actual local date
				today := task.NewDate(time.Date(0, 1, 1, 0, 0, 0, 0, time.Local))
				p.StartDate = &today
				continue
			}
			project := body
			p.ProjectRef = &project
		case strings.HasPrefix(tok.text, "!"):
			body := strings.TrimPrefix(tok.text, "!")
			t, err := time.ParseInLocation("2006-01-02", body, time.Local)
			if err != nil {
				return Parsed{}, &ParseError{Position: tok.pos, Token: tok.text, Reason: "invalid date"}
			}
			d := task.NewDate(t)
			p.Deadline = &d
		default:
			titleParts = append(titleParts, tok.text)
		}
	}
	p.Title = strings.Join(titleParts, " ")
	if strings.TrimSpace(p.Title) == "" {
		return Parsed{}, &ParseError{Position: 0, Token: input, Reason: "title must not be empty after token extraction"}
	}
	return p, nil
}

type token struct {
	text string
	pos  int
}

func tokenize(s string) []token {
	tokens := make([]token, 0, 8)
	start := -1
	for i, r := range s {
		if r == ' ' || r == '\t' || r == '\n' {
			if start >= 0 {
				tokens = append(tokens, token{text: s[start:i], pos: start})
				start = -1
			}
			continue
		}
		if start < 0 {
			start = i
		}
	}
	if start >= 0 {
		tokens = append(tokens, token{text: s[start:], pos: start})
	}
	return tokens
}
