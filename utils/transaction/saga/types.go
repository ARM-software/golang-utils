package saga

import (
	"crypto/rand"
	"fmt"

	"github.com/ARM-software/golang-utils/utils/reflection"
)

type stepIdentifier struct {
	name      string
	namespace string
}

func (i *stepIdentifier) String() string {
	return fmt.Sprintf("%s@%s", i.name, i.namespace)
}

func (i *stepIdentifier) GetName() string {
	return i.name
}

func (i *stepIdentifier) GetNamespace() string {
	return i.namespace
}

func NewStepIdentifier(name, namespace string) IActionIdentifier {
	return &stepIdentifier{
		name:      name,
		namespace: namespace,
	}
}

type stepArguments struct {
	idemKey string
	args    map[string]any
}

func (a *stepArguments) GetIdemKey() string {
	return a.idemKey
}

func (a *stepArguments) GetArguments() map[string]any {
	return a.args
}

func NewStepArgumentsWithIdempotentKey(idemKey string, args map[string]any) IActionArguments {
	m := args
	if reflection.IsEmpty(m) {
		m = map[string]any{}
	}
	return &stepArguments{
		idemKey: idemKey,
		args:    m,
	}
}

func NewStepArguments(args map[string]any) IActionArguments {
	return NewStepArgumentsWithIdempotentKey(rand.Text(), args) //nolint:gosec
}

func NoStepArguments() IActionArguments {
	return NewStepArguments(nil)
}
