Description of the fee pot algorithm

Modify blockchain class 

https://github.com/ethereum-pocr/go-ethereum/blob/4f4c660842d081daceaaadeff2d8fad17094e69b/core/blockchain.go#L256

has the state process that computes fees.

Add the Engine interface here: 
https://github.com/ethereum-pocr/go-ethereum/blob/4f4c660842d081daceaaadeff2d8fad17094e69b/core/blockchain.go#L1671

You would then have an Engine interface here:
https://github.com/ethereum-pocr/go-ethereum/blob/4f4c660842d081daceaaadeff2d8fad17094e69b/core/state_processor.go#L59

Modify the applyTransaction to use the Engine as a parameter
https://github.com/ethereum-pocr/go-ethereum/blob/4f4c660842d081daceaaadeff2d8fad17094e69b/core/state_processor.go#L82

You would get to modify one of the core function of the EVM, ApplyMessage, to add the Engine interfrace

https://github.com/ethereum-pocr/go-ethereum/blob/4f4c660842d081daceaaadeff2d8fad17094e69b/core/state_transition.go#L180

In the TransitionDB method, add the Engine interface as well (it's empty parameter)
https://github.com/ethereum-pocr/go-ethereum/blob/4f4c660842d081daceaaadeff2d8fad17094e69b/core/state_transition.go#L181

Modify the gas attribution function here:
https://github.com/ethereum-pocr/go-ethereum/blob/4f4c660842d081daceaaadeff2d8fad17094e69b/core/state_transition.go#L333

Instave of just calling the evm, call the engine first, asking how much gas must be allocated to the current sealer. The logic of the algorithm would then be:
If the sealer is top 1 CF, he gets 100% of the gas.
If not, it gets the racerank reduction function (0.9 for the second, 0.81 for the third etc) and the remaining gas is bur,t.



