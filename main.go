package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"zarate.co/tggobot/element"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile | log.Lmsgprefix)
	log.Println("Hello World")
	if os.Getenv("TELEGRAM_APITOKEN") == "" {
		fmt.Println("TELEGRAM_APITOKEN is not set")
		os.Exit(1)
	}

	if os.Getenv("BRAIN_LOCATION") == "" {
		fmt.Println("BRAIN_LOCATION is not set")
		os.Exit(1)
	}

	bot, err := tg.NewBotAPI(os.Getenv("TELEGRAM_APITOKEN"))
	if err != nil {
		panic(err)
	}

	value, exists := getEnv("DEBUG")
	if exists {
		boolean, err := strconv.ParseBool(value)
		if err != nil {
			panic(err)
		}
		bot.Debug = boolean
	}

	log.Printf("Authorized on account %s", bot.Self.UserName)

	// Create a new UpdateConfig struct with an offset of 0. Offsets are used
	// to make sure Telegram knows we've handled previous values and we don't
	// need them repeated.
	updateConfig := tg.NewUpdate(0)

	// Tell Telegram we should wait up to 30 seconds on each request for an
	// update. This way we can get information just as quickly as making many
	// frequent requests without having to send nearly as many.
	updateConfig.Timeout = 30

	// Start polling Telegram for updates.
	updates := bot.GetUpdatesChan(updateConfig)

	// Let's go through each update that we're getting from Telegram.
	for update := range updates {
		go updateHandler(&update, bot)
	}

}

func updateHandler(update *tg.Update, bot *tg.BotAPI) {
	// Telegram can send many types of updates depending on what your Bot
	// is up to. We only want to look at messages for now, so we can
	// discard any other updates.
	if update.Message == nil {
		return
	}

	if update.Message.IsCommand() { // ignore any non-command Messages
		processCommand(update.Message, bot)
		return
	}

	reply := tg.NewMessage(update.Message.Chat.ID, "Got message, processing")
	reply.ReplyToMessageID = update.Message.MessageID

	response, err := bot.Send(reply)
	if err != nil {
		panic(err)
	}

	messageID := response.MessageID
	processUpdate(messageID, *update.Message, bot)
	updateMessage(messageID, response.Chat.ID, bot, "Has been processed")
}

func processCommand(m *tg.Message, b *tg.BotAPI) {

	msg := tg.NewMessage(m.Chat.ID, "")
	switch m.Command() {
	case "ping":
		msg.Text = "pong"
	default:
		msg.Text = "I don't know that command, I know /ping"
	}
	_, err := b.Send(msg)
	if err != nil {
		log.Panic(err)
	}
}

func processUpdate(messageID int, m tg.Message, b *tg.BotAPI) {
	log.Println("got message from:", m.From.FirstName)
	log.Println("I am:", b.Self.FirstName)

	Brain := newBrain(os.Getenv("BRAIN_LOCATION"))
	Brain.Text = strings.ReplaceAll(m.Text, "\n", " ")

	// lets extract the entities, we only care about links
	for _, entity := range m.Entities {
		switch entity.Type {
		case "url":
			e := element.GenericElement{Element: element.Element{Message: m}}
			Brain.Text = e.MakeMarkdown()

		case "text_link":
			// get the string slice for a given entity
			var link_to string
			// check if its a message from HN use what it provides
			if m.ForwardFromChat != nil && m.ForwardFromChat.UserName == "hackernewslive" {
				log.Println("Link is from Hacker News - Making it simple")
				e := element.HNElement{Element: element.Element{Message: m}}
				link_to = e.MakeMarkdown(entity)
			} else {
				log.Println("Its text with a link")
				e := element.TextLinkElement{Element: element.Element{Message: m}}
				link_to = e.MakeMarkdown(entity)
			}
			Brain.Links = append(Brain.Links, link_to)

		default:
			log.Println("Got something else of type: ", entity.Type)
		}
	}

	log.Println("Contents of brain links:\t", Brain.Links)
	log.Println("Contents of brain Text:\t", Brain.Text)
	chatID := m.Chat.ID
	updateMessage(messageID, chatID, b, "updating...")
	if saveToBrain(Brain) {
		updateMessage(messageID, chatID, b, "Sucessfully appended to brain")
	}

}

func updateMessage(messageID int, chatID int64, b *tg.BotAPI, t string) {
	rpl := tg.NewEditMessageText(chatID, messageID, t)
	_, err := b.Send(rpl)
	if err != nil {
		panic(err)
	}
}

func getEnv(key string) (string, bool) {
	if value, ok := os.LookupEnv(key); ok {
		return value, true
	}
	return "", false
}
