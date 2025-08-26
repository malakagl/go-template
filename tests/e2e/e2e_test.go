package e2e

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	config2 "github.com/malakagl/kart-challenge/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var configPath string
var cfg *config2.Config
var dbPool *pgxpool.Pool

func TestMain(m *testing.M) {
	loadConfig()
	if !waitForPostgres() {
		log.Fatal("timed out waiting for postgres")
	}

	seedPostgresData()
	code := m.Run()
	tearDownTestData()

	os.Exit(code)
}

func tearDownTestData() {
	log.Println("tearing down test database")
	ctx := context.Background()
	_, err := dbPool.Exec(ctx, `
        TRUNCATE TABLE 
            kart_challenge_it.products, 
            kart_challenge_it.orders
        RESTART IDENTITY CASCADE;
    `)
	if err != nil {
		log.Println("tear down data encountered error, ", err)
	}
}

func loadConfig() {
	flag.StringVar(&configPath, "config", "./config/config.default.yaml", "Path to config file")
	flag.Parse()
	log.Println("Loading config from ", configPath)
	var err error
	cfg, err = config2.LoadConfig(configPath)
	if err != nil {
		log.Fatal(err)
	}
}

func waitForPostgres() bool {
	var db *pgxpool.Pool
	poolMaxWait := 60 * time.Second
	poolStart := time.Now()
	for {
		dbURL := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
			cfg.Database.User, cfg.Database.Password, "localhost", cfg.Database.Port, cfg.Database.Name)
		var err error
		db, err = pgxpool.New(context.Background(), dbURL)
		if err == nil {
			err = db.Ping(context.Background())
		}
		if err == nil {
			log.Println("Postgres is ready!")
			break
		}

		if time.Since(poolStart) > poolMaxWait {
			log.Fatalf("Postgres did not start in %v: %v", poolMaxWait, err)
		}

		log.Println("Waiting for Postgres...", err)
		time.Sleep(1 * time.Second)
	}
	dbPool = db
	return db != nil
}

func seedPostgresData() {
	ctx := context.Background()
	_, err := dbPool.Exec(ctx, "SET search_path TO kart_challenge_it")
	if err != nil {
		log.Fatal("failed to set test schema", err)
	}

	_, err = dbPool.Exec(ctx, `INSERT INTO products (name, price, category) 
								VALUES ('Chicken Waffle', 13.25, 'Waffle')`)
	if err != nil {
		log.Fatal("seed product failed with error, ", err)
	}

	_, err = dbPool.Exec(ctx, `INSERT INTO product_images (product_id, thumbnail, mobile, tablet, desktop) 
								VALUES (1, '1/thumbnail.jpg', '1/mobile.jpg', '1/tablet.jpg', '1/desktop.jpg')`)
	if err != nil {
		log.Fatal("seed product images failed with error, ", err)
	}
}

func TestProductsAPI(t *testing.T) {
	type args struct {
		productId string
		apiKey    string
	}
	type expected struct {
		statusCode int
		body       string
	}
	tests := []struct {
		name     string
		args     args
		expected expected
	}{
		{
			name:     "invalid api key",
			args:     args{apiKey: "invalid"},
			expected: expected{statusCode: http.StatusUnauthorized, body: "Unauthorized"},
		},
		{
			name:     "success get all products",
			args:     args{apiKey: "apitest"},
			expected: expected{statusCode: http.StatusOK, body: "OK"},
		},
		{
			name:     "success get one product",
			args:     args{apiKey: "apitest", productId: "/1"},
			expected: expected{statusCode: http.StatusOK, body: "OK"},
		},
	}
	for _, tt := range tests {
		url := fmt.Sprintf("http://%s:%d/products"+tt.args.productId, cfg.Server.Host, cfg.Server.Port)
		status, body := doRequest(t, http.MethodGet, url, tt.args.apiKey, nil)
		assert.Equal(t, tt.expected.statusCode, status, tt.name)
		assert.Contains(t, body, tt.expected.body, tt.name)
	}
}

func TestOrderAPI(t *testing.T) {
	type arg struct {
		couponCode string
		productID  string
		apiKey     string
	}
	type expect struct {
		statusCode int
		body       string
	}
	tests := []struct {
		name     string
		args     arg
		expected expect
	}{
		{
			name:     "invalid api key",
			args:     arg{apiKey: "invalid"},
			expected: expect{statusCode: http.StatusUnauthorized, body: "Unauthorized"},
		},
		{
			name:     "invalid coupon code",
			args:     arg{apiKey: "create_order", productID: "1", couponCode: "invalid"},
			expected: expect{statusCode: http.StatusUnprocessableEntity, body: "invalid coupon code"},
		},
		{
			name:     "invalid request body",
			args:     arg{apiKey: "create_order", productID: ""},
			expected: expect{statusCode: http.StatusBadRequest, body: "Invalid request dat"},
		},
		{
			name:     "success order",
			args:     arg{apiKey: "create_order", productID: "1", couponCode: "FIFTYOFF"},
			expected: expect{statusCode: http.StatusCreated, body: "OK"},
		},
	}
	for _, tt := range tests {
		url := fmt.Sprintf("http://%s:%d/orders", cfg.Server.Host, cfg.Server.Port)
		b := []byte(`{
    			"couponCode": "` + tt.args.couponCode + `",
    			"items": [
        			{
            			"productId": "` + tt.args.productID + `",
            			"quantity": 10
        			}
    			]
			}`)
		status, body := doRequest(t, http.MethodPost, url, tt.args.apiKey, b)
		assert.Equal(t, tt.expected.statusCode, status, tt.name)
		assert.Contains(t, body, tt.expected.body, tt.name)
	}
}

func doRequest(t *testing.T, method, url, apiKey string, body []byte) (int, string) {
	req, err := http.NewRequest(method, url, bytes.NewReader(body))
	require.NoError(t, err)

	client := &http.Client{Timeout: 60 * time.Second}
	req.Header.Set("x-api-key", apiKey)
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, string(respBody)
}
