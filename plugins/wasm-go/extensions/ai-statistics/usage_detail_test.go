package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCountJSONOutputImagesOpenAIImageResponse(t *testing.T) {
	body := []byte(`{
		"created": 1776767451,
		"data": [
			{"url": "https://example.com/a.png"},
			{"url": "https://example.com/b.png"}
		]
	}`)

	require.Equal(t, int64(2), countJSONOutputImages(body))
}
