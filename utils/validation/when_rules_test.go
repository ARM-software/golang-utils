package validation

import (
	"regexp"
	"testing"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ARM-software/golang-utils/utils/collection"
	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/commonerrors/errortest"
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

	t.Run("property not equals", func(t *testing.T) {
		rule := WhenPropertyNotEquals("mode", "strict", RequiredProperties("name"))
		assert.Error(t, validation.Validate(map[string]any{"mode": "relaxed"}, rule))
		assert.NoError(t, validation.Validate(map[string]any{"mode": "relaxed", "name": "ok"}, rule))
		assert.NoError(t, validation.Validate(map[string]any{"mode": "strict"}, rule))
	})

	t.Run("field not equals", func(t *testing.T) {
		type config struct {
			Mode string
			Name string
		}
		cfg := &config{Mode: "relaxed"}
		rule := WhenFieldNotEquals(&cfg.Mode, "strict", PatternProperties(
			PatternProperty{Pattern: regexp.MustCompile(`^Name$`), Rule: MinLength(1)},
		))
		err := validation.Validate(cfg, rule)
		require.Error(t, err)

		cfg.Name = "ok"
		assert.NoError(t, validation.Validate(cfg, rule))

		cfg.Mode = "strict"
		cfg.Name = ""
		assert.NoError(t, validation.Validate(cfg, rule))
	})

	t.Run("field not equals value", func(t *testing.T) {
		type config struct {
			Mode string
			Name string
		}
		cfg := &config{Mode: "relaxed"}
		rule := WhenFieldNotEqualsValue(&cfg.Mode, "strict", PatternProperties(
			PatternProperty{Pattern: regexp.MustCompile(`^Name$`), Rule: MinLength(1)},
		))
		err := validation.Validate(cfg, rule)
		require.Error(t, err)
		cfg.Name = "ok"
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

	t.Run("property absent and present", func(t *testing.T) {
		absentRule := WhenPropertyAbsent("name", RequiredProperties("mode"))
		presentRule := WhenPropertyPresent("name", RequiredProperties("mode"))

		assert.Error(t, validation.Validate(map[string]any{"name": ""}, absentRule))
		assert.Error(t, validation.Validate(map[string]any{}, absentRule))
		assert.NoError(t, validation.Validate(map[string]any{"name": "", "mode": "strict"}, absentRule))

		assert.Error(t, validation.Validate(map[string]any{"name": "ok"}, presentRule))
		assert.NoError(t, validation.Validate(map[string]any{"name": "ok", "mode": "strict"}, presentRule))
		assert.NoError(t, validation.Validate(map[string]any{}, presentRule))
	})

	t.Run("field absent and present", func(t *testing.T) {
		type config struct {
			Mode string
			Name string
		}

		absentCfg := &config{}
		absentRule := WhenFieldAbsent(&absentCfg.Name, PatternProperties(
			PatternProperty{Pattern: regexp.MustCompile(`^Mode$`), Rule: MinLength(1)},
		))
		err := validation.Validate(absentCfg, absentRule)
		require.Error(t, err)
		absentCfg.Mode = "strict"
		assert.NoError(t, validation.Validate(absentCfg, absentRule))

		presentCfg := &config{Name: "ok"}
		presentRule := WhenFieldPresent(&presentCfg.Name, PatternProperties(
			PatternProperty{Pattern: regexp.MustCompile(`^Mode$`), Rule: MinLength(1)},
		))
		err = validation.Validate(presentCfg, presentRule)
		require.Error(t, err)
		presentCfg.Mode = "strict"
		assert.NoError(t, validation.Validate(presentCfg, presentRule))

		presentCfg.Name = ""
		presentCfg.Mode = ""
		assert.NoError(t, validation.Validate(presentCfg, presentRule))
	})

	t.Run("field in and not in values", func(t *testing.T) {
		type config struct {
			Mode string
			Name string
		}
		cfg := &config{Mode: "strict"}
		inRule := WhenFieldInValues(&cfg.Mode, []any{"strict", "relaxed"}, PatternProperties(
			PatternProperty{Pattern: regexp.MustCompile(`^Name$`), Rule: MinLength(1)},
		))
		err := validation.Validate(cfg, inRule)
		require.Error(t, err)
		cfg.Name = "ok"
		assert.NoError(t, validation.Validate(cfg, inRule))

		notInCfg := &config{Mode: "archive"}
		notInRule := WhenFieldNotInValues(&notInCfg.Mode, []any{"strict", "relaxed"}, PatternProperties(
			PatternProperty{Pattern: regexp.MustCompile(`^Name$`), Rule: MinLength(1)},
		))
		err = validation.Validate(notInCfg, notInRule)
		require.Error(t, err)
		notInCfg.Name = "ok"
		assert.NoError(t, validation.Validate(notInCfg, notInRule))
	})

	t.Run("field equals value and not empty do not recurse inside Validate", func(t *testing.T) {
		recursionErr := commonerrors.New(commonerrors.ErrUnexpected, "recursive validate call")
		type config struct {
			Summary     string
			SummaryFile string
			validating  bool
		}
		validateConfig := func(cfg *config) error {
			if cfg.validating {
				return recursionErr
			}
			cfg.validating = true
			defer func() {
				cfg.validating = false
			}()
			return NewAllRule(
				WhenFieldEqualsValue(&cfg.Summary, "file", RequiredFieldsBy(&cfg.SummaryFile)),
				WhenFieldPresent(&cfg.SummaryFile, RequiredFieldsBy(&cfg.Summary)),
			).Validate(cfg)
		}

		cfg := &config{Summary: "file"}
		err := validateConfig(cfg)
		require.Error(t, err)
		errortest.AssertErrorDescription(t, err, "SummaryFile")
		assert.False(t, commonerrors.Any(err, commonerrors.ErrUnexpected))

		cfg.SummaryFile = "summary.md"
		require.NoError(t, validateConfig(cfg))

		cfg = &config{SummaryFile: "summary.md"}
		err = validateConfig(cfg)
		require.Error(t, err)
		errortest.AssertErrorDescription(t, err, "Summary")
		assert.False(t, commonerrors.Any(err, commonerrors.ErrUnexpected))
	})
}
