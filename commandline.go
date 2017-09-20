package main

import (
	"flag"
)

type Commandline struct {
	URL        string
	CaCert     string
	SSL        bool
	Repetition int
}

func (c *Commandline) Parse(options []string) (err error) {
	if len(options) == 1 {
		return
	}

	flagSet := flag.NewFlagSet("", flag.ExitOnError)
	ssl := flagSet.Bool("ssl", false, "")
	cacert := flagSet.String("cacert", "", "")
	url := flagSet.String("url", "", "")
	repetition := flagSet.Int("rep", 1500, "")

	err = flagSet.Parse(options[1:])
	if err != nil {
		panic(err)
	}

	c.URL = *url
	c.CaCert = *cacert
	if c.CaCert != "" {
		c.SSL = true
	} else {
		c.SSL = *ssl
	}

	c.Repetition = *repetition

	return
}

func NewCommandline(options []string) *Commandline {
	c := new(Commandline)
	c.Parse(options)
	return c
}
