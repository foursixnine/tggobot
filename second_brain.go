package main

import (
	"log"
	"os"
)

type Brain struct {
	Text          string
	Links         []string
	BrainLocation string
}

func newBrain(location string) Brain {
	return Brain{
		BrainLocation: location,
	}
}

func saveToBrain(b Brain) bool {
	f, err := os.OpenFile(b.BrainLocation, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		panic(err)
	}
	// start at the top, next elements are in second level of a list
	text := "- " + b.Text + "\n"

	for _, link := range b.Links {
		text += "    - " + link + "\n"
	}

	defer f.Close()

	if _, err = f.WriteString(text); err != nil {
		panic(err)
	} // save the brain to a file

	log.Println("Link saved to brain")

	return true
}
