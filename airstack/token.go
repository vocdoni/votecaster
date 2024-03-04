package airstack

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	gql "github.com/vocdoni/vote-frame/airstack/graphql"
)

// TokenDetails wraps useful information about a token
type TokenDetails struct {
	Decimals    int8
	Name        string
	Symbol      string
	TotalSupply *big.Int
}

// GetTokenDetails gets a token details given its address and blockchain
func (c *Client) GetTokenDetails(
	tokenAddress common.Address,
	blockchain gql.TokenBlockchain,
) (*TokenDetails, error) {
	cctx, cancel := context.WithTimeout(c.ctx, apiTimeout)
	defer cancel()
	resp, err := gql.GetTokenDetails(cctx, c.Client, tokenAddress, blockchain)
	if err != nil {
		return nil, fmt.Errorf("cannot get token details from Airstack: %w", err)
	}
	totalSupply := new(big.Int)
	totalSupply.SetString(resp.Token.TotalSupply, 10)
	return &TokenDetails{
		Decimals:    int8(resp.Token.Decimals),
		Name:        resp.Token.Name,
		Symbol:      resp.Token.Symbol,
		TotalSupply: totalSupply,
	}, nil
}

// TokenHolder wraps a token holder with its address and balance of a certain token
type TokenHolder struct {
	Address common.Address
	Balance *big.Int
}

// getTokenBalances is a wrapper around the generated function for GraphQL query GetTokenBalances.
func (c *Client) getTokenBalances(
	tokenAddress common.Address,
	blockchain gql.TokenBlockchain,
	limit int,
	cursor string,
) (*gql.GetTokenBalancesResponse, error) {
	cctx, cancel := context.WithTimeout(c.ctx, apiTimeout)
	defer cancel()
	return gql.GetTokenBalances(cctx, c.Client, tokenAddress, blockchain, limit, cursor)
}

// GetTokenBalances gets all the token holders of a given token in a given chain
// calling the Airstack API. This function also takes care of Airstack API pagination.
func (c *Client) GetTokenBalances(tokenAddress common.Address, blockchain gql.TokenBlockchain) ([]*TokenHolder, error) {
	hasNextPage := true
	cursor := ""
	th := make([]*TokenHolder, 0)
	for hasNextPage {
		resp, err := c.getTokenBalances(tokenAddress, blockchain, airstackAPIlimit, cursor)
		if err != nil {
			return nil, fmt.Errorf("cannot get token balances from Airstack: %w", err)
		}
		for _, holder := range resp.TokenBalances.TokenBalance {
			balance := new(big.Int)
			balance.SetString(holder.Amount, 10)
			th = append(th, &TokenHolder{
				Address: holder.Owner.Addresses[0],
				Balance: balance,
			})
		}
		cursor = resp.TokenBalances.PageInfo.NextCursor
		if resp.TokenBalances.PageInfo.NextCursor == "" {
			hasNextPage = false
		}
	}
	return th, nil
}
