package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/brandonau24/emoji-data-generator/parsers"
	"github.com/brandonau24/emoji-data-generator/readers"
)

type EmojiHandler struct{}

func (h *EmojiHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		var buffer bytes.Buffer
		encoder := json.NewEncoder(&buffer)
		encoder.SetEscapeHTML(false)

		emojiDataFile := readers.ReadEmojiDataFile()
		emojiAnnotationsFile := readers.ReadEmojiAnnotationsFile()

		emojiAnnotations := parsers.ParseAnnotations(emojiAnnotationsFile)

		emojis := parsers.ParseEmojis(emojiDataFile, emojiAnnotations)
		err := encoder.Encode(emojis)

		if err == nil {
			w.Write(buffer.Bytes()) //nolint:errcheck
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("500 - Could not parse emoji data")) //nolint:errcheck
		}
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte(fmt.Sprintf("%v request not allowed", r.Method))) //nolint:errcheck
	}
}

func main() {
	mux := http.NewServeMux()
	mux.Handle("/", &EmojiHandler{})

	log.Fatal(http.ListenAndServe(":8080", mux))
}
