package element

import (
	"net/http"
	"net/http/httptest"
	"testing"

	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func TestGenericElement_MakeMarkdown_fetchTitle(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("<html><head><title>Test Title</title></head><body>ok</body></html>"))
	}))
	defer ts.Close()

	m := tg.Message{Text: ts.URL}
	e := GenericElement{Element: Element{Message: m}}

	got := e.MakeMarkdown()
	want := "[Test Title](" + ts.URL + ")"

	if got != want {
		t.Fatalf("GenericElement.MakeMarkdown() = %q, want %q", got, want)
	}
}

func TestGenericElement_MakeMarkdown_fetchTitle_nonhttpok(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("<html><head><title>Test Title</title></head><body>ok</body></html>"))
	}))
	defer ts.Close()

	m := tg.Message{Text: ts.URL}
	e := GenericElement{Element: Element{Message: m}}

	got := e.MakeMarkdown()
	want := ts.URL

	if got != want {
		t.Fatalf("nonhttpok: GenericElement.MakeMarkdown() = %q, want %q", got, want)
	}
}

func TestGenericElement_MakeMarkdown_fetchTitle_noTitle(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("<html><head></head><body>ok</body></html>"))
	}))
	defer ts.Close()

	m := tg.Message{Text: ts.URL}
	e := GenericElement{Element: Element{Message: m}}

	got := e.MakeMarkdown()
	want := ts.URL

	if got != want {
		t.Fatalf("GenericElement.MakeMarkdown() = %q, want %q", got, want)
	}
}

func TestTextLinkElement_MakeMarkdown_fetchTitle(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("<html><head><title>Another Title</title></head><body>ok</body></html>"))
	}))
	defer ts.Close()

	m := tg.Message{Text: ""}
	e := TextLinkElement{Element: Element{Message: m}}
	entity := tg.MessageEntity{URL: ts.URL}

	got := e.MakeMarkdown(entity)
	want := "[Another Title](" + ts.URL + ")"

	if got != want {
		t.Fatalf("TextLinkElement.MakeMarkdown() = %q, want %q", got, want)
	}
}

func TestHNElement_MakeMarkdown(t *testing.T) {
	text := "Check out golang"
	// link text is "golang" starting at offset 10 length 6
	m := tg.Message{Text: text}
	e := HNElement{Element: Element{Message: m}}
	entity := tg.MessageEntity{Offset: 10, Length: 6, URL: "https://golang.org"}

	got := e.MakeMarkdown(entity)
	want := "[golang](https://golang.org)"

	if got != want {
		t.Fatalf("HNElement.MakeMarkdown() = %q, want %q", got, want)
	}
}
