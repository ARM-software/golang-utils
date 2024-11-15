package pagination

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/commonerrors/errortest"
)

func TestStreamPaginator(t *testing.T) {
	tests := []struct {
		paginator    func(context.Context, IStaticPageStream) (IGenericStreamPaginator, error)
		name         string
		generateFunc func() (firstPage IStream, itemTotal int64, err error)
		dryOut       bool
	}{
		{
			paginator: func(ctx context.Context, collection IStaticPageStream) (IGenericStreamPaginator, error) {
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
				return paginator, err
			},
			generateFunc: GenerateMockStreamWithEnding,
			name:         "stream paginator over a stream of static pages with known ending",
		},
		{
			paginator: func(ctx context.Context, collection IStaticPageStream) (IGenericStreamPaginator, error) {
				paginator, err := NewStreamPaginator(ctx, time.Second, 10*time.Millisecond, func(context.Context) (IStream, error) {
					return toDynamicStream(collection)
				})
				return paginator, err
			},
			generateFunc: GenerateMockStreamWithEnding,
			name:         "stream paginator over a stream of dynamic pages but with a known ending",
		},
		{
			paginator: func(ctx context.Context, collection IStaticPageStream) (IGenericStreamPaginator, error) {
				paginator, err := NewStreamPaginator(ctx, time.Second, 10*time.Millisecond, func(context.Context) (IStream, error) {
					return toDynamicStream(collection)
				})
				return paginator, err
			},
			name:         "stream paginator over a running dry stream of dynamic pages",
			generateFunc: GenerateMockStream,
			dryOut:       true,
		},
		{
			paginator: func(ctx context.Context, collection IStaticPageStream) (IGenericStreamPaginator, error) {
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
			name:         "stream paginator over a running dry stream of static pages",
			generateFunc: GenerateMockStream,
			dryOut:       true,
		},
	}

	for te := range tests {
		test := tests[te]
		for i := 0; i < 10; i++ {
			mockPages, expectedCount, err := test.generateFunc()
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
					mockItem, ok := item.(*MockItem)
					require.True(t, ok)
					assert.Equal(t, int(count-1), mockItem.Index)
					if count >= expectedCount%2 {
						require.NoError(t, paginator.DryUp())
					}
				}
				assert.Equal(t, expectedCount, count)
			})
		}
	}
}

func TestEmptyStream(t *testing.T) {
	mockPages, expectedCount, err := GenerateMockEmptyStream()
	require.NoError(t, err)
	require.Zero(t, expectedCount)
	require.NotNil(t, mockPages)
	paginator, err := NewStreamPaginator(context.Background(), time.Second, 10*time.Millisecond, func(context.Context) (IStream, error) {
		return toDynamicStream(mockPages)
	})
	require.NoError(t, err)
	assert.False(t, paginator.HasNext())
	assert.False(t, paginator.IsRunningDry())
	item, err := paginator.GetNext()
	errortest.AssertError(t, err, commonerrors.ErrNotFound)
	assert.Nil(t, item)
}

func TestDryOutStream(t *testing.T) {
	mockPages, expectedCount, err := GenerateMockEmptyStream()
	require.NoError(t, err)
	require.Zero(t, expectedCount)
	require.NotNil(t, mockPages)
	paginator, err := NewStreamPaginator(context.Background(), time.Millisecond, 10*time.Millisecond, func(context.Context) (IStream, error) {
		return toDynamicStream(mockPages)
	})
	require.NoError(t, err)
	require.NoError(t, paginator.DryUp())
	assert.False(t, paginator.HasNext())
	assert.True(t, paginator.IsRunningDry())
}
