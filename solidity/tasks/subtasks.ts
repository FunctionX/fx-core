import {subtask} from "hardhat/config";
import {LedgerSigner} from "@anders-t/ethers-ledger";
import {BigNumber} from "ethers";
import axios from "axios";
import {ConfigurableTaskDefinition} from "hardhat/types";
import {boolean, string} from "hardhat/internal/core/params/argumentTypes";

const inquirer = require('inquirer')

// sub task name
export const SUB_CHECK_PRIVATE_KEY: string = "sub:check-private-key";
export const SUB_PRIVATE_KEY_WALLET: string = "sub:generate-wallet";
export const SUB_GET_NODE_URL: string = "sub:get-eth-node-url";
export const SUB_CREATE_LEDGER_WALLET: string = "sub:create-ledger-wallet";
export const SUB_CREATE_TRANSACTION: string = "sub:create-transaction";
export const SUB_CONFIRM_TRANSACTION: string = "sub:confirm-transaction";
export const SUB_MNEMONIC_WALLET: string = "sub:mnemonic-wallet";
export const SUB_SEND_ETH: string = "sub:send-eth";
export const SUB_GET_CONTRACT_ADDR: string = "sub:get-contract-addr";

// public flag
export const DISABLE_CONFIRM_FLAG: string = "disableConfirm";
export const PRIVATE_KEY_FLAG = "privateKey";
export const MNEMONIC_FLAG = "mnemonic";
export const INDEX_FLAG = "index";
export const IS_LEDGER_FLAG = "isLedger";
export const DRIVER_PATH_FLAG = "driverPath";
export const NONCE_FLAG = "nonce";
export const GAS_PRICE_FLAG = "gasPrice";
export const MAX_FEE_PER_GAS_FLAG = "maxFeePerGas";
export const MAX_PRIORITY_FEE_PER_GAS_FLAG = "maxPriorityFeePerGas";
export const GAS_LIMIT_FLAG = "gasLimit";
export const VALUE_FLAG = "value";

export const DEFAULT_DRIVE_PATH = "m/44'/60'/0'/0/0";
export const DEFAULT_PRIORITY_FEE = "1500000000";

subtask(SUB_SEND_ETH, "send eth").setAction(
    async (taskArgs, hre) => {
        const {to, value, wallet, gasPrice, maxFeePerGas, maxPriorityFeePerGas, nonce, gasLimit, chainId} = taskArgs;
        const transaction: Transaction = await hre.run(SUB_CREATE_TRANSACTION, {
            from: wallet.address,
            to: to,
            value: value,
            gasPrice: gasPrice,
            maxFeePerGas: maxFeePerGas,
            maxPriorityFeePerGas: maxPriorityFeePerGas,
            nonce: nonce,
            gasLimit: gasLimit || 21000,
            chainId: chainId
        });
        const {answer} = await hre.run(SUB_CONFIRM_TRANSACTION, {
            message: `\n${TransactionToJson(transaction)}\n`,
            disableConfirm: taskArgs.disableConfirm,
        });
        if (!answer) {
            return
        }
        const tx = await wallet.sendTransaction(transaction)
        console.log(`${tx.hash}`)
        await tx.wait()
        return
    }
);

subtask(SUB_CREATE_TRANSACTION, "create transaction").setAction(
    async (taskArgs, hre) => {
        let {from, to, value, data, gasPrice, maxFeePerGas, maxPriorityFeePerGas, nonce, gasLimit, chainId} = taskArgs;
        if (gasPrice && maxFeePerGas) {
            throw new Error("Please provide only one of gasPrice or maxFeePerGas and maxPriorityFeePerGas");
        }
        if (!gasPrice && !maxFeePerGas) {
            await hre.ethers.provider.getBlock("latest").then(
                async (block) => {
                    if (block.baseFeePerGas) {
                        maxPriorityFeePerGas = maxPriorityFeePerGas ? maxPriorityFeePerGas : BigNumber.from(DEFAULT_PRIORITY_FEE);
                        maxFeePerGas = block.baseFeePerGas.add(maxPriorityFeePerGas);
                    } else {
                        gasPrice = await hre.ethers.provider.getGasPrice()
                    }
                }
            );
        }
        if (maxFeePerGas) {
            maxPriorityFeePerGas = maxPriorityFeePerGas ? maxPriorityFeePerGas : BigNumber.from(DEFAULT_PRIORITY_FEE);
            maxFeePerGas = BigNumber.from(maxFeePerGas).add(maxPriorityFeePerGas);
        }
        const transaction: Transaction = {
            from: from,
            to: to,
            value: value,
            data: data,
            nonce: nonce ? nonce : await hre.ethers.provider.getTransactionCount(from),
            gasLimit: gasLimit ? gasLimit : await hre.ethers.provider.estimateGas({
                from: from,
                to: to,
                data: data,
                value: value
            }),
            chainId: chainId ? chainId : await hre.ethers.provider.getNetwork().then(network => network.chainId)
        }
        if (gasPrice) {
            transaction.gasPrice = gasPrice;
        }
        if (maxFeePerGas) {
            transaction.maxFeePerGas = maxFeePerGas;
            transaction.maxPriorityFeePerGas = maxPriorityFeePerGas;
        }
        return transaction;
    }
);

subtask(SUB_GET_CONTRACT_ADDR, "get contract address").setAction(
    async (taskArgs, hre) => {
        const {from} = taskArgs;

        const nodeUrl = await hre.run(SUB_GET_NODE_URL);
        const provider = await new hre.ethers.providers.JsonRpcProvider(nodeUrl);
        const nonce = await provider.getTransactionCount(from);

        return hre.ethers.utils.getContractAddress({
            from: from,
            nonce: nonce,
        });
    });

subtask(SUB_CHECK_PRIVATE_KEY, "check the method of getting private key").setAction(
    async (taskArgs, hre) => {
        const {privateKey, isLedger, mnemonic} = taskArgs;
        if (
            privateKey && isLedger || privateKey && mnemonic || isLedger && mnemonic
        ) {
            throw new Error("Please provide only one of private key or ledger or mnemonic");
        }
        if (privateKey) {
            const {wallet} = await hre.run(SUB_PRIVATE_KEY_WALLET, taskArgs);
            return {wallet}
        }
        if (mnemonic) {
            return await hre.run(SUB_MNEMONIC_WALLET, taskArgs);
        }
        if (isLedger) {
            return await hre.run(SUB_CREATE_LEDGER_WALLET, taskArgs);
        }
        return (await hre.ethers.getSigners())[0];
    }
);

subtask(SUB_CREATE_LEDGER_WALLET, "create ledger wallet").setAction(
    async (taskArgs, hre) => {
        const {driverPath} = taskArgs;
        const nodeUrl = await hre.run(SUB_GET_NODE_URL);
        const provider = await new hre.ethers.providers.JsonRpcProvider(nodeUrl);

        const _path = driverPath ? driverPath : DEFAULT_DRIVE_PATH;

        const wallet = new LedgerSigner(provider, _path);
        return {wallet};
    });

subtask(SUB_PRIVATE_KEY_WALLET, "private key wallet account").setAction(
    async (taskArgs, hre) => {
        const {privateKey} = taskArgs;
        const nodeUrl = await hre.run(SUB_GET_NODE_URL);
        const provider = await new hre.ethers.providers.JsonRpcProvider(nodeUrl);
        const wallet = new hre.ethers.Wallet(privateKey, provider);
        return {provider, wallet};
    });

subtask(SUB_MNEMONIC_WALLET, "mnemonic wallet account").setAction(
    async (taskArgs, hre) => {
        const {mnemonic, driverPath, index} = taskArgs;
        const nodeUrl = await hre.run(SUB_GET_NODE_URL);
        const provider = await new hre.ethers.providers.JsonRpcProvider(nodeUrl);

        let _path = DEFAULT_DRIVE_PATH
        if (driverPath) {
            _path = driverPath
        }
        if (index) {
            _path = `m/44'/60'/0'/0/${index}`
        }
        const wallet = hre.ethers.Wallet.fromMnemonic(mnemonic, _path).connect(provider);
        return {provider, wallet};
    }
);

subtask(SUB_GET_NODE_URL, "get node url form hardhat.network").setAction(
    async (taskArgs, hre) => {
        return "url" in hre.network.config ? hre.network.config.url : "";
    },
);

subtask(SUB_CONFIRM_TRANSACTION, "confirm transaction").setAction(
    async (taskArgs, _) => {
        const {message, disableConfirm} = taskArgs;
        let _answer;
        if (!disableConfirm) {
            const {answer} = await inquirer.createPromptModule()({
                type: "confirm",
                name: "answer",
                message,
            });
            _answer = answer;
        } else {
            _answer = true;
        }
        return {answer: _answer};
    });

type Transaction = {
    from: string,
    to?: string,
    value?: BigNumber,
    data?: string,
    gasPrice?: BigNumber,
    maxFeePerGas?: BigNumber,
    maxPriorityFeePerGas?: BigNumber,
    nonce: number,
    gasLimit?: number,
    chainId: number
}

// function Transaction to json string
export function TransactionToJson(transaction: Transaction): string {
    return JSON.stringify({
        from: transaction.from,
        to: transaction.to,
        value: transaction.value,
        data: transaction.data,
        gasPrice: transaction.gasPrice ? transaction.gasPrice.toString() : undefined,
        maxFeePerGas: transaction.maxFeePerGas ? transaction.maxFeePerGas.toString() : undefined,
        maxPriorityFeePerGas: transaction.maxPriorityFeePerGas ? transaction.maxPriorityFeePerGas.toString() : undefined,
        nonce: transaction.nonce,
        gasLimit: transaction.gasLimit ? transaction.gasLimit.toString() : undefined,
        chainId: transaction.chainId
    }, null, 2);
}

export const vote_power = 2834678415

type Oracle = {
    power: number;
    external_address: string
};
type OracleSet = {
    members: Oracle[];
    nonce: number;
};

export async function GetOracleSet(restRpc: string, chainName: string): Promise<OracleSet> {
    const request_string = restRpc + `/fx/crosschain/v1/oracle_set/current?chain_name=${chainName}`
    const response = await axios.get(request_string);
    return response.data.oracle_set;
}

export async function GetGravityId(restRpc: string, chainName: string): Promise<string> {
    const request_string = restRpc + `/fx/crosschain/v1/params?chain_name=${chainName}`
    const response = await axios.get(request_string);
    return response.data.params.gravity_id;
}

export function AddTxParam(tasks: ConfigurableTaskDefinition[]) {
    tasks.forEach((task) => {
        task.addParam(NONCE_FLAG, "nonce", undefined, string, true)
            .addParam(GAS_PRICE_FLAG, "gas price", undefined, string, true)
            .addParam(MAX_FEE_PER_GAS_FLAG, "max fee per gas", undefined, string, true)
            .addParam(MAX_PRIORITY_FEE_PER_GAS_FLAG, "max priority fee per gas", undefined, string, true)
            .addParam(GAS_LIMIT_FLAG, "gas limit", undefined, string, true)
            .addParam(VALUE_FLAG, "value", undefined, string, true)
            .addParam(PRIVATE_KEY_FLAG, "send tx by private key", undefined, string, true)
            .addParam(MNEMONIC_FLAG, "send tx by mnemonic", undefined, string, true)
            .addParam(INDEX_FLAG, "mnemonic index", undefined, string, true)
            .addParam(IS_LEDGER_FLAG, "ledger to send tx", false, boolean, true)
            .addParam(DRIVER_PATH_FLAG, "manual HD Path derivation (overrides BIP44 config)", "m/44'/60'/0'/0/0", string, true)
            .addParam(DISABLE_CONFIRM_FLAG, "disable confirm", false, boolean, true)
    })
}