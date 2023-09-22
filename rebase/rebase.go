package rebase

import (
	"math/big"
	"time"

	"math/rand"

	"github.com/ethereum/go-ethereum/log"
)

var BLOCKS_PER_EPOCH = big.NewInt(10) // 17280
var START_TX_GOAL = uint64(3)         // 10000
var EPOCH_TX_INCREMENT = uint64(1)    // 500
var INTEREST_PER_EPOCH = uint64(100087671)
var INITIAL_SUPPLY, _ = new(big.Int).SetString("50000000000000000000000000", 10)
var MAX_SUPPLY, _ = new(big.Int).SetString("250000000000000000000000000", 10)
var PERKS_MAX_EPOCHS = 720
var PERKS_EPOCH_COINS, _ = new(big.Int).SetString("13888888888888888889000", 10)
var PERKS_POOL = "0xbc1fBEbE0184446C9aac772E58A085b7cf13B543"

var UINT64_DIVISOR = uint64(100000000)
var DIVISOR = new(big.Int).Exp(big.NewInt(10), big.NewInt(8), nil)

var rndSemaphore = uint64(0)

type RebaseInfo struct {
	Epoch    uint64
	EpochTx  uint64
	Rbx      uint64
	RbxEpoch uint64
	Supply   *big.Int
	Tx       uint64
}

func GetRebasedAmount(amount *big.Int, rbx uint64) *big.Int {
	rebasedAmount := new(big.Int).Mul(amount, new(big.Int).SetUint64(rbx))
	//log.Warn("GetRebased", "rebased", rebasedAmount, "rbx", rbx)
	rebasedAmount.Div(rebasedAmount, DIVISOR)

	//	log.Info("RebasedAmount", "address", "block", blockNumber, "rebase", rebase, "amount", amount, "rebasedAmount", rebasedAmount)
	return rebasedAmount
}

func GetTransferAmount(amount *big.Int, rbx uint64) *big.Int {

	expandAmount := new(big.Int).Mul(amount, DIVISOR)
	senderAmount := new(big.Int).Div(expandAmount, new(big.Int).SetUint64(rbx))

	return senderAmount
}

func ProcessRebase(blockNumber *big.Int, last RebaseInfo, current RebaseInfo) (uint64, uint64, uint64, uint64, *big.Int) {

	epoch := last.Epoch
	epochTx := last.EpochTx
	rbx := last.Rbx
	rbxEpoch := last.RbxEpoch
	supply := GetRebasedAmount(INITIAL_SUPPLY, current.Rbx)

	// A new epoch occurs when the block number is evenly divisible by Blocks_Per_Epoch
	newEpoch := new(big.Int).Mod(blockNumber, BLOCKS_PER_EPOCH)
	if newEpoch.Sign() == 0 {

		// Epoch increment is conditional on meeting at least 75% of TX goal for epoch
		txGoal := START_TX_GOAL + (rbxEpoch * EPOCH_TX_INCREMENT)

		// Only for testing, we compute the transactions as a random number
		if rndSemaphore == 0 {
			rand.Seed(time.Now().UnixNano()) // Seed the random number generator
			epochTx = uint64(0 + rand.Intn(int(txGoal+(txGoal/2))))
			rndSemaphore = epochTx
		} else {
			epochTx = rndSemaphore
			rndSemaphore = 0
		}

		txRatio := (epochTx * 100) / txGoal

		// TX Goal was met or exceeded
		if txRatio >= 75 {
			// Upper limit is 125%
			if txRatio > 125 {
				txRatio = 125
			}

			// Increment the rebase epoch
			rbxEpoch = rbxEpoch + 1
			interest := ((INTEREST_PER_EPOCH - UINT64_DIVISOR) * txRatio / 100) + UINT64_DIVISOR

			rbx = rbx * interest / UINT64_DIVISOR

			// Add perks coins

			log.Warn("Rebase Success 🎉🎉🎉", "Epoch", epoch, "RbxEpoch", rbxEpoch, "Rbx", rbx, "Ratio", txRatio)

		} else {
			log.Warn("Rebase Skipped 🙁", "Goal", txGoal, "TX", epochTx, "Ratio", txRatio)
		}

		// At every epoch the transaction count is always reset
		epochTx = current.Tx
		epoch = epoch + 1

	} else {

		epochTx = epochTx + current.Tx

	}

	return epoch, epochTx, rbx, rbxEpoch, supply
}
