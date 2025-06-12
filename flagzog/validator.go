package flagzog

import (
	"context"
	"log/slog"

	"github.com/spf13/pflag"
)

type FlagValidator interface {
	Name() string
	Validate() ([]string, error)
	Usage() string
	AddToCommandFlags(flagset *pflag.FlagSet, shorthand string, value interface{})
}

func ValidateFlags(ctx context.Context, logger *slog.Logger, flags []FlagValidator) ([]string, error) {
	var validatedFlags []string
	var err error

	for _, flag := range flags {
		var issues []string
		if issues, err = flag.Validate(); err != nil {
			logger.Log(ctx, slog.LevelError, "validation failed", slog.String("flag", flag.Name()), slog.Any("issues", issues))
			return validatedFlags, err
		}
		validatedFlags = append(validatedFlags, flag.Name())
	}
	return validatedFlags, nil
}
