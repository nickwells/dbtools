package main

import (
	"fmt"
)

// showReleases this shows all the releases in the releaseScripts directory
func showReleases() {
	releases, err := findReleases()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	if len(releases) == 0 {
		fmt.Println("No releases")
	} else {
		fmt.Println("releases:")
		for _, r := range releases {
			fmt.Println("  ", r)
		}
	}
}
