package main

import (
	"fmt"
	"io/ioutil"
	"strconv"
    "strings"
)

func main() {
	var str string
	var index int
	fmt.Scan(&str)
	r := []rune(str)
	text, err := ioutil.ReadFile("res/lastTeachIndex")
	if err != nil {
		index = 0
	} else {
        tempStr := strings.Trim(string(text), " \n")
		index, err = strconv.Atoi(tempStr)
		if err != nil {
			index = 0
		}
	}
	for i := range r {
		ioutil.WriteFile("file"+fmt.Sprint(i+index)+".txt", []byte(string(r[i])), 0644)
	}
	ioutil.WriteFile("res/lastTeachIndex", []byte(strconv.Itoa(index + len(r))), 0644)
}
