package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/krognol/go-wolfram"
	"github.com/tidwall/gjson"

	"github.com/joho/godotenv"
	"github.com/shomali11/slacker"
	witai "github.com/wit-ai/wit-go"
)

func printCommandEvents(analyticsChannel <-chan *slacker.CommandEvent) {
	for event := range analyticsChannel {
		fmt.Println("Command Events")
		fmt.Println(event.Timestamp)
		fmt.Println(event.Command)
		fmt.Println(event.Parameters)
		fmt.Println(event.Event)
		fmt.Println()
	}
}

func main() {
	
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	
	bot := slacker.NewClient(os.Getenv("SLACK_BOT_TOKEN"), os.Getenv("SLACK_APP_TOKEN"))
	fmt.Println("SLACK_BOT_TOKEN:", os.Getenv("SLACK_BOT_TOKEN"))
	fmt.Println("SLACK_APP_TOKEN:", os.Getenv("SLACK_APP_TOKEN"))

	client := witai.NewClient((os.Getenv("WIT_AI_TOKEN")))
	wolframClient := &wolfram.Client{AppID: os.Getenv("WOLFRAM_APP_ID")}

	go printCommandEvents(bot.CommandEvents())


	bot.Command("query for bot - <message>", &slacker.CommandDefinition{
		Description: "Send any question to Wolfram",

		Handler: func(botCtx slacker.BotContext, request slacker.Request, response slacker.ResponseWriter) {
			query := request.Param("message")
			fmt.Println(query)
			msg, _ := client.Parse(&witai.MessageRequest{
				Query: query,
			})

			data, _ := json.MarshalIndent(msg, "", "    ")
			rough := string(data[:])
			value := gjson.Get(rough, "entities.wit$wolfram_search_query:wolfram_search_query.0.value")
			answer := value.String()
			res, err := wolframClient.GetSpokentAnswerQuery(answer, wolfram.Metric, 1000)
			if err != nil {
				fmt.Println("there is an error")
			}
			fmt.Println(value)

			response.Reply(res)
		},
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start the bot
	err = bot.Listen(ctx)
	if err != nil {
		log.Fatal(err)
	}
}
