package main

import (
	"log"
	"os"
)

// Reload the domain list from path each time a signal arrives. A list that
// fails to load leaves the current one in force, so a bad edit cannot open or
// close the filter. Each result is sent to results when it is not nil.
func reloadDomainList(path string, signals <-chan os.Signal, results chan<- error) {
	for sig := range signals {
		err := loadDomainList(path)
		if err != nil {
			log.Printf("Failed to reload domain list on %v, keeping the current list: %v", sig, err)
		} else {
			log.Printf("Reloaded domain list from %s on %v", path, sig)
		}

		if results != nil {
			results <- err
		}
	}
}
