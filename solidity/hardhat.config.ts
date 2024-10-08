import { HardhatUserConfig } from "hardhat/config";
import "hardhat-dependency-compiler";
import "@nomicfoundation/hardhat-ethers";
import "@typechain/hardhat";
import "hardhat-gas-reporter";
import "@nomicfoundation/hardhat-verify";
import "@nomicfoundation/hardhat-chai-matchers";
import "hardhat-ignore-warnings";

import "./tasks/task";

const config: HardhatUserConfig = {
  defaultNetwork: "hardhat",
  networks: {
    hardhat: {
      chainId: process.env.CHAIN_ID ? parseInt(process.env.CHAIN_ID) : 1337,
    },
    ethereum: {
      url: `${process.env.ETHEREUM_URL || "https://1rpc.io/eth"}`,
      chainId: 1,
    },
    base: {
      url: `${process.env.BASE_URL || "https://mainnet.base.org"}`,
      chainId: 8453,
    },
    sepolia: {
      url: `${process.env.SEPOLIA_URL || "https://rpc.sepolia.org"}`,
      chainId: 11155111,
    },
    arbitrumSepolia: {
      url: `${
        process.env.ARBITRUM_URL || "https://sepolia-rollup.arbitrum.io/rpc"
      }`,
      chainId: 421614,
    },
    optimisticSepolia: {
      url: `${process.env.OPTIMISTIC_URL || "https://sepolia.optimism.io"}`,
      chainId: 11155420,
    },
    baseSepolia: {
      url: `${process.env.BASE_URL || "https://sepolia.base.org"}`,
      chainId: 84532,
    },
    polygonAmoy: {
      url: `${
        process.env.POLYGON_URL || "https://rpc-amoy.polygon.technology"
      }`,
      chainId: 80002,
    },
    fxcore: {
      url: `${
        process.env.FXCORE_URL || "https://fx-json-web3.functionx.io:8545"
      }`,
      chainId: 530,
    },
    dhobyghaut: {
      url: `${
        process.env.DHOBYGHAUT_URL ||
        "https://testnet-fx-json-web3.functionx.io:8545"
      }`,
      chainId: 90001,
    },
    localhost: {
      url: `${process.env.LOCAL_URL || "http://127.0.0.1:8545"}`,
    },
  },
  solidity: {
    compilers: [
      {
        version: "0.8.0",
        settings: {
          optimizer: {
            enabled: true,
            runs: 200,
          },
        },
      },
      {
        version: "0.8.1",
        settings: {
          optimizer: {
            enabled: true,
            runs: 200,
          },
        },
      },
      {
        version: "0.8.2",
        settings: {
          optimizer: {
            enabled: true,
            runs: 200,
          },
        },
      },
      {
        version: "0.8.10",
        settings: {
          optimizer: {
            enabled: true,
            runs: 200,
          },
        },
      },
    ],
  },
  etherscan: {
    apiKey: {
      ethereum: `${process.env.ETHERSCAN_API_KEY}`,
      sepolia: `${process.env.ETHERSCAN_API_KEY}`,
      arbitrumSepolia: `${process.env.ETHERSCAN_API_KEY}`,
      optimisticSepolia: `${process.env.ETHERSCAN_API_KEY}`,
      baseSepolia: `${process.env.ETHERSCAN_API_KEY}`,
    },
    customChains: [
      {
        network: "ethereum",
        chainId: 1,
        urls: {
          apiURL: "https://api.etherscan.io/api",
          browserURL: "https://etherscan.io",
        },
      },
      {
        network: "optimisticSepolia",
        chainId: 11155420,
        urls: {
          apiURL: "https://api-sepolia-optimistic.etherscan.io/api",
          browserURL: "https://sepolia-optimism.etherscan.io/",
        },
      },
    ],
  },
  dependencyCompiler: {
    paths: [
      "@openzeppelin/contracts/proxy/ERC1967/ERC1967Proxy.sol",
      "@openzeppelin/contracts/proxy/transparent/TransparentUpgradeableProxy.sol",
    ],
  },
  gasReporter: {
    enabled: false,
    currency: "USD",
    gasPrice: 30,
  },
  warnings: {
    "@openzeppelin/contracts/**/*": "off",
    "@openzeppelin/contracts-upgradeable/**/*": "off",
  },
};

export default config;
