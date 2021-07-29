package smartexposure

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"

	"github.com/barnbridge/internal-api/query"
	"github.com/barnbridge/internal-api/response"
	"github.com/barnbridge/internal-api/smartexposure/types"
	"github.com/barnbridge/internal-api/utils"
)

func (s *SmartExposure) handleTransactions(ctx *gin.Context) {
	builder := query.New()
	poolAddress := ctx.DefaultQuery("poolAddress", "")
	if poolAddress != "" {
		poolAddress, err := utils.ValidateAccount(poolAddress)
		if err != nil {
			response.Error(ctx, err)
			return
		}
		err, exists := s.checkPoolExists(ctx, poolAddress)
		if err != nil {
			response.Error(ctx, err)
			return
		}

		if !exists {
			response.NotFound(ctx)
			return
		}

		builder.Filters.Add("(select pool_address from smart_exposure.tranches t where t.etoken_address = etoken_address)", poolAddress)
	}

	eTokenAddress := ctx.DefaultQuery("eTokenAddress", "")
	if eTokenAddress != "" {
		eTokenAddress, err := utils.ValidateAccount(eTokenAddress)
		if err != nil {
			response.BadRequest(ctx, err)
			return
		}

		err, exists := s.checkTrancheExists(ctx, eTokenAddress)
		if err != nil {
			response.Error(ctx, err)
			return
		}

		if !exists {
			response.NotFound(ctx)
			return
		}

		builder.Filters.Add("etoken_address", eTokenAddress)
	}

	accountAddress := strings.ToLower(ctx.DefaultQuery("accountAddress", ""))
	if accountAddress != "" {
		accountAddress, err := utils.ValidateAccount(accountAddress)
		if err != nil {
			response.BadRequest(ctx, errors.New("invalid accountAddress"))
			return
		}
		builder.Filters.Add("user_address", accountAddress)
	}

	transactionType := strings.ToUpper(ctx.DefaultQuery("transactionType", "ALL"))
	if transactionType != "ALL" {
		if !checkTxType(transactionType) {
			response.BadRequest(ctx, errors.New("invalid transaction type"))
			return
		}
		builder.Filters.Add("transaction_type", transactionType)
	}

	err := builder.SetLimitFromCtx(ctx)
	if err != nil {
		response.BadRequest(ctx, err)
		return
	}

	err = builder.SetOffsetFromCtx(ctx)
	if err != nil {
		response.BadRequest(ctx, err)
		return
	}

	q, params := builder.WithPaginationFromCtx(ctx).Run(`
		select t.etoken_address,
			   t.user_address,
			   p.token_a_address,
			   p.token_a_symbol,
			   p.token_a_decimals,
			   p.token_b_address,
			   p.token_b_symbol,
			   p.token_b_decimals,
			   t.amount_a,
			   t.amount_b,
			   t.amount,
			   t.transaction_type,
			   t.tx_hash,
			   t.block_timestamp,
			   t.included_in_block,
			   coalesce((select price_usd
						 from public.token_prices tp
						 where token_address = (select token_a_address
												from smart_exposure.pools
												where pool_address = (select pool_address
																	  from smart_exposure.tranches
																	  where etoken_address = etoken_address))
						   and tp.block_timestamp <= t.block_timestamp
						 limit 1), 0)                                                                    as token_a_price,
			   coalesce((select price_usd
						 from public.token_prices tp
						 where token_address = (select token_b_address
												from smart_exposure.pools
												where pool_address = (select pool_address
																	  from smart_exposure.tranches
																	  where etoken_address = etoken_address))
						   and tp.block_timestamp <= t.block_timestamp
						 limit 1), 0)                                                                    as token_b_price,
			   coalesce((select etoken_price
						 from smart_exposure.tranche_state ts
						 where ts.etoken_address = t.etoken_address
						   and ts.block_timestamp <= t.block_timestamp
						 limit 1), 0)                                                                    as etoken_price,
			   (select etoken_symbol from smart_exposure.tranches where etoken_address = etoken_address) as etoken_symbol
		from smart_exposure.transaction_history t
				 inner join smart_exposure.pools p on pool_address = (select pool_address
																	  from smart_exposure.tranches
																	  where etoken_address = t.etoken_address)
		$filters$
		order by included_in_block desc, tx_index desc, log_index desc
		$offset$ $limit$`)

	rows, err := s.db.Connection().Query(ctx, q, params...)
	if err != nil && err != pgx.ErrNoRows {
		response.Error(ctx, err)
		return
	}
	defer rows.Close()
	var history []types.Transaction

	for rows.Next() {
		var h types.Transaction
		err := rows.Scan(&h.ETokenAddress, &h.AccountAddress, &h.TokenA.TokenAddress, &h.TokenA.TokenSymbol, &h.TokenA.TokenDecimals, &h.TokenB.TokenAddress, &h.TokenB.TokenSymbol, &h.TokenB.TokenDecimals,
			&h.AmountA, &h.AmountB, &h.AmountEToken, &h.TransactionType, &h.TransactionHash, &h.BlockTimestamp, &h.BlockNumber, &h.TokenAPrice, &h.TokenBPrice, &h.ETokenPrice, &h.ETokenSymbol)
		if err != nil {
			response.Error(ctx, err)
			return
		}

		history = append(history, h)
	}

	q, params = builder.Run(`select count(*) from smart_exposure.transaction_history $filters$`)
	var count int64

	err = s.db.Connection().QueryRow(ctx, q, params...).Scan(&count)
	if err != nil {
		response.Error(ctx, err)
		return
	}

	response.OKWithBlock(ctx, s.db, history, response.Meta().Set("count", count))
}
