package utils

import "math/big"

func WeiToETH(wei *big.Int) string {
	//1 ETH = 10^18 Wei
	base := new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)

	weiFloat := new(big.Float).SetInt(wei)
	baseFloat := new(big.Float).SetInt(base)

	ethValue := new(big.Float).Quo(weiFloat, baseFloat)

	return ethValue.Text('f', 18)
}
