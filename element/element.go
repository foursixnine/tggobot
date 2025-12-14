package element

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type IElement interface {
	GetText() string
	MakeMarkdown() string
	getTitleofLink()
}

type Element struct {
	Message tg.Message
}

func (e Element) GetText() string {
	return strings.ReplaceAll(e.Message.Text, "\n", " ")
}

func (e Element) MakeMarkdown() string {
	title, err := e.getTitleofLink()
	if err != nil {
		log.Println("GE Error getting title of link, leaving entry as is")
		return e.Message.Text
	}

	return "[" + title + "](" + e.Message.Text + ")" // lets reassign, we don't have a problem with simple urls

}

func (e Element) getTitleofLink() (string, error) {
	response, err := http.Get(e.Message.Text)

	if err != nil {
		log.Println("Element: Error getting link: ", err)
		return "", err
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		log.Println("Element: Error getting link: ", response.Status)
		return "", fmt.Errorf("Http status response: %s", response.Status)
	}

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		log.Fatal("Element: Could not read document", err)
		return "", errors.New("Document can't be read")
	}

	title := doc.Find("title").Text()
	if title == "" {
		log.Fatal("Element: title element is empty or not found: ", title, e.Message.Text)
		return e.Message.Text, nil
	}

	return title, nil
}

type GenericElement struct {
	Element
}

type HNElement struct {
	Element
}

func (e HNElement) MakeMarkdown(entity tg.MessageEntity) string {
	link_to := fmt.Sprintf("[%s](%s)", e.Message.Text[entity.Offset:entity.Offset+entity.Length], entity.URL)
	return link_to

}

type TextLinkElement struct {
	Element
}

func (e TextLinkElement) MakeMarkdown(entity tg.MessageEntity) string {
	var link_to string
	title, err := e.getTitleofLink(entity.URL)
	if err != nil {
		log.Println("TextLinkElement Error getting title of link")
		log.Println("TextLinkElement leaving entry as is")
		link_to = entity.URL
	} else {
		link_to = fmt.Sprintf("[%s](%s)", title, entity.URL)
		log.Println("TextLinkElement Formatting with: ", title, entity.URL, link_to)
	}
	return link_to

}

func (e TextLinkElement) getTitleofLink(link string) (string, error) {
	response, err := http.Get(link)

	if err != nil {
		log.Println("Error getting link: ", err)
		return "", err
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		log.Println("Error getting link: ", response.Status)
		return "", err
	}

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		log.Fatal("Could not read document", err)
	}

	title := doc.Find("title").Text()
	log.Println("found title:", title)

	return title, nil

}
