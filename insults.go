package main

import (
	"math/rand"
	"os"
	"strings"
)

type InsultFactory interface {
	GetInsult() string
}

type basicInsultFactory struct {
	insults []string
}

func (bif *basicInsultFactory) GetInsult() string {
	i := rand.Intn(len(bif.insults))
	return bif.insults[i]
}

func NewInsultFactory(fname string) (InsultFactory, error) {
	bytes, err := os.ReadFile(fname)
	if err != nil {
		return nil, err
	}

	bif := &basicInsultFactory{
		insults: strings.Split(string(bytes), "\n"),
	}

	return bif, nil
}
