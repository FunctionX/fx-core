import { task } from "hardhat/config";
import { boolean, string } from "hardhat/internal/core/params/argumentTypes";
import {
  AddTxParam,
  GetGravityId,
  GetOracleSet,
  SUB_CHECK_PRIVATE_KEY,
  SUB_CONFIRM_TRANSACTION,
  SUB_CREATE_TRANSACTION,
  TransactionToJson,
  vote_power,
} from "./subtasks";
import { bech32 } from "bech32";
import { ZeroAddress } from "ethers";

const sendToFx = task("send-to-fx", "call bridge contract sendToFx()")
  .addParam("bridgeContract", "bridge token address", undefined, string, false)
  .addParam("bridgeToken", "bridge token address", undefined, string, false)
  .addParam("amount", "amount to bridge", undefined, string, false)
  .addParam("destination", "destination address", undefined, string, false)
  .addParam("targetIbc", "target ibc address", "", string, true)
  .setAction(async (taskArgs, hre) => {
    const { wallet } = await hre.run(SUB_CHECK_PRIVATE_KEY, taskArgs);

    const bridgeTokenContract = await hre.ethers.getContractAt(
      "ERC20TokenTest",
      taskArgs.bridgeToken,
      wallet
    );
    const from = await wallet.getAddress();

    const allowanceAmount = await bridgeTokenContract.allowance(
      from,
      taskArgs.bridgeContract
    );

    if (hre.ethers.getBigInt(taskArgs.amount) > allowanceAmount) {
      const erc20_factory = await hre.ethers.getContractFactory(
        "ERC20TokenTest"
      );
      const data = erc20_factory.interface.encodeFunctionData("approve", [
        taskArgs.bridgeContract,
        taskArgs.amount,
      ]);

      const tx = await hre.run(SUB_CREATE_TRANSACTION, {
        from: from,
        to: taskArgs.bridgeToken,
        data: data,
        value: taskArgs.value,
        gasPrice: taskArgs.gasPrice,
        maxFeePerGas: taskArgs.maxFeePerGas,
        maxPriorityFeePerGas: taskArgs.maxPriorityFeePerGas,
        nonce: taskArgs.nonce,
        gasLimit: taskArgs.gasLimit,
      });

      const { answer } = await hre.run(SUB_CONFIRM_TRANSACTION, {
        message: `\n${TransactionToJson(tx)}\n`,
        disableConfirm: taskArgs.disableConfirm,
      });
      if (!answer) return;

      try {
        const approveTx = await wallet.sendTransaction(tx);
        await approveTx.wait();
        console.log(`Approve success, ${approveTx.hash}`);
      } catch (e) {
        console.log(`Approve failed, ${e}`);
        return;
      }
    }
    const bridge_factory = await hre.ethers.getContractFactory("FxBridgeLogic");

    const destination_bc = bech32.fromWords(
      bech32.decode(taskArgs.destination).words
    );
    const destination_bc_hex = (
      "0x" +
      "0".repeat(24) +
      Buffer.from(destination_bc).toString("hex")
    ).toString();

    const target = hre.ethers.encodeBytes32String(taskArgs.targetIbc);

    const data = bridge_factory.interface.encodeFunctionData("sendToFx", [
      taskArgs.bridgeToken,
      destination_bc_hex,
      target,
      taskArgs.amount,
    ]);

    const tx = await hre.run(SUB_CREATE_TRANSACTION, {
      from: from,
      to: taskArgs.bridgeContract,
      data: data,
      value: taskArgs.value,
      gasPrice: taskArgs.gasPrice,
      maxFeePerGas: taskArgs.maxFeePerGas,
      maxPriorityFeePerGas: taskArgs.maxPriorityFeePerGas,
      nonce: taskArgs.nonce,
      gasLimit: taskArgs.gasLimit,
    });

    const { answer } = await hre.run(SUB_CONFIRM_TRANSACTION, {
      message: `\n${TransactionToJson(tx)}\n`,
      disableConfirm: taskArgs.disableConfirm,
    });
    if (!answer) return;

    try {
      const sendToFxTx = await wallet.sendTransaction(tx);
      await sendToFxTx.wait();
      console.log(`SendToFx success, ${sendToFxTx.hash}`);
    } catch (e) {
      console.log(`SendToFx failed, ${e}`);
    }
  });

const initBridge = task("init-bridge", "init bridge contract")
  .addParam(
    "bridgeLogic",
    "bridge logic contract address",
    undefined,
    string,
    true
  )
  .addParam(
    "bridgeContract",
    "init bridge contract address",
    undefined,
    string,
    false
  )
  .addParam("restUrl", "fx node rest rpc url", undefined, string, false)
  .addParam("chainName", "init cross chain name", undefined, string, false)
  .setAction(async (taskArgs, hre) => {
    const { bridgeLogic, bridgeContract, restUrl, chainName } = taskArgs;
    const { wallet } = await hre.run(SUB_CHECK_PRIVATE_KEY, taskArgs);
    const from = await wallet.getAddress();

    const bridge_logic_factory = await hre.ethers.getContractFactory(
      "FxBridgeLogic"
    );

    const oracle_set = await GetOracleSet(restUrl, chainName);
    const gravity_id_str = await GetGravityId(restUrl, chainName);
    const gravity_id = hre.ethers.encodeBytes32String(gravity_id_str);

    const external_addresses = [];
    const powers = [];
    let powers_sum = 0;

    for (let i = 0; i < oracle_set.members.length; i++) {
      external_addresses.push(oracle_set.members[i].external_address);
      powers.push(oracle_set.members[i].power);
      powers_sum += oracle_set.members[i].power;
    }

    if (powers_sum < vote_power) {
      console.error("Incorrect power! Please inspect the oracle set");
      console.log(`Current oracle set:\n${oracle_set}`);
      return;
    }

    let data = bridge_logic_factory.interface.encodeFunctionData("init", [
      gravity_id,
      vote_power,
      external_addresses,
      powers,
    ]);

    if (bridgeLogic) {
      const proxy_factory = await hre.ethers.getContractAt(
        "ITransparentUpgradeableProxy",
        bridgeContract,
        wallet
      );
      data = proxy_factory.interface.encodeFunctionData("upgradeToAndCall", [
        bridgeLogic,
        data,
      ]);
    }

    const tx = await hre.run(SUB_CREATE_TRANSACTION, {
      from: from,
      to: bridgeContract,
      data: data,
      value: taskArgs.value,
      gasPrice: taskArgs.gasPrice,
      maxFeePerGas: taskArgs.maxFeePerGas,
      maxPriorityFeePerGas: taskArgs.maxPriorityFeePerGas,
      nonce: taskArgs.nonce,
      gasLimit: taskArgs.gasLimit,
    });

    const { answer } = await hre.run(SUB_CONFIRM_TRANSACTION, {
      message: `\n${TransactionToJson(tx)}\n`,
      disableConfirm: taskArgs.disableConfirm,
    });
    if (!answer) return;

    try {
      const initTx = await wallet.sendTransaction(tx);
      await initTx.wait();
      console.log(`Init success, ${initTx.hash}`);
    } catch (e) {
      console.log(`Init failed, ${e}`);
    }
  });

const addBridgeToken = task(
  "add-bridge-token",
  "add bridge token into bridge contract"
)
  .addParam(
    "bridgeContract",
    "bridge proxy contract address",
    undefined,
    string,
    false
  )
  .addParam("tokenContract", "token contract address", undefined, string, false)
  .addParam(
    "isOriginal",
    "bridge token target ibc for bridge token",
    false,
    boolean,
    true
  )
  .addParam(
    "targetIbc",
    "bridge token target ibc for bridge token",
    "",
    string,
    true
  )
  .setAction(async (taskArgs, hre) => {
    const { bridgeContract, tokenContract, isOriginal, targetIbc } = taskArgs;
    const { wallet } = await hre.run(SUB_CHECK_PRIVATE_KEY, taskArgs);
    const from = await wallet.getAddress();

    const bridge_factory = await hre.ethers.getContractFactory("FxBridgeLogic");
    const ibc = hre.ethers.encodeBytes32String(targetIbc);
    const data = bridge_factory.interface.encodeFunctionData("addBridgeToken", [
      tokenContract,
      ibc,
      isOriginal,
    ]);

    const tx = await hre.run(SUB_CREATE_TRANSACTION, {
      from: from,
      to: bridgeContract,
      data: data,
      value: taskArgs.value,
      gasPrice: taskArgs.gasPrice,
      maxFeePerGas: taskArgs.maxFeePerGas,
      maxPriorityFeePerGas: taskArgs.maxPriorityFeePerGas,
      nonce: taskArgs.nonce,
      gasLimit: taskArgs.gasLimit,
    });

    const { answer } = await hre.run(SUB_CONFIRM_TRANSACTION, {
      message: `\n${TransactionToJson(tx)}\n`,
      disableConfirm: taskArgs.disableConfirm,
    });
    if (!answer) return;

    try {
      const addBridgeTokenTx = await wallet.sendTransaction(tx);
      await addBridgeTokenTx.wait();
      console.log(`AddBridgeToken success, ${addBridgeTokenTx.hash}`);
    } catch (e) {
      console.log(`AddBridgeToken failed, ${e}`);
    }
  });

const bridgeCall = task("bridge-call", "bridge call function")
  .addParam(
    "bridgeContract",
    "bridge contract address",
    undefined,
    string,
    false
  )
  .addParam("dstChainId", "destination chain id", undefined, string, false)
  .addParam("refund", "call receiver", undefined, string, false)
  .addParam("tokens", "bridge token address list", undefined, string, false)
  .addParam("amounts", "bridge token amount list", undefined, string, false)
  .addParam("to", "call to", undefined, string, false)
  .addParam("data", "call data", undefined, string, false)
  .addParam("callValue", "call value", "0", string, true)
  .addParam("memo", "call memo", undefined, string, false)

  .setAction(async (taskArgs, hre) => {
    let {
      bridgeContract,
      dstChainId,
      refund,
      tokens,
      amounts,
      to,
      data,
      callValue,
      memo,
    } = taskArgs;
    const { wallet } = await hre.run(SUB_CHECK_PRIVATE_KEY, taskArgs);
    const from = await wallet.getAddress();

    let tokenArr: string[];
    let amountArr: string[];
    if (tokens === "") {
      tokenArr = [];
      amountArr = [];
    } else {
      tokenArr = tokens.split(",");
      amountArr = amounts.split(",");
    }
    if (tokenArr.length !== amountArr.length) {
      console.log("tokens and amounts do not match");
      return;
    }

    if (to === "") {
      to = ZeroAddress;
    }

    const bridge_logic_factory = await hre.ethers.getContractFactory(
      "FxBridgeLogic"
    );
    const bridgeCallData = bridge_logic_factory.interface.encodeFunctionData(
      "bridgeCall",
      [dstChainId, refund, tokenArr, amountArr, to, data, callValue, memo]
    );

    const tx = await hre.run(SUB_CREATE_TRANSACTION, {
      from: from,
      to: bridgeContract,
      data: bridgeCallData,
      value: taskArgs.value,
      gasPrice: taskArgs.gasPrice,
      maxFeePerGas: taskArgs.maxFeePerGas,
      maxPriorityFeePerGas: taskArgs.maxPriorityFeePerGas,
      nonce: taskArgs.nonce,
      gasLimit: taskArgs.gasLimit,
    });

    const { answer } = await hre.run(SUB_CONFIRM_TRANSACTION, {
      message: `\n${TransactionToJson(tx)}\n`,
      disableConfirm: taskArgs.disableConfirm,
    });
    if (!answer) return;

    try {
      const bridgeCallTx = await wallet.sendTransaction(tx);
      await bridgeCallTx.wait();
      console.log(`bridge call success, ${bridgeCallTx.hash}`);
    } catch (e) {
      console.log(`bridge call failed, ${e}`);
    }
  });

AddTxParam([sendToFx, initBridge, addBridgeToken, bridgeCall]);
