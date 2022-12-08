/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// Package pagination provides a ways to iterate over collections.
// For that purposes, it defines iterators and paginators, which act as an abstraction over the process of iterating over an entire result set of a truncated API operation returning pages.
package pagination

import (
	"context"
	"io"
)

//go:generate mockgen -destination=../../mocks/mock_$GOPACKAGE.go -package=mocks github.com/ARM-software/golang-utils/utils/collection/$GOPACKAGE IPage,IStream,IIterator,IPaginator,IStreamPaginator

// IIterator defines an iterator over a collection of items.
type IIterator interface {
	// HasNext returns whether there are more items available or not.
	HasNext() bool
	// GetNext returns the next item.
	GetNext() (*interface{}, error)
}

// IPage defines a generic page for a collection.
type IPage interface {
	// HasNext states whether more pages are accessible.
	HasNext() bool
	// GetNext returns the next page.
	GetNext(ctx context.Context) (IPage, error)
	// GetItemIterator returns an iterator over the page's items.
	GetItemIterator() (IIterator, error)
	// GetItemCount returns the number of items in this page
	GetItemCount() (int64, error)
}

// IStream defines a page for a collection which does not have any known ending.
type IStream interface {
	IPage
	// HasFuture states whether there may be future items.
	HasFuture() bool
	// GetFuture returns the future page.
	GetFuture(ctx context.Context) (IPage, error)
}

// IPaginator is an iterator over multiple pages
type IPaginator interface {
	io.Closer
	IIterator
	// Stop returns a stop function which stops the iteration.
	Stop() context.CancelFunc
	// GetCurrentPage returns the current page.
	GetCurrentPage() (IPage, error)
}

// IStreamPaginator is an iterator over a stream. A stream is a collection without any know ending.
type IStreamPaginator interface {
	IPaginator
	//DryUp indicates to the stream that it will soon run out.
	DryUp() error
	// IsRunningDry indicates whether the stream is about to run out.
	IsRunningDry() bool
}
