// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.0;

interface IBridgeCallback {
    function bridgeCallback(
        address,
        address,
        address[] memory,
        uint256[] memory,
        address,
        bytes memory,
        bytes memory,
        uint256
    ) external;
}
