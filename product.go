package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"runtime"
	"sync"

	"github.com/jszwec/csvutil"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type Product struct {
	ASIN         string           `csv:"asin" json:"asin"`
	Title        string           `csv:"title" json:"title"`
	ProductURL   string           `csv:"productURL" json:"product_url"`
	Stars        float64          `csv:"stars" json:"stars"`
	Reviews      int              `csv:"reviews" json:"reviews"`
	Price        float64          `csv:"price" json:"price"`
	ListPrice    float64          `csv:"listPrice" json:"list_price"`
	CategoryID   int              `csv:"category_id" json:"category_id"`
	Category     *ProductCategory `json:"category"`
	IsBestSeller bool             `csv:"isBestSeller" json:"is_best_seller"`
}

type ProductCategory struct {
	ID   int    `csv:"id" json:"id"`
	Name string `csv:"category_name" json:"name"`
}

const DATA_SIZE = 1_426_337

func LoadProductData() ([]*Product, map[int]*ProductCategory, error) {
	log.Info().Msg("Loading products and categories...")
	productsFilePath := "data/amazon_products.csv"
	categoriesFilePath := "data/amazon_categories.csv"

	products := make([]*Product, DATA_SIZE)
	categoryMap := make(map[int]*ProductCategory)

	categoriesFile, err := os.Open(categoriesFilePath)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed reading categories file")
	}
	defer categoriesFile.Close()

	productsFile, err := os.Open(productsFilePath)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed reading products file")
	}
	defer productsFile.Close()

	categoriesDecoder, err := csvutil.NewDecoder(csv.NewReader(categoriesFile))
	if err != nil {
		return nil, nil, err
	}

	for {
		var category *ProductCategory
		if err := categoriesDecoder.Decode(&category); err == io.EOF {
			break
		} else if err != nil {
			return nil, nil, errors.Wrap(err, "failed decoding category")
		}

		if _, ok := categoryMap[category.ID]; ok {
			continue
		}

		categoryMap[category.ID] = category
	}

	productsDecoder, err := csvutil.NewDecoder(csv.NewReader(productsFile))
	if err != nil {
		return nil, nil, err
	}

	i := 0
	for {
		var product *Product
		if err := productsDecoder.Decode(&product); err == io.EOF {
			break
		} else if err != nil {
			return nil, nil, errors.Wrap(err, "failed decoding product")
		}

		product.Category = categoryMap[product.CategoryID]
		products[i] = product
		i++
	}

	nPartition := 5
	productsPerPartition := len(products) / nPartition
	for i := 0; i < nPartition; i++ {
		idx := i * productsPerPartition
		log.Info().Dict(fmt.Sprintf("products[%d]_%s", idx, products[idx].ASIN), zerolog.Dict().
			Str("name", products[idx].Title).
			Float64("price", products[idx].Price).
			Any("category", products[idx].Category),
		).Send()
	}

	return products, categoryMap, nil
}

func DumpProducts(search *MeilisearchClient, products []*Product) error {
	index, _ := search.client.GetIndex("products")
	if index != nil {
		return nil
	}

	index = search.client.Index("products")

	n := (len(products) + runtime.NumCPU() - 1) / runtime.NumCPU()
	chunks := splitProductsIntoChunks(products, n)

	log.Info().Int("chunks_length", len(chunks)).Send()
	for i, chunk := range chunks {
		log.Info().Int(fmt.Sprintf("chunk[%d].length", i), len(chunk)).Send()
	}

	var wg sync.WaitGroup
	wg.Add(n)
	for i, chunk := range chunks {
		go func(i int, chunk []*Product) {
			log.Info().Msg("Dumping into products index...")
			result, err := index.AddDocuments(chunk)
			if err != nil {
				log.Fatal().Err(err).Msg("failed to dump products into index")
			}

			log.Info().Any("dump_result", result).Send()
		}(i, chunk)
	}

	wg.Wait()

	return nil
}

func splitProductsIntoChunks(products []*Product, chunkSize int) [][]*Product {
	divided := make([][]*Product, (len(products)+chunkSize-1)/chunkSize)
	prev := 0
	i := 0
	till := len(products) - chunkSize
	for prev < till {
		next := prev + chunkSize
		divided[i] = products[prev:next]
		prev = next
		i++
	}
	divided[i] = products[prev:]
	return divided
}
