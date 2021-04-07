package main

import (
	"github.com/MikkelHJuul/ld/client/impl"
	"github.com/desertbit/grumble"
)

var app = grumble.New(&grumble.Config{
	Name:        "ld-client",
	Description: "ld-client is an interactive client and executable to do non-\"client-side\" streaming requests",
        Flags:       func(f *grumble.Flags) {
	               f.String("t", "target", "localhost:5326", "the target ld server")
	               f.String("p", "protofile", "", "the protofile to serialize from, if unset plain bytes are posted")
        },
})

var getCmd = &grumble.Command{
	Name:    "get",
	Help:    "get a single record",
	Aliases: []string{"fetch", "read"},
	Args: func(a *grumble.Args) {
		a.String("key", "the key to fetch")
        },
	Usage: "get <key>",
	Run:   impl.Get,
}

var setCmd = &grumble.Command{
	Name:    "set",
	Help:    "set a single record",
	Aliases: []string{"add", "create"},
	Args: func(a *grumble.Args) {
		a.String("key", "the key to fetch")
		a.String("value", "the value to set, or to serialize if protofile is set")
	},
	Usage: "set <key> <value>",
	Run:   impl.Set,
}

var getRangeCmd = &grumble.Command{
	Name: "get-range",
	Help: "get a range of records",
	Flags: func(f *grumble.Flags) {
		f.String("", "prefix", "", "key prefix")
		f.String("", "from", "", "scan range from this key, inclusive")
		f.String("", "to", "", "scan range to this key, inclusive")
		f.String("", "pattern", "", "key pattern to query using")
	},
	Run: impl.GetRange,
}

var deleteCmd = &grumble.Command{
	Name:    "delete",
	Help:    "get a single record",
	Aliases: []string{"del", "remove", "rem"},
	Args: func(a *grumble.Args) {
		a.String("key", "the key to fetch")
	},
	Run:   impl.Delete,
}

var deleteRangeCmd = &grumble.Command{
	Name: "delete-range",
	Help: "delete a range of records",
	Flags: func(f *grumble.Flags) {
		f.String("", "prefix", "", "key prefix")
		f.String("", "from", "", "scan range from this key, inclusive")
		f.String("", "to", "", "scan range to this key, inclusive")
		f.String("", "pattern", "", "key pattern to query using")
	},
	Run: impl.DeleteRange,
}

func init() {
	app.AddCommand(setCmd)
	app.AddCommand(getCmd)
	app.AddCommand(getRangeCmd)
	app.AddCommand(deleteCmd)
	app.AddCommand(deleteRangeCmd)
}
