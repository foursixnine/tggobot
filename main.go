package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"

	"time"

	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {
	fmt.Println("Hello World")
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

	bot.Debug = false
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
		// Telegram can send many types of updates depending on what your Bot
		// is up to. We only want to look at messages for now, so we can
		// discard any other updates.
		if update.Message == nil {
			continue
		}

		if update.Message.IsCommand() { // ignore any non-command Messages
			processCommand(update.Message, bot)
			continue
		}

		reply := tg.NewMessage(update.Message.Chat.ID, "Got message, processing")
		reply.ReplyToMessageID = update.Message.MessageID

		var messageID int
		response, err := bot.Send(reply)
		if err != nil {
			panic(err)
		}

		messageID = response.MessageID
		processUpdate(messageID, *update.Message, bot)
		updateMessage(messageID, response.Chat.ID, bot, "Has been processed")

	}

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
	fmt.Println("got message from:", m.From.FirstName)
	fmt.Println("I am:", b.Self.FirstName)

	Brain := newBrain(os.Getenv("BRAIN_LOCATION"))
	Brain.Text = strings.Replace(m.Text, "\n", " ", -1)

	// lets extract the entities, we only care about links
	for _, entity := range m.Entities {
		switch entity.Type {
		case "url":
			log.Println("Got a link: ", m.Text) // simple URL has nothing
			title, err := getTitleofLink(m.Text)
			if err != nil {
				log.Println("Error getting title of link")
				log.Println("leaving entry as is")
				Brain.Text = m.Text // lets reassign, we don't have a problem with simple urls
				break
			}

			Brain.Text = "[" + title + "](" + m.Text + ")" // lets reassign, we don't have a problem with simple urls

		case "text_link":
			// get the string slice for a given entity
			link_to := fmt.Sprintf("[%s](%s)", m.Text[entity.Offset:entity.Offset+entity.Length], entity.URL)
			Brain.Links = append(Brain.Links, link_to)
		default:
			log.Println("Got something else of type ", entity.Type)
		}
	}

	log.Println("Contents of brain links:", Brain.Links)
	log.Println("Contents of brain Text:", Brain.Text)
	chatID := m.Chat.ID
	updateMessage(messageID, chatID, b, "Message is edited again")
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

	time.Sleep(2 * time.Second)
}

func getTitleofLink(s string) (string, error) {
	response, err := http.Get(s)
	defer response.Body.Close()
	if err != nil {
		fmt.Println("Error getting link", err)
		return "", err
	}

	if response.StatusCode != http.StatusOK {
		fmt.Println("Error getting link", response.Status)
		return "", err
	}

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		log.Fatal(err)
	}

	title := doc.Find("title").Text()
	fmt.Println("found title", title)

	return title, nil

}
