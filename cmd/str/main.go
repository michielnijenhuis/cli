package main

import (
	"log"
)

func main() {
	if err := RootCmd.Execute(); err != nil {
		log.Fatalln(err)
	}
}
