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

type badgerFromIterator struct {
	badgerPrefixIterator
}

type badgerToIterator struct {
	badgerFromToIterator
}

// a badgerPrefixToIterator is simply a FromToIterator,
//while badgerPrefixFromIterator is slightly faster than
//a badgerFromToIterator iterator, and the logic wouldn't work anyway
type badgerPrefixFromIterator struct {
	badgerPrefixIterator
	from []byte
}

func (b badgerPrefixFromIterator) Rewind() {
	b.Seek(b.from)
}

func (b *badgerToIterator) Rewind() {
	b.Iterator.Rewind()
}

func (b *badgerFromIterator) Valid() bool {
	return b.Iterator.Valid()
}

func (b *badgerPrefixIterator) Rewind() {
	b.Seek(b.prefix)
}

func (b *badgerPrefixIterator) Valid() bool {
	return b.ValidForPrefix(b.prefix)
}

func (b *badgerFromToIterator) Valid() bool {
	return b.Iterator.Valid() && keyLtEq(b.Item().Key(), b.to)
}

func keyLtEq(key, upperBound []byte) bool {
	return 1 > bytes.Compare(key, upperBound)
}

// KeyRangeIterator returns an Iterator interface implementation, that wraps the badger.Iterator
// in order to simplify iterating with from-to and/or prefix.
func KeyRangeIterator(it *badger.Iterator, prefix, from, to string) Iterator {
	switch {
	case prefix+from+to == "":
		return it
	case prefix != "" && from+to == "":
		return &badgerPrefixIterator{*it, []byte(prefix)}
	case to != "" && prefix+from == "":
		return &badgerToIterator{badgerFromToIterator{badgerPrefixIterator{*it, nil}, []byte(to)}}
	case from != "" && prefix+to == "":
		return &badgerFromIterator{badgerPrefixIterator{*it, []byte(from)}}
	case from != "" && to != "" && prefix == "":
		return &badgerFromToIterator{badgerPrefixIterator{*it, []byte(from)}, []byte(to)}
	case from == "":
		lastInPrefix := lastInPrefix(prefix, to)
		if keyLtEq(lastInPrefix, []byte(to)) {
			return &badgerPrefixIterator{*it, []byte(prefix)}
		}
		return &badgerFromToIterator{badgerPrefixIterator{*it, []byte(prefix)}, []byte(to)}
	case to == "":
		if prefix > from {
			return &badgerPrefixIterator{*it, []byte(prefix)}
		}
		return &badgerPrefixFromIterator{badgerPrefixIterator{*it, []byte(prefix)}, []byte(from)}

	default: // to != "" && prefix != "" && from != "":
		f, t := from, to
		if prefix > from {
			f = prefix
		}
		lastInPrefix := lastInPrefix(prefix, to)
		if keyLtEq(lastInPrefix, []byte(to)) {
			if f == prefix {
				return &badgerPrefixIterator{*it, []byte(prefix)}
			}
			return &badgerPrefixFromIterator{badgerPrefixIterator{*it, []byte(prefix)}, []byte(f)}
		}
		return &badgerFromToIterator{badgerPrefixIterator{*it, []byte(f)}, []byte(t)}
	}
}

func lastInPrefix(prefix, to string) []byte {
	lastValueInPrefix := []byte(prefix)
	if len(to)-len(prefix) > 0 {
		padding := make([]byte, len(to)-len(prefix))
		for i := range padding {
			padding[i] = uint8(255)
		}
		lastValueInPrefix = append(lastValueInPrefix, padding...)
	}
	return lastValueInPrefix
}

func keyRangeIterator(it *badger.Iterator, keyRange *proto.KeyRange) Iterator {
	return KeyRangeIterator(it, keyRange.Prefix, keyRange.From, keyRange.To)
}
