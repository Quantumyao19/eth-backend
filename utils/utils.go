package utils

import (
	"math"
	"math/big"
)

func WeiToETH(wei *big.Int) string {
	//1 ETH = 10^18 Wei
	base := new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)

	weiFloat := new(big.Float).SetInt(wei)
	baseFloat := new(big.Float).SetInt(base)

	ethValue := new(big.Float).Quo(weiFloat, baseFloat)

	return ethValue.Text('f', 18)
}

func FormatTokenAmount(value *big.Int, decimals uint8) string {
	divisor := new(big.Float).SetFloat64(math.Pow10(int(decimals)))
	v := new(big.Float).SetInt(value)
	result := new(big.Float).Quo(v, divisor)
	return result.Text('f', int(decimals))
}
