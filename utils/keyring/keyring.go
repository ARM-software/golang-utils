package keyring

import (
	"context"

	"github.com/zalando/go-keyring"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/parallelisation"
	"github.com/ARM-software/golang-utils/utils/reflection"
	"github.com/ARM-software/golang-utils/utils/serialization/maps" //nolint:misspell
)

// StorePointer is like Store but working on a pointer.
func StorePointer[T any](ctx context.Context, prefix string, cfg T) (err error) {
	m, err := maps.ToMapFromPointer[T](cfg)
	if err != nil {
		return
	}
	err = parallelisation.DetermineContextError(ctx)
	if err != nil {
		return
	}
	for k, v := range m {
		if reflection.IsEmpty(v) {
			continue
		}
		subErr := parallelisation.DetermineContextError(ctx)
		if subErr != nil {
			err = subErr
			return
		}
		subErr = convertError(keyring.Set(prefix, k, v))
		if subErr != nil {
			if commonerrors.Any(subErr, commonerrors.ErrUnsupported) {
				err = subErr
				return
			}
			err = commonerrors.Join(err, commonerrors.WrapErrorf(commonerrors.ErrUnexpected, subErr, "failed to store '%s' in keyring", k))
		}
	}
	if err != nil {
		err = commonerrors.WrapErrorf(commonerrors.ErrUnexpected, err, "failed to store %v credentials in keyring", prefix)
	}
	return
}

// Store stores a configuration into system's keyring service.
func Store[T any](ctx context.Context, prefix string, cfg *T) error {
	return StorePointer[*T](ctx, prefix, cfg)
}

// FetchPointer is like Fetch but working on a pointer.
func FetchPointer[T any](ctx context.Context, prefix string, cfg T) (err error) {
	m, err := maps.ToMapFromPointer[T](cfg)
	if err != nil {
		return
	}
	err = parallelisation.DetermineContextError(ctx)
	if err != nil {
		return
	}
	for k := range m {
		subErr := parallelisation.DetermineContextError(ctx)
		if subErr != nil {
			err = subErr
			return
		}
		value, _ := keyring.Get(prefix, k)
		if reflection.IsEmpty(value) {
			continue
		}
		m[k] = value

	}
	err = maps.FromMapToPointer[T](m, cfg)
	if err != nil {
		err = commonerrors.WrapErrorf(commonerrors.ErrUnexpected, err, "failed retrieving %v credentials from keyring", prefix)
	}
	return
}

// Fetch fetches a configuration from system's keyring service.
func Fetch[T any](ctx context.Context, prefix string, cfg *T) error {
	return FetchPointer[*T](ctx, prefix, cfg)
}

// Clear removes any entry related to a configuration identified by a prefix.
func Clear(_ context.Context, prefix string) (err error) {
	err = commonerrors.Ignore(convertError(keyring.DeleteAll(prefix)), commonerrors.ErrNotFound, commonerrors.ErrUnsupported)
	if err != nil {
		err = commonerrors.WrapErrorf(commonerrors.ErrUnexpected, err, "failed to clear the keyring of %v credentials", prefix)
	}
	return
}

func convertError(err error) error {
	err = commonerrors.ConvertContextError(err)
	switch {
	case err == nil:
		return nil
	case commonerrors.Any(err, keyring.ErrNotFound):
		return commonerrors.WrapError(commonerrors.ErrNotFound, err, "")
	case commonerrors.Any(err, keyring.ErrSetDataTooBig):
		return commonerrors.WrapError(commonerrors.ErrTooLarge, err, "")
	case commonerrors.Any(err, keyring.ErrUnsupportedPlatform) || commonerrors.CorrespondTo(err, "was not provided by any .service files"):
		return commonerrors.WrapError(commonerrors.ErrUnsupported, err, "")
	default:
		return err
	}
}
