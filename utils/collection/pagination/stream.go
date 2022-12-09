/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package pagination

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/atomic"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/parallelisation"
)

type StreamPaginator struct {
	AbstractPaginator
	runningOut    *atomic.Bool
	timeReachLast *atomic.Time
	timeOut       time.Duration
	backoff       time.Duration
}

func (s *StreamPaginator) Close() error {
	return s.AbstractPaginator.Close()
}

func (s *StreamPaginator) HasNext() bool {
	if s.AbstractPaginator.HasNext() {
		s.timeReachLast.Store(time.Now())
		return true
	}
	page, err := s.AbstractPaginator.GetCurrentPage()
	if err != nil {
		return false
	}
	stream, ok := page.(IStream)
	if !ok {
		return false
	}
	if !stream.HasFuture() {
		return false
	}
	if s.IsRunningDry() {
		if time.Since(s.timeReachLast.Load()) >= s.timeOut {
			return false
		}
		future, err := stream.GetFuture(s.GetContext())
		if err != nil {
			return false
		}
		err = s.AbstractPaginator.SetCurrentPage(future)
		if err != nil {
			return false
		}
	} else {
		s.timeReachLast.Store(time.Now())
	}
	return s.HasNext()
}

func (s *StreamPaginator) GetNext() (*interface{}, error) {
	for {
		item, err := s.AbstractPaginator.GetNext()
		if commonerrors.Any(err, nil, commonerrors.ErrCancelled, commonerrors.ErrTimeout) {
			return item, err
		}

		if !s.HasNext() {
			err = fmt.Errorf("%w: there is not any next item", commonerrors.ErrNotFound)
			return nil, err
		}

		parallelisation.SleepWithContext(s.GetContext(), s.backoff)

	}
}

func (s *StreamPaginator) Stop() context.CancelFunc {
	return s.AbstractPaginator.Stop()
}

func (s *StreamPaginator) GetCurrentPage() (IPage, error) {
	return s.AbstractPaginator.GetCurrentPage()
}

func (s *StreamPaginator) DryUp() error {
	s.runningOut.Store(true)
	return nil
}

func (s *StreamPaginator) IsRunningDry() bool {
	return s.runningOut.Load()
}

// NewStreamPaginator creates a paginator over a stream.
// runOutTimeOut corresponds to the grace period between the stream being marked as running dry and the iteration actually ending
// backoff corresponds to the backoff time between page iteration.
func NewStreamPaginator(ctx context.Context, runOutTimeOut, backoff time.Duration, fetchFirstPage func(context.Context) (IStream, error)) (paginator IStreamPaginator, err error) {
	firstPage, err := fetchFirstPage(ctx)
	if err != nil {
		return
	}
	parent := NewAbstractPaginator(ctx, firstPage).(*AbstractPaginator)
	if parent == nil {
		err = fmt.Errorf("%w: missing abstract paginator", commonerrors.ErrUndefined)
		return
	}
	paginator = &StreamPaginator{
		AbstractPaginator: *parent,
		runningOut:        atomic.NewBool(false),
		timeReachLast:     atomic.NewTime(time.Now()),
		timeOut:           runOutTimeOut,
		backoff:           backoff,
	}
	return
}
