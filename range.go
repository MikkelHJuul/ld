package main

import "bytes"

type ranger interface {
	Start(seeker) (key, value)
	Valid(key) bool
	Accept(key) bool
	Next(scanner) (key, value)
}

//bbolt's Cursor implements seeker
type seeker interface {
	First() ([]byte, []byte)
	Last() ([]byte, []byte)
	Seek([]byte) ([]byte, []byte)
}

//bbolt's Cursor implements seeker
type scanner interface {
	Next() ([]byte, []byte)
	Prev() ([]byte, []byte)
}

type acceptor func(key) bool
type scanFunc func(scanner) (key, value)
type seekFunc func(seeker) (key, value)

type CompRanger struct {
	seek      seekFunc
	validator acceptor
	acceptor  acceptor
	scanFunc  scanFunc
}

var _ ranger = (*CompRanger)(nil)

func (d CompRanger) Start(s seeker) (key, value) {
	return d.seek(s)
}

func (d CompRanger) Accept(k key) bool {
	return d.acceptor(k)
}

func (d CompRanger) Valid(k key) bool {
	return d.validator(k)
}

func (d CompRanger) Next(s scanner) (key, value) {
	return d.scanFunc(s)
}

func accept() func(key) bool {
	return func(key) bool {
		return true
	}
}

type RangerOpt func(*CompRanger)

func NewRanger(options ...RangerOpt) CompRanger {
	c := &CompRanger{
		seek:      SeekerFirst(),
		validator: accept(),
		acceptor:  accept(),
		scanFunc:  ForwardScanning(),
	}
	for _, opt := range options {
		opt(c)
	}
	return *c
}

func WithValidatorPrefix(prefix key) RangerOpt {
	return func(r *CompRanger) {
		r.validator = func(k key) bool {
			return bytes.HasPrefix(k, prefix)
		}
	}
}

func WithSeek(s seekFunc) RangerOpt {
	return func(r *CompRanger) {
		r.seek = s
	}
}

func WithAcceptor(a acceptor) RangerOpt {
	return func(r *CompRanger) {
		r.acceptor = a
	}
}

func WithValidatorMax(max key) RangerOpt {
	return func(r *CompRanger) {
		r.validator = func(k key) bool {
			return bytes.Compare(k, max) <= 0
		}
	}
}

func WithValidatorMin(min key) RangerOpt {
	return func(r *CompRanger) {
		r.validator = func(k key) bool {
			return bytes.Compare(k, min) >= 0
		}
	}
}

func WithPrefix(prefix key) (RangerOpt, RangerOpt) {
	return WithSeek(SeekerTo(prefix)), WithValidatorPrefix(prefix)
}

func WithScanner(scanner scanFunc) RangerOpt {
	return func(r *CompRanger) {
		r.scanFunc = scanner
	}
}

func ForwardScanning() scanFunc {
	return func(s scanner) (key, value) {
		return s.Next()
	}
}

func ReverseScanning() scanFunc {
	return func(s scanner) (key, value) {
		return s.Prev()
	}
}

func SeekerFirst() seekFunc {
	return func(s seeker) (key, value) {
		return s.First()
	}
}

func SeekerLast() seekFunc {
	return func(s seeker) (key, value) {
		return s.Last()
	}
}

func SeekerTo(k key) seekFunc {
	return func(s seeker) (key, value) {
		return s.Seek(k)
	}
}
