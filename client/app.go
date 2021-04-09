package main

import (
	"github.com/MikkelHJuul/ld/client/impl"
	"github.com/desertbit/grumble"
)

var (
	sharedFlags = func(f *grumble.Flags) {
		f.String("t", "target", "localhost:5326", "the target ld server")
		f.String("p", "protofile", "", "the protofile to serialize from, if unset plain bytes are sent, and the received values are not marshalled to JSON")
	}

	app = grumble.New(&grumble.Config{
		Name:        "ld-client",
		Description: "ld-client is an interactive client and executable to do non-\"client-side\" streaming requests",
		Flags:       sharedFlags,
	})

	getCmd = &grumble.Command{
		Name:    "get",
		Help:    "get a single record",
		Aliases: []string{"fetch", "read"},
		Args: func(a *grumble.Args) {
			a.String("key", "the key to fetch")
		},
		Flags: sharedFlags,
		Run:   impl.Get,
	}

	setCmd = &grumble.Command{
		Name:    "set",
		Help:    "set a single record",
		Aliases: []string{"add", "create"},
		Args: func(a *grumble.Args) {
			a.String("key", "the key to fetch")
			a.String("value", "the value to set, or to serialize if protofile is set")
		},
		Flags: sharedFlags,
		Run:   impl.Set,
	}

	getRangeCmd = &grumble.Command{
		Name: "get-range",
		Help: "get a range of records, empty implies all",
		Flags: func(f *grumble.Flags) {
			f.String("", "prefix", "", "key prefix")
			f.String("", "from", "", "scan range from this key, inclusive")
			f.String("", "to", "", "scan range to this key, inclusive")
			f.String("", "pattern", "", "key pattern to query using")
			sharedFlags(f)
		},
		Aliases: []string{"getran"},
		Run:     impl.GetRange,
	}

	deleteCmd = &grumble.Command{
		Name:    "delete",
		Help:    "get a single record",
		Aliases: []string{"del", "remove", "rem"},
		Args: func(a *grumble.Args) {
			a.String("key", "the key to fetch")
		},
		Flags: sharedFlags,
		Run:   impl.Delete,
	}

	deleteRangeCmd = &grumble.Command{
		Name: "delete-range",
		Help: "delete a range of records, empty implies all",
		Flags: func(f *grumble.Flags) {
			f.String("", "prefix", "", "key prefix")
			f.String("", "from", "", "scan range from this key, inclusive")
			f.String("", "to", "", "scan range to this key, inclusive")
			f.String("", "pattern", "", "key pattern to query using")
			sharedFlags(f)
		},
		Aliases: []string{"delran"},
		Run:     impl.DeleteRange,
	}

	// Version acts as a target for compile-time linking the project version into the code
	Version = "unset"

	versionCmd = &grumble.Command{
		Name: "version",
		Help: "print version info",
		Run: func(ctx *grumble.Context) error {
			ctx.App.Println("Version:", Version)
			return nil
		},
	}
)

func init() {
	app.AddCommand(setCmd)
	app.AddCommand(getCmd)
	app.AddCommand(getRangeCmd)
	app.AddCommand(deleteCmd)
	app.AddCommand(deleteRangeCmd)
	app.AddCommand(versionCmd)
}
