package anilist

import (
	"context"
	"github.com/Yamashou/gqlgenc/clientv2"
	"net/http"
)

func NewAuthedClient(token string) *Client {
	return &Client{
		Client: clientv2.NewClient(http.DefaultClient, "https://graphql.anilist.co", nil,
			func(ctx context.Context, req *http.Request, gqlInfo *clientv2.GQLRequestInfo, res interface{}, next clientv2.RequestInterceptorFunc) error {
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Accept", "application/json")
				if len(token) > 0 {
					req.Header.Set("Authorization", "Bearer "+token)
				}
				return next(ctx, req, gqlInfo, res)
			}),
	}
}
