//nolint:errcheck
package parsers

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

const mockUnicodeEmojisPath = "/mock/emoji.txt"

type MockDataUrlProvider struct {
	baseUrl string
}

func (p MockDataUrlProvider) GetUnicodeEmojisDataUrl() string {
	return p.baseUrl + mockUnicodeEmojisPath
}

func (p MockDataUrlProvider) GetUnicodeAnnotationsUrl() string {
	return ""
}

func TestParseEmojisSkipsComments(t *testing.T) {
	mockHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == mockUnicodeEmojisPath {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`# This is a comment
# This is another comment
# This is the last comment`))
		}
	}))
	defer mockHttpServer.Close()

	emojis, _ := ParseEmojis(MockDataUrlProvider{
		baseUrl: mockHttpServer.URL,
	}, map[string]Annotation{})

	if len(emojis) != 0 {
		t.Errorf("Failed to parse comments")
	}
}

func TestParseEmojisSetsCodepoint(t *testing.T) {
	mockHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == mockUnicodeEmojisPath {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`# group: group1
1F600                                                  ; fully-qualified     # 😀 E1.0 grinning face`))
		}
	}))

	defer mockHttpServer.Close()

	emojis, _ := ParseEmojis(MockDataUrlProvider{
		baseUrl: mockHttpServer.URL,
	}, map[string]Annotation{
		"😀": {
			Default: []string{"one"},
			Tts:     []string{"tts"},
		},
	})
	emoji := emojis["group1"][0]

	if emoji.Codepoints != "1F600" {
		t.Errorf("Failed to parse codepoint. Received %v, expected 😀", emoji.Codepoints)
	}
}

func TestParseEmojisSetsCodepoints(t *testing.T) {
	mockHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == mockUnicodeEmojisPath {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`# group: group1
1F62E 200D 1F4A8                                       ; fully-qualified     # 😮‍💨 E13.1 face exhaling`))
		}
	}))

	defer mockHttpServer.Close()

	emojis, _ := ParseEmojis(MockDataUrlProvider{
		baseUrl: mockHttpServer.URL,
	}, map[string]Annotation{
		"😮‍💨": {
			Default: []string{"one"},
			Tts:     []string{"tts"},
		},
	})
	emoji := emojis["group1"][0]

	if emoji.Codepoints != "1F62E 200D 1F4A8" {
		t.Errorf("Failed to parse codepoint. Received %v, expected 😮‍💨", emoji.Codepoints)
	}
}

func TestParseEmojisSetsName(t *testing.T) {
	mockHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == mockUnicodeEmojisPath {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`# group: group1
1F600                                                  ; fully-qualified     # 😀 E1.0 grinning face`))
		}
	}))

	defer mockHttpServer.Close()

	emojis, _ := ParseEmojis(MockDataUrlProvider{
		baseUrl: mockHttpServer.URL,
	}, map[string]Annotation{
		"😀": {
			Default: []string{"one"},
			Tts:     []string{"grinning face"},
		},
	})
	emoji := emojis["group1"][0]

	if emoji.Name != "grinning face" {
		t.Errorf("Failed to parse emoji name. Received %v, expected grinning face", emoji.Name)
	}
}

func TestParseEmojisSetsFirstNameFromTtsList(t *testing.T) {
	mockHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == mockUnicodeEmojisPath {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`# group: group1
1F600                                                  ; fully-qualified     # 😀 E1.0 grinning face`))
		}
	}))

	defer mockHttpServer.Close()

	emojis, _ := ParseEmojis(MockDataUrlProvider{
		baseUrl: mockHttpServer.URL,
	}, map[string]Annotation{
		"😀": {
			Default: []string{"one"},
			Tts:     []string{"grinning face", "not a name to be used", "another name not to be used"},
		},
	})
	emoji := emojis["group1"][0]

	if emoji.Name != "grinning face" {
		t.Errorf("Failed to parse emoji name. Received %v, expected grinning face", emoji.Name)
	}
}

func TestParseEmojiSelectsFullyQualifiedEmojis(t *testing.T) {
	mockHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == mockUnicodeEmojisPath {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(` # group: group1
F636 200D 1F32B FE0F                                  ; fully-qualified     # 😶‍🌫️ E13.1 face in clouds
1F636 200D 1F32B                                       ; minimally-qualified # 😶‍🌫 E13.1 face in clouds
2620                                                   ; unqualified         # ☠ E1.0 skull and crossbones
`))
		}
	}))

	defer mockHttpServer.Close()

	emojis, _ := ParseEmojis(MockDataUrlProvider{
		baseUrl: mockHttpServer.URL,
	}, map[string]Annotation{
		"😀": {
			Default: []string{"one"},
			Tts:     []string{"tts"},
		},
	})
	emojisInGroup1 := emojis["group1"]

	if len(emojisInGroup1) != 1 {
		t.Errorf("Other emojis that are not fully qualified have been parsed")
	}
}

func TestParseEmojisGroupsEmojis(t *testing.T) {
	mockHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == mockUnicodeEmojisPath {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`# group: Smileys & Emotion
1F600                                                  ; fully-qualified     # 😀 E1.0 grinning face
1F603                                                  ; fully-qualified     # 😃 E0.6 grinning face with big eyes
1F636 200D 1F32B                                       ; minimally-qualified # 😶‍🌫 E13.1 face in clouds

# group: People & Body
1F44B                                                  ; fully-qualified     # 👋 E0.6 waving hand
1F44B 1F3FB                                            ; fully-qualified     # 👋🏻 E1.0 waving hand: light skin tone
1F590                                                  ; unqualified         # 🖐 E0.7 hand with fingers splayed
`))
		}
	}))

	defer mockHttpServer.Close()

	emojis, _ := ParseEmojis(MockDataUrlProvider{
		baseUrl: mockHttpServer.URL,
	}, map[string]Annotation{
		"😀": {
			Default: []string{"one"},
			Tts:     []string{"tts"},
		},
	})

	smileyAndEmotionsGroup, smileyAndEmotionOk := emojis["Smileys & Emotion"]

	if !smileyAndEmotionOk {
		t.Errorf("Could not parse Smileys & Emotion group")
	}

	if len(smileyAndEmotionsGroup) != 2 {
		t.Errorf("Expected 2 emojis in Smiley & Emotion group, received %v", len(smileyAndEmotionsGroup))
	}

	peopleAndBodyGroup, peopleAndBodyOk := emojis["People & Body"]

	if !peopleAndBodyOk {
		t.Errorf("Could not parse People & Body group")
	}

	if len(peopleAndBodyGroup) != 2 {
		t.Errorf("Expected 2 emojis in People & Body group, received %v", len(peopleAndBodyGroup))
	}
}

func TestParseEmojisSetsAnnotations(t *testing.T) {
	mockHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == mockUnicodeEmojisPath {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`# group: Smileys & Emotion
1F600                                                  ; fully-qualified     # 😀 E1.0 grinning face
1F636 200D 1F32B FE0F                                  ; fully-qualified     # 😶‍🌫️ E13.1 face in clouds
1F636 200D 1F32B                                       ; minimally-qualified # 😶‍🌫 E13.1 face in clouds
`))
		}
	}))

	defer mockHttpServer.Close()

	smileyAnnotations := []string{"face", "grin", "grinning face"}
	faceCloudAnnotations := []string{"absentminded", "face in clouds", "face in the fog", "head in clouds"}

	emojis, _ := ParseEmojis(MockDataUrlProvider{
		baseUrl: mockHttpServer.URL,
	}, map[string]Annotation{
		"😀": {
			Default: smileyAnnotations,
			Tts:     []string{"grinning face"},
		},
		"😶‍🌫️": {
			Default: faceCloudAnnotations,
			Tts:     []string{"face in clouds"},
		},
	})

	smileyAndEmotionsGroup, ok := emojis["Smileys & Emotion"]
	if !ok {
		t.Errorf("Could not parse Smileys & Emotion group")
	}

	smileyEmoji := smileyAndEmotionsGroup[0]
	if !areAnnotationsEqual(smileyEmoji.Annotations, smileyAnnotations) {
		t.Errorf("Failed to map annotations. Received %v, expected %v", smileyEmoji.Annotations, smileyAnnotations)

	}

	faceInCloudEmoji := smileyAndEmotionsGroup[1]
	if !areAnnotationsEqual(faceInCloudEmoji.Annotations, faceCloudAnnotations) {
		t.Errorf("Failed to map annotations. Received %v, expected %v", faceInCloudEmoji.Annotations, faceCloudAnnotations)

	}
}

func TestParseEmojisSetsCharacter(t *testing.T) {
	mockHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == mockUnicodeEmojisPath {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(` # group: group1
1F62E 200D 1F4A8                                       ; fully-qualified     # 😮‍💨 E13.1 face exhaling
1F44B 1F3FB                                            ; fully-qualified     # 👋🏻 E1.0 waving hand: light skin tone
`))
		}
	}))

	defer mockHttpServer.Close()

	emojis, _ := ParseEmojis(MockDataUrlProvider{
		baseUrl: mockHttpServer.URL,
	}, map[string]Annotation{
		"😮‍💨": {
			Default: []string{"one"},
			Tts:     []string{"tts"},
		},
	})

	faceExhalingEmoji := emojis["group1"][0]

	if faceExhalingEmoji.Character != "😮‍💨" {
		t.Errorf("Failed to get emoji character. Received %v, expected %v", faceExhalingEmoji.Character, "😮‍💨")
	}

	lightSkinWavingHandEmoji := emojis["group1"][1]

	if lightSkinWavingHandEmoji.Character != "👋🏻" {
		t.Errorf("Failed to get emoji character. Received %v, expected %v", lightSkinWavingHandEmoji.Character, "👋🏻")
	}
}

func TestFetchEmojiDataFileFails(t *testing.T) {
	mockHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == mockUnicodeEmojisPath {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Failed test request"))
		}
	}))

	defer mockHttpServer.Close()

	_, err := ParseEmojis(MockDataUrlProvider{
		baseUrl: mockHttpServer.URL,
	}, nil)

	if err == nil {
		t.Errorf("Expected parser to return error on a failed request")
	}
}
