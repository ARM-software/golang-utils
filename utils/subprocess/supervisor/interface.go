package supervisor

import "context"

//go:generate mockgen -destination=../../mocks/mock_$GOPACKAGE.go -package=mocks github.com/ARM-software/golang-utils/utils/subprocess/$GOPACKAGE ISupervisor

type ISupervisor interface {
	Run(ctx context.Context) error
}
