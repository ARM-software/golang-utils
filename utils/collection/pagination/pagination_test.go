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
		paginator func(context.Context, IStaticPageStream) (IGenericPaginator, error)
		name      string
		useStream bool
	}{
		{
			paginator: func(ctx context.Context, collection IStaticPageStream) (IGenericPaginator, error) {
				return NewAbstractPaginator(ctx, collection, func(fCtx context.Context, current IStaticPage) (IStaticPage, error) {
					c, err := toDynamicPage(current)
					if err != nil {
						return nil, err
					}
					return c.GetNext(fCtx)
				})
			},
			name: "Abstract paginator",
		},
		{
			paginator: func(ctx context.Context, collection IStaticPageStream) (IGenericPaginator, error) {
				return NewStaticPagePaginator(ctx, func(context.Context) (IStaticPage, error) {
					return collection, nil
				}, func(fCtx context.Context, current IStaticPage) (IStaticPage, error) {
					c, err := toDynamicPage(current)
					if err != nil {
						return nil, err
					}
					return c.GetNext(fCtx)
				})
			},
			name: "Static page paginator",
		},
		{
			paginator: func(ctx context.Context, collection IStaticPageStream) (IGenericPaginator, error) {
				return NewCollectionPaginator(ctx, func(context.Context) (IPage, error) {
					return toDynamicPage(collection)
				})
			},
			name: "paginator over a collection of dynamic pages",
		},
		{
			paginator: func(ctx context.Context, collection IStaticPageStream) (IGenericPaginator, error) {
				return NewStaticPageStreamPaginator(ctx, time.Second, 10*time.Millisecond, func(context.Context) (IStaticPageStream, error) {
					return collection, nil
				}, func(fCtx context.Context, current IStaticPage) (IStaticPage, error) {
					c, err := toDynamicPage(current)
					if err != nil {
						return nil, err
					}
					return c.GetNext(fCtx)
				}, func(fCtx context.Context, current IStaticPageStream) (IStaticPageStream, error) {
					s, err := toDynamicStream(current)
					if err != nil {
						return nil, err
					}
					return s.GetFuture(fCtx)
				})
			},
			name: "stream paginator over a collection of static pages",
		},
		{
			paginator: func(ctx context.Context, collection IStaticPageStream) (IGenericPaginator, error) {
				return NewStreamPaginator(ctx, time.Second, 10*time.Millisecond, func(context.Context) (IStream, error) {
					return toDynamicStream(collection)
				})
			},
			name: "stream paginator over a collection of dynamic pages",
		},
		{
			paginator: func(ctx context.Context, collection IStaticPageStream) (IGenericPaginator, error) {
				paginator, err := NewStaticPageStreamPaginator(ctx, time.Second, 10*time.Millisecond, func(context.Context) (IStaticPageStream, error) {
					return collection, nil
				}, func(fCtx context.Context, current IStaticPage) (IStaticPage, error) {
					c, err := toDynamicPage(current)
					if err != nil {
						return nil, err
					}
					return c.GetNext(fCtx)
				}, func(fCtx context.Context, current IStaticPageStream) (IStaticPageStream, error) {
					s, err := toDynamicStream(current)
					if err != nil {
						return nil, err
					}
					return s.GetFuture(fCtx)
				})
				if paginator != nil {
					// Indicate the stream will run out.
					err = paginator.DryUp()
				}
				return paginator, err
			},
			name:      "stream paginator over a running dry stream of static pages",
			useStream: true,
		},
		{
			paginator: func(ctx context.Context, collection IStaticPageStream) (IGenericPaginator, error) {
				paginator, err := NewStreamPaginator(ctx, 50*time.Millisecond, 10*time.Millisecond, func(context.Context) (IStream, error) {
					return toDynamicStream(collection)
				})
				if paginator != nil {
					// Indicate the stream will run out.
					err = paginator.DryUp()
				}
				return paginator, err
			},
			name:      "stream paginator over a running dry stream of dynamic pages",
			useStream: true,
		},
	}

	for te := range tests {
		test := tests[te]
		for i := 0; i < 50; i++ {
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
					if !paginator.HasNext() { //nolint:staticcheck
						break
					}
					count += 1
					item, err := paginator.GetNext()
					require.NoError(t, err)
					require.NotNil(t, item)
					mockItem, ok := item.(*MockItem)
					require.True(t, ok)
					assert.Equal(t, int(count-1), mockItem.Index)
				}
				assert.Equal(t, expectedCount, count)
			})
		}
	}
}

// TestPaginator_InitialisationError tests whether errors are correctly handled if API returns some error
func TestPaginator_InitialisationError(t *testing.T) {
	tests := []struct {
		paginator                 func(context.Context, IStaticPageStream) (IGenericPaginator, error)
		name                      string
		expectInitialisationError bool
	}{
		{
			paginator: func(ctx context.Context, collection IStaticPageStream) (IGenericPaginator, error) {
				return NewAbstractPaginator(ctx, collection, func(fCtx context.Context, current IStaticPage) (IStaticPage, error) {
					return nil, commonerrors.ErrUnexpected
				})
			},
			name:                      "Abstract paginator",
			expectInitialisationError: false,
		},
		{
			paginator: func(ctx context.Context, collection IStaticPageStream) (IGenericPaginator, error) {
				return NewStaticPagePaginator(ctx, func(context.Context) (IStaticPage, error) {
					return nil, commonerrors.ErrUnexpected
				}, func(fCtx context.Context, current IStaticPage) (IStaticPage, error) {
					return nil, commonerrors.ErrUnexpected
				})
			},
			name:                      "Static page paginator",
			expectInitialisationError: true,
		},
		{
			paginator: func(ctx context.Context, collection IStaticPageStream) (IGenericPaginator, error) {
				return NewCollectionPaginator(ctx, func(context.Context) (IPage, error) {
					return nil, commonerrors.ErrUnexpected
				})
			},
			name:                      "paginator over a collection of dynamic pages",
			expectInitialisationError: true,
		},
		{
			paginator: func(ctx context.Context, collection IStaticPageStream) (IGenericPaginator, error) {
				return NewStaticPageStreamPaginator(ctx, time.Second, 10*time.Millisecond, func(context.Context) (IStaticPageStream, error) {
					return nil, commonerrors.ErrUnexpected
				}, func(fCtx context.Context, current IStaticPage) (IStaticPage, error) {
					return nil, commonerrors.ErrUnexpected
				}, func(fCtx context.Context, current IStaticPageStream) (IStaticPageStream, error) {
					return nil, commonerrors.ErrUnexpected
				})
			},
			name:                      "stream paginator over a collection of static pages",
			expectInitialisationError: true,
		},
		{
			paginator: func(ctx context.Context, collection IStaticPageStream) (IGenericPaginator, error) {
				return NewStreamPaginator(ctx, time.Second, 10*time.Millisecond, func(context.Context) (IStream, error) {
					return nil, commonerrors.ErrUnexpected
				})
			},
			name:                      "stream paginator over a collection of dynamic pages",
			expectInitialisationError: true,
		},
		{
			paginator: func(ctx context.Context, collection IStaticPageStream) (IGenericPaginator, error) {
				paginator, err := NewStaticPageStreamPaginator(ctx, time.Second, 10*time.Millisecond, func(context.Context) (IStaticPageStream, error) {
					return nil, commonerrors.ErrUnexpected
				}, func(fCtx context.Context, current IStaticPage) (IStaticPage, error) {
					return nil, commonerrors.ErrUnexpected
				}, func(fCtx context.Context, current IStaticPageStream) (IStaticPageStream, error) {
					return nil, commonerrors.ErrUnexpected
				})
				if paginator != nil {
					// Indicate the stream will run out.
					err = paginator.DryUp()
				}
				return paginator, err
			},
			name:                      "stream paginator over a running dry stream of static pages",
			expectInitialisationError: true,
		},
		{
			paginator: func(ctx context.Context, collection IStaticPageStream) (IGenericPaginator, error) {
				paginator, err := NewStreamPaginator(ctx, 50*time.Millisecond, 10*time.Millisecond, func(context.Context) (IStream, error) {
					return nil, commonerrors.ErrUnexpected
				})
				if paginator != nil {
					// Indicate the stream will run out.
					err = paginator.DryUp()
				}
				return paginator, err
			},
			name:                      "stream paginator over a running dry stream of dynamic pages",
			expectInitialisationError: true,
		},
	}

	for te := range tests {
		test := tests[te]
		for i := 0; i < 50; i++ {
			var mockPages IStream
			t.Run(fmt.Sprintf("%v-#%v", test.name, i), func(t *testing.T) {
				paginator, err := test.paginator(context.TODO(), mockPages)
				if test.expectInitialisationError {
					assert.Error(t, err)
					assert.Nil(t, paginator)
				} else {
					assert.NoError(t, err)
					assert.NotNil(t, paginator)
					assert.False(t, paginator.HasNext())
					require.NoError(t, paginator.Close())
				}

			})
		}
	}
}

func TestPaginator_stop(t *testing.T) {
	tests := []struct {
		paginator func(context.Context, IStaticPageStream) (IGenericPaginator, error)
		name      string
		useStream bool
	}{
		{
			paginator: func(ctx context.Context, collection IStaticPageStream) (IGenericPaginator, error) {
				return NewAbstractPaginator(ctx, collection, func(fCtx context.Context, current IStaticPage) (IStaticPage, error) {
					c, err := toDynamicPage(current)
					if err != nil {
						return nil, err
					}
					return c.GetNext(fCtx)
				})
			},
			name: "Abstract paginator",
		},
		{
			paginator: func(ctx context.Context, collection IStaticPageStream) (IGenericPaginator, error) {
				return NewStaticPagePaginator(ctx, func(context.Context) (IStaticPage, error) {
					return collection, nil
				}, func(fCtx context.Context, current IStaticPage) (IStaticPage, error) {
					c, err := toDynamicPage(current)
					if err != nil {
						return nil, err
					}
					return c.GetNext(fCtx)
				})
			},
			name: "Static page paginator",
		},
		{
			paginator: func(ctx context.Context, collection IStaticPageStream) (IGenericPaginator, error) {
				return NewCollectionPaginator(ctx, func(context.Context) (IPage, error) {
					return toDynamicPage(collection)
				})
			},
			name: "paginator over a collection of dynamic pages",
		},
		{
			paginator: func(ctx context.Context, collection IStaticPageStream) (IGenericPaginator, error) {
				return NewStaticPageStreamPaginator(ctx, time.Second, 10*time.Millisecond, func(context.Context) (IStaticPageStream, error) {
					return collection, nil
				}, func(fCtx context.Context, current IStaticPage) (IStaticPage, error) {
					c, err := toDynamicPage(current)
					if err != nil {
						return nil, err
					}
					return c.GetNext(fCtx)
				}, func(fCtx context.Context, current IStaticPageStream) (IStaticPageStream, error) {
					s, err := toDynamicStream(current)
					if err != nil {
						return nil, err
					}
					return s.GetFuture(fCtx)
				})
			},
			name: "stream paginator over a collection of static pages",
		},
		{
			paginator: func(ctx context.Context, collection IStaticPageStream) (IGenericPaginator, error) {
				return NewStreamPaginator(ctx, time.Second, 10*time.Millisecond, func(context.Context) (IStream, error) {
					return toDynamicStream(collection)
				})
			},
			name: "stream paginator over a collection of dynamic pages",
		},
		{
			paginator: func(ctx context.Context, collection IStaticPageStream) (IGenericPaginator, error) {
				paginator, err := NewStaticPageStreamPaginator(ctx, time.Second, 10*time.Millisecond, func(context.Context) (IStaticPageStream, error) {
					return collection, nil
				}, func(fCtx context.Context, current IStaticPage) (IStaticPage, error) {
					c, err := toDynamicPage(current)
					if err != nil {
						return nil, err
					}
					return c.GetNext(fCtx)
				}, func(fCtx context.Context, current IStaticPageStream) (IStaticPageStream, error) {
					s, err := toDynamicStream(current)
					if err != nil {
						return nil, err
					}
					return s.GetFuture(fCtx)
				})
				if paginator != nil {
					// Indicate the stream will run out.
					err = paginator.DryUp()
				}
				return paginator, err
			},
			name:      "stream paginator over a running dry stream of static pages",
			useStream: true,
		},
		{
			paginator: func(ctx context.Context, collection IStaticPageStream) (IGenericPaginator, error) {
				paginator, err := NewStreamPaginator(ctx, 50*time.Millisecond, 10*time.Millisecond, func(context.Context) (IStream, error) {
					return toDynamicStream(collection)
				})
				if paginator != nil {
					// Indicate the stream will run out.
					err = paginator.DryUp()
				}
				return paginator, err
			},
			name:      "stream paginator over a running dry stream of dynamic pages",
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
