package main

import (
	"strings"
)

type calendarIdsFlag []string

func (i *calendarIdsFlag) String() string {
	return strings.Join(*i, ",")
}

func (i *calendarIdsFlag) Set(value string) error {
	*i = append(*i, value)
	return nil
}
