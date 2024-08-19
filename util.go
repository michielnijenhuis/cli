package cli

import "log"

func checkPtr(ptr any, name string) {
	if ptr == nil {
		log.Fatalf("expected \"%s\" pointer not te be nil", name)
	}
}
