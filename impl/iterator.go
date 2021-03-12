package impl

import (
	"bytes"
	"github.com/MikkelHJuul/ld/proto"
	"github.com/dgraph-io/badger/v3"
)

type Iterator interface {
	Rewind()
	Valid() bool
	Next()
	Item() *badger.Item
}

type badgerPrefixIterator struct {
	*badger.Iterator
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

func keyRangeIterator(it *badger.Iterator, keyRange *proto.KeyRange) Iterator {
	if keyRange.Prefix+keyRange.From+keyRange.To != "" {
		from, to := keyRange.Prefix, keyRange.Prefix
		if keyRange.Prefix < keyRange.From {
			from = keyRange.From
		}
		if keyRange.Prefix > keyRange.To {
			to = keyRange.To
		}
		if from == to {
			return &badgerPrefixIterator{it, []byte(from)} //faster than from-to Iteration
		}
		return &badgerFromToIterator{
			badgerPrefixIterator: badgerPrefixIterator{it, []byte(from)},
			to:                   []byte(to),
		}
	}
	return it
}
