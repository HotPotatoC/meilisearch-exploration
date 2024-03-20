package main

import (
	"context"
	"os"

	"github.com/HotPotatoC/meilisearch-exploration/config"
	"github.com/gofiber/fiber/v2"
	_ "github.com/joho/godotenv/autoload"
	"github.com/meilisearch/meilisearch-go"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
)

func init() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack

	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	log.Logger = log.With().Caller().Stack().Logger().Output(zerolog.ConsoleWriter{Out: os.Stderr})

	log.Info().Msg("Logger initialized")
	config.Init()
	log.Info().Msg("Config initialized")
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	log.Info().Msg("Starting application...")

	search := NewMeilisearchClient()
	if !search.IndexExists("products") {
		go func() {
			products, categoryMap, err := LoadProductData()
			if err != nil {
				log.Fatal().Err(err).Msg("failed to load product data")
				return
			}

			log.Info().Msgf("Loaded %d products", len(products))
			log.Info().Msgf("Loaded %d categories", len(categoryMap))

			if err := DumpProducts(search, products); err != nil {
				cancel()
				log.Fatal().Err(err).Msg("failed to dump products into index")
			}
		}()
	}

	log.Info().Msg("Starting HTTP Server...")

	srv := NewHTTPServer()

	srv.app.Get("/search", func(c *fiber.Ctx) error {
		query := c.Query("query", "")
		limit := c.QueryInt("limit", 10)
		page := c.QueryInt("page", 1)

		response, err := search.client.Index("products").Search(query, &meilisearch.SearchRequest{
			AttributesToHighlight: []string{"*"},
			Page:                  int64(page),
			Limit:                 int64(limit),
			HitsPerPage:           int64(limit),
		})
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}

		return c.JSON(response)
	})

	signal := <-srv.Start(config.Host(), config.Port())
	srv.Shutdown(ctx, signal)
}
