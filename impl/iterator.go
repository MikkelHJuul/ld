package impl

import (
	"bytes"
	"github.com/MikkelHJuul/ld/proto"
	"github.com/dgraph-io/badger/v3"
)

// Iterator adds an implementable target for variations of different Iterator's
// for simplification of functional code, that you can then implement this reduced
// interface such that primarily methods Rewind and Valid and be overloaded.
// usage:
// 		for iterator.Rewind(); iterator.Valid(); iterator.Next() {
//			iterator.Item()
//			...
//		}
// is a general snippet of code that using this interface may have overloaded
// ... Rewind to seek to a prefix or a value
// ... Valid to validate the key is still within bounds using fx bytes.Compare
type Iterator interface {
	Rewind()
	Valid() bool
	Next()
	Item() *badger.Item
}

type badgerPrefixIterator struct {
	badger.Iterator
	prefix []byte
}

//prefix is used as from
type badgerFromToIterator struct {
	badgerPrefixIterator
	to []byte
}

func (b *badgerPrefixIterator) Rewind() {
	b.Seek(b.prefix)
}

func (b *badgerPrefixIterator) Valid() bool {
	return b.ValidForPrefix(b.prefix)
}

func (b *badgerFromToIterator) Valid() bool {
	return b.Iterator.Valid() && 1 > bytes.Compare(b.Item().Key(), b.to)
}

// KeyRangeIterator returns an Iterator interface implementation, that wraps the badger.Iterator
// in order to simplify iterating with from-to and/or prefix.
func KeyRangeIterator(it *badger.Iterator, prefix, from, to string) Iterator {
	if prefix+from+to != "" {
		f, t := prefix, prefix
		if from != "" && prefix < from {
			f = from
		}
		if to != "" && prefix > to {
			t = to
		}
		if f == t {
			return &badgerPrefixIterator{*it, []byte(f)}
		}
		return &badgerFromToIterator{
			badgerPrefixIterator: badgerPrefixIterator{*it, []byte(f)},
			to:                   []byte(t),
		}
	}
	return it
}

func keyRangeIterator(it *badger.Iterator, keyRange *proto.KeyRange) Iterator {
	return KeyRangeIterator(it, keyRange.Prefix, keyRange.From, keyRange.To)
}
