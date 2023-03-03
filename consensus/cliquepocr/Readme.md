The cliquePocr module is a clique-compliant consensus engine that introduces a new reward mechanism defined on a smart contract deployed in the genesis file.
The reward mechanism is described in the white paper 'Proof of Climate awaReness'.

The design of the cliquepocr new consensus is aiming at:
1. Not impacting the clique engine (except for one configuration section, "ispocr", which is a boolean targetting the new cliquepocr consensus engine)
2. Having a minimum impact on the eth/backend.go code. Unfortunately, has no dependency injection was defined in it, it has been required to add "if" code in this code to target the case of the new cliquepocr engine.
3. To reuse as much as possible the clique engine, overriding only reward mechanisms. For this purpose, a clique engine is instantiated in the cliquepocr engine and most of the engine lifecycle methods are directly redirected to the clique engine behind.

