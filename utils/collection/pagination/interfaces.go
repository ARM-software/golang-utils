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

//go:generate go tool mockgen -destination=../../mocks/mock_$GOPACKAGE.go -package=mocks github.com/ARM-software/golang-utils/utils/collection/$GOPACKAGE IStaticPage,IPage,IStaticPageStream,IStream,IIterator,IPaginator,IPaginatorAndPageFetcher,IStreamPaginator,IStreamPaginatorAndPageFetcher

// IIterator defines an iterator over a collection of items.
type IIterator interface {
	// HasNext returns whether there are more items available or not.
	HasNext() bool
	// GetNext returns the next item.
	GetNext() (interface{}, error)
}

// IStaticPage defines a generic page for a collection. A page is marked as static when it cannot retrieve next pages on its own.
type IStaticPage interface {
	// HasNext states whether more pages are accessible.
	HasNext() bool
	// GetItemIterator returns a new iterator over the page's items.
	GetItemIterator() (IIterator, error)
	// GetItemCount returns the number of items in this page
	GetItemCount() (int64, error)
}

// IPage defines a page with the ability to access next pages.
type IPage interface {
	IStaticPage
	// GetNext returns the next page.
	GetNext(ctx context.Context) (IPage, error)
}

// IStaticPageStream defines a page for a collection which does not have any known ending.
type IStaticPageStream interface {
	IStaticPage
	// HasFuture states whether there may be future items.
	HasFuture() bool
}

// IStream defines a stream with the ability to access future pages.
type IStream interface {
	IPage
	IStaticPageStream
	// GetFuture returns the future page.
	GetFuture(ctx context.Context) (IStream, error)
}

// IGenericPaginator defines a generic paginator.
type IGenericPaginator interface {
	io.Closer
	IIterator
	// Stop returns a stop function which stops the iteration.
	Stop() context.CancelFunc
}

// IPaginator is an iterator over multiple pages
type IPaginator interface {
	IGenericPaginator
	// GetCurrentPage returns the current page.
	GetCurrentPage() (IPage, error)
}

// IPaginatorAndPageFetcher is a paginator dealing with static pages
type IPaginatorAndPageFetcher interface {
	IGenericPaginator
	// GetCurrentPage returns the current page.
	GetCurrentPage() (IStaticPage, error)
	// FetchNextPage fetches the next page.
	FetchNextPage(ctx context.Context, currentPage IStaticPage) (IStaticPage, error)
}

// IGenericStreamPaginator is an iterator over a stream. A stream is a collection without any known ending.
type IGenericStreamPaginator interface {
	IGenericPaginator
	// DryUp indicates to the stream that it will soon run out.
	DryUp() error
	// IsRunningDry indicates whether the stream is about to run out.
	IsRunningDry() bool
}

// IStreamPaginator is stream paginator over dynamic pages.
type IStreamPaginator interface {
	IGenericStreamPaginator
	IPaginator
}

// IStreamPaginatorAndPageFetcher is a stream paginator over static pages.
type IStreamPaginatorAndPageFetcher interface {
	IGenericStreamPaginator
	IPaginatorAndPageFetcher
	// FetchFuturePage returns the future page.
	FetchFuturePage(ctx context.Context, currentPage IStaticPageStream) (IStaticPageStream, error)
}
