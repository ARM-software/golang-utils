package pagination

import (
	"context"
	"fmt"
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ARM-software/golang-utils/utils/safecast"
)

func TestGenerateEmptyPage(t *testing.T) {
	page := GenerateEmptyPage()
	require.NotNil(t, page)
	assert.False(t, page.HasNext())
	assert.False(t, page.HasFuture())
	count, err := page.GetItemCount()
	require.NoError(t, err)
	assert.Equal(t, int64(0), count)
}

func TestGenerateMockPage(t *testing.T) {
	r, err := faker.RandomInt(2, 50)
	require.NoError(t, err)
	for i := 0; i < r[0]; i++ {
		page, count, err := GenerateMockPage()
		require.NoError(t, err)
		t.Run(fmt.Sprintf("%d_items#%d", i, count), func(t *testing.T) {
			require.NotNil(t, page)
			intCount, err := page.GetItemCount()
			require.NoError(t, err)
			assert.Equal(t, count, intCount)
		})
	}
}

func TestGenerateMockCollection(t *testing.T) {
	r, err := faker.RandomInt(2, 50)
	require.NoError(t, err)
	for i := 0; i < r[0]; i++ {
		firstPage, count, err := GenerateMockCollection()
		require.NoError(t, err)
		t.Run(fmt.Sprintf("%d_items#%d", i, count), func(t *testing.T) {
			require.NotNil(t, firstPage)
			cCount, err := firstPage.GetItemCount()
			require.NoError(t, err)
			assert.True(t, cCount <= count)
			size := safecast.ToInt64(cCount)
			page := firstPage.(IPage)
			for {
				if !page.HasNext() {
					break
				}
				page, err = page.GetNext(context.Background())
				require.NoError(t, err)
				cCount, err := page.GetItemCount()
				require.NoError(t, err)
				size += cCount
			}
			assert.Equal(t, count, size)
		})
	}
}

func TestGenerateMockStream(t *testing.T) {
	r, err := faker.RandomInt(2, 50)
	require.NoError(t, err)
	for i := 0; i < r[0]; i++ {
		firstPage, count, err := GenerateMockStream()
		require.NoError(t, err)
		t.Run(fmt.Sprintf("%d_items#%d", i, count), func(t *testing.T) {
			require.NotNil(t, firstPage)
			cCount, err := firstPage.GetItemCount()
			require.NoError(t, err)
			assert.True(t, cCount <= count)
			size := safecast.ToInt64(cCount)
			page := firstPage.(IStream)
			for {
				if !page.HasNext() && !page.HasFuture() {
					break
				}
				if page.HasNext() {
					nextPage, err := page.GetNext(context.Background())
					require.NoError(t, err)
					page = nextPage.(IStream)
				} else {
					page, err = page.GetFuture(context.Background())
					require.NoError(t, err)
				}

				cCount, err := page.GetItemCount()
				require.NoError(t, err)
				size += cCount
			}
			assert.Equal(t, count, size)
		})
	}
}

func TestNewMockPageIterator(t *testing.T) {
	type args struct {
		page *MockPage
	}
	tests := []struct {
		name    string
		args    args
		want    IIterator
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewMockPageIterator(tt.args.page)
			if !tt.wantErr(t, err, fmt.Sprintf("NewMockPageIterator(%v)", tt.args.page)) {
				return
			}
			assert.Equalf(t, tt.want, got, "NewMockPageIterator(%v)", tt.args.page)
		})
	}
}
