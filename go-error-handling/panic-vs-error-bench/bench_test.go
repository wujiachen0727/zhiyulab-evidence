package main

import (
	"errors"
	"fmt"
	"testing"
)

// ========================================
// panic+recover 路径
// ========================================

func mayPanic() {
	panic("something went wrong")
}

func callWithPanicRecover() (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("recovered: %v", r)
		}
	}()
	mayPanic()
	return nil
}

// ========================================
// error return 路径
// ========================================

var errSomething = errors.New("something went wrong")

func mayError() error {
	return errSomething
}

func callWithError() error {
	return mayError()
}

// ========================================
// Benchmark
// ========================================

func BenchmarkPanicRecover(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = callWithPanicRecover()
	}
}

func BenchmarkErrorReturn(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = callWithError()
	}
}

// 正常路径（无错误）对比
func noPanic() int {
	return 42
}

func callNoPanic() int {
	return noPanic()
}

func noError() (int, error) {
	return 42, nil
}

func callNoError() (int, error) {
	return noError()
}

func BenchmarkNoPanicPath(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = callNoPanic()
	}
}

func BenchmarkNoErrorPath(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = callNoError()
	}
}
