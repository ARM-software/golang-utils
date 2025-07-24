/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package logs

import (
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/rs/zerolog/diode"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/parallelisation"
)

type MultipleWritersWithSource struct {
	mu           sync.RWMutex
	writers      []WriterWithSource
	closeWriters bool
	closerStore  *parallelisation.CloserStore
}

func (w *MultipleWritersWithSource) GetWriters() ([]WriterWithSource, error) {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.writers, nil
}

func (w *MultipleWritersWithSource) AddWriters(writers ...WriterWithSource) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.writers = append(w.writers, writers...)
	if w.closeWriters {
		for i := range writers {
			w.closerStore.RegisterCloser(writers[i])
		}
	}
	return nil
}
func (w *MultipleWritersWithSource) Write(p []byte) (n int, err error) {
	writers, err := w.GetWriters()
	if err != nil {
		return
	}
	for i := range writers {
		n, _ = writers[i].Write(p)
	}
	return
}

func (w *MultipleWritersWithSource) SetSource(source string) (err error) {
	writers, err := w.GetWriters()
	if err != nil {
		return
	}
	for i := range writers {
		err = writers[i].SetSource(source)
	}
	return
}

func (w *MultipleWritersWithSource) Close() (err error) {
	err = w.closerStore.Close()
	return
}

// NewMultipleWritersWithSource returns a writer which writes using multiple writers.
// On close, all sub writers are also closed.
func NewMultipleWritersWithSource(writers ...WriterWithSource) (*MultipleWritersWithSource, error) {
	return newWritersWithSource(true, writers...)
}

// NewMultipleWriterWithSourceWithoutClosingWriters returns a writer which writes using multiple writers.
// It is similar to NewMultipleWritersWithSource but differs when closing as the sub writers are not closed and, it is the responsibility of their creator to do so.
func NewMultipleWriterWithSourceWithoutClosingWriters(writers ...WriterWithSource) (*MultipleWritersWithSource, error) {
	return newWritersWithSource(false, writers...)
}

func newWritersWithSource(closeWriterOnClose bool, writers ...WriterWithSource) (writer *MultipleWritersWithSource, err error) {
	writer = &MultipleWritersWithSource{
		writers:      nil,
		closeWriters: closeWriterOnClose,
		closerStore:  parallelisation.NewCloserStore(false),
	}
	err = writer.AddWriters(writers...)
	return
}

// CreateMultipleWritersWithSource creates a compound writer with source.
//
// Deprecated: Use NewMultipleWritersWithSource instead
func CreateMultipleWritersWithSource(writers ...WriterWithSource) (writer *MultipleWritersWithSource, err error) {
	return NewMultipleWritersWithSource(writers...)
}

type DiodeWriter struct {
	WriterWithSource
	diodeWriter io.Writer
	slowWriter  WriterWithSource
	closeStore  *parallelisation.CloserStore
}

func (w *DiodeWriter) Write(p []byte) (n int, err error) {
	return w.diodeWriter.Write(p)
}

func (w *DiodeWriter) Close() error {
	err := w.slowWriter.Close()
	if err != nil {
		return err
	}
	if diodeCloser, ok := w.diodeWriter.(io.Closer); ok {
		return diodeCloser.Close()
	}
	return err
}

func (w *DiodeWriter) SetSource(source string) error {
	return w.slowWriter.SetSource(source)
}

// NewDiodeWriterForSlowWriter returns a thread-safe, lock-free, non-blocking WriterWithSource using a diode. On close, the writer is also closed.
func NewDiodeWriterForSlowWriter(slowWriter WriterWithSource, ringBufferSize int, pollInterval time.Duration, droppedMessagesLogger Loggers) WriterWithSource {
	return newDiodeWriterForSlowWriter(true, slowWriter, ringBufferSize, pollInterval, droppedMessagesLogger)
}

// NewDiodeWriterForSlowWriterWithoutClosing returns a thread-safe, lock-free, non-blocking WriterWithSource using a diode. It is similar to NewDiodeWriterForSlowWriter but differs in that the writer is not closed when closing, only the internal diode.
func NewDiodeWriterForSlowWriterWithoutClosing(slowWriter WriterWithSource, ringBufferSize int, pollInterval time.Duration, droppedMessagesLogger Loggers) WriterWithSource {
	return newDiodeWriterForSlowWriter(false, slowWriter, ringBufferSize, pollInterval, droppedMessagesLogger)
}

func newDiodeWriterForSlowWriter(closeWriterOnClose bool, slowWriter WriterWithSource, ringBufferSize int, pollInterval time.Duration, droppedMessagesLogger Loggers) WriterWithSource {
	closerStore := parallelisation.NewCloserStore(false)
	if closeWriterOnClose {
		closerStore.RegisterCloser(slowWriter)
	}
	d := diode.NewWriter(slowWriter, ringBufferSize, pollInterval, func(missed int) {
		if droppedMessagesLogger != nil {
			droppedMessagesLogger.LogError(fmt.Sprintf("Logger dropped %d messages", missed))
		}
	})
	var diodeWriter io.Writer //nolint:gosimple // S1021: should merge variable declaration with assignment on next line (gosimple)
	diodeWriter = d
	if diodeCloser, ok := diodeWriter.(io.Closer); ok {
		closerStore.RegisterCloser(diodeCloser)
	}

	return &DiodeWriter{
		diodeWriter: diodeWriter,
		slowWriter:  slowWriter,
		closeStore:  closerStore,
	}
}

type infoWriter struct {
	loggers Loggers
}

func (w *infoWriter) Write(p []byte) (n int, err error) {
	if w.loggers == nil {
		err = commonerrors.ErrNoLogger
		return
	}
	n = len(p)
	w.loggers.Log(string(p))
	return
}

// NewInfoWriterFromLoggers returns a io.Writer from a Loggers by only returning INFO messages
func NewInfoWriterFromLoggers(l Loggers) (w io.Writer, err error) {
	if l == nil {
		err = commonerrors.ErrNoLogger
		return
	}
	w = &infoWriter{
		loggers: l,
	}
	return
}

type errWriter struct {
	loggers Loggers
}

func (w *errWriter) Write(p []byte) (n int, err error) {
	if w.loggers == nil {
		err = commonerrors.ErrNoLogger
		return
	}
	n = len(p)
	w.loggers.LogError(string(p))
	return
}

// NewErrorWriterFromLoggers returns a io.Writer from a Loggers by only returning ERROR messages
func NewErrorWriterFromLoggers(l Loggers) (w io.Writer, err error) {
	if l == nil {
		err = commonerrors.ErrNoLogger
		return
	}
	w = &errWriter{
		loggers: l,
	}
	return
}
