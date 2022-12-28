/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package pagination

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/bxcodec/faker/v3"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/parallelisation"
)

type MockItem struct {
	Index  int
	ID     string
	Value1 string
	Value2 string
}

func GenerateMockItem() *MockItem {
	return &MockItem{
		ID:     faker.UUIDHyphenated(),
		Value1: faker.Name(),
		Value2: faker.Sentence(),
	}
}

type MockPageIterator struct {
	elements     []MockItem
	currentIndex int
}

func (m *MockPageIterator) HasNext() bool {
	return m.currentIndex < len(m.elements)
}

func (m *MockPageIterator) GetNext() (item *interface{}, err error) {
	if m.currentIndex < 0 {
		err = fmt.Errorf("%w: incorrect element index", commonerrors.ErrInvalid)
		return
	}
	if !m.HasNext() {
		err = fmt.Errorf("%w: there is no more items", commonerrors.ErrNotFound)
		return
	}
	element := interface{}(m.elements[m.currentIndex])
	item = &element
	m.currentIndex++
	return
}

func NewMockPageIterator(page *MockPage) (IIterator, error) {
	if page == nil {
		return nil, commonerrors.ErrUndefined
	}
	return &MockPageIterator{
		elements:     page.elements,
		currentIndex: 0,
	}, nil
}

type MockPage struct {
	elements   []MockItem
	nextPage   IStream
	futurePage IStream
}

func (m *MockPage) HasNext() bool {
	return m.nextPage != nil
}

func (m *MockPage) GetNext(ctx context.Context) (page IPage, err error) {
	err = parallelisation.DetermineContextError(ctx)
	if err != nil {
		return
	}
	if !m.HasNext() {
		err = fmt.Errorf("%w: there is no more pages", commonerrors.ErrNotFound)
	}
	page = m.nextPage
	return
}

func (m *MockPage) GetItemIterator() (IIterator, error) {
	return NewMockPageIterator(m)
}

func (m *MockPage) AppendItem(i *MockItem) error {
	if i == nil {
		return commonerrors.ErrUndefined
	}
	m.elements = append(m.elements, *i)
	return nil
}

func (m *MockPage) SetNext(next IStream) error {
	if next == nil {
		return commonerrors.ErrUndefined
	}
	m.nextPage = next
	return nil
}

func (m *MockPage) SetFuture(future IStream) error {
	if future == nil {
		return commonerrors.ErrUndefined
	}
	m.futurePage = future
	return nil
}

func (m *MockPage) SetIndexes(firstIndex int) {
	for i := 0; i < len(m.elements); i++ {
		m.elements[i].Index = i + firstIndex
	}
	if m.nextPage != nil {
		nPage := m.nextPage.(*MockPage)
		nPage.SetIndexes(firstIndex + len(m.elements))
	}
	if m.futurePage != nil {
		fPage := m.futurePage.(*MockPage)
		fPage.SetIndexes(firstIndex + len(m.elements))
	}
}

func (m *MockPage) GetItemCount() (int64, error) {
	return int64(len(m.elements)), nil
}

func (m *MockPage) HasFuture() bool {
	return m.futurePage != nil
}

func (m *MockPage) GetFuture(ctx context.Context) (future IStream, err error) {
	err = parallelisation.DetermineContextError(ctx)
	if err != nil {
		return
	}
	if !m.HasFuture() {
		err = fmt.Errorf("%w: there is no future page", commonerrors.ErrNotFound)
	}
	future = m.futurePage
	return
}

func GenerateEmptyPage() IStream {
	return &MockPage{}
}

func GenerateMockPage() (IStream, error) {
	page := GenerateEmptyPage().(*MockPage)
	n := rand.Intn(50) //nolint:gosec //causes G404: Use of weak random number generator (math/rand instead of crypto/rand) (gosec), So disable gosec as this is just for testing
	for i := 0; i < n; i++ {
		err := page.AppendItem(GenerateMockItem())
		if err != nil {
			return nil, err
		}
	}
	return page, nil
}

func GenerateMockCollection() (firstPage IStream, itemTotal int64, err error) {
	rand.Seed(int64(time.Now().Nanosecond())) //nolint:gosec //causes G404: Use of weak random number generator (math/rand instead of crypto/rand) (gosec), So disable gosec as this is just for testing
	n := rand.Intn(50)                        //nolint:gosec //causes G404: Use of weak random number generator (math/rand instead of crypto/rand) (gosec), So disable gosec as this is just for testing
	var next IStream
	for i := 0; i < n; i++ {
		currentPage, subErr := GenerateMockPage()
		if subErr != nil {
			err = subErr
			return
		}
		currentCount, subErr := currentPage.GetItemCount()
		if subErr != nil {
			err = subErr
			return
		}
		itemTotal += currentCount

		if next != nil {
			mockP := currentPage.(*MockPage)
			subErr = mockP.SetNext(next)
			if subErr != nil {
				err = subErr
				return
			}

		}
		firstPage = currentPage
		next = firstPage
	}
	if firstPage == nil {
		return
	}
	mockP := firstPage.(*MockPage)
	mockP.SetIndexes(0)
	return
}

func GenerateMockStream() (firstPage IStream, itemTotal int64, err error) {
	rand.Seed(int64(time.Now().Nanosecond())) //nolint:gosec //causes G404: Use of weak random number generator (math/rand instead of crypto/rand) (gosec), So disable gosec as this is just for testing
	n := rand.Intn(50)                        //nolint:gosec //causes G404: Use of weak random number generator (math/rand instead of crypto/rand) (gosec), So disable gosec as this is just for testing
	var future IStream
	for i := 0; i < n; i++ {
		currentPage, subErr := GenerateMockPage()
		if subErr != nil {
			err = subErr
			return
		}
		currentCount, subErr := currentPage.GetItemCount()
		if subErr != nil {
			err = subErr
			return
		}
		itemTotal += currentCount

		mockP := currentPage.(*MockPage)
		if future == nil {
			subErr = mockP.SetFuture(GenerateEmptyPage())
		} else {
			subErr = mockP.SetFuture(future)
		}
		if subErr != nil {
			err = subErr
			return
		}

		firstPage = currentPage
		future = firstPage
	}
	if firstPage == nil {
		return
	}
	mockP := firstPage.(*MockPage)
	mockP.SetIndexes(0)
	return
}
