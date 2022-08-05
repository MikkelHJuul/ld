package main

type key []byte
type value []byte
type kv struct {
	key
	value
	error
}

type sender interface {
	Send(kv) error
}

type receiver interface {
	Receive() kv
}

type kvDatabase interface {
	get(key) (value, error)
	getMany(sender, receiver) error
	getRange(ranger, sender) error

	set(kv) (kv, error)
	setMany(sender, receiver) error

	delete(key) (value, error)
	deleteMany(sender, receiver) error
	deleteRange(ranger, sender) error
}
