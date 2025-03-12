/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package pagination

import (
	"context"
	"time"

	"go.uber.org/atomic"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/parallelisation"
)

type AbstractStreamPaginator struct {
	AbstractPaginator
	fetchFutureFunc func(context.Context, IStaticPageStream) (IStaticPageStream, error)
	runningOut      *atomic.Bool
	timeReachLast   *atomic.Time
	timeOut         time.Duration
	backoff         time.Duration
}

func (s *AbstractStreamPaginator) FetchFuturePage(ctx context.Context, currentPage IStaticPageStream) (IStaticPageStream, error) {
	return s.fetchFutureFunc(ctx, currentPage)
}

func (s *AbstractStreamPaginator) Close() error {
	return s.AbstractPaginator.Close()
}

func (s *AbstractStreamPaginator) HasNext() bool {
	for {
		if s.AbstractPaginator.HasNext() {
			s.timeReachLast.Store(time.Now())
			return true
		}
		page, err := s.AbstractPaginator.FetchCurrentPage()
		if err != nil {
			return false
		}
		stream, ok := page.(IStaticPageStream)
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
		} else {
			s.timeReachLast.Store(time.Now())
		}
		future, err := s.FetchFuturePage(s.GetContext(), stream)
		if err != nil {
			return false
		}
		err = s.AbstractPaginator.SetCurrentPage(future)
		if err != nil {
			return false
		}

		parallelisation.SleepWithContext(s.GetContext(), s.backoff)
	}
}

func (s *AbstractStreamPaginator) GetNext() (interface{}, error) {
	for {
		item, err := s.AbstractPaginator.GetNext()
		if commonerrors.Any(err, nil, commonerrors.ErrCancelled, commonerrors.ErrTimeout) {
			return item, err
		}

		if !s.HasNext() {
			err = commonerrors.New(commonerrors.ErrNotFound, "there is not any next item")
			return nil, err
		}
		parallelisation.SleepWithContext(s.GetContext(), s.backoff)
	}
}

func (s *AbstractStreamPaginator) Stop() context.CancelFunc {
	return s.AbstractPaginator.Stop()
}

func (s *AbstractStreamPaginator) DryUp() error {
	s.runningOut.Store(true)
	return nil
}

func (s *AbstractStreamPaginator) IsRunningDry() bool {
	return s.runningOut.Load()
}

// newAbstractStreamPaginator creates a generic paginator over a stream of static pages.
// runOutTimeOut corresponds to the grace period between the stream being marked as running dry and the iteration actually ending
// backoff corresponds to the backoff time between page iteration.
func newAbstractStreamPaginator(ctx context.Context, runOutTimeOut, backoff time.Duration, fetchFirstPageFunc func(context.Context) (IStaticPage, error), fetchNextFunc func(context.Context, IStaticPage) (IStaticPage, error), fetchFutureFunc func(context.Context, IStaticPageStream) (IStaticPageStream, error)) (paginator *AbstractStreamPaginator, err error) {
	firstPage, err := fetchFirstPageFunc(ctx)
	if err != nil {
		return
	}
	parent, err := NewAbstractPaginator(ctx, firstPage, fetchNextFunc)
	if err != nil {
		return
	}
	paginator = &AbstractStreamPaginator{
		AbstractPaginator: *parent,
		fetchFutureFunc:   fetchFutureFunc,
		runningOut:        atomic.NewBool(false),
		timeReachLast:     atomic.NewTime(time.Now()),
		timeOut:           runOutTimeOut,
		backoff:           backoff,
	}
	return
}

func toDynamicStream(stream IStaticPageStream) (dynamicStream IStream, err error) {
	if stream == nil {
		return
	}
	dynamicStream, ok := stream.(IStream)
	if !ok {
		err = commonerrors.New(commonerrors.ErrInvalid, "current stream is not dynamic i.e. it is not possible to fetch next pages from it")
	}
	return
}

// DynamicPageStreamPaginator defines a paginator over a stream of dynamic pages i.e. pages from which it is possible to access the next one.
type DynamicPageStreamPaginator struct {
	AbstractStreamPaginator
}

func (d *DynamicPageStreamPaginator) GetCurrentPage() (dynamicPage IPage, err error) {
	p, err := d.FetchCurrentPage()
	if err != nil {
		return
	}
	dynamicPage, err = toDynamicPage(p)
	return
}

// NewStreamPaginator creates a paginator over a stream of dynamic pages i.e. pages can access next and future pages.
// runOutTimeOut corresponds to the grace period between the stream being marked as running dry and the iteration actually ending
// backoff corresponds to the backoff time between page iteration.
func NewStreamPaginator(ctx context.Context, runOutTimeOut, backoff time.Duration, fetchFirstPageFunc func(context.Context) (IStream, error)) (paginator IStreamPaginator, err error) {
	parent, err := newAbstractStreamPaginator(ctx, runOutTimeOut, backoff, func(fCtx context.Context) (IStaticPage, error) {
		return fetchFirstPageFunc(fCtx)
	}, func(fCtx context.Context, current IStaticPage) (nextPage IStaticPage, err error) {
		p, err := toDynamicPage(current)
		if err != nil {
			return
		}
		nextPage, err = p.GetNext(fCtx)
		return
	}, func(fCtx context.Context, current IStaticPageStream) (future IStaticPageStream, err error) {
		s, err := toDynamicStream(current)
		if err != nil {
			return
		}
		future, err = s.GetFuture(fCtx)
		return
	})
	if err != nil {
		return
	}
	paginator = &DynamicPageStreamPaginator{
		AbstractStreamPaginator: *parent,
	}
	return
}

// StaticPageStreamPaginator defines a paginator over a stream of static pages i.e. pages from which it is not possible to access the next nor the future pages.
type StaticPageStreamPaginator struct {
	AbstractStreamPaginator
}

func (d *StaticPageStreamPaginator) GetCurrentPage() (IStaticPage, error) {
	return d.FetchCurrentPage()
}

// NewStaticPageStreamPaginator creates a paginator over a stream but the pages are static i.e. they cannot access future and next pages from themselves.
// runOutTimeOut corresponds to the grace period between the stream being marked as running dry and the iteration actually ending
// backoff corresponds to the backoff time between page iteration.
func NewStaticPageStreamPaginator(ctx context.Context, runOutTimeOut, backoff time.Duration, fetchFirstPageFunc func(context.Context) (IStaticPageStream, error), fetchNextPageFunc func(context.Context, IStaticPage) (IStaticPage, error), fetchFutureFunc func(context.Context, IStaticPageStream) (IStaticPageStream, error)) (paginator IStreamPaginatorAndPageFetcher, err error) {
	parent, err := newAbstractStreamPaginator(ctx, runOutTimeOut, backoff, func(fCtx context.Context) (IStaticPage, error) {
		return fetchFirstPageFunc(fCtx)
	}, fetchNextPageFunc, fetchFutureFunc)
	if err != nil {
		return
	}
	paginator = &StaticPageStreamPaginator{
		AbstractStreamPaginator: *parent,
	}
	return
}
