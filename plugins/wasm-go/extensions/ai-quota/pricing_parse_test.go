package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseTokenPriceMicroYuanPerToken(t *testing.T) {
	t.Run("accepts zero modern price", func(t *testing.T) {
		value, err := parseTokenPriceMicroYuanPerToken(map[string]string{
			"input_price_micro_yuan_per_token": "0",
		}, "input_price_micro_yuan_per_token", "input_price_per_1k_micro_yuan")
		require.NoError(t, err)
		require.Zero(t, value)
	})

	t.Run("accepts zero legacy price", func(t *testing.T) {
		value, err := parseTokenPriceMicroYuanPerToken(map[string]string{
			"output_price_per_1k_micro_yuan": "0",
		}, "output_price_micro_yuan_per_token", "output_price_per_1k_micro_yuan")
		require.NoError(t, err)
		require.Zero(t, value)
	})

	t.Run("rejects negative values", func(t *testing.T) {
		_, err := parseTokenPriceMicroYuanPerToken(map[string]string{
			"output_price_micro_yuan_per_token": "-1",
		}, "output_price_micro_yuan_per_token", "output_price_per_1k_micro_yuan")
		require.Error(t, err)
	})
}
