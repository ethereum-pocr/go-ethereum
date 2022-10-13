package cliquepcr
import (
	"math/big"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm/runtime"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/params"
)
func NewCarbonFootPrintContract(nodeAddress common.Address, config *params.ChainConfig, state *state.StateDB, header *types.Header) CarbonFootprintContract {
	contract := CarbonFootprintContract{}
	contract.ContractAddress = common.HexToAddress(proofOfCarbonReductionContractAddress)
	block := big.NewInt(0).Sub(header.Number, big.NewInt(1))
	stateCopy := state.Copy() // necessary to work on the copy of the state when performing a call
	cfg := runtime.Config{ChainConfig: config, Origin: nodeAddress, GasLimit: 1000000, State: stateCopy, BlockNumber: block}
	contract.RuntimeConfig = &cfg
	return contract
}
func (contract *CarbonFootprintContract) totalFootprint() (*big.Int, error) {
	input := common.Hex2Bytes("b6c3dcf8")
	result, _, err := runtime.Call(contract.ContractAddress, input, contract.RuntimeConfig)
	// log.Info("Result/Err", "Result", common.Bytes2Hex(result), "Err", err.Error())
	if err != nil {
		log.Error("Impossible to get the total carbon footprint", "err", err.Error(), "block", contract.RuntimeConfig.BlockNumber.Int64())
		return nil, err
	} else {
		// log.Info("Total Carbon footprint", "result", common.Bytes2Hex(result))
		return common.BytesToHash(result).Big(), nil
	}
}
func (contract *CarbonFootprintContract) nbNodes() (*big.Int, error) {
	input := common.Hex2Bytes("03b2ec98")
	result, _, err := runtime.Call(contract.ContractAddress, input, contract.RuntimeConfig)
	// log.Info("Result/Err", "Result", common.Bytes2Hex(result), "Err", err.Error())
	if err != nil {
		log.Error("Impossible to get the number of nodes in carbon footprint contract", "err", err.Error(), "block", contract.RuntimeConfig.BlockNumber.Int64())
		return nil, err
	} else {
		// log.Info("Carbon footprint nb nodes", "result", common.Bytes2Hex(result))
		return common.BytesToHash(result).Big(), nil
	}
}
func (contract *CarbonFootprintContract) footprint(ofNode common.Address) (*big.Int, error) {
	addressString := ofNode.String()
	addressString = addressString[2:]

	input := common.Hex2Bytes("79f85816000000000000000000000000" + addressString)
	result, _, err := runtime.Call(contract.ContractAddress, input, contract.RuntimeConfig)
	// log.Info("Result/Err", "Result", common.Bytes2Hex(result), "Err", err.Error())
	if err != nil {
		log.Error("Impossible to get the carbon footprint", "err", err.Error(), "node", ofNode.String(), "block", contract.RuntimeConfig.BlockNumber.Int64())
		return nil, err
	} else {
		// log.Info("Carbon footprint node", "result", common.Bytes2Hex(result), "node", ofNode.String())
		return common.BytesToHash(result).Big(), nil
	}
}


// The allFootprints return the addess and the CF of a sealer index 0, and returns false if there is no more to parse, true otherwise
// func (contract *CarbonFootprintContract) allFootprints(index big.Int) (common.Address, big.Int, bool,  error) {
//	return common.Address,nil,nil,nil
// }
