// Package logformat provides some ad-hoc log formatting for some of my projects.
package logformat

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/tj/go-logformat/internal/colors"
)

// FormatFunc is a function for formatting values.
type FormatFunc func(string) string

// Formatters is a map of formatting functions.
type Formatters map[string]FormatFunc

// defaultFormatters is a set of default formatters.
var defaultFormatters = Formatters{
	// Levels.
	"debug":     colors.Gray,
	"info":      colors.Purple,
	"warn":      colors.Yellow,
	"warning":   colors.Yellow,
	"error":     colors.Red,
	"fatal":     colors.Red,
	"critical":  colors.Red,
	"emergency": colors.Red,

	// Values.
	"string": colors.None,
	"number": colors.None,
	"bool":   colors.None,
	"date":   colors.None,

	// Fields.
	"object.key":       colors.Purple,
	"object.separator": colors.Gray,
	"object.value":     colors.None,

	// Arrays.
	"array.delimiter": colors.Gray,
	"array.separator": colors.Gray,
}

// rgb struct.
type rgb struct {
	r uint8
	g uint8
	b uint8
}

// config is the formatter configuration.
type config struct {
	format Formatters
}

// Option function.
type Option func(*config)

// WithFormatters option sets the formatters used.
func WithFormatters(f Formatters) Option {
	return func(v *config) {
		v.format = f
	}
}

// Compact returns a value in the compact format.
func Compact(v map[string]interface{}, options ...Option) string {
	c := config{
		format: defaultFormatters,
	}
	for _, o := range options {
		o(&c)
	}
	return compact(v, &c)
}

// compact returns a formatted value.
func compact(v interface{}, c *config) string {
	switch v := v.(type) {
	case map[string]interface{}:
		return compactMap(v, c)
	case []interface{}:
		return compactSlice(v, c)
	default:
		return primitive(v, c)
	}
}

// compactMap returns a formatted map.
func compactMap(m map[string]interface{}, c *config) string {
	s := ""
	keys := mapKeys(m)
	for i, k := range keys {
		v := m[k]
		s += c.format["object.key"](k)
		s += c.format["object.separator"]("=")
		if isComposite(v) {
			s += c.format["object.value"](compact(v, c))
		} else {
			s += compact(v, c)
		}
		if i < len(keys)-1 {
			s += " "
		}
	}
	return s
}

// compactSlice returns a formatted slice.
func compactSlice(v []interface{}, c *config) string {
	s := c.format["array.delimiter"]("[")
	for i, v := range v {
		if i > 0 {
			s += c.format["array.separator"](", ")
		}
		s += compact(v, c)
	}
	return s + c.format["array.delimiter"]("]")
}

// Expanded returns a value in the expanded format.
func Expanded(v map[string]interface{}, options ...Option) string {
	c := config{
		format: defaultFormatters,
	}
	for _, o := range options {
		o(&c)
	}
	return expanded(v, "  ", &c)
}

// expanded returns a formatted value with prefix.
func expanded(v interface{}, prefix string, c *config) string {
	switch v := v.(type) {
	case map[string]interface{}:
		return expandedMap(v, prefix, c)
	case []interface{}:
		return expandedSlice(v, prefix, c)
	default:
		return primitive(v, c)
	}
}

// expandedMap returns a formatted map.
func expandedMap(m map[string]interface{}, prefix string, c *config) string {
	s := ""
	keys := mapKeys(m)
	for _, k := range keys {
		v := m[k]
		k = c.format["object.key"](k)
		d := c.format["object.separator"](":")
		if isComposite(v) {
			s += fmt.Sprintf("%s%s%s\n%s", prefix, k, d, expanded(v, prefix+"  ", c))
		} else {
			s += fmt.Sprintf("%s%s%s %s\n", prefix, k, d, expanded(v, prefix+"  ", c))
		}
	}
	return s
}

// expandedSlice returns a formatted slice.
func expandedSlice(v []interface{}, prefix string, c *config) string {
	s := ""
	for _, v := range v {
		d := c.format["array.separator"]("-")
		if isComposite(v) {
			s += fmt.Sprintf("%s%s\n%s", prefix, d, expanded(v, prefix+"  ", c))
		} else {
			s += fmt.Sprintf("%s%s %v\n", prefix, d, primitive(v, c))
		}
	}
	return s
}

// primitive returns a formatted value.
func primitive(v interface{}, c *config) string {
	switch v := v.(type) {
	case string:
		if strings.ContainsAny(v, " \n\t") || strings.TrimSpace(v) == "" {
			return c.format["string"](strconv.Quote(v))
		} else {
			return c.format["string"](v)
		}
	case time.Time:
		return c.format["date"](formatDate(v))
	case bool:
		return c.format["bool"](strconv.FormatBool(v))
	case float64:
		return c.format["number"](strconv.FormatFloat(v, 'f', -1, 64))
	default:
		return fmt.Sprintf("%v", v)
	}
}

// isComposite returns true if the value is a composite.
func isComposite(v interface{}) bool {
	switch v.(type) {
	case map[string]interface{}:
		return true
	case []interface{}:
		return true
	default:
		return false
	}
}

// mapKeys returns map keys, sorted ascending.
func mapKeys(m map[string]interface{}) (keys []string) {
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return
}

// formatDate formats t relative to now.
func formatDate(t time.Time) string {
	return t.Format(`Jan 2` + dateSuffix(t) + ` 03:04:05pm`)
}

// bold string.
func bold(s string) string {
	return fmt.Sprintf("\033[1m%s\033[0m", s)
}

// dateSuffix returns the date suffix for t.
func dateSuffix(t time.Time) string {
	switch t.Day() {
	case 1, 21, 31:
		return "st"
	case 2, 22:
		return "nd"
	case 3, 23:
		return "rd"
	default:
		return "th"
	}
}
