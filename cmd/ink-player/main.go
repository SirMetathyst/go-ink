package main

import (
	"bufio"
	"fmt"
	"github.com/SirMetathyst/go-ink/runtime"
	"io"
	"log"
	"os"
	"strconv"
)

func main() {

	jsonBytes, err := os.ReadFile("TheIntercept.json")
	if err != nil {
		log.Fatalln(err)
	}

	story := runtime.NewStory(string(jsonBytes))

	//fmt.Println(story.Continue())

	for {

		for story.CanContinue() {
			fmt.Println(story.Continue())
		}

		choices := story.CurrentChoices()
		if len(choices) == 0 {
			return
		}

		for i, choice := range choices {
			fmt.Printf("%d: %s\n", i, choice.Text)
		}

		reader := bufio.NewReader(os.Stdin)
		fmt.Print(">")
		v, err := reader.ReadString('\n')
		if err == io.EOF {
			return
		}

		choiceIndex, err := strconv.Atoi(v[:len(v)-1])
		if err != nil || choiceIndex < 0 || choiceIndex > len(choices) {
			fmt.Println(err)
		}

		story.ChooseChoiceIndex(choiceIndex)
	}
}
