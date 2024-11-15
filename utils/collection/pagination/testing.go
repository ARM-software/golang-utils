/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package pagination

import (
	"context"
	"fmt"

	"github.com/go-faker/faker/v4"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/parallelisation"
	"github.com/ARM-software/golang-utils/utils/safecast"
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

func (m *MockPageIterator) GetNext() (item interface{}, err error) {
	if m.currentIndex < 0 {
		err = fmt.Errorf("%w: incorrect element index", commonerrors.ErrInvalid)
		return
	}
	if !m.HasNext() {
		err = fmt.Errorf("%w: there is no more items", commonerrors.ErrNotFound)
		return
	}
	element := m.elements[m.currentIndex]
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
	return safecast.ToInt64(len(m.elements)), nil
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

func GenerateMockPage() (IStream, int64, error) {
	randoms, err := faker.RandomInt(0, 50)
	if err != nil {
		return nil, 0, err
	}
	n := randoms[2]
	page := GenerateEmptyPage().(*MockPage)
	for i := 0; i < n; i++ {
		subErr := page.AppendItem(GenerateMockItem())
		if subErr != nil {
			return nil, 0, subErr
		}
	}
	return page, safecast.ToInt64(n), nil
}

func GenerateMockCollection() (firstPage IStream, itemTotal int64, err error) {
	randoms, err := faker.RandomInt(0, 50)
	if err != nil {
		return
	}
	n := randoms[1]
	var next IStream
	for i := 0; i < n; i++ {
		currentPage, _, subErr := GenerateMockPage()
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
		firstPage = GenerateEmptyPage()
	}
	mockP := firstPage.(*MockPage)
	mockP.SetIndexes(0)
	return
}

// GenerateMockStream creates a mock stream which could never end (as in, a future link will be always present)
func GenerateMockStream() (firstPage IStream, itemTotal int64, err error) {
	randoms, err := faker.RandomInt(1, 3)
	if err != nil {
		return
	}
	n := randoms[0]
	var future IStream
	for i := 0; i < n; i++ {
		currentPage, _, subErr := GenerateMockPage()
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
		firstPage = GenerateEmptyPage()
	}
	mockP := firstPage.(*MockPage)
	mockP.SetIndexes(0)
	return
}

// GenerateMockStreamWithEnding generates a stream which will end itself (as in the future link will disappear).
func GenerateMockStreamWithEnding() (firstPage IStream, itemTotal int64, err error) {
	randoms, err := faker.RandomInt(1, 50)
	if err != nil {
		return
	}
	n := randoms[0]
	var future IStream
	for i := 0; i < n; i++ {
		currentPage, _, subErr := GenerateMockPage()
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
			subErr = mockP.SetNext(GenerateEmptyPage())
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
		firstPage = GenerateEmptyPage()
	}
	mockP := firstPage.(*MockPage)
	mockP.SetIndexes(0)
	return
}

// GenerateMockEmptyStream generates an empty stream (as in stream of pages with no element).
func GenerateMockEmptyStream() (firstPage IStream, itemTotal int64, err error) {
	randoms, err := faker.RandomInt(1, 50)
	if err != nil {
		return
	}
	n := randoms[0]
	var future IStream
	for i := 0; i < n; i++ {
		currentPage := GenerateEmptyPage()
		currentCount, subErr := currentPage.GetItemCount()
		if subErr != nil {
			err = subErr
			return
		}
		itemTotal += currentCount

		mockP := currentPage.(*MockPage)
		if future == nil {
			subErr = mockP.SetNext(GenerateEmptyPage())
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
		firstPage = GenerateEmptyPage()
	}
	mockP := firstPage.(*MockPage)
	mockP.SetIndexes(0)
	return
}
