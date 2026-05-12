package main

import (
	"fmt"
	"strings"
)

func runDeps(args []string) error {
	fs := subFlagSet("deps", "show dependency overview for all items in the tap")
	plain := fs.Bool("plain", false, "machine-readable output (one line per dep, TSV)")
	if err := parseFlags(fs, args); err != nil {
		return err
	}

	c := LoadConfig()
	tap, err := TapDir(c)
	if err != nil {
		return err
	}
	items, err := ScanTap(tap)
	if err != nil {
		return err
	}
	if len(items) == 0 {
		ok("tap is empty — nothing to report")
		return nil
	}

	if *plain {
		for _, it := range items {
			if len(it.Deps) == 0 {
				fmt.Printf("%s\t%s\t-\n", it.Name, it.Kind)
				continue
			}
			for _, d := range it.Deps {
				fmt.Printf("%s\t%s\t%s\n", it.Name, it.Kind, d)
			}
		}
		return nil
	}

	banner("dependency overview — %d item(s)", len(items))
	nameW, kindW := colWidth(items, func(it Item) string { return it.Name }, 4),
		colWidth(items, func(it Item) string { return it.Kind }, 4)
	for _, it := range items {
		dep := strings.Join(it.Deps, ", ")
		if dep == "" {
			dep = "(none)"
		}
		fmt.Printf("  %-*s  %-*s  %s\n", nameW, it.Name, kindW, it.Kind, dep)
	}
	return nil
}

func colWidth(items []Item, get func(Item) string, min int) int {
	w := min
	for _, it := range items {
		if n := len(get(it)); n > w {
			w = n
		}
	}
	return w
}
