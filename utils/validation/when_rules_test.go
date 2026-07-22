package validation

import (
	"regexp"
	"testing"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ARM-software/golang-utils/utils/collection"
)

func TestWhenRules(t *testing.T) {
	t.Run("property equals", func(t *testing.T) {
		rule := WhenPropertyEquals("mode", "strict", RequiredProperties("name"))
		assert.NoError(t, validation.Validate(map[string]any{"mode": "relaxed"}, rule))
		assert.Error(t, validation.Validate(map[string]any{"mode": "strict"}, rule))
		assert.NoError(t, validation.Validate(map[string]any{"mode": "strict", "name": "ok"}, rule))
	})

	t.Run("field equals", func(t *testing.T) {
		type config struct {
			Mode string
			Name string
		}
		cfg := &config{Mode: "strict"}
		rule := WhenFieldEquals(&cfg.Mode, "strict", PatternProperties(
			PatternProperty{Pattern: regexp.MustCompile(`^Name$`), Rule: MinLength(1)},
		))
		err := validation.Validate(cfg, rule)
		require.Error(t, err)

		cfg.Name = "ok"
		assert.NoError(t, validation.Validate(cfg, rule))

		cfg.Mode = "relaxed"
		cfg.Name = ""
		assert.NoError(t, validation.Validate(cfg, rule))
	})

	t.Run("property matches", func(t *testing.T) {
		rule := WhenPropertyMatches("mode", "strict", collection.StringCaseInsensitiveMatch, RequiredProperties("name"))
		assert.Error(t, validation.Validate(map[string]any{"mode": "STRICT"}, rule))
		assert.NoError(t, validation.Validate(map[string]any{"mode": "STRICT", "name": "ok"}, rule))
	})

	t.Run("field matches", func(t *testing.T) {
		type config struct {
			Mode string
			Name string
		}
		cfg := &config{Mode: "STRICT"}
		rule := WhenFieldMatches(&cfg.Mode, "strict", collection.StringCaseInsensitiveMatch, PatternProperties(
			PatternProperty{Pattern: regexp.MustCompile(`^Name$`), Rule: MinLength(1)},
		))
		err := validation.Validate(cfg, rule)
		require.Error(t, err)

		cfg.Name = "ok"
		assert.NoError(t, validation.Validate(cfg, rule))
	})
}
