/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package pagination

import (
	"context"
	"fmt"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/parallelisation"
)

type AbstractPaginator struct {
	currentPage       IPage
	cancellationStore *parallelisation.CancelFunctionStore
	ctx               context.Context
}

func (a *AbstractPaginator) Close() error {
	a.Stop()()
	return nil
}

func (a *AbstractPaginator) HasNext() bool {
	if parallelisation.DetermineContextError(a.ctx) != nil {
		return false
	}
	currentPage, err := a.GetCurrentPage()
	if err != nil {
		return false
	}
	currentIt, err := currentPage.GetItemIterator()
	if err != nil {
		return false
	}
	if currentIt.HasNext() {
		return true
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

func (a *AbstractPaginator) fetchNextPage() (err error) {
	currentPage, err := a.GetCurrentPage()
	if err != nil {
		return
	}
	if !currentPage.HasNext() {
		return
	}
	newPage, err := currentPage.GetNext(a.ctx)
	if err != nil {
		return
	}
	a.currentPage = newPage
	return
}

func (a *AbstractPaginator) GetNext() (item *interface{}, err error) {
	err = parallelisation.DetermineContextError(a.ctx)
	if err != nil {
		return
	}
	if !a.HasNext() {
		err = fmt.Errorf("%w: there is not any next item", commonerrors.ErrNotFound)
		return
	}
	currentPage, err := a.GetCurrentPage()
	if err != nil {
		return
	}
	currentIt, err := currentPage.GetItemIterator()
	if err != nil {
		return
	}
	item, err = currentIt.GetNext()
	return
}

func (a *AbstractPaginator) Stop() context.CancelFunc {
	return a.cancellationStore.Cancel
}

func (a *AbstractPaginator) SetCurrentPage(page IPage) (err error) {
	if page == nil {
		err = fmt.Errorf("%w: missing page", commonerrors.ErrUndefined)
		return
	}
	a.currentPage = page
	return
}

func (a *AbstractPaginator) GetCurrentPage() (page IPage, err error) {
	page = a.currentPage
	if page == nil {
		err = fmt.Errorf("%w: missing page", commonerrors.ErrUndefined)
	}
	return
}

func (a *AbstractPaginator) GetContext() context.Context {
	return a.ctx
}

func NewAbstractPaginator(ctx context.Context, firstPage IPage) IPaginator {
	store := parallelisation.NewCancelFunctionsStore()
	cancelCtx, cancel := context.WithCancel(ctx)

	store.RegisterCancelFunction(cancel)
	return &AbstractPaginator{
		currentPage:       firstPage,
		cancellationStore: store,
		ctx:               cancelCtx,
	}
}

// NewCollectionPaginator creates a paginator over a collection.
func NewCollectionPaginator(ctx context.Context, fetchFirstPage func(context.Context) (IPage, error)) (paginator IPaginator, err error) {
	firstPage, err := fetchFirstPage(ctx)
	if err != nil {
		return
	}
	paginator = NewAbstractPaginator(ctx, firstPage)
	return
}
