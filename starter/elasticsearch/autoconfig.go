// Package elasticsearch 提供 Elasticsearch 搜索引擎自动配置。
package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"

	"github.com/xudefa/enhance/boot"
	"github.com/xudefa/enhance/condition"
	"github.com/xudefa/enhance/config/environment"
	"github.com/xudefa/enhance/core"
	"github.com/xudefa/enhance/log"
)

func init() {
	boot.RegisterAutoConfigWith(&ElasticsearchAutoConfiguration{},
		boot.WithConditions(
			condition.OnProperty(ElasticsearchEnabled, ConditionTrue),
		),
		boot.WithOrder(int(boot.OrderPriorityDataLayer)),
	)
}

// ElasticsearchAutoConfiguration Elasticsearch 自动配置类。
type ElasticsearchAutoConfiguration struct {
	logger log.Logger
	client *elasticsearch.Client
	config *ElasticsearchConfig
}

// Configure 配置 Elasticsearch 连接。
func (c *ElasticsearchAutoConfiguration) Configure(ctx boot.ApplicationContext) error {
	env := ctx.Environment()

	if logger, err := core.GetByName[log.Logger](ctx.Container(), ""); err == nil {
		c.logger = logger
	} else {
		c.logger = log.Build()
	}

	cfg, err := c.loadConfig(env)
	if err != nil {
		return fmt.Errorf("failed to load Elasticsearch config: %w", err)
	}

	c.config = cfg

	addresses := make([]string, len(cfg.Addresses))
	for i, addr := range cfg.Addresses {
		if !strings.HasPrefix(addr, "http") {
			addresses[i] = fmt.Sprintf("http://%s", addr)
		} else {
			addresses[i] = addr
		}
	}

	esCfg := elasticsearch.Config{
		Addresses: addresses,
		Username:  cfg.Username,
		Password:  cfg.Password,
		Transport: &http.Transport{
			MaxIdleConnsPerHost:   cfg.MaxIdleConns,
			ResponseHeaderTimeout: time.Duration(cfg.Timeout) * time.Second,
		},
	}

	client, err := elasticsearch.NewClient(esCfg)
	if err != nil {
		return fmt.Errorf("failed to create Elasticsearch client: %w", err)
	}

	res, err := client.Ping()
	if err != nil {
		return fmt.Errorf("failed to connect to Elasticsearch: %w", err)
	}
	defer res.Body.Close()

	c.client = client

	if err := ctx.Container().RegisterInstance(c.client, reflect.TypeFor[*elasticsearch.Client]()); err != nil {
		return fmt.Errorf("failed to register Elasticsearch Client: %w", err)
	}

	c.logger.Info(ctx.Context(), "Elasticsearch connected successfully",
		log.KeyValue{Key: "addresses", Value: cfg.Addresses},
	)

	return nil
}

// GetClient 获取 Elasticsearch Client 实例。
func (c *ElasticsearchAutoConfiguration) GetClient() *elasticsearch.Client {
	return c.client
}

// Index 索引文档。
func (c *ElasticsearchAutoConfiguration) Index(ctx context.Context, index string, doc map[string]interface{}, id string) error {
	// 使用 JSON 序列化文档
	data, err := json.Marshal(doc)
	if err != nil {
		return fmt.Errorf("failed to serialize document as JSON: %w", err)
	}

	req := esapi.IndexRequest{
		Index:      index,
		DocumentID: id,
		Body:       bytes.NewReader(data),
	}

	res, err := req.Do(ctx, c.client)
	if err != nil {
		return fmt.Errorf("failed to index document: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("failed to index document, status: %d, response: %s", res.StatusCode, res.String())
	}

	return nil
}

// Search 搜索文档。
func (c *ElasticsearchAutoConfiguration) Search(ctx context.Context, index string, query map[string]interface{}) (*esapi.Response, error) {
	return c.client.Search(
		c.client.Search.WithIndex(index),
		c.client.Search.WithBody(nil),
	)
}

// ElasticsearchConfig Elasticsearch 配置。
type ElasticsearchConfig struct {
	Enabled      bool     `json:"enabled" mapstructure:"enabled"`
	Addresses    []string `json:"addresses" mapstructure:"addresses"`
	Username     string   `json:"username" mapstructure:"username"`
	Password     string   `json:"password" mapstructure:"password"`
	MaxIdleConns int      `json:"max-idle-conns" mapstructure:"max-idle-conns"`
	Timeout      int      `json:"timeout" mapstructure:"timeout"`
}

// 配置常量。
const (
	ElasticsearchEnabled = "elasticsearch.enabled"
	DefaultTimeout       = 10
	DefaultMaxIdleConns  = 10
	ConditionTrue        = "true"
)

// loadConfig 从 Environment 加载 Elasticsearch 配置。
func (c *ElasticsearchAutoConfiguration) loadConfig(env *environment.Environment) (*ElasticsearchConfig, error) {
	cfg := &ElasticsearchConfig{
		Timeout:      DefaultTimeout,
		MaxIdleConns: DefaultMaxIdleConns,
	}

	if err := env.BindPrefix("elasticsearch", cfg); err != nil {
		return nil, fmt.Errorf("failed to bind Elasticsearch config: %w", err)
	}

	if len(cfg.Addresses) == 0 {
		return nil, fmt.Errorf("elasticsearch addresses must not be empty")
	}

	return cfg, nil
}
