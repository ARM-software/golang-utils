package supervisor

import "context"

//go:generate mockgen -destination=../../mocks/mock_$GOPACKAGE.go -package=mocks github.com/ARM-software/golang-utils/utils/subprocess/$GOPACKAGE ISupervisor

// ISupervisor  will run a command and automatically restart it if it exits. Hooks can be used to execute code at
// different points in the execution lifecyle. Restarts can be delayed
type ISupervisor interface {
	// Run will run the supervisor and execute any of the command hooks. If it receives a halting error or the context is cancelled then it will exit
	Run(ctx context.Context) error
	Stop() error
}
