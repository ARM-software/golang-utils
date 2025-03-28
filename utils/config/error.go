package config

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"golang.org/x/exp/maps"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/field"
	"github.com/ARM-software/golang-utils/utils/reflection"
)

// WrapFieldValidationError creates an error resulting from the validation of a field in a structure
func WrapFieldValidationError(fieldName string, mapStructure, prefix *string, err error) IValidationError {
	vErr := newValidationError(err)
	if vErr == nil {
		return nil
	}
	vErr.RecordField(fieldName, mapStructure, prefix)
	return vErr
}

// WrapValidationError creates an error resulting from the validation of a structure
func WrapValidationError(prefix *string, err error) IValidationError {
	vErr := newValidationError(err)
	if vErr == nil {
		return nil
	}
	if !reflection.IsEmpty(prefix) {
		vErr.RecordPrefix(*prefix)
	}
	return vErr
}

// IValidationError defines a typical structure validation error.
type IValidationError interface {
	error
	GetMapStructurePath() string
	GetTreePath() string
	GetReason() string
	Unwrap() error
	RecordField(fieldName string, mapStructureFieldName *string, mapStructurePrefix *string)
	RecordPrefix(mapStructurePrefix string)
	fmt.Stringer
	GetTree() []string
	GetMapStructureTree() []string
	GetMapStructurePrefix() *string
}

type validationError struct {
	tree               []string
	mapStructureTree   []string
	mapStructurePrefix *string
	reason             string
}

func (v *validationError) GetTree() []string {
	return v.tree
}

func (v *validationError) GetMapStructureTree() []string {
	return v.mapStructureTree
}

func (v *validationError) GetMapStructurePrefix() *string {
	return v.mapStructurePrefix
}

func (v *validationError) RecordField(fieldName string, mapStructureFieldName *string, mapStructurePrefix *string) {
	tree := make([]string, 0, len(v.tree)+1)
	tree = append(tree, strings.TrimSpace(fieldName))
	tree = append(tree, v.tree...)
	v.tree = tree
	if mapStructureFieldName != nil {
		treeMap := make([]string, 0, len(v.mapStructureTree)+1)
		treeMap = append(treeMap, strings.ToUpper(strings.TrimSpace(*mapStructureFieldName)))
		treeMap = append(treeMap, v.mapStructureTree...)
		v.mapStructureTree = treeMap
	}
	v.mapStructurePrefix = mapStructurePrefix

}

func (v *validationError) RecordPrefix(mapStructurePrefix string) {
	v.mapStructurePrefix = field.ToOptionalString(mapStructurePrefix)
}

func (v *validationError) Error() string {
	mapstructureStr := v.GetMapStructurePath()
	if mapstructureStr != "" {
		mapstructureStr = fmt.Sprintf(" [%v]", mapstructureStr)
	}
	treeStr := v.GetTreePath()
	if treeStr != "" {
		treeStr = fmt.Sprintf(" (%v)", treeStr)
	}
	reasonStr := v.GetReason()
	if reasonStr != "" {
		reasonStr = fmt.Sprintf(" %v", reasonStr)
	}

	return commonerrors.Newf(v.Unwrap(), "structure failed validation:%v%v%v", treeStr, mapstructureStr, reasonStr).Error()
}

func (v *validationError) GetMapStructurePath() string {
	mapstructureStr := ""
	if len(v.mapStructureTree) > 0 {
		mapstructureStr = strings.Join(v.mapStructureTree, "_")
		mapstructureStr = strings.ReplaceAll(mapstructureStr, "-", "_")
		if v.mapStructurePrefix != nil {
			mapstructureStr = fmt.Sprintf("%v_%v", strings.ToUpper(strings.TrimSpace(*v.mapStructurePrefix)), mapstructureStr)
		}
	}
	return mapstructureStr
}

func (v *validationError) GetTreePath() string {
	treeStr := ""
	if len(v.tree) > 0 {
		treeStr = strings.Join(v.tree, "->")
	}
	return treeStr
}

func (v *validationError) GetReason() string {
	return v.reason
}

func (v *validationError) Unwrap() error {
	return commonerrors.ErrInvalid
}

func (v *validationError) String() string {
	return v.Error()
}

func newValidationError(err error) *validationError {
	if err == nil {
		return nil
	}
	var vErr *validationError
	if errors.As(err, &vErr) {
		return vErr
	}
	var ve IValidationError
	if errors.As(err, &ve) {
		return newValidationErrorFromIValidationError(ve)
	}
	var oe validation.Error
	if errors.As(err, &oe) {
		return newValidationErrorFromOzzoValidation(oe)
	}
	var oes validation.Errors
	if errors.As(err, &oes) {
		return newValidationErrorFromOzzoValidationErrors(oes)
	}

	reason, subErr := commonerrors.GetCommonErrorReason(err)
	if subErr != nil {
		reason = err.Error()
	}
	return &validationError{
		reason: reason,
	}

}

func newValidationErrorFromOzzoValidationErrors(oes validation.Errors) *validationError {
	if len(oes) == 0 {
		return &validationError{
			reason: oes.Error(),
		}
	}
	// Only store the one parameter
	params := maps.Keys(oes)
	slices.Sort(params)
	param := params[0]
	veo := &validationError{
		reason: oes[param].Error(),
	}
	veo.RecordField(param, nil, nil)
	return veo
}

func newValidationErrorFromOzzoValidation(oe validation.Error) *validationError {
	veo := &validationError{
		reason: oe.Message(),
	}
	// Only store the one parameter
	params := maps.Keys(oe.Params())
	slices.Sort(params)
	if len(params) > 0 {
		veo.RecordField(params[0], nil, nil)
	}
	return veo
}

func newValidationErrorFromIValidationError(ve IValidationError) *validationError {
	return &validationError{
		tree:               ve.GetTree(),
		mapStructureTree:   ve.GetMapStructureTree(),
		mapStructurePrefix: ve.GetMapStructurePrefix(),
		reason:             ve.GetReason(),
	}
}
