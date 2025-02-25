package main

import (
	"fmt"
)

// showReleases this shows all the releases in the releaseScripts directory
func (prog *Prog) showReleases(indent, relIndent string) {
	releases, err := prog.findReleases()
	if err != nil {
		fmt.Println(indent+"Error:", err)
		return
	}

	if len(releases) == 0 {
		fmt.Println(indent + "There are no releases")
	} else {
		fmt.Println(indent + "Available releases:")

		for _, r := range releases {
			fmt.Println(relIndent, r)
		}
	}
}
