package main

import (
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/charmbracelet/bubbles/textinput"
)

func printInputs(inputs []textinput.Model) {
	fmt.Println()
	fmt.Println()
	fmt.Println("Your specifications are...")
	for _, input := range inputs {
		fmt.Printf("%s: %s\n", input.Placeholder, input.Value())
	}
	fmt.Println()
}

func houseHunt(inputs []textinput.Model) {
	areas := strings.Split(inputs[0].Value(), ",")
	priceMin := convertStringToNum(inputs[1].Value())
	priceMax := convertStringToNum(inputs[2].Value())
	bedMin := convertStringToNum(inputs[3].Value())
	bathMin := convertStringToNum(inputs[4].Value())
	sqftMin := convertStringToNum(inputs[5].Value())
	status := calculateStatus(inputs[6].Value())
	numResults := convertStringToNum(inputs[7].Value())

	ch := make(chan string)
	var wg sync.WaitGroup

	for _, area := range areas {
		wg.Add(1)
		go callApi(area, priceMin, priceMax, bedMin, bathMin, sqftMin, status, numResults, ch, &wg)
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	for res := range ch {
		fmt.Println(res)
	}
}

func calculateStatus(lookingFor string) []string {
	if strings.EqualFold(lookingFor, "Buy") {
		status := []string{"for_sale", "ready_to_build"}
		return status
	} else {
		status := []string{"for_rent"}
		return status
	}
}

func convertStringToNum(numStr string) (num int) {
	num, err := strconv.Atoi(numStr)
	if err != nil {
		panic(err)
	}
	return num
}
