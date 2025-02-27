/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package pagination

import (
	"context"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/parallelisation"
)

type AbstractPaginator struct {
	currentPage         IStaticPage
	currentPageIterator IIterator
	cancellationStore   *parallelisation.CancelFunctionStore
	fetchNextFunc       func(context.Context, IStaticPage) (IStaticPage, error)
	ctx                 context.Context
}

func (a *AbstractPaginator) Close() error {
	a.Stop()()
	return nil
}

func (a *AbstractPaginator) HasNext() bool {
	if parallelisation.DetermineContextError(a.ctx) != nil {
		return false
	}
	currentIt, err := a.FetchCurrentPageIterator()
	if err != nil {
		return false
	}
	if currentIt.HasNext() {
		return true
	}
	currentPage, err := a.FetchCurrentPage()
	if err != nil {
		return false
	}
	if !currentPage.HasNext() {
		return false
	}
	err = a.fetchNextPage()
	if err != nil {
		return false
	}
	return a.HasNext()
}

func (a *AbstractPaginator) FetchNextPage(ctx context.Context, currentPage IStaticPage) (IStaticPage, error) {
	return a.fetchNextFunc(ctx, currentPage)
}

func (a *AbstractPaginator) fetchNextPage() (err error) {
	currentPage, err := a.FetchCurrentPage()
	if err != nil {
		return
	}
	if !currentPage.HasNext() {
		return
	}
	newPage, err := a.FetchNextPage(a.ctx, currentPage)
	if err != nil {
		return
	}
	err = a.setCurrentPage(newPage)
	return
}

func (a *AbstractPaginator) GetNext() (item interface{}, err error) {
	err = parallelisation.DetermineContextError(a.ctx)
	if err != nil {
		return
	}
	if !a.HasNext() {
		err = commonerrors.New(commonerrors.ErrNotFound, "there is not any next item")
		return
	}
	currentIt, err := a.FetchCurrentPageIterator()
	if err != nil {
		return
	}
	item, err = currentIt.GetNext()
	return
}

func (a *AbstractPaginator) Stop() context.CancelFunc {
	return a.cancellationStore.Cancel
}

func (a *AbstractPaginator) setCurrentPage(page IStaticPage) (err error) {
	a.currentPage = page
	a.currentPageIterator = nil
	if page != nil {
		a.currentPageIterator, err = page.GetItemIterator()
	}
	return
}

func (a *AbstractPaginator) SetCurrentPage(page IStaticPage) (err error) {
	if page == nil {
		err = commonerrors.New(commonerrors.ErrUndefined, "missing page")
		return
	}
	err = a.setCurrentPage(page)
	return
}

func (a *AbstractPaginator) FetchCurrentPage() (page IStaticPage, err error) {
	page = a.currentPage
	if page == nil {
		err = commonerrors.New(commonerrors.ErrUndefined, "missing page")
	}
	return
}

func (a *AbstractPaginator) FetchCurrentPageIterator() (it IIterator, err error) {
	it = a.currentPageIterator
	if it == nil {
		err = commonerrors.New(commonerrors.ErrUndefined, "missing page iterator")
	}
	return
}

func (a *AbstractPaginator) GetContext() context.Context {
	return a.ctx
}

func NewAbstractPaginator(ctx context.Context, firstPage IStaticPage, fetchNextFunc func(context.Context, IStaticPage) (IStaticPage, error)) (p *AbstractPaginator, err error) {
	store := parallelisation.NewCancelFunctionsStore()
	cancelCtx, cancel := context.WithCancel(ctx)

	store.RegisterCancelFunction(cancel)
	p = &AbstractPaginator{
		cancellationStore: store,
		fetchNextFunc:     fetchNextFunc,
		ctx:               cancelCtx,
	}
	err = p.setCurrentPage(firstPage)
	return
}

// DynamicPagePaginator defines a paginator over dynamic pages i.e. pages from which it is possible to access the next one.
type DynamicPagePaginator struct {
	AbstractPaginator
}

func (d *DynamicPagePaginator) GetCurrentPage() (dynamicPage IPage, err error) {
	p, err := d.FetchCurrentPage()
	if err != nil {
		return
	}
	dynamicPage, err = toDynamicPage(p)
	return
}

func toDynamicPage(page IStaticPage) (dynamicPage IPage, err error) {
	if page == nil {
		return
	}
	dynamicPage, ok := page.(IPage)
	if !ok {
		err = commonerrors.New(commonerrors.ErrInvalid, "current page is not dynamic i.e. it is not possible to fetch next pages from it")
	}
	return
}

func newDynamicPagePaginator(ctx context.Context, firstPage IPage) (p IPaginator, err error) {
	a, err := NewAbstractPaginator(ctx, firstPage, func(fCtx context.Context, current IStaticPage) (nextPage IStaticPage, err error) {
		p, err := toDynamicPage(current)
		if err != nil {
			return
		}
		nextPage, err = p.GetNext(fCtx)
		return
	})
	if err != nil {
		return
	}
	p = &DynamicPagePaginator{
		AbstractPaginator: *a,
	}
	return
}

// StaticPagePaginator defines a paginator over static pages i.e. pages from which it is not possible to access the next one and the paginator must define how to fetch it.
type StaticPagePaginator struct {
	AbstractPaginator
}

func (d *StaticPagePaginator) GetCurrentPage() (IStaticPage, error) {
	return d.FetchCurrentPage()
}

func newStaticPagePaginator(ctx context.Context, firstPage IStaticPage, fetchNextPageFunc func(context.Context, IStaticPage) (IStaticPage, error)) (p IPaginatorAndPageFetcher, er error) {
	a, err := NewAbstractPaginator(ctx, firstPage, fetchNextPageFunc)
	if err != nil {
		return
	}
	p = &StaticPagePaginator{
		AbstractPaginator: *a,
	}
	return
}

// NewCollectionPaginator creates a paginator over a collection of dynamic pages.
func NewCollectionPaginator(ctx context.Context, fetchFirstPageFunc func(context.Context) (IPage, error)) (paginator IPaginator, err error) {
	firstPage, err := fetchFirstPageFunc(ctx)
	if err != nil {
		return
	}
	paginator, err = newDynamicPagePaginator(ctx, firstPage)
	return
}

// NewStaticPagePaginator creates a paginator over a collection but only dealing with static pages i.e. pages from which it is not possible to access the next one and the paginator must define how to fetch them using the fetchNextPageFunc function.
func NewStaticPagePaginator(ctx context.Context, fetchFirstPageFunc func(context.Context) (IStaticPage, error), fetchNextPageFunc func(context.Context, IStaticPage) (IStaticPage, error)) (paginator IPaginatorAndPageFetcher, err error) {
	firstPage, err := fetchFirstPageFunc(ctx)
	if err != nil {
		return
	}
	paginator, err = newStaticPagePaginator(ctx, firstPage, fetchNextPageFunc)
	return
}
