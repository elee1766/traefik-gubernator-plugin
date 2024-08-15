package traefik_gubernator_plugin

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type gubernatorClient struct {
	endpoint string
	headers  map[string]string

	client http.Client
}

func newClient(conf *Config) (*gubernatorClient, error) {
	gc := &gubernatorClient{}

	gc.endpoint = conf.Remote
	gc.headers = conf.Headers
	if gc.headers == nil {
		gc.headers = map[string]string{}
	}
	return gc, nil
}

func (c *gubernatorClient) GetRateLimits(ctx context.Context, req *GetRateLimitsReq) (*GetRateLimitsResp, error) {

	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Post(c.endpoint, "application/json", bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode > 299 || resp.StatusCode < 200 {
		return nil, fmt.Errorf("gubernator http error: %s", resp.Status)
	}
	rez := &GetRateLimitsResp{}
	err = json.NewDecoder(resp.Body).Decode(rez)
	if err != nil {
		return nil, fmt.Errorf("unmarshal gubernator response: %w", err)
	}
	return rez, nil
}

func (c *gubernatorClient) Close() {
}
