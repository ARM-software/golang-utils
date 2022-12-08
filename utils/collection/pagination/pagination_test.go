/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package pagination

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
)

func TestPaginator(t *testing.T) {
	tests := []struct {
		paginator func(context.Context, IStream) (IPaginator, error)
		name      string
		useStream bool
	}{
		{
			paginator: func(ctx context.Context, collection IStream) (IPaginator, error) {
				return NewAbstractPaginator(ctx, collection), nil
			},
			name: "Abstract paginator",
		},
		{
			paginator: func(ctx context.Context, collection IStream) (IPaginator, error) {
				return NewCollectionPaginator(ctx, func(context.Context) (IPage, error) {
					return collection, nil
				})
			},
			name: "paginator over a collection",
		},
		{
			paginator: func(ctx context.Context, collection IStream) (IPaginator, error) {
				return NewStreamPaginator(ctx, time.Second, 10*time.Millisecond, func(context.Context) (IStream, error) {
					return collection, nil
				})
			},
			name: "stream paginator over a collection",
		},
		{
			paginator: func(ctx context.Context, collection IStream) (IPaginator, error) {
				paginator, err := NewStreamPaginator(ctx, 50*time.Millisecond, 10*time.Millisecond, func(context.Context) (IStream, error) {
					return collection, nil
				})
				if paginator != nil {
					// Indicate the stream will run out.
					err = paginator.DryUp()
				}
				return paginator, err
			},
			name:      "stream paginator over a running dry stream",
			useStream: true,
		},
	}

	for te := range tests {
		test := tests[te]
		for i := 0; i < 10; i++ {
			var mockPages IStream
			var expectedCount int64
			var err error
			if test.useStream {
				mockPages, expectedCount, err = GenerateMockStream()
			} else {
				mockPages, expectedCount, err = GenerateMockCollection()
			}
			require.NoError(t, err)
			t.Run(fmt.Sprintf("%v-#%v-[%v items]", test.name, i, expectedCount), func(t *testing.T) {
				paginator, err := test.paginator(context.TODO(), mockPages)
				require.NoError(t, err)
				count := int64(0)
				for {
					if !paginator.HasNext() {
						break
					}
					count += 1
					item, err := paginator.GetNext()
					require.NoError(t, err)
					require.NotNil(t, item)
					mockItem, ok := (*item).(MockItem)
					require.True(t, ok)
					assert.Equal(t, int(count-1), mockItem.Index)
				}
				assert.Equal(t, expectedCount, count)
			})
		}
	}
}

func TestPaginator_stop(t *testing.T) {
	tests := []struct {
		paginator func(context.Context, IStream) (IPaginator, error)
		name      string
		useStream bool
	}{
		{
			paginator: func(ctx context.Context, collection IStream) (IPaginator, error) {
				return NewAbstractPaginator(ctx, collection), nil
			},
			name: "Abstract paginator",
		},
		{
			paginator: func(ctx context.Context, collection IStream) (IPaginator, error) {
				return NewCollectionPaginator(ctx, func(context.Context) (IPage, error) {
					return collection, nil
				})
			},
			name: "paginator over a collection",
		},
		{
			paginator: func(ctx context.Context, collection IStream) (IPaginator, error) {
				return NewStreamPaginator(ctx, time.Second, 10*time.Millisecond, func(context.Context) (IStream, error) {
					return collection, nil
				})
			},
			name: "stream paginator over a collection",
		},
		{
			paginator: func(ctx context.Context, collection IStream) (IPaginator, error) {
				paginator, err := NewStreamPaginator(ctx, 50*time.Millisecond, 10*time.Millisecond, func(context.Context) (IStream, error) {
					return collection, nil
				})
				if paginator != nil {
					// Indicate the stream will run out.
					err = paginator.DryUp()
				}
				return paginator, err
			},
			name:      "stream paginator over a running dry stream",
			useStream: true,
		},
	}

	for te := range tests {
		test := tests[te]
		var mockPages IStream
		var expectedCount int64
		var err error
		// Ensuring there are some items
		for {
			if test.useStream {
				mockPages, expectedCount, err = GenerateMockStream()
			} else {
				mockPages, expectedCount, err = GenerateMockCollection()
			}
			if expectedCount > 0 {
				break
			}
		}
		require.NoError(t, err)
		t.Run(fmt.Sprintf("%v", test.name), func(t *testing.T) {
			require.NoError(t, err)
			require.NotEmpty(t, expectedCount)
			paginator, err := test.paginator(context.TODO(), mockPages)
			require.NoError(t, err)
			paginator.Stop()()

			assert.False(t, paginator.HasNext())
			item, err := paginator.GetNext()
			assert.Error(t, err)
			assert.Empty(t, item)
			assert.True(t, commonerrors.Any(err, commonerrors.ErrTimeout, commonerrors.ErrCancelled))

			require.NoError(t, paginator.Close())
		})
	}
}
