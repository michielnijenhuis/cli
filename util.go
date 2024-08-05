package cli

import "log"

func checkPtr(ptr any, name string) {
	if ptr == nil {
		log.Fatalf("Expected \"%s\" pointer not te be nil", name)
	}
}
