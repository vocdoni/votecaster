// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package DelegateRegistry

import (
	"errors"
	"math/big"
	"strings"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = errors.New
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
	_ = abi.ConvertType
)

// IDelegateRegistryDelegation is an auto generated low-level Go binding around an user-defined struct.
type IDelegateRegistryDelegation struct {
	Type     uint8
	To       common.Address
	From     common.Address
	Rights   [32]byte
	Contract common.Address
	TokenId  *big.Int
	Amount   *big.Int
}

// DelegateRegistryMetaData contains all meta data concerning the DelegateRegistry contract.
var DelegateRegistryMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[],\"name\":\"MulticallFailed\",\"type\":\"error\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"bytes32\",\"name\":\"rights\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"bool\",\"name\":\"enable\",\"type\":\"bool\"}],\"name\":\"DelegateAll\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"contract_\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"bytes32\",\"name\":\"rights\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"bool\",\"name\":\"enable\",\"type\":\"bool\"}],\"name\":\"DelegateContract\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"contract_\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"bytes32\",\"name\":\"rights\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"DelegateERC1155\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"contract_\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"bytes32\",\"name\":\"rights\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"DelegateERC20\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"contract_\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"bytes32\",\"name\":\"rights\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"bool\",\"name\":\"enable\",\"type\":\"bool\"}],\"name\":\"DelegateERC721\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"rights\",\"type\":\"bytes32\"}],\"name\":\"checkDelegateForAll\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"valid\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"contract_\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"rights\",\"type\":\"bytes32\"}],\"name\":\"checkDelegateForContract\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"valid\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"contract_\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"},{\"internalType\":\"bytes32\",\"name\":\"rights\",\"type\":\"bytes32\"}],\"name\":\"checkDelegateForERC1155\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"contract_\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"rights\",\"type\":\"bytes32\"}],\"name\":\"checkDelegateForERC20\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"contract_\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"},{\"internalType\":\"bytes32\",\"name\":\"rights\",\"type\":\"bytes32\"}],\"name\":\"checkDelegateForERC721\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"valid\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"rights\",\"type\":\"bytes32\"},{\"internalType\":\"bool\",\"name\":\"enable\",\"type\":\"bool\"}],\"name\":\"delegateAll\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"hash\",\"type\":\"bytes32\"}],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"contract_\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"rights\",\"type\":\"bytes32\"},{\"internalType\":\"bool\",\"name\":\"enable\",\"type\":\"bool\"}],\"name\":\"delegateContract\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"hash\",\"type\":\"bytes32\"}],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"contract_\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"},{\"internalType\":\"bytes32\",\"name\":\"rights\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"delegateERC1155\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"hash\",\"type\":\"bytes32\"}],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"contract_\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"rights\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"delegateERC20\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"hash\",\"type\":\"bytes32\"}],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"contract_\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"},{\"internalType\":\"bytes32\",\"name\":\"rights\",\"type\":\"bytes32\"},{\"internalType\":\"bool\",\"name\":\"enable\",\"type\":\"bool\"}],\"name\":\"delegateERC721\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"hash\",\"type\":\"bytes32\"}],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32[]\",\"name\":\"hashes\",\"type\":\"bytes32[]\"}],\"name\":\"getDelegationsFromHashes\",\"outputs\":[{\"components\":[{\"internalType\":\"enumIDelegateRegistry.DelegationType\",\"name\":\"type_\",\"type\":\"uint8\"},{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"rights\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"contract_\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"internalType\":\"structIDelegateRegistry.Delegation[]\",\"name\":\"delegations_\",\"type\":\"tuple[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"}],\"name\":\"getIncomingDelegationHashes\",\"outputs\":[{\"internalType\":\"bytes32[]\",\"name\":\"delegationHashes\",\"type\":\"bytes32[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"}],\"name\":\"getIncomingDelegations\",\"outputs\":[{\"components\":[{\"internalType\":\"enumIDelegateRegistry.DelegationType\",\"name\":\"type_\",\"type\":\"uint8\"},{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"rights\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"contract_\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"internalType\":\"structIDelegateRegistry.Delegation[]\",\"name\":\"delegations_\",\"type\":\"tuple[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"}],\"name\":\"getOutgoingDelegationHashes\",\"outputs\":[{\"internalType\":\"bytes32[]\",\"name\":\"delegationHashes\",\"type\":\"bytes32[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"}],\"name\":\"getOutgoingDelegations\",\"outputs\":[{\"components\":[{\"internalType\":\"enumIDelegateRegistry.DelegationType\",\"name\":\"type_\",\"type\":\"uint8\"},{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"rights\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"contract_\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"internalType\":\"structIDelegateRegistry.Delegation[]\",\"name\":\"delegations_\",\"type\":\"tuple[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes[]\",\"name\":\"data\",\"type\":\"bytes[]\"}],\"name\":\"multicall\",\"outputs\":[{\"internalType\":\"bytes[]\",\"name\":\"results\",\"type\":\"bytes[]\"}],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"location\",\"type\":\"bytes32\"}],\"name\":\"readSlot\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"contents\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32[]\",\"name\":\"locations\",\"type\":\"bytes32[]\"}],\"name\":\"readSlots\",\"outputs\":[{\"internalType\":\"bytes32[]\",\"name\":\"contents\",\"type\":\"bytes32[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes4\",\"name\":\"interfaceId\",\"type\":\"bytes4\"}],\"name\":\"supportsInterface\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"sweep\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
}

// DelegateRegistryABI is the input ABI used to generate the binding from.
// Deprecated: Use DelegateRegistryMetaData.ABI instead.
var DelegateRegistryABI = DelegateRegistryMetaData.ABI

// DelegateRegistry is an auto generated Go binding around an Ethereum contract.
type DelegateRegistry struct {
	DelegateRegistryCaller     // Read-only binding to the contract
	DelegateRegistryTransactor // Write-only binding to the contract
	DelegateRegistryFilterer   // Log filterer for contract events
}

// DelegateRegistryCaller is an auto generated read-only Go binding around an Ethereum contract.
type DelegateRegistryCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// DelegateRegistryTransactor is an auto generated write-only Go binding around an Ethereum contract.
type DelegateRegistryTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// DelegateRegistryFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type DelegateRegistryFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// DelegateRegistrySession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type DelegateRegistrySession struct {
	Contract     *DelegateRegistry // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// DelegateRegistryCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type DelegateRegistryCallerSession struct {
	Contract *DelegateRegistryCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts           // Call options to use throughout this session
}

// DelegateRegistryTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type DelegateRegistryTransactorSession struct {
	Contract     *DelegateRegistryTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts           // Transaction auth options to use throughout this session
}

// DelegateRegistryRaw is an auto generated low-level Go binding around an Ethereum contract.
type DelegateRegistryRaw struct {
	Contract *DelegateRegistry // Generic contract binding to access the raw methods on
}

// DelegateRegistryCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type DelegateRegistryCallerRaw struct {
	Contract *DelegateRegistryCaller // Generic read-only contract binding to access the raw methods on
}

// DelegateRegistryTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type DelegateRegistryTransactorRaw struct {
	Contract *DelegateRegistryTransactor // Generic write-only contract binding to access the raw methods on
}

// NewDelegateRegistry creates a new instance of DelegateRegistry, bound to a specific deployed contract.
func NewDelegateRegistry(address common.Address, backend bind.ContractBackend) (*DelegateRegistry, error) {
	contract, err := bindDelegateRegistry(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &DelegateRegistry{DelegateRegistryCaller: DelegateRegistryCaller{contract: contract}, DelegateRegistryTransactor: DelegateRegistryTransactor{contract: contract}, DelegateRegistryFilterer: DelegateRegistryFilterer{contract: contract}}, nil
}

// NewDelegateRegistryCaller creates a new read-only instance of DelegateRegistry, bound to a specific deployed contract.
func NewDelegateRegistryCaller(address common.Address, caller bind.ContractCaller) (*DelegateRegistryCaller, error) {
	contract, err := bindDelegateRegistry(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &DelegateRegistryCaller{contract: contract}, nil
}

// NewDelegateRegistryTransactor creates a new write-only instance of DelegateRegistry, bound to a specific deployed contract.
func NewDelegateRegistryTransactor(address common.Address, transactor bind.ContractTransactor) (*DelegateRegistryTransactor, error) {
	contract, err := bindDelegateRegistry(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &DelegateRegistryTransactor{contract: contract}, nil
}

// NewDelegateRegistryFilterer creates a new log filterer instance of DelegateRegistry, bound to a specific deployed contract.
func NewDelegateRegistryFilterer(address common.Address, filterer bind.ContractFilterer) (*DelegateRegistryFilterer, error) {
	contract, err := bindDelegateRegistry(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &DelegateRegistryFilterer{contract: contract}, nil
}

// bindDelegateRegistry binds a generic wrapper to an already deployed contract.
func bindDelegateRegistry(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := DelegateRegistryMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_DelegateRegistry *DelegateRegistryRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _DelegateRegistry.Contract.DelegateRegistryCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_DelegateRegistry *DelegateRegistryRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _DelegateRegistry.Contract.DelegateRegistryTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_DelegateRegistry *DelegateRegistryRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _DelegateRegistry.Contract.DelegateRegistryTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_DelegateRegistry *DelegateRegistryCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _DelegateRegistry.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_DelegateRegistry *DelegateRegistryTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _DelegateRegistry.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_DelegateRegistry *DelegateRegistryTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _DelegateRegistry.Contract.contract.Transact(opts, method, params...)
}

// CheckDelegateForAll is a free data retrieval call binding the contract method 0xe839bd53.
//
// Solidity: function checkDelegateForAll(address to, address from, bytes32 rights) view returns(bool valid)
func (_DelegateRegistry *DelegateRegistryCaller) CheckDelegateForAll(opts *bind.CallOpts, to common.Address, from common.Address, rights [32]byte) (bool, error) {
	var out []interface{}
	err := _DelegateRegistry.contract.Call(opts, &out, "checkDelegateForAll", to, from, rights)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// CheckDelegateForAll is a free data retrieval call binding the contract method 0xe839bd53.
//
// Solidity: function checkDelegateForAll(address to, address from, bytes32 rights) view returns(bool valid)
func (_DelegateRegistry *DelegateRegistrySession) CheckDelegateForAll(to common.Address, from common.Address, rights [32]byte) (bool, error) {
	return _DelegateRegistry.Contract.CheckDelegateForAll(&_DelegateRegistry.CallOpts, to, from, rights)
}

// CheckDelegateForAll is a free data retrieval call binding the contract method 0xe839bd53.
//
// Solidity: function checkDelegateForAll(address to, address from, bytes32 rights) view returns(bool valid)
func (_DelegateRegistry *DelegateRegistryCallerSession) CheckDelegateForAll(to common.Address, from common.Address, rights [32]byte) (bool, error) {
	return _DelegateRegistry.Contract.CheckDelegateForAll(&_DelegateRegistry.CallOpts, to, from, rights)
}

// CheckDelegateForContract is a free data retrieval call binding the contract method 0x8988eea9.
//
// Solidity: function checkDelegateForContract(address to, address from, address contract_, bytes32 rights) view returns(bool valid)
func (_DelegateRegistry *DelegateRegistryCaller) CheckDelegateForContract(opts *bind.CallOpts, to common.Address, from common.Address, contract_ common.Address, rights [32]byte) (bool, error) {
	var out []interface{}
	err := _DelegateRegistry.contract.Call(opts, &out, "checkDelegateForContract", to, from, contract_, rights)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// CheckDelegateForContract is a free data retrieval call binding the contract method 0x8988eea9.
//
// Solidity: function checkDelegateForContract(address to, address from, address contract_, bytes32 rights) view returns(bool valid)
func (_DelegateRegistry *DelegateRegistrySession) CheckDelegateForContract(to common.Address, from common.Address, contract_ common.Address, rights [32]byte) (bool, error) {
	return _DelegateRegistry.Contract.CheckDelegateForContract(&_DelegateRegistry.CallOpts, to, from, contract_, rights)
}

// CheckDelegateForContract is a free data retrieval call binding the contract method 0x8988eea9.
//
// Solidity: function checkDelegateForContract(address to, address from, address contract_, bytes32 rights) view returns(bool valid)
func (_DelegateRegistry *DelegateRegistryCallerSession) CheckDelegateForContract(to common.Address, from common.Address, contract_ common.Address, rights [32]byte) (bool, error) {
	return _DelegateRegistry.Contract.CheckDelegateForContract(&_DelegateRegistry.CallOpts, to, from, contract_, rights)
}

// CheckDelegateForERC1155 is a free data retrieval call binding the contract method 0xb8705875.
//
// Solidity: function checkDelegateForERC1155(address to, address from, address contract_, uint256 tokenId, bytes32 rights) view returns(uint256 amount)
func (_DelegateRegistry *DelegateRegistryCaller) CheckDelegateForERC1155(opts *bind.CallOpts, to common.Address, from common.Address, contract_ common.Address, tokenId *big.Int, rights [32]byte) (*big.Int, error) {
	var out []interface{}
	err := _DelegateRegistry.contract.Call(opts, &out, "checkDelegateForERC1155", to, from, contract_, tokenId, rights)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// CheckDelegateForERC1155 is a free data retrieval call binding the contract method 0xb8705875.
//
// Solidity: function checkDelegateForERC1155(address to, address from, address contract_, uint256 tokenId, bytes32 rights) view returns(uint256 amount)
func (_DelegateRegistry *DelegateRegistrySession) CheckDelegateForERC1155(to common.Address, from common.Address, contract_ common.Address, tokenId *big.Int, rights [32]byte) (*big.Int, error) {
	return _DelegateRegistry.Contract.CheckDelegateForERC1155(&_DelegateRegistry.CallOpts, to, from, contract_, tokenId, rights)
}

// CheckDelegateForERC1155 is a free data retrieval call binding the contract method 0xb8705875.
//
// Solidity: function checkDelegateForERC1155(address to, address from, address contract_, uint256 tokenId, bytes32 rights) view returns(uint256 amount)
func (_DelegateRegistry *DelegateRegistryCallerSession) CheckDelegateForERC1155(to common.Address, from common.Address, contract_ common.Address, tokenId *big.Int, rights [32]byte) (*big.Int, error) {
	return _DelegateRegistry.Contract.CheckDelegateForERC1155(&_DelegateRegistry.CallOpts, to, from, contract_, tokenId, rights)
}

// CheckDelegateForERC20 is a free data retrieval call binding the contract method 0xba63c817.
//
// Solidity: function checkDelegateForERC20(address to, address from, address contract_, bytes32 rights) view returns(uint256 amount)
func (_DelegateRegistry *DelegateRegistryCaller) CheckDelegateForERC20(opts *bind.CallOpts, to common.Address, from common.Address, contract_ common.Address, rights [32]byte) (*big.Int, error) {
	var out []interface{}
	err := _DelegateRegistry.contract.Call(opts, &out, "checkDelegateForERC20", to, from, contract_, rights)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// CheckDelegateForERC20 is a free data retrieval call binding the contract method 0xba63c817.
//
// Solidity: function checkDelegateForERC20(address to, address from, address contract_, bytes32 rights) view returns(uint256 amount)
func (_DelegateRegistry *DelegateRegistrySession) CheckDelegateForERC20(to common.Address, from common.Address, contract_ common.Address, rights [32]byte) (*big.Int, error) {
	return _DelegateRegistry.Contract.CheckDelegateForERC20(&_DelegateRegistry.CallOpts, to, from, contract_, rights)
}

// CheckDelegateForERC20 is a free data retrieval call binding the contract method 0xba63c817.
//
// Solidity: function checkDelegateForERC20(address to, address from, address contract_, bytes32 rights) view returns(uint256 amount)
func (_DelegateRegistry *DelegateRegistryCallerSession) CheckDelegateForERC20(to common.Address, from common.Address, contract_ common.Address, rights [32]byte) (*big.Int, error) {
	return _DelegateRegistry.Contract.CheckDelegateForERC20(&_DelegateRegistry.CallOpts, to, from, contract_, rights)
}

// CheckDelegateForERC721 is a free data retrieval call binding the contract method 0xb9f36874.
//
// Solidity: function checkDelegateForERC721(address to, address from, address contract_, uint256 tokenId, bytes32 rights) view returns(bool valid)
func (_DelegateRegistry *DelegateRegistryCaller) CheckDelegateForERC721(opts *bind.CallOpts, to common.Address, from common.Address, contract_ common.Address, tokenId *big.Int, rights [32]byte) (bool, error) {
	var out []interface{}
	err := _DelegateRegistry.contract.Call(opts, &out, "checkDelegateForERC721", to, from, contract_, tokenId, rights)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// CheckDelegateForERC721 is a free data retrieval call binding the contract method 0xb9f36874.
//
// Solidity: function checkDelegateForERC721(address to, address from, address contract_, uint256 tokenId, bytes32 rights) view returns(bool valid)
func (_DelegateRegistry *DelegateRegistrySession) CheckDelegateForERC721(to common.Address, from common.Address, contract_ common.Address, tokenId *big.Int, rights [32]byte) (bool, error) {
	return _DelegateRegistry.Contract.CheckDelegateForERC721(&_DelegateRegistry.CallOpts, to, from, contract_, tokenId, rights)
}

// CheckDelegateForERC721 is a free data retrieval call binding the contract method 0xb9f36874.
//
// Solidity: function checkDelegateForERC721(address to, address from, address contract_, uint256 tokenId, bytes32 rights) view returns(bool valid)
func (_DelegateRegistry *DelegateRegistryCallerSession) CheckDelegateForERC721(to common.Address, from common.Address, contract_ common.Address, tokenId *big.Int, rights [32]byte) (bool, error) {
	return _DelegateRegistry.Contract.CheckDelegateForERC721(&_DelegateRegistry.CallOpts, to, from, contract_, tokenId, rights)
}

// GetDelegationsFromHashes is a free data retrieval call binding the contract method 0x4705ed38.
//
// Solidity: function getDelegationsFromHashes(bytes32[] hashes) view returns((uint8,address,address,bytes32,address,uint256,uint256)[] delegations_)
func (_DelegateRegistry *DelegateRegistryCaller) GetDelegationsFromHashes(opts *bind.CallOpts, hashes [][32]byte) ([]IDelegateRegistryDelegation, error) {
	var out []interface{}
	err := _DelegateRegistry.contract.Call(opts, &out, "getDelegationsFromHashes", hashes)

	if err != nil {
		return *new([]IDelegateRegistryDelegation), err
	}

	out0 := *abi.ConvertType(out[0], new([]IDelegateRegistryDelegation)).(*[]IDelegateRegistryDelegation)

	return out0, err

}

// GetDelegationsFromHashes is a free data retrieval call binding the contract method 0x4705ed38.
//
// Solidity: function getDelegationsFromHashes(bytes32[] hashes) view returns((uint8,address,address,bytes32,address,uint256,uint256)[] delegations_)
func (_DelegateRegistry *DelegateRegistrySession) GetDelegationsFromHashes(hashes [][32]byte) ([]IDelegateRegistryDelegation, error) {
	return _DelegateRegistry.Contract.GetDelegationsFromHashes(&_DelegateRegistry.CallOpts, hashes)
}

// GetDelegationsFromHashes is a free data retrieval call binding the contract method 0x4705ed38.
//
// Solidity: function getDelegationsFromHashes(bytes32[] hashes) view returns((uint8,address,address,bytes32,address,uint256,uint256)[] delegations_)
func (_DelegateRegistry *DelegateRegistryCallerSession) GetDelegationsFromHashes(hashes [][32]byte) ([]IDelegateRegistryDelegation, error) {
	return _DelegateRegistry.Contract.GetDelegationsFromHashes(&_DelegateRegistry.CallOpts, hashes)
}

// GetIncomingDelegationHashes is a free data retrieval call binding the contract method 0x063182a5.
//
// Solidity: function getIncomingDelegationHashes(address to) view returns(bytes32[] delegationHashes)
func (_DelegateRegistry *DelegateRegistryCaller) GetIncomingDelegationHashes(opts *bind.CallOpts, to common.Address) ([][32]byte, error) {
	var out []interface{}
	err := _DelegateRegistry.contract.Call(opts, &out, "getIncomingDelegationHashes", to)

	if err != nil {
		return *new([][32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([][32]byte)).(*[][32]byte)

	return out0, err

}

// GetIncomingDelegationHashes is a free data retrieval call binding the contract method 0x063182a5.
//
// Solidity: function getIncomingDelegationHashes(address to) view returns(bytes32[] delegationHashes)
func (_DelegateRegistry *DelegateRegistrySession) GetIncomingDelegationHashes(to common.Address) ([][32]byte, error) {
	return _DelegateRegistry.Contract.GetIncomingDelegationHashes(&_DelegateRegistry.CallOpts, to)
}

// GetIncomingDelegationHashes is a free data retrieval call binding the contract method 0x063182a5.
//
// Solidity: function getIncomingDelegationHashes(address to) view returns(bytes32[] delegationHashes)
func (_DelegateRegistry *DelegateRegistryCallerSession) GetIncomingDelegationHashes(to common.Address) ([][32]byte, error) {
	return _DelegateRegistry.Contract.GetIncomingDelegationHashes(&_DelegateRegistry.CallOpts, to)
}

// GetIncomingDelegations is a free data retrieval call binding the contract method 0x42f87c25.
//
// Solidity: function getIncomingDelegations(address to) view returns((uint8,address,address,bytes32,address,uint256,uint256)[] delegations_)
func (_DelegateRegistry *DelegateRegistryCaller) GetIncomingDelegations(opts *bind.CallOpts, to common.Address) ([]IDelegateRegistryDelegation, error) {
	var out []interface{}
	err := _DelegateRegistry.contract.Call(opts, &out, "getIncomingDelegations", to)

	if err != nil {
		return *new([]IDelegateRegistryDelegation), err
	}

	out0 := *abi.ConvertType(out[0], new([]IDelegateRegistryDelegation)).(*[]IDelegateRegistryDelegation)

	return out0, err

}

// GetIncomingDelegations is a free data retrieval call binding the contract method 0x42f87c25.
//
// Solidity: function getIncomingDelegations(address to) view returns((uint8,address,address,bytes32,address,uint256,uint256)[] delegations_)
func (_DelegateRegistry *DelegateRegistrySession) GetIncomingDelegations(to common.Address) ([]IDelegateRegistryDelegation, error) {
	return _DelegateRegistry.Contract.GetIncomingDelegations(&_DelegateRegistry.CallOpts, to)
}

// GetIncomingDelegations is a free data retrieval call binding the contract method 0x42f87c25.
//
// Solidity: function getIncomingDelegations(address to) view returns((uint8,address,address,bytes32,address,uint256,uint256)[] delegations_)
func (_DelegateRegistry *DelegateRegistryCallerSession) GetIncomingDelegations(to common.Address) ([]IDelegateRegistryDelegation, error) {
	return _DelegateRegistry.Contract.GetIncomingDelegations(&_DelegateRegistry.CallOpts, to)
}

// GetOutgoingDelegationHashes is a free data retrieval call binding the contract method 0x01a920a0.
//
// Solidity: function getOutgoingDelegationHashes(address from) view returns(bytes32[] delegationHashes)
func (_DelegateRegistry *DelegateRegistryCaller) GetOutgoingDelegationHashes(opts *bind.CallOpts, from common.Address) ([][32]byte, error) {
	var out []interface{}
	err := _DelegateRegistry.contract.Call(opts, &out, "getOutgoingDelegationHashes", from)

	if err != nil {
		return *new([][32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([][32]byte)).(*[][32]byte)

	return out0, err

}

// GetOutgoingDelegationHashes is a free data retrieval call binding the contract method 0x01a920a0.
//
// Solidity: function getOutgoingDelegationHashes(address from) view returns(bytes32[] delegationHashes)
func (_DelegateRegistry *DelegateRegistrySession) GetOutgoingDelegationHashes(from common.Address) ([][32]byte, error) {
	return _DelegateRegistry.Contract.GetOutgoingDelegationHashes(&_DelegateRegistry.CallOpts, from)
}

// GetOutgoingDelegationHashes is a free data retrieval call binding the contract method 0x01a920a0.
//
// Solidity: function getOutgoingDelegationHashes(address from) view returns(bytes32[] delegationHashes)
func (_DelegateRegistry *DelegateRegistryCallerSession) GetOutgoingDelegationHashes(from common.Address) ([][32]byte, error) {
	return _DelegateRegistry.Contract.GetOutgoingDelegationHashes(&_DelegateRegistry.CallOpts, from)
}

// GetOutgoingDelegations is a free data retrieval call binding the contract method 0x51525e9a.
//
// Solidity: function getOutgoingDelegations(address from) view returns((uint8,address,address,bytes32,address,uint256,uint256)[] delegations_)
func (_DelegateRegistry *DelegateRegistryCaller) GetOutgoingDelegations(opts *bind.CallOpts, from common.Address) ([]IDelegateRegistryDelegation, error) {
	var out []interface{}
	err := _DelegateRegistry.contract.Call(opts, &out, "getOutgoingDelegations", from)

	if err != nil {
		return *new([]IDelegateRegistryDelegation), err
	}

	out0 := *abi.ConvertType(out[0], new([]IDelegateRegistryDelegation)).(*[]IDelegateRegistryDelegation)

	return out0, err

}

// GetOutgoingDelegations is a free data retrieval call binding the contract method 0x51525e9a.
//
// Solidity: function getOutgoingDelegations(address from) view returns((uint8,address,address,bytes32,address,uint256,uint256)[] delegations_)
func (_DelegateRegistry *DelegateRegistrySession) GetOutgoingDelegations(from common.Address) ([]IDelegateRegistryDelegation, error) {
	return _DelegateRegistry.Contract.GetOutgoingDelegations(&_DelegateRegistry.CallOpts, from)
}

// GetOutgoingDelegations is a free data retrieval call binding the contract method 0x51525e9a.
//
// Solidity: function getOutgoingDelegations(address from) view returns((uint8,address,address,bytes32,address,uint256,uint256)[] delegations_)
func (_DelegateRegistry *DelegateRegistryCallerSession) GetOutgoingDelegations(from common.Address) ([]IDelegateRegistryDelegation, error) {
	return _DelegateRegistry.Contract.GetOutgoingDelegations(&_DelegateRegistry.CallOpts, from)
}

// ReadSlot is a free data retrieval call binding the contract method 0xe8e834a9.
//
// Solidity: function readSlot(bytes32 location) view returns(bytes32 contents)
func (_DelegateRegistry *DelegateRegistryCaller) ReadSlot(opts *bind.CallOpts, location [32]byte) ([32]byte, error) {
	var out []interface{}
	err := _DelegateRegistry.contract.Call(opts, &out, "readSlot", location)

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// ReadSlot is a free data retrieval call binding the contract method 0xe8e834a9.
//
// Solidity: function readSlot(bytes32 location) view returns(bytes32 contents)
func (_DelegateRegistry *DelegateRegistrySession) ReadSlot(location [32]byte) ([32]byte, error) {
	return _DelegateRegistry.Contract.ReadSlot(&_DelegateRegistry.CallOpts, location)
}

// ReadSlot is a free data retrieval call binding the contract method 0xe8e834a9.
//
// Solidity: function readSlot(bytes32 location) view returns(bytes32 contents)
func (_DelegateRegistry *DelegateRegistryCallerSession) ReadSlot(location [32]byte) ([32]byte, error) {
	return _DelegateRegistry.Contract.ReadSlot(&_DelegateRegistry.CallOpts, location)
}

// ReadSlots is a free data retrieval call binding the contract method 0x61451a30.
//
// Solidity: function readSlots(bytes32[] locations) view returns(bytes32[] contents)
func (_DelegateRegistry *DelegateRegistryCaller) ReadSlots(opts *bind.CallOpts, locations [][32]byte) ([][32]byte, error) {
	var out []interface{}
	err := _DelegateRegistry.contract.Call(opts, &out, "readSlots", locations)

	if err != nil {
		return *new([][32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([][32]byte)).(*[][32]byte)

	return out0, err

}

// ReadSlots is a free data retrieval call binding the contract method 0x61451a30.
//
// Solidity: function readSlots(bytes32[] locations) view returns(bytes32[] contents)
func (_DelegateRegistry *DelegateRegistrySession) ReadSlots(locations [][32]byte) ([][32]byte, error) {
	return _DelegateRegistry.Contract.ReadSlots(&_DelegateRegistry.CallOpts, locations)
}

// ReadSlots is a free data retrieval call binding the contract method 0x61451a30.
//
// Solidity: function readSlots(bytes32[] locations) view returns(bytes32[] contents)
func (_DelegateRegistry *DelegateRegistryCallerSession) ReadSlots(locations [][32]byte) ([][32]byte, error) {
	return _DelegateRegistry.Contract.ReadSlots(&_DelegateRegistry.CallOpts, locations)
}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) pure returns(bool)
func (_DelegateRegistry *DelegateRegistryCaller) SupportsInterface(opts *bind.CallOpts, interfaceId [4]byte) (bool, error) {
	var out []interface{}
	err := _DelegateRegistry.contract.Call(opts, &out, "supportsInterface", interfaceId)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) pure returns(bool)
func (_DelegateRegistry *DelegateRegistrySession) SupportsInterface(interfaceId [4]byte) (bool, error) {
	return _DelegateRegistry.Contract.SupportsInterface(&_DelegateRegistry.CallOpts, interfaceId)
}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) pure returns(bool)
func (_DelegateRegistry *DelegateRegistryCallerSession) SupportsInterface(interfaceId [4]byte) (bool, error) {
	return _DelegateRegistry.Contract.SupportsInterface(&_DelegateRegistry.CallOpts, interfaceId)
}

// DelegateAll is a paid mutator transaction binding the contract method 0x30ff3140.
//
// Solidity: function delegateAll(address to, bytes32 rights, bool enable) payable returns(bytes32 hash)
func (_DelegateRegistry *DelegateRegistryTransactor) DelegateAll(opts *bind.TransactOpts, to common.Address, rights [32]byte, enable bool) (*types.Transaction, error) {
	return _DelegateRegistry.contract.Transact(opts, "delegateAll", to, rights, enable)
}

// DelegateAll is a paid mutator transaction binding the contract method 0x30ff3140.
//
// Solidity: function delegateAll(address to, bytes32 rights, bool enable) payable returns(bytes32 hash)
func (_DelegateRegistry *DelegateRegistrySession) DelegateAll(to common.Address, rights [32]byte, enable bool) (*types.Transaction, error) {
	return _DelegateRegistry.Contract.DelegateAll(&_DelegateRegistry.TransactOpts, to, rights, enable)
}

// DelegateAll is a paid mutator transaction binding the contract method 0x30ff3140.
//
// Solidity: function delegateAll(address to, bytes32 rights, bool enable) payable returns(bytes32 hash)
func (_DelegateRegistry *DelegateRegistryTransactorSession) DelegateAll(to common.Address, rights [32]byte, enable bool) (*types.Transaction, error) {
	return _DelegateRegistry.Contract.DelegateAll(&_DelegateRegistry.TransactOpts, to, rights, enable)
}

// DelegateContract is a paid mutator transaction binding the contract method 0xd90e73ab.
//
// Solidity: function delegateContract(address to, address contract_, bytes32 rights, bool enable) payable returns(bytes32 hash)
func (_DelegateRegistry *DelegateRegistryTransactor) DelegateContract(opts *bind.TransactOpts, to common.Address, contract_ common.Address, rights [32]byte, enable bool) (*types.Transaction, error) {
	return _DelegateRegistry.contract.Transact(opts, "delegateContract", to, contract_, rights, enable)
}

// DelegateContract is a paid mutator transaction binding the contract method 0xd90e73ab.
//
// Solidity: function delegateContract(address to, address contract_, bytes32 rights, bool enable) payable returns(bytes32 hash)
func (_DelegateRegistry *DelegateRegistrySession) DelegateContract(to common.Address, contract_ common.Address, rights [32]byte, enable bool) (*types.Transaction, error) {
	return _DelegateRegistry.Contract.DelegateContract(&_DelegateRegistry.TransactOpts, to, contract_, rights, enable)
}

// DelegateContract is a paid mutator transaction binding the contract method 0xd90e73ab.
//
// Solidity: function delegateContract(address to, address contract_, bytes32 rights, bool enable) payable returns(bytes32 hash)
func (_DelegateRegistry *DelegateRegistryTransactorSession) DelegateContract(to common.Address, contract_ common.Address, rights [32]byte, enable bool) (*types.Transaction, error) {
	return _DelegateRegistry.Contract.DelegateContract(&_DelegateRegistry.TransactOpts, to, contract_, rights, enable)
}

// DelegateERC1155 is a paid mutator transaction binding the contract method 0xab764683.
//
// Solidity: function delegateERC1155(address to, address contract_, uint256 tokenId, bytes32 rights, uint256 amount) payable returns(bytes32 hash)
func (_DelegateRegistry *DelegateRegistryTransactor) DelegateERC1155(opts *bind.TransactOpts, to common.Address, contract_ common.Address, tokenId *big.Int, rights [32]byte, amount *big.Int) (*types.Transaction, error) {
	return _DelegateRegistry.contract.Transact(opts, "delegateERC1155", to, contract_, tokenId, rights, amount)
}

// DelegateERC1155 is a paid mutator transaction binding the contract method 0xab764683.
//
// Solidity: function delegateERC1155(address to, address contract_, uint256 tokenId, bytes32 rights, uint256 amount) payable returns(bytes32 hash)
func (_DelegateRegistry *DelegateRegistrySession) DelegateERC1155(to common.Address, contract_ common.Address, tokenId *big.Int, rights [32]byte, amount *big.Int) (*types.Transaction, error) {
	return _DelegateRegistry.Contract.DelegateERC1155(&_DelegateRegistry.TransactOpts, to, contract_, tokenId, rights, amount)
}

// DelegateERC1155 is a paid mutator transaction binding the contract method 0xab764683.
//
// Solidity: function delegateERC1155(address to, address contract_, uint256 tokenId, bytes32 rights, uint256 amount) payable returns(bytes32 hash)
func (_DelegateRegistry *DelegateRegistryTransactorSession) DelegateERC1155(to common.Address, contract_ common.Address, tokenId *big.Int, rights [32]byte, amount *big.Int) (*types.Transaction, error) {
	return _DelegateRegistry.Contract.DelegateERC1155(&_DelegateRegistry.TransactOpts, to, contract_, tokenId, rights, amount)
}

// DelegateERC20 is a paid mutator transaction binding the contract method 0x003c2ba6.
//
// Solidity: function delegateERC20(address to, address contract_, bytes32 rights, uint256 amount) payable returns(bytes32 hash)
func (_DelegateRegistry *DelegateRegistryTransactor) DelegateERC20(opts *bind.TransactOpts, to common.Address, contract_ common.Address, rights [32]byte, amount *big.Int) (*types.Transaction, error) {
	return _DelegateRegistry.contract.Transact(opts, "delegateERC20", to, contract_, rights, amount)
}

// DelegateERC20 is a paid mutator transaction binding the contract method 0x003c2ba6.
//
// Solidity: function delegateERC20(address to, address contract_, bytes32 rights, uint256 amount) payable returns(bytes32 hash)
func (_DelegateRegistry *DelegateRegistrySession) DelegateERC20(to common.Address, contract_ common.Address, rights [32]byte, amount *big.Int) (*types.Transaction, error) {
	return _DelegateRegistry.Contract.DelegateERC20(&_DelegateRegistry.TransactOpts, to, contract_, rights, amount)
}

// DelegateERC20 is a paid mutator transaction binding the contract method 0x003c2ba6.
//
// Solidity: function delegateERC20(address to, address contract_, bytes32 rights, uint256 amount) payable returns(bytes32 hash)
func (_DelegateRegistry *DelegateRegistryTransactorSession) DelegateERC20(to common.Address, contract_ common.Address, rights [32]byte, amount *big.Int) (*types.Transaction, error) {
	return _DelegateRegistry.Contract.DelegateERC20(&_DelegateRegistry.TransactOpts, to, contract_, rights, amount)
}

// DelegateERC721 is a paid mutator transaction binding the contract method 0xb18e2bbb.
//
// Solidity: function delegateERC721(address to, address contract_, uint256 tokenId, bytes32 rights, bool enable) payable returns(bytes32 hash)
func (_DelegateRegistry *DelegateRegistryTransactor) DelegateERC721(opts *bind.TransactOpts, to common.Address, contract_ common.Address, tokenId *big.Int, rights [32]byte, enable bool) (*types.Transaction, error) {
	return _DelegateRegistry.contract.Transact(opts, "delegateERC721", to, contract_, tokenId, rights, enable)
}

// DelegateERC721 is a paid mutator transaction binding the contract method 0xb18e2bbb.
//
// Solidity: function delegateERC721(address to, address contract_, uint256 tokenId, bytes32 rights, bool enable) payable returns(bytes32 hash)
func (_DelegateRegistry *DelegateRegistrySession) DelegateERC721(to common.Address, contract_ common.Address, tokenId *big.Int, rights [32]byte, enable bool) (*types.Transaction, error) {
	return _DelegateRegistry.Contract.DelegateERC721(&_DelegateRegistry.TransactOpts, to, contract_, tokenId, rights, enable)
}

// DelegateERC721 is a paid mutator transaction binding the contract method 0xb18e2bbb.
//
// Solidity: function delegateERC721(address to, address contract_, uint256 tokenId, bytes32 rights, bool enable) payable returns(bytes32 hash)
func (_DelegateRegistry *DelegateRegistryTransactorSession) DelegateERC721(to common.Address, contract_ common.Address, tokenId *big.Int, rights [32]byte, enable bool) (*types.Transaction, error) {
	return _DelegateRegistry.Contract.DelegateERC721(&_DelegateRegistry.TransactOpts, to, contract_, tokenId, rights, enable)
}

// Multicall is a paid mutator transaction binding the contract method 0xac9650d8.
//
// Solidity: function multicall(bytes[] data) payable returns(bytes[] results)
func (_DelegateRegistry *DelegateRegistryTransactor) Multicall(opts *bind.TransactOpts, data [][]byte) (*types.Transaction, error) {
	return _DelegateRegistry.contract.Transact(opts, "multicall", data)
}

// Multicall is a paid mutator transaction binding the contract method 0xac9650d8.
//
// Solidity: function multicall(bytes[] data) payable returns(bytes[] results)
func (_DelegateRegistry *DelegateRegistrySession) Multicall(data [][]byte) (*types.Transaction, error) {
	return _DelegateRegistry.Contract.Multicall(&_DelegateRegistry.TransactOpts, data)
}

// Multicall is a paid mutator transaction binding the contract method 0xac9650d8.
//
// Solidity: function multicall(bytes[] data) payable returns(bytes[] results)
func (_DelegateRegistry *DelegateRegistryTransactorSession) Multicall(data [][]byte) (*types.Transaction, error) {
	return _DelegateRegistry.Contract.Multicall(&_DelegateRegistry.TransactOpts, data)
}

// Sweep is a paid mutator transaction binding the contract method 0x35faa416.
//
// Solidity: function sweep() returns()
func (_DelegateRegistry *DelegateRegistryTransactor) Sweep(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _DelegateRegistry.contract.Transact(opts, "sweep")
}

// Sweep is a paid mutator transaction binding the contract method 0x35faa416.
//
// Solidity: function sweep() returns()
func (_DelegateRegistry *DelegateRegistrySession) Sweep() (*types.Transaction, error) {
	return _DelegateRegistry.Contract.Sweep(&_DelegateRegistry.TransactOpts)
}

// Sweep is a paid mutator transaction binding the contract method 0x35faa416.
//
// Solidity: function sweep() returns()
func (_DelegateRegistry *DelegateRegistryTransactorSession) Sweep() (*types.Transaction, error) {
	return _DelegateRegistry.Contract.Sweep(&_DelegateRegistry.TransactOpts)
}

// DelegateRegistryDelegateAllIterator is returned from FilterDelegateAll and is used to iterate over the raw logs and unpacked data for DelegateAll events raised by the DelegateRegistry contract.
type DelegateRegistryDelegateAllIterator struct {
	Event *DelegateRegistryDelegateAll // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *DelegateRegistryDelegateAllIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(DelegateRegistryDelegateAll)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(DelegateRegistryDelegateAll)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *DelegateRegistryDelegateAllIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *DelegateRegistryDelegateAllIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// DelegateRegistryDelegateAll represents a DelegateAll event raised by the DelegateRegistry contract.
type DelegateRegistryDelegateAll struct {
	From   common.Address
	To     common.Address
	Rights [32]byte
	Enable bool
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterDelegateAll is a free log retrieval operation binding the contract event 0xda3ef6410e30373a9137f83f9781a8129962b6882532b7c229de2e39de423227.
//
// Solidity: event DelegateAll(address indexed from, address indexed to, bytes32 rights, bool enable)
func (_DelegateRegistry *DelegateRegistryFilterer) FilterDelegateAll(opts *bind.FilterOpts, from []common.Address, to []common.Address) (*DelegateRegistryDelegateAllIterator, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}

	logs, sub, err := _DelegateRegistry.contract.FilterLogs(opts, "DelegateAll", fromRule, toRule)
	if err != nil {
		return nil, err
	}
	return &DelegateRegistryDelegateAllIterator{contract: _DelegateRegistry.contract, event: "DelegateAll", logs: logs, sub: sub}, nil
}

// WatchDelegateAll is a free log subscription operation binding the contract event 0xda3ef6410e30373a9137f83f9781a8129962b6882532b7c229de2e39de423227.
//
// Solidity: event DelegateAll(address indexed from, address indexed to, bytes32 rights, bool enable)
func (_DelegateRegistry *DelegateRegistryFilterer) WatchDelegateAll(opts *bind.WatchOpts, sink chan<- *DelegateRegistryDelegateAll, from []common.Address, to []common.Address) (event.Subscription, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}

	logs, sub, err := _DelegateRegistry.contract.WatchLogs(opts, "DelegateAll", fromRule, toRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(DelegateRegistryDelegateAll)
				if err := _DelegateRegistry.contract.UnpackLog(event, "DelegateAll", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseDelegateAll is a log parse operation binding the contract event 0xda3ef6410e30373a9137f83f9781a8129962b6882532b7c229de2e39de423227.
//
// Solidity: event DelegateAll(address indexed from, address indexed to, bytes32 rights, bool enable)
func (_DelegateRegistry *DelegateRegistryFilterer) ParseDelegateAll(log types.Log) (*DelegateRegistryDelegateAll, error) {
	event := new(DelegateRegistryDelegateAll)
	if err := _DelegateRegistry.contract.UnpackLog(event, "DelegateAll", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// DelegateRegistryDelegateContractIterator is returned from FilterDelegateContract and is used to iterate over the raw logs and unpacked data for DelegateContract events raised by the DelegateRegistry contract.
type DelegateRegistryDelegateContractIterator struct {
	Event *DelegateRegistryDelegateContract // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *DelegateRegistryDelegateContractIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(DelegateRegistryDelegateContract)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(DelegateRegistryDelegateContract)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *DelegateRegistryDelegateContractIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *DelegateRegistryDelegateContractIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// DelegateRegistryDelegateContract represents a DelegateContract event raised by the DelegateRegistry contract.
type DelegateRegistryDelegateContract struct {
	From     common.Address
	To       common.Address
	Contract common.Address
	Rights   [32]byte
	Enable   bool
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterDelegateContract is a free log retrieval operation binding the contract event 0x021be15e24de4afc43cfb5d0ba95ca38e0783571e05c12bbe6aece8842ae82df.
//
// Solidity: event DelegateContract(address indexed from, address indexed to, address indexed contract_, bytes32 rights, bool enable)
func (_DelegateRegistry *DelegateRegistryFilterer) FilterDelegateContract(opts *bind.FilterOpts, from []common.Address, to []common.Address, contract_ []common.Address) (*DelegateRegistryDelegateContractIterator, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}
	var contract_Rule []interface{}
	for _, contract_Item := range contract_ {
		contract_Rule = append(contract_Rule, contract_Item)
	}

	logs, sub, err := _DelegateRegistry.contract.FilterLogs(opts, "DelegateContract", fromRule, toRule, contract_Rule)
	if err != nil {
		return nil, err
	}
	return &DelegateRegistryDelegateContractIterator{contract: _DelegateRegistry.contract, event: "DelegateContract", logs: logs, sub: sub}, nil
}

// WatchDelegateContract is a free log subscription operation binding the contract event 0x021be15e24de4afc43cfb5d0ba95ca38e0783571e05c12bbe6aece8842ae82df.
//
// Solidity: event DelegateContract(address indexed from, address indexed to, address indexed contract_, bytes32 rights, bool enable)
func (_DelegateRegistry *DelegateRegistryFilterer) WatchDelegateContract(opts *bind.WatchOpts, sink chan<- *DelegateRegistryDelegateContract, from []common.Address, to []common.Address, contract_ []common.Address) (event.Subscription, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}
	var contract_Rule []interface{}
	for _, contract_Item := range contract_ {
		contract_Rule = append(contract_Rule, contract_Item)
	}

	logs, sub, err := _DelegateRegistry.contract.WatchLogs(opts, "DelegateContract", fromRule, toRule, contract_Rule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(DelegateRegistryDelegateContract)
				if err := _DelegateRegistry.contract.UnpackLog(event, "DelegateContract", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseDelegateContract is a log parse operation binding the contract event 0x021be15e24de4afc43cfb5d0ba95ca38e0783571e05c12bbe6aece8842ae82df.
//
// Solidity: event DelegateContract(address indexed from, address indexed to, address indexed contract_, bytes32 rights, bool enable)
func (_DelegateRegistry *DelegateRegistryFilterer) ParseDelegateContract(log types.Log) (*DelegateRegistryDelegateContract, error) {
	event := new(DelegateRegistryDelegateContract)
	if err := _DelegateRegistry.contract.UnpackLog(event, "DelegateContract", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// DelegateRegistryDelegateERC1155Iterator is returned from FilterDelegateERC1155 and is used to iterate over the raw logs and unpacked data for DelegateERC1155 events raised by the DelegateRegistry contract.
type DelegateRegistryDelegateERC1155Iterator struct {
	Event *DelegateRegistryDelegateERC1155 // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *DelegateRegistryDelegateERC1155Iterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(DelegateRegistryDelegateERC1155)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(DelegateRegistryDelegateERC1155)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *DelegateRegistryDelegateERC1155Iterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *DelegateRegistryDelegateERC1155Iterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// DelegateRegistryDelegateERC1155 represents a DelegateERC1155 event raised by the DelegateRegistry contract.
type DelegateRegistryDelegateERC1155 struct {
	From     common.Address
	To       common.Address
	Contract common.Address
	TokenId  *big.Int
	Rights   [32]byte
	Amount   *big.Int
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterDelegateERC1155 is a free log retrieval operation binding the contract event 0x27ab1adc9bca76301ed7a691320766dfa4b4b1aa32c9e05cf789611be7f8c75f.
//
// Solidity: event DelegateERC1155(address indexed from, address indexed to, address indexed contract_, uint256 tokenId, bytes32 rights, uint256 amount)
func (_DelegateRegistry *DelegateRegistryFilterer) FilterDelegateERC1155(opts *bind.FilterOpts, from []common.Address, to []common.Address, contract_ []common.Address) (*DelegateRegistryDelegateERC1155Iterator, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}
	var contract_Rule []interface{}
	for _, contract_Item := range contract_ {
		contract_Rule = append(contract_Rule, contract_Item)
	}

	logs, sub, err := _DelegateRegistry.contract.FilterLogs(opts, "DelegateERC1155", fromRule, toRule, contract_Rule)
	if err != nil {
		return nil, err
	}
	return &DelegateRegistryDelegateERC1155Iterator{contract: _DelegateRegistry.contract, event: "DelegateERC1155", logs: logs, sub: sub}, nil
}

// WatchDelegateERC1155 is a free log subscription operation binding the contract event 0x27ab1adc9bca76301ed7a691320766dfa4b4b1aa32c9e05cf789611be7f8c75f.
//
// Solidity: event DelegateERC1155(address indexed from, address indexed to, address indexed contract_, uint256 tokenId, bytes32 rights, uint256 amount)
func (_DelegateRegistry *DelegateRegistryFilterer) WatchDelegateERC1155(opts *bind.WatchOpts, sink chan<- *DelegateRegistryDelegateERC1155, from []common.Address, to []common.Address, contract_ []common.Address) (event.Subscription, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}
	var contract_Rule []interface{}
	for _, contract_Item := range contract_ {
		contract_Rule = append(contract_Rule, contract_Item)
	}

	logs, sub, err := _DelegateRegistry.contract.WatchLogs(opts, "DelegateERC1155", fromRule, toRule, contract_Rule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(DelegateRegistryDelegateERC1155)
				if err := _DelegateRegistry.contract.UnpackLog(event, "DelegateERC1155", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseDelegateERC1155 is a log parse operation binding the contract event 0x27ab1adc9bca76301ed7a691320766dfa4b4b1aa32c9e05cf789611be7f8c75f.
//
// Solidity: event DelegateERC1155(address indexed from, address indexed to, address indexed contract_, uint256 tokenId, bytes32 rights, uint256 amount)
func (_DelegateRegistry *DelegateRegistryFilterer) ParseDelegateERC1155(log types.Log) (*DelegateRegistryDelegateERC1155, error) {
	event := new(DelegateRegistryDelegateERC1155)
	if err := _DelegateRegistry.contract.UnpackLog(event, "DelegateERC1155", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// DelegateRegistryDelegateERC20Iterator is returned from FilterDelegateERC20 and is used to iterate over the raw logs and unpacked data for DelegateERC20 events raised by the DelegateRegistry contract.
type DelegateRegistryDelegateERC20Iterator struct {
	Event *DelegateRegistryDelegateERC20 // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *DelegateRegistryDelegateERC20Iterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(DelegateRegistryDelegateERC20)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(DelegateRegistryDelegateERC20)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *DelegateRegistryDelegateERC20Iterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *DelegateRegistryDelegateERC20Iterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// DelegateRegistryDelegateERC20 represents a DelegateERC20 event raised by the DelegateRegistry contract.
type DelegateRegistryDelegateERC20 struct {
	From     common.Address
	To       common.Address
	Contract common.Address
	Rights   [32]byte
	Amount   *big.Int
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterDelegateERC20 is a free log retrieval operation binding the contract event 0x6ebd000dfc4dc9df04f723f827bae7694230795e8f22ed4af438e074cc982d18.
//
// Solidity: event DelegateERC20(address indexed from, address indexed to, address indexed contract_, bytes32 rights, uint256 amount)
func (_DelegateRegistry *DelegateRegistryFilterer) FilterDelegateERC20(opts *bind.FilterOpts, from []common.Address, to []common.Address, contract_ []common.Address) (*DelegateRegistryDelegateERC20Iterator, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}
	var contract_Rule []interface{}
	for _, contract_Item := range contract_ {
		contract_Rule = append(contract_Rule, contract_Item)
	}

	logs, sub, err := _DelegateRegistry.contract.FilterLogs(opts, "DelegateERC20", fromRule, toRule, contract_Rule)
	if err != nil {
		return nil, err
	}
	return &DelegateRegistryDelegateERC20Iterator{contract: _DelegateRegistry.contract, event: "DelegateERC20", logs: logs, sub: sub}, nil
}

// WatchDelegateERC20 is a free log subscription operation binding the contract event 0x6ebd000dfc4dc9df04f723f827bae7694230795e8f22ed4af438e074cc982d18.
//
// Solidity: event DelegateERC20(address indexed from, address indexed to, address indexed contract_, bytes32 rights, uint256 amount)
func (_DelegateRegistry *DelegateRegistryFilterer) WatchDelegateERC20(opts *bind.WatchOpts, sink chan<- *DelegateRegistryDelegateERC20, from []common.Address, to []common.Address, contract_ []common.Address) (event.Subscription, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}
	var contract_Rule []interface{}
	for _, contract_Item := range contract_ {
		contract_Rule = append(contract_Rule, contract_Item)
	}

	logs, sub, err := _DelegateRegistry.contract.WatchLogs(opts, "DelegateERC20", fromRule, toRule, contract_Rule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(DelegateRegistryDelegateERC20)
				if err := _DelegateRegistry.contract.UnpackLog(event, "DelegateERC20", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseDelegateERC20 is a log parse operation binding the contract event 0x6ebd000dfc4dc9df04f723f827bae7694230795e8f22ed4af438e074cc982d18.
//
// Solidity: event DelegateERC20(address indexed from, address indexed to, address indexed contract_, bytes32 rights, uint256 amount)
func (_DelegateRegistry *DelegateRegistryFilterer) ParseDelegateERC20(log types.Log) (*DelegateRegistryDelegateERC20, error) {
	event := new(DelegateRegistryDelegateERC20)
	if err := _DelegateRegistry.contract.UnpackLog(event, "DelegateERC20", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// DelegateRegistryDelegateERC721Iterator is returned from FilterDelegateERC721 and is used to iterate over the raw logs and unpacked data for DelegateERC721 events raised by the DelegateRegistry contract.
type DelegateRegistryDelegateERC721Iterator struct {
	Event *DelegateRegistryDelegateERC721 // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *DelegateRegistryDelegateERC721Iterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(DelegateRegistryDelegateERC721)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(DelegateRegistryDelegateERC721)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *DelegateRegistryDelegateERC721Iterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *DelegateRegistryDelegateERC721Iterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// DelegateRegistryDelegateERC721 represents a DelegateERC721 event raised by the DelegateRegistry contract.
type DelegateRegistryDelegateERC721 struct {
	From     common.Address
	To       common.Address
	Contract common.Address
	TokenId  *big.Int
	Rights   [32]byte
	Enable   bool
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterDelegateERC721 is a free log retrieval operation binding the contract event 0x15e7a1bdcd507dd632d797d38e60cc5a9c0749b9a63097a215c4d006126825c6.
//
// Solidity: event DelegateERC721(address indexed from, address indexed to, address indexed contract_, uint256 tokenId, bytes32 rights, bool enable)
func (_DelegateRegistry *DelegateRegistryFilterer) FilterDelegateERC721(opts *bind.FilterOpts, from []common.Address, to []common.Address, contract_ []common.Address) (*DelegateRegistryDelegateERC721Iterator, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}
	var contract_Rule []interface{}
	for _, contract_Item := range contract_ {
		contract_Rule = append(contract_Rule, contract_Item)
	}

	logs, sub, err := _DelegateRegistry.contract.FilterLogs(opts, "DelegateERC721", fromRule, toRule, contract_Rule)
	if err != nil {
		return nil, err
	}
	return &DelegateRegistryDelegateERC721Iterator{contract: _DelegateRegistry.contract, event: "DelegateERC721", logs: logs, sub: sub}, nil
}

// WatchDelegateERC721 is a free log subscription operation binding the contract event 0x15e7a1bdcd507dd632d797d38e60cc5a9c0749b9a63097a215c4d006126825c6.
//
// Solidity: event DelegateERC721(address indexed from, address indexed to, address indexed contract_, uint256 tokenId, bytes32 rights, bool enable)
func (_DelegateRegistry *DelegateRegistryFilterer) WatchDelegateERC721(opts *bind.WatchOpts, sink chan<- *DelegateRegistryDelegateERC721, from []common.Address, to []common.Address, contract_ []common.Address) (event.Subscription, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}
	var contract_Rule []interface{}
	for _, contract_Item := range contract_ {
		contract_Rule = append(contract_Rule, contract_Item)
	}

	logs, sub, err := _DelegateRegistry.contract.WatchLogs(opts, "DelegateERC721", fromRule, toRule, contract_Rule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(DelegateRegistryDelegateERC721)
				if err := _DelegateRegistry.contract.UnpackLog(event, "DelegateERC721", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseDelegateERC721 is a log parse operation binding the contract event 0x15e7a1bdcd507dd632d797d38e60cc5a9c0749b9a63097a215c4d006126825c6.
//
// Solidity: event DelegateERC721(address indexed from, address indexed to, address indexed contract_, uint256 tokenId, bytes32 rights, bool enable)
func (_DelegateRegistry *DelegateRegistryFilterer) ParseDelegateERC721(log types.Log) (*DelegateRegistryDelegateERC721, error) {
	event := new(DelegateRegistryDelegateERC721)
	if err := _DelegateRegistry.contract.UnpackLog(event, "DelegateERC721", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}