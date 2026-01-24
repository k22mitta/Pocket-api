package plaidclient

import (
	"github.com/plaid/plaid-go/v29/plaid"
)

type Client struct {
	api *plaid.APIClient
}

func NewClient(clientID, secret, env string) *Client {
	cfg := plaid.NewConfiguration()
	cfg.AddDefaultHeader("PLAID-CLIENT-ID", clientID)
	cfg.AddDefaultHeader("PLAID-SECRET", secret)

	switch env {
	case "production":
		cfg.UseEnvironment(plaid.Production)
	case "development":
		cfg.UseEnvironment(plaid.Environment("https://development.plaid.com"))
	default:
		cfg.UseEnvironment(plaid.Sandbox)
	}

	return &Client{api: plaid.NewAPIClient(cfg)}
}

func (c *Client) API() *plaid.APIClient {
	return c.api
}
