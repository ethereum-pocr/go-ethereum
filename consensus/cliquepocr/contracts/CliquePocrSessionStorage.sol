// SPDX-License-Identifier: GPL-3.0

pragma solidity >=0.7.0 <0.9.0;

/**
 * @title Storage
 * @dev Retrieve the value of CLiquePOCR engine session variables
 */
contract CliquePocrSessionStorage {
    /**
     * @dev Return value 
     * @return value of the current session variable
     */
    function retrieveSessionVariable(string calldata variableName) public view returns (uint256) {
        (bool success, bytes memory data) = address(this).staticcall(
      abi.encodeWithSelector(
        bytes4(keccak256(abi.encodePacked(variableName,"()")))
      )
    );
    if (success) {
      return abi.decode(data, (uint256));
    }
    // returning max INT to indicate there is no variable with that key
    return 0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff;
    }
}