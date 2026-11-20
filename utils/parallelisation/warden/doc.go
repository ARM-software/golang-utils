/*
 * Copyright (C) 2020-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// Package warden provides lifecycle tracking for one or more goroutines.
//
// A warden tracks whether a set of goroutines is still alive, is in the
// process of shutting down, or is fully dead, and records the reason for that
// termination.
//
// This is primarily useful because goroutines cannot be forcibly interrupted,
// preemptively reclaimed, or otherwise directly controlled by application code
// once they have been spawned. They must cooperate by observing shared state,
// cancellation, or shutdown signals and then returning on their own.
//
// A warden provides a small lifecycle layer around that reality so related
// goroutines can share:
//   - a shutdown signal
//   - a derived context model
//   - a final death reason
//   - a way to wait until all tracked goroutines have really stopped
//
// This is useful when you need to:
//   - start several related goroutines and wait for them as a unit
//   - propagate shutdown through derived contexts
//   - stop all tracked work when one goroutine fails
//   - distinguish between graceful completion and error-driven termination
//
// The design is intentionally similar to the following prior art:
//   - tomb:
//     https://pkg.go.dev/gopkg.in/tomb.v2
//   - Go contexts:
//     https://pkg.go.dev/context
//   - Go blog, Share Memory By Communicating:
//     https://go.dev/blog/codelab-share
//   - Go blog, Context:
//     https://go.dev/blog/context
//
// Compared with a plain [context.Context], a warden additionally tracks
// goroutine membership and exposes lifecycle signals such as [IState.Dying] and
// [IState.Dead].
//
// Compared with `sync.WaitGroup`, a warden does more than wait for completion:
// it also tracks a shared lifecycle state, captures the reason for shutdown,
// and can derive contexts that are cancelled when the tracked state starts
// dying.
//
// Compared with `errgroup.Group`, a warden exposes explicit lifecycle signals
// and state transitions such as alive, dying, and dead. It is therefore useful
// when callers need to observe shutdown as an event in its own right rather
// than only waiting for a final aggregated error.
package warden
