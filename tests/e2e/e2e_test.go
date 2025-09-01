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
	config2 "github.com/malakagl/go-template/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	configPath string
	cfg        *config2.Config
	dbPool     *pgxpool.Pool
)

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
		drop table if exists kart_challenge_it.api_key_endpoints cascade;
		drop table if exists kart_challenge_it.api_keys cascade;
		drop table if exists kart_challenge_it.endpoints cascade;
		drop table if exists kart_challenge_it.products cascade;
		drop table if exists kart_challenge_it.orders cascade;
		drop table if exists kart_challenge_it.product_images cascade;
		drop table if exists kart_challenge_it.order_products cascade;
		drop table if exists kart_challenge_it.coupon_codes cascade;
		drop table if exists kart_challenge_it.files cascade;
		drop table if exists kart_challenge_it.schema_migrations cascade;
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
		log.Println("failed to set test schema", err)
		return
	}

	_, err = dbPool.Exec(ctx, `
		INSERT INTO products (name, price, category) 
			VALUES ('Chicken Waffle', 13.25, 'Waffle');
		INSERT INTO product_images (product_id, thumbnail, mobile, tablet, desktop) 
			VALUES (1, '1/thumbnail.jpg', '1/mobile.jpg', '1/tablet.jpg', '1/desktop.jpg');
		INSERT INTO api_keys (client_id, api_key) 
			VALUES ('q87w3qPEoFk','$2a$10$ju0j1dHc7zJf/nF.cz6WQ.TZHNfGTOXmXIdmBoem25uzhKYAaILYK'),
				   ('-G6a8hn7Rac','$2a$10$D594dp02sYvf336dnl606OVZPiPeL.NKtDA4FHS5FSWxWm909NSe2');
		INSERT INTO api_key_endpoints (api_key_id, endpoint_id, is_active) 
			VALUES (1,3,true), (1,4,true),(1,5,true);`)
	if err != nil {
		log.Println("seed api key endpoints for test client failed with error, ", err)
		return
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
			args:     args{apiKey: "q87w3qPEoFk.wiYU5t4RZHG_axVkKgKVFRexITBTdppZsKH6eKZFh8s"},
			expected: expected{statusCode: http.StatusOK, body: "OK"},
		},
		{
			name:     "success get one product",
			args:     args{apiKey: "q87w3qPEoFk.wiYU5t4RZHG_axVkKgKVFRexITBTdppZsKH6eKZFh8s", productId: "/1"},
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
			args:     arg{apiKey: "q87w3qPEoFk.wiYU5t4RZHG_axVkKgKVFRexITBTdppZsKH6eKZFh8s", productID: "1", couponCode: "invalid"},
			expected: expect{statusCode: http.StatusUnprocessableEntity, body: "invalid coupon code"},
		},
		{
			name:     "invalid request body",
			args:     arg{apiKey: "q87w3qPEoFk.wiYU5t4RZHG_axVkKgKVFRexITBTdppZsKH6eKZFh8s", productID: ""},
			expected: expect{statusCode: http.StatusBadRequest, body: "Invalid request dat"},
		},
		{
			name:     "success order",
			args:     arg{apiKey: "q87w3qPEoFk.wiYU5t4RZHG_axVkKgKVFRexITBTdppZsKH6eKZFh8s", productID: "1", couponCode: "FIFTYOFF"},
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

	client := &http.Client{Timeout: 120 * time.Second}
	req.Header.Set("x-api-key", apiKey)
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, string(respBody)
}
