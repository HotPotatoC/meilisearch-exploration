package main

import (
	"github.com/HotPotatoC/meilisearch-exploration/config"
	"github.com/meilisearch/meilisearch-go"
	"github.com/rs/zerolog/log"
)

type MeilisearchClient struct {
	client *meilisearch.Client
}

func NewMeilisearchClient() *MeilisearchClient {
	client := meilisearch.NewClient(meilisearch.ClientConfig{
		Host:   config.MeiliHost(),
		APIKey: config.MeiliMasterkey(),
	})

	health, err := client.Health()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to meilisearch")
	}

	log.Info().Str("status", health.Status).Msg("connected to meilisearch")

	return &MeilisearchClient{client: client}
}

func (mc *MeilisearchClient) IndexExists(name string) bool {
	index, err := mc.client.GetIndex(name)
	if err != nil {
		return false
	}

	return index != nil
}
