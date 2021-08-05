package types

import (
	globalTypes "github.com/barnbridge/internal-api/types"
)

type Pool struct {
	PoolName    string            `json:"poolName"`
	PoolAddress string            `json:"poolAddress"`
	PoolToken   globalTypes.Token `json:"poolToken"`

	JuniorTokenAddress string `json:"juniorTokenAddress"`
	SeniorTokenAddress string `json:"seniorTokenAddress"`

	OracleAddress     string `json:"oracleAddress"`
	OracleAssetSymbol string `json:"oracleAssetSymbol"`

	Epoch1Start   int64 `json:"epoch1Start"`
	EpochDuration int64 `json:"epochDuration"`

	State PoolState `json:"state"`
}
