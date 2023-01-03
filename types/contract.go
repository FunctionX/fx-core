package types

import (
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"

	"github.com/functionx/fx-core/v3/types/contract"
)

const (
	EmptyEvmAddress   = "0x0000000000000000000000000000000000000000"
	FIP20LogicAddress = "0x0000000000000000000000000000000000001001"
	WFXLogicAddress   = "0x0000000000000000000000000000000000001002"
)

type Contract struct {
	Address common.Address
	ABI     abi.ABI
	Bin     []byte
	Code    []byte
}

var (
	initERC20Code = mustDecodeHex("0x60806040526004361061011f5760003560e01c8063715018a6116100a0578063b86d529811610064578063b86d52981461031c578063c5cb9b511461033a578063dd62ed3e1461035a578063de7ea79d146103a0578063f2fde38b146103c057600080fd5b8063715018a6146102805780638da5cb5b1461029557806395d89b41146102c75780639dc29fac146102dc578063a9059cbb146102fc57600080fd5b80633659cfe6116100e75780633659cfe6146101e057806340c10f19146102025780634f1ef2861461022257806352d1902d1461023557806370a082311461024a57600080fd5b806306fdde0314610124578063095ea7b31461014f57806318160ddd1461017f57806323b872dd1461019e578063313ce567146101be575b600080fd5b34801561013057600080fd5b506101396103e0565b60405161014691906118a7565b60405180910390f35b34801561015b57600080fd5b5061016f61016a366004611743565b610472565b6040519015158152602001610146565b34801561018b57600080fd5b5060cc545b604051908152602001610146565b3480156101aa57600080fd5b5061016f6101b93660046116a9565b6104c8565b3480156101ca57600080fd5b5060cb5460405160ff9091168152602001610146565b3480156101ec57600080fd5b506102006101fb36600461165d565b610577565b005b34801561020e57600080fd5b5061020061021d366004611743565b610657565b6102006102303660046116e4565b61068f565b34801561024157600080fd5b5061019061075c565b34801561025657600080fd5b5061019061026536600461165d565b6001600160a01b0316600090815260cd602052604090205490565b34801561028c57600080fd5b5061020061080f565b3480156102a157600080fd5b506097546001600160a01b03165b6040516001600160a01b039091168152602001610146565b3480156102d357600080fd5b50610139610845565b3480156102e857600080fd5b506102006102f7366004611743565b610854565b34801561030857600080fd5b5061016f610317366004611743565b610888565b34801561032857600080fd5b5060cf546001600160a01b03166102af565b34801561034657600080fd5b5061016f61035536600461180d565b61089e565b34801561036657600080fd5b50610190610375366004611677565b6001600160a01b03918216600090815260ce6020908152604080832093909416825291909152205490565b3480156103ac57600080fd5b506102006103bb366004611784565b61090d565b3480156103cc57600080fd5b506102006103db36600461165d565b610a2c565b606060c980546103ef90611a5c565b80601f016020809104026020016040519081016040528092919081815260200182805461041b90611a5c565b80156104685780601f1061043d57610100808354040283529160200191610468565b820191906000526020600020905b81548152906001019060200180831161044b57829003601f168201915b5050505050905090565b600061047f338484610ac4565b6040518281526001600160a01b0384169033907f8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b9259060200160405180910390a350600192915050565b6001600160a01b038316600090815260ce602090815260408083203384529091528120548281101561054b5760405162461bcd60e51b815260206004820152602160248201527f7472616e7366657220616d6f756e74206578636565647320616c6c6f77616e636044820152606560f81b60648201526084015b60405180910390fd5b61055f853361055a8685611a19565b610ac4565b61056a858585610b46565b60019150505b9392505050565b306001600160a01b037f00000000000000000000000000000000000000000000000000000000000010011614156105c05760405162461bcd60e51b8152600401610542906118e9565b7f00000000000000000000000000000000000000000000000000000000000010016001600160a01b0316610609600080516020611ac4833981519152546001600160a01b031690565b6001600160a01b03161461062f5760405162461bcd60e51b815260040161054290611935565b61063881610cf5565b6040805160008082526020820190925261065491839190610d1f565b50565b6097546001600160a01b031633146106815760405162461bcd60e51b815260040161054290611981565b61068b8282610e9e565b5050565b306001600160a01b037f00000000000000000000000000000000000000000000000000000000000010011614156106d85760405162461bcd60e51b8152600401610542906118e9565b7f00000000000000000000000000000000000000000000000000000000000010016001600160a01b0316610721600080516020611ac4833981519152546001600160a01b031690565b6001600160a01b0316146107475760405162461bcd60e51b815260040161054290611935565b61075082610cf5565b61068b82826001610d1f565b6000306001600160a01b037f000000000000000000000000000000000000000000000000000000000000100116146107fc5760405162461bcd60e51b815260206004820152603860248201527f555550535570677261646561626c653a206d757374206e6f742062652063616c60448201527f6c6564207468726f7567682064656c656761746563616c6c00000000000000006064820152608401610542565b50600080516020611ac483398151915290565b6097546001600160a01b031633146108395760405162461bcd60e51b815260040161054290611981565b6108436000610f7d565b565b606060ca80546103ef90611a5c565b6097546001600160a01b0316331461087e5760405162461bcd60e51b815260040161054290611981565b61068b8282610fcf565b6000610895338484610b46565b50600192915050565b600063ffffffff333b16156108f55760405162461bcd60e51b815260206004820152601960248201527f63616c6c65722063616e6e6f7420626520636f6e7472616374000000000000006044820152606401610542565b6109023386868686611111565b506001949350505050565b600054610100900460ff166109285760005460ff161561092c565b303b155b61098f5760405162461bcd60e51b815260206004820152602e60248201527f496e697469616c697a61626c653a20636f6e747261637420697320616c72656160448201526d191e481a5b9a5d1a585b1a5e995960921b6064820152608401610542565b600054610100900460ff161580156109b1576000805461ffff19166101011790555b84516109c49060c9906020880190611513565b5083516109d89060ca906020870190611513565b5060cb805460ff191660ff851617905560cf80546001600160a01b0319166001600160a01b038416179055610a0b611259565b610a13611288565b8015610a25576000805461ff00191690555b5050505050565b6097546001600160a01b03163314610a565760405162461bcd60e51b815260040161054290611981565b6001600160a01b038116610abb5760405162461bcd60e51b815260206004820152602660248201527f4f776e61626c653a206e6577206f776e657220697320746865207a65726f206160448201526564647265737360d01b6064820152608401610542565b61065481610f7d565b6001600160a01b038316610b1a5760405162461bcd60e51b815260206004820152601d60248201527f617070726f76652066726f6d20746865207a65726f20616464726573730000006044820152606401610542565b6001600160a01b03928316600090815260ce602090815260408083209490951682529290925291902055565b6001600160a01b038316610b9c5760405162461bcd60e51b815260206004820152601e60248201527f7472616e736665722066726f6d20746865207a65726f206164647265737300006044820152606401610542565b6001600160a01b038216610bf25760405162461bcd60e51b815260206004820152601c60248201527f7472616e7366657220746f20746865207a65726f2061646472657373000000006044820152606401610542565b6001600160a01b038316600090815260cd602052604090205481811015610c5b5760405162461bcd60e51b815260206004820152601f60248201527f7472616e7366657220616d6f756e7420657863656564732062616c616e6365006044820152606401610542565b610c658282611a19565b6001600160a01b03808616600090815260cd60205260408082209390935590851681529081208054849290610c9b908490611a01565b92505081905550826001600160a01b0316846001600160a01b03167fddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef84604051610ce791815260200190565b60405180910390a350505050565b6097546001600160a01b031633146106545760405162461bcd60e51b815260040161054290611981565b7f4910fdfa16fed3260ed0e7147f7cc6da11a60208b5b9406d12a635614ffd91435460ff1615610d5757610d52836112af565b505050565b826001600160a01b03166352d1902d6040518163ffffffff1660e01b815260040160206040518083038186803b158015610d9057600080fd5b505afa925050508015610dc0575060408051601f3d908101601f19168201909252610dbd9181019061176c565b60015b610e235760405162461bcd60e51b815260206004820152602e60248201527f45524331393637557067726164653a206e657720696d706c656d656e7461746960448201526d6f6e206973206e6f74205555505360901b6064820152608401610542565b600080516020611ac48339815191528114610e925760405162461bcd60e51b815260206004820152602960248201527f45524331393637557067726164653a20756e737570706f727465642070726f786044820152681a58589b195555525160ba1b6064820152608401610542565b50610d5283838361134b565b6001600160a01b038216610ef45760405162461bcd60e51b815260206004820152601860248201527f6d696e7420746f20746865207a65726f206164647265737300000000000000006044820152606401610542565b8060cc6000828254610f069190611a01565b90915550506001600160a01b038216600090815260cd602052604081208054839290610f33908490611a01565b90915550506040518181526001600160a01b038316906000907fddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef9060200160405180910390a35050565b609780546001600160a01b038381166001600160a01b0319831681179093556040519116919082907f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e090600090a35050565b6001600160a01b0382166110255760405162461bcd60e51b815260206004820152601a60248201527f6275726e2066726f6d20746865207a65726f20616464726573730000000000006044820152606401610542565b6001600160a01b038216600090815260cd60205260409020548181101561108e5760405162461bcd60e51b815260206004820152601b60248201527f6275726e20616d6f756e7420657863656564732062616c616e636500000000006044820152606401610542565b6110988282611a19565b6001600160a01b038416600090815260cd602052604081209190915560cc80548492906110c6908490611a19565b90915550506040518281526000906001600160a01b038516907fddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef9060200160405180910390a3505050565b6001600160a01b0385166111675760405162461bcd60e51b815260206004820152601e60248201527f7472616e736665722066726f6d20746865207a65726f206164647265737300006044820152606401610542565b60008451116111ac5760405162461bcd60e51b81526020600482015260116024820152701a5b9d985b1a59081c9958da5c1a595b9d607a1b6044820152606401610542565b806111ea5760405162461bcd60e51b815260206004820152600e60248201526d1a5b9d985b1a59081d185c99d95d60921b6044820152606401610542565b60cf5461120b9086906001600160a01b03166112068587611a01565b610b46565b846001600160a01b03167f282dd1817b996776123a00596764d4d54cc16460c9854f7a23f6be020ba0463d8585858560405161124a94939291906118ba565b60405180910390a25050505050565b600054610100900460ff166112805760405162461bcd60e51b8152600401610542906119b6565b610843611376565b600054610100900460ff166108435760405162461bcd60e51b8152600401610542906119b6565b6001600160a01b0381163b61131c5760405162461bcd60e51b815260206004820152602d60248201527f455243313936373a206e657720696d706c656d656e746174696f6e206973206e60448201526c1bdd08184818dbdb9d1c9858dd609a1b6064820152608401610542565b600080516020611ac483398151915280546001600160a01b0319166001600160a01b0392909216919091179055565b611354836113a6565b6000825111806113615750805b15610d525761137083836113e6565b50505050565b600054610100900460ff1661139d5760405162461bcd60e51b8152600401610542906119b6565b61084333610f7d565b6113af816112af565b6040516001600160a01b038216907fbc7cd75a20ee27fd9adebab32041f755214dbc6bffa90cc0225b39da2e5c2d3b90600090a250565b60606001600160a01b0383163b61144e5760405162461bcd60e51b815260206004820152602660248201527f416464726573733a2064656c65676174652063616c6c20746f206e6f6e2d636f6044820152651b9d1c9858dd60d21b6064820152608401610542565b600080846001600160a01b031684604051611469919061188b565b600060405180830381855af49150503d80600081146114a4576040519150601f19603f3d011682016040523d82523d6000602084013e6114a9565b606091505b50915091506114d18282604051806060016040528060278152602001611ae4602791396114da565b95945050505050565b606083156114e9575081610570565b8251156114f95782518084602001fd5b8160405162461bcd60e51b815260040161054291906118a7565b82805461151f90611a5c565b90600052602060002090601f0160209004810192826115415760008555611587565b82601f1061155a57805160ff1916838001178555611587565b82800160010185558215611587579182015b8281111561158757825182559160200191906001019061156c565b50611593929150611597565b5090565b5b808211156115935760008155600101611598565b600067ffffffffffffffff808411156115c7576115c7611aad565b604051601f8501601f19908116603f011681019082821181831017156115ef576115ef611aad565b8160405280935085815286868601111561160857600080fd5b858560208301376000602087830101525050509392505050565b80356001600160a01b038116811461163957600080fd5b919050565b600082601f83011261164e578081fd5b610570838335602085016115ac565b60006020828403121561166e578081fd5b61057082611622565b60008060408385031215611689578081fd5b61169283611622565b91506116a060208401611622565b90509250929050565b6000806000606084860312156116bd578081fd5b6116c684611622565b92506116d460208501611622565b9150604084013590509250925092565b600080604083850312156116f6578182fd5b6116ff83611622565b9150602083013567ffffffffffffffff81111561171a578182fd5b8301601f8101851361172a578182fd5b611739858235602084016115ac565b9150509250929050565b60008060408385031215611755578182fd5b61175e83611622565b946020939093013593505050565b60006020828403121561177d578081fd5b5051919050565b60008060008060808587031215611799578081fd5b843567ffffffffffffffff808211156117b0578283fd5b6117bc8883890161163e565b955060208701359150808211156117d1578283fd5b506117de8782880161163e565b935050604085013560ff811681146117f4578182fd5b915061180260608601611622565b905092959194509250565b60008060008060808587031215611822578384fd5b843567ffffffffffffffff811115611838578485fd5b6118448782880161163e565b97602087013597506040870135966060013595509350505050565b60008151808452611877816020860160208601611a30565b601f01601f19169290920160200192915050565b6000825161189d818460208701611a30565b9190910192915050565b602081526000610570602083018461185f565b6080815260006118cd608083018761185f565b6020830195909552506040810192909252606090910152919050565b6020808252602c908201527f46756e6374696f6e206d7573742062652063616c6c6564207468726f7567682060408201526b19195b1959d85d1958d85b1b60a21b606082015260800190565b6020808252602c908201527f46756e6374696f6e206d7573742062652063616c6c6564207468726f7567682060408201526b6163746976652070726f787960a01b606082015260800190565b6020808252818101527f4f776e61626c653a2063616c6c6572206973206e6f7420746865206f776e6572604082015260600190565b6020808252602b908201527f496e697469616c697a61626c653a20636f6e7472616374206973206e6f74206960408201526a6e697469616c697a696e6760a81b606082015260800190565b60008219821115611a1457611a14611a97565b500190565b600082821015611a2b57611a2b611a97565b500390565b60005b83811015611a4b578181015183820152602001611a33565b838111156113705750506000910152565b600181811c90821680611a7057607f821691505b60208210811415611a9157634e487b7160e01b600052602260045260246000fd5b50919050565b634e487b7160e01b600052601160045260246000fd5b634e487b7160e01b600052604160045260246000fdfe360894a13ba1a3210667c828492db98dca3e2076cc3735a920a3ca505d382bbc416464726573733a206c6f772d6c6576656c2064656c65676174652063616c6c206661696c6564a2646970667358221220fcaf066833d727d01e7acbf759899143615dfd717f285f84b2f7c2fd419211ed64736f6c63430008040033")
	initWFXCode   = mustDecodeHex("0x6080604052600436106101395760003560e01c80638da5cb5b116100ab578063c5cb9b511161006f578063c5cb9b5114610364578063d0e30db014610148578063dd62ed3e14610377578063de7ea79d146103bd578063f2fde38b146103dd578063f3fef3a3146103fd57610148565b80638da5cb5b146102bf57806395d89b41146102f15780639dc29fac14610306578063a9059cbb14610326578063b86d52981461034657610148565b80633659cfe6116100fd5780633659cfe61461020c57806340c10f191461022c5780634f1ef2861461024c57806352d1902d1461025f57806370a0823114610274578063715018a6146102aa57610148565b806306fdde0314610150578063095ea7b31461017b57806318160ddd146101ab57806323b872dd146101ca578063313ce567146101ea57610148565b366101485761014661041d565b005b61014661041d565b34801561015c57600080fd5b5061016561045e565b60405161017291906119b4565b60405180910390f35b34801561018757600080fd5b5061019b610196366004611865565b6104f0565b6040519015158152602001610172565b3480156101b757600080fd5b5060cc545b604051908152602001610172565b3480156101d657600080fd5b5061019b6101e53660046117c4565b610546565b3480156101f657600080fd5b5060cb5460405160ff9091168152602001610172565b34801561021857600080fd5b50610146610227366004611745565b6105f5565b34801561023857600080fd5b50610146610247366004611865565b6106d5565b61014661025a366004611804565b61070d565b34801561026b57600080fd5b506101bc6107da565b34801561028057600080fd5b506101bc61028f366004611745565b6001600160a01b0316600090815260cd602052604090205490565b3480156102b657600080fd5b5061014661088d565b3480156102cb57600080fd5b506097546001600160a01b03165b6040516001600160a01b039091168152602001610172565b3480156102fd57600080fd5b506101656108c3565b34801561031257600080fd5b50610146610321366004611865565b6108d2565b34801561033257600080fd5b5061019b610341366004611865565b610906565b34801561035257600080fd5b5060cf546001600160a01b03166102d9565b61019b61037236600461191a565b61091c565b34801561038357600080fd5b506101bc61039236600461178c565b6001600160a01b03918216600090815260ce6020908152604080832093909416825291909152205490565b3480156103c957600080fd5b506101466103d836600461188f565b6109e0565b3480156103e957600080fd5b506101466103f8366004611745565b610aff565b34801561040957600080fd5b50610146610418366004611761565b610b97565b6104273334610c1d565b60405134815233907fe1fffcc4923d04b559f4d29a8bfc6cda04eb5b0d3c460751c2402c5c5cc9109c9060200160405180910390a2565b606060c9805461046d90611b69565b80601f016020809104026020016040519081016040528092919081815260200182805461049990611b69565b80156104e65780601f106104bb576101008083540402835291602001916104e6565b820191906000526020600020905b8154815290600101906020018083116104c957829003601f168201915b5050505050905090565b60006104fd338484610cf5565b6040518281526001600160a01b0384169033907f8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b9259060200160405180910390a350600192915050565b6001600160a01b038316600090815260ce60209081526040808320338452909152812054828110156105c95760405162461bcd60e51b815260206004820152602160248201527f7472616e7366657220616d6f756e74206578636565647320616c6c6f77616e636044820152606560f81b60648201526084015b60405180910390fd5b6105dd85336105d88685611b26565b610cf5565b6105e8858585610d77565b60019150505b9392505050565b306001600160a01b037f000000000000000000000000000000000000000000000000000000000000100216141561063e5760405162461bcd60e51b81526004016105c0906119f6565b7f00000000000000000000000000000000000000000000000000000000000010026001600160a01b0316610687600080516020611be6833981519152546001600160a01b031690565b6001600160a01b0316146106ad5760405162461bcd60e51b81526004016105c090611a42565b6106b681610f26565b604080516000808252602082019092526106d291839190610f50565b50565b6097546001600160a01b031633146106ff5760405162461bcd60e51b81526004016105c090611a8e565b6107098282610c1d565b5050565b306001600160a01b037f00000000000000000000000000000000000000000000000000000000000010021614156107565760405162461bcd60e51b81526004016105c0906119f6565b7f00000000000000000000000000000000000000000000000000000000000010026001600160a01b031661079f600080516020611be6833981519152546001600160a01b031690565b6001600160a01b0316146107c55760405162461bcd60e51b81526004016105c090611a42565b6107ce82610f26565b61070982826001610f50565b6000306001600160a01b037f0000000000000000000000000000000000000000000000000000000000001002161461087a5760405162461bcd60e51b815260206004820152603860248201527f555550535570677261646561626c653a206d757374206e6f742062652063616c60448201527f6c6564207468726f7567682064656c656761746563616c6c000000000000000060648201526084016105c0565b50600080516020611be683398151915290565b6097546001600160a01b031633146108b75760405162461bcd60e51b81526004016105c090611a8e565b6108c160006110cf565b565b606060ca805461046d90611b69565b6097546001600160a01b031633146108fc5760405162461bcd60e51b81526004016105c090611a8e565b6107098282611121565b6000610913338484610d77565b50600192915050565b600063ffffffff333b16156109735760405162461bcd60e51b815260206004820152601960248201527f63616c6c65722063616e6e6f7420626520636f6e74726163740000000000000060448201526064016105c0565b34156109815761098161041d565b61098e3386868686611263565b336001600160a01b03167f282dd1817b996776123a00596764d4d54cc16460c9854f7a23f6be020ba0463d868686866040516109cd94939291906119c7565b60405180910390a2506001949350505050565b600054610100900460ff166109fb5760005460ff16156109ff565b303b155b610a625760405162461bcd60e51b815260206004820152602e60248201527f496e697469616c697a61626c653a20636f6e747261637420697320616c72656160448201526d191e481a5b9a5d1a585b1a5e995960921b60648201526084016105c0565b600054610100900460ff16158015610a84576000805461ffff19166101011790555b8451610a979060c9906020880190611617565b508351610aab9060ca906020870190611617565b5060cb805460ff191660ff851617905560cf80546001600160a01b0319166001600160a01b038416179055610ade61135d565b610ae661138c565b8015610af8576000805461ff00191690555b5050505050565b6097546001600160a01b03163314610b295760405162461bcd60e51b81526004016105c090611a8e565b6001600160a01b038116610b8e5760405162461bcd60e51b815260206004820152602660248201527f4f776e61626c653a206e6577206f776e657220697320746865207a65726f206160448201526564647265737360d01b60648201526084016105c0565b6106d2816110cf565b610ba13382611121565b6040516001600160a01b0383169082156108fc029083906000818181858888f19350505050158015610bd7573d6000803e3d6000fd5b506040518181526001600160a01b0383169033907f9b1bfa7fa9ee420a16e124f794c35ac9f90472acc99140eb2f6447c714cad8eb906020015b60405180910390a35050565b6001600160a01b038216610c735760405162461bcd60e51b815260206004820152601860248201527f6d696e7420746f20746865207a65726f2061646472657373000000000000000060448201526064016105c0565b8060cc6000828254610c859190611b0e565b90915550506001600160a01b038216600090815260cd602052604081208054839290610cb2908490611b0e565b90915550506040518181526001600160a01b038316906000907fddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef90602001610c11565b6001600160a01b038316610d4b5760405162461bcd60e51b815260206004820152601d60248201527f617070726f76652066726f6d20746865207a65726f206164647265737300000060448201526064016105c0565b6001600160a01b03928316600090815260ce602090815260408083209490951682529290925291902055565b6001600160a01b038316610dcd5760405162461bcd60e51b815260206004820152601e60248201527f7472616e736665722066726f6d20746865207a65726f2061646472657373000060448201526064016105c0565b6001600160a01b038216610e235760405162461bcd60e51b815260206004820152601c60248201527f7472616e7366657220746f20746865207a65726f20616464726573730000000060448201526064016105c0565b6001600160a01b038316600090815260cd602052604090205481811015610e8c5760405162461bcd60e51b815260206004820152601f60248201527f7472616e7366657220616d6f756e7420657863656564732062616c616e63650060448201526064016105c0565b610e968282611b26565b6001600160a01b03808616600090815260cd60205260408082209390935590851681529081208054849290610ecc908490611b0e565b92505081905550826001600160a01b0316846001600160a01b03167fddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef84604051610f1891815260200190565b60405180910390a350505050565b6097546001600160a01b031633146106d25760405162461bcd60e51b81526004016105c090611a8e565b7f4910fdfa16fed3260ed0e7147f7cc6da11a60208b5b9406d12a635614ffd91435460ff1615610f8857610f83836113b3565b505050565b826001600160a01b03166352d1902d6040518163ffffffff1660e01b815260040160206040518083038186803b158015610fc157600080fd5b505afa925050508015610ff1575060408051601f3d908101601f19168201909252610fee91810190611877565b60015b6110545760405162461bcd60e51b815260206004820152602e60248201527f45524331393637557067726164653a206e657720696d706c656d656e7461746960448201526d6f6e206973206e6f74205555505360901b60648201526084016105c0565b600080516020611be683398151915281146110c35760405162461bcd60e51b815260206004820152602960248201527f45524331393637557067726164653a20756e737570706f727465642070726f786044820152681a58589b195555525160ba1b60648201526084016105c0565b50610f8383838361144f565b609780546001600160a01b038381166001600160a01b0319831681179093556040519116919082907f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e090600090a35050565b6001600160a01b0382166111775760405162461bcd60e51b815260206004820152601a60248201527f6275726e2066726f6d20746865207a65726f206164647265737300000000000060448201526064016105c0565b6001600160a01b038216600090815260cd6020526040902054818110156111e05760405162461bcd60e51b815260206004820152601b60248201527f6275726e20616d6f756e7420657863656564732062616c616e6365000000000060448201526064016105c0565b6111ea8282611b26565b6001600160a01b038416600090815260cd602052604081209190915560cc8054849290611218908490611b26565b90915550506040518281526000906001600160a01b038516907fddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef9060200160405180910390a3505050565b6001600160a01b0385166112b95760405162461bcd60e51b815260206004820152601e60248201527f7472616e736665722066726f6d20746865207a65726f2061646472657373000060448201526064016105c0565b60008451116112fe5760405162461bcd60e51b81526020600482015260116024820152701a5b9d985b1a59081c9958da5c1a595b9d607a1b60448201526064016105c0565b8061133c5760405162461bcd60e51b815260206004820152600e60248201526d1a5b9d985b1a59081d185c99d95d60921b60448201526064016105c0565b60cf54610af89086906001600160a01b03166113588587611b0e565b610d77565b600054610100900460ff166113845760405162461bcd60e51b81526004016105c090611ac3565b6108c161147a565b600054610100900460ff166108c15760405162461bcd60e51b81526004016105c090611ac3565b6001600160a01b0381163b6114205760405162461bcd60e51b815260206004820152602d60248201527f455243313936373a206e657720696d706c656d656e746174696f6e206973206e60448201526c1bdd08184818dbdb9d1c9858dd609a1b60648201526084016105c0565b600080516020611be683398151915280546001600160a01b0319166001600160a01b0392909216919091179055565b611458836114aa565b6000825111806114655750805b15610f835761147483836114ea565b50505050565b600054610100900460ff166114a15760405162461bcd60e51b81526004016105c090611ac3565b6108c1336110cf565b6114b3816113b3565b6040516001600160a01b038216907fbc7cd75a20ee27fd9adebab32041f755214dbc6bffa90cc0225b39da2e5c2d3b90600090a250565b60606001600160a01b0383163b6115525760405162461bcd60e51b815260206004820152602660248201527f416464726573733a2064656c65676174652063616c6c20746f206e6f6e2d636f6044820152651b9d1c9858dd60d21b60648201526084016105c0565b600080846001600160a01b03168460405161156d9190611998565b600060405180830381855af49150503d80600081146115a8576040519150601f19603f3d011682016040523d82523d6000602084013e6115ad565b606091505b50915091506115d58282604051806060016040528060278152602001611c06602791396115de565b95945050505050565b606083156115ed5750816105ee565b8251156115fd5782518084602001fd5b8160405162461bcd60e51b81526004016105c091906119b4565b82805461162390611b69565b90600052602060002090601f016020900481019282611645576000855561168b565b82601f1061165e57805160ff191683800117855561168b565b8280016001018555821561168b579182015b8281111561168b578251825591602001919060010190611670565b5061169792915061169b565b5090565b5b80821115611697576000815560010161169c565b600067ffffffffffffffff808411156116cb576116cb611bba565b604051601f8501601f19908116603f011681019082821181831017156116f3576116f3611bba565b8160405280935085815286868601111561170c57600080fd5b858560208301376000602087830101525050509392505050565b600082601f830112611736578081fd5b6105ee838335602085016116b0565b600060208284031215611756578081fd5b81356105ee81611bd0565b60008060408385031215611773578081fd5b823561177e81611bd0565b946020939093013593505050565b6000806040838503121561179e578182fd5b82356117a981611bd0565b915060208301356117b981611bd0565b809150509250929050565b6000806000606084860312156117d8578081fd5b83356117e381611bd0565b925060208401356117f381611bd0565b929592945050506040919091013590565b60008060408385031215611816578182fd5b823561182181611bd0565b9150602083013567ffffffffffffffff81111561183c578182fd5b8301601f8101851361184c578182fd5b61185b858235602084016116b0565b9150509250929050565b60008060408385031215611773578182fd5b600060208284031215611888578081fd5b5051919050565b600080600080608085870312156118a4578081fd5b843567ffffffffffffffff808211156118bb578283fd5b6118c788838901611726565b955060208701359150808211156118dc578283fd5b506118e987828801611726565b935050604085013560ff811681146118ff578182fd5b9150606085013561190f81611bd0565b939692955090935050565b6000806000806080858703121561192f578384fd5b843567ffffffffffffffff811115611945578485fd5b61195187828801611726565b97602087013597506040870135966060013595509350505050565b60008151808452611984816020860160208601611b3d565b601f01601f19169290920160200192915050565b600082516119aa818460208701611b3d565b9190910192915050565b6020815260006105ee602083018461196c565b6080815260006119da608083018761196c565b6020830195909552506040810192909252606090910152919050565b6020808252602c908201527f46756e6374696f6e206d7573742062652063616c6c6564207468726f7567682060408201526b19195b1959d85d1958d85b1b60a21b606082015260800190565b6020808252602c908201527f46756e6374696f6e206d7573742062652063616c6c6564207468726f7567682060408201526b6163746976652070726f787960a01b606082015260800190565b6020808252818101527f4f776e61626c653a2063616c6c6572206973206e6f7420746865206f776e6572604082015260600190565b6020808252602b908201527f496e697469616c697a61626c653a20636f6e7472616374206973206e6f74206960408201526a6e697469616c697a696e6760a81b606082015260800190565b60008219821115611b2157611b21611ba4565b500190565b600082821015611b3857611b38611ba4565b500390565b60005b83811015611b58578181015183820152602001611b40565b838111156114745750506000910152565b600181811c90821680611b7d57607f821691505b60208210811415611b9e57634e487b7160e01b600052602260045260246000fd5b50919050565b634e487b7160e01b600052601160045260246000fd5b634e487b7160e01b600052604160045260246000fd5b6001600160a01b03811681146106d257600080fdfe360894a13ba1a3210667c828492db98dca3e2076cc3735a920a3ca505d382bbc416464726573733a206c6f772d6c6576656c2064656c65676174652063616c6c206661696c6564a26469706673582212203c27e83d0b8c9b8527404ed6d99dbdd9ee3310b5e1ae03b9506a430d642661b564736f6c63430008040033")
)

var (
	fip20Init = Contract{
		Address: common.HexToAddress(FIP20LogicAddress),
		ABI:     mustABIJson(contract.FIP20ABI),
		Bin:     mustDecodeHex(contract.FIP20Bin),
		Code:    initERC20Code,
	}
	wfxInit = Contract{
		Address: common.HexToAddress(WFXLogicAddress),
		ABI:     mustABIJson(contract.WFXABI),
		Bin:     mustDecodeHex(contract.WFXBin),
		Code:    initWFXCode,
	}
	erc1967Proxy = Contract{
		Address: common.Address{},
		ABI:     mustABIJson(contract.ERC1967ProxyABI),
		Bin:     mustDecodeHex(contract.ERC1967ProxyBin),
		Code:    []byte{},
	}
)

func GetERC20() Contract {
	return fip20Init
}

func GetWFX() Contract {
	return wfxInit
}

func GetERC1967Proxy() Contract {
	return erc1967Proxy
}

func mustDecodeHex(str string) []byte {
	bz, err := hexutil.Decode(str)
	if err != nil {
		panic(err)
	}
	return bz
}

func mustABIJson(str string) abi.ABI {
	j, err := abi.JSON(strings.NewReader(str))
	if err != nil {
		panic(err)
	}
	return j
}
