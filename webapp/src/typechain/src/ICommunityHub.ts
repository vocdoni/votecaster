/* Autogenerated file. Do not edit manually. */
/* tslint:disable */
/* eslint-disable */
import type {
  BaseContract,
  BigNumberish,
  BytesLike,
  FunctionFragment,
  Result,
  Interface,
  EventFragment,
  AddressLike,
  ContractRunner,
  ContractMethod,
  Listener,
} from "ethers";
import type {
  TypedContractEvent,
  TypedDeferredTopicFilter,
  TypedEventLog,
  TypedLogDescription,
  TypedListener,
  TypedContractMethod,
} from "../common";

export declare namespace ICommunityHub {
  export type TokenStruct = {
    blockchain: string;
    contractAddress: AddressLike;
  };

  export type TokenStructOutput = [
    blockchain: string,
    contractAddress: string
  ] & { blockchain: string; contractAddress: string };

  export type CensusStruct = {
    censusType: BigNumberish;
    tokens: ICommunityHub.TokenStruct[];
    channel: string;
  };

  export type CensusStructOutput = [
    censusType: bigint,
    tokens: ICommunityHub.TokenStructOutput[],
    channel: string
  ] & {
    censusType: bigint;
    tokens: ICommunityHub.TokenStructOutput[];
    channel: string;
  };

  export type CommunityMetadataStruct = {
    name: string;
    imageURI: string;
    groupChatURL: string;
    channels: string[];
    notifications: boolean;
  };

  export type CommunityMetadataStructOutput = [
    name: string,
    imageURI: string,
    groupChatURL: string,
    channels: string[],
    notifications: boolean
  ] & {
    name: string;
    imageURI: string;
    groupChatURL: string;
    channels: string[];
    notifications: boolean;
  };

  export type CommunityStruct = {
    metadata: ICommunityHub.CommunityMetadataStruct;
    census: ICommunityHub.CensusStruct;
    guardians: BigNumberish[];
    createElectionPermission: BigNumberish;
    disabled: boolean;
    funds: BigNumberish;
  };

  export type CommunityStructOutput = [
    metadata: ICommunityHub.CommunityMetadataStructOutput,
    census: ICommunityHub.CensusStructOutput,
    guardians: bigint[],
    createElectionPermission: bigint,
    disabled: boolean,
    funds: bigint
  ] & {
    metadata: ICommunityHub.CommunityMetadataStructOutput;
    census: ICommunityHub.CensusStructOutput;
    guardians: bigint[];
    createElectionPermission: bigint;
    disabled: boolean;
    funds: bigint;
  };
}

export interface ICommunityHubInterface extends Interface {
  getFunction(
    nameOrSignature:
      | "addGuardian"
      | "adminManageCommunity"
      | "adminSetCommunityPrice"
      | "adminSetPricePerElection"
      | "createCommunity"
      | "deposit"
      | "getCommunity"
      | "getCreateCommunityPrice"
      | "getNextCommunityId"
      | "getPricePerElection"
      | "removeGuardian"
      | "setCensus"
      | "setCreateElectionPermission"
      | "setMetadata"
      | "setNotifiableElections"
      | "withdraw"
  ): FunctionFragment;

  getEvent(
    nameOrSignatureOrTopic:
      | "AdminCommunityManaged"
      | "CensusSet"
      | "CommunityCreated"
      | "CommunityDeposit"
      | "CommunityDisabled"
      | "CommunityEnabled"
      | "CreateCommunityPriceSet"
      | "CreateElectionPermissionSet"
      | "Deposit"
      | "GuardianAdded"
      | "GuardianRemoved"
      | "MetadataSet"
      | "NotifiableElectionsSet"
      | "PricePerElectionSet"
      | "Withdrawal"
  ): EventFragment;

  encodeFunctionData(
    functionFragment: "addGuardian",
    values: [BigNumberish, BigNumberish]
  ): string;
  encodeFunctionData(
    functionFragment: "adminManageCommunity",
    values: [
      BigNumberish,
      ICommunityHub.CommunityMetadataStruct,
      ICommunityHub.CensusStruct,
      BigNumberish[],
      BigNumberish,
      boolean
    ]
  ): string;
  encodeFunctionData(
    functionFragment: "adminSetCommunityPrice",
    values: [BigNumberish]
  ): string;
  encodeFunctionData(
    functionFragment: "adminSetPricePerElection",
    values: [BigNumberish]
  ): string;
  encodeFunctionData(
    functionFragment: "createCommunity",
    values: [
      ICommunityHub.CommunityMetadataStruct,
      ICommunityHub.CensusStruct,
      BigNumberish[],
      BigNumberish
    ]
  ): string;
  encodeFunctionData(
    functionFragment: "deposit",
    values: [BigNumberish]
  ): string;
  encodeFunctionData(
    functionFragment: "getCommunity",
    values: [BigNumberish]
  ): string;
  encodeFunctionData(
    functionFragment: "getCreateCommunityPrice",
    values?: undefined
  ): string;
  encodeFunctionData(
    functionFragment: "getNextCommunityId",
    values?: undefined
  ): string;
  encodeFunctionData(
    functionFragment: "getPricePerElection",
    values?: undefined
  ): string;
  encodeFunctionData(
    functionFragment: "removeGuardian",
    values: [BigNumberish, BigNumberish]
  ): string;
  encodeFunctionData(
    functionFragment: "setCensus",
    values: [BigNumberish, ICommunityHub.CensusStruct]
  ): string;
  encodeFunctionData(
    functionFragment: "setCreateElectionPermission",
    values: [BigNumberish, BigNumberish]
  ): string;
  encodeFunctionData(
    functionFragment: "setMetadata",
    values: [BigNumberish, ICommunityHub.CommunityMetadataStruct]
  ): string;
  encodeFunctionData(
    functionFragment: "setNotifiableElections",
    values: [BigNumberish, boolean]
  ): string;
  encodeFunctionData(functionFragment: "withdraw", values?: undefined): string;

  decodeFunctionResult(
    functionFragment: "addGuardian",
    data: BytesLike
  ): Result;
  decodeFunctionResult(
    functionFragment: "adminManageCommunity",
    data: BytesLike
  ): Result;
  decodeFunctionResult(
    functionFragment: "adminSetCommunityPrice",
    data: BytesLike
  ): Result;
  decodeFunctionResult(
    functionFragment: "adminSetPricePerElection",
    data: BytesLike
  ): Result;
  decodeFunctionResult(
    functionFragment: "createCommunity",
    data: BytesLike
  ): Result;
  decodeFunctionResult(functionFragment: "deposit", data: BytesLike): Result;
  decodeFunctionResult(
    functionFragment: "getCommunity",
    data: BytesLike
  ): Result;
  decodeFunctionResult(
    functionFragment: "getCreateCommunityPrice",
    data: BytesLike
  ): Result;
  decodeFunctionResult(
    functionFragment: "getNextCommunityId",
    data: BytesLike
  ): Result;
  decodeFunctionResult(
    functionFragment: "getPricePerElection",
    data: BytesLike
  ): Result;
  decodeFunctionResult(
    functionFragment: "removeGuardian",
    data: BytesLike
  ): Result;
  decodeFunctionResult(functionFragment: "setCensus", data: BytesLike): Result;
  decodeFunctionResult(
    functionFragment: "setCreateElectionPermission",
    data: BytesLike
  ): Result;
  decodeFunctionResult(
    functionFragment: "setMetadata",
    data: BytesLike
  ): Result;
  decodeFunctionResult(
    functionFragment: "setNotifiableElections",
    data: BytesLike
  ): Result;
  decodeFunctionResult(functionFragment: "withdraw", data: BytesLike): Result;
}

export namespace AdminCommunityManagedEvent {
  export type InputTuple = [communityId: BigNumberish];
  export type OutputTuple = [communityId: bigint];
  export interface OutputObject {
    communityId: bigint;
  }
  export type Event = TypedContractEvent<InputTuple, OutputTuple, OutputObject>;
  export type Filter = TypedDeferredTopicFilter<Event>;
  export type Log = TypedEventLog<Event>;
  export type LogDescription = TypedLogDescription<Event>;
}

export namespace CensusSetEvent {
  export type InputTuple = [
    communityId: BigNumberish,
    census: ICommunityHub.CensusStruct
  ];
  export type OutputTuple = [
    communityId: bigint,
    census: ICommunityHub.CensusStructOutput
  ];
  export interface OutputObject {
    communityId: bigint;
    census: ICommunityHub.CensusStructOutput;
  }
  export type Event = TypedContractEvent<InputTuple, OutputTuple, OutputObject>;
  export type Filter = TypedDeferredTopicFilter<Event>;
  export type Log = TypedEventLog<Event>;
  export type LogDescription = TypedLogDescription<Event>;
}

export namespace CommunityCreatedEvent {
  export type InputTuple = [communityId: BigNumberish, creator: AddressLike];
  export type OutputTuple = [communityId: bigint, creator: string];
  export interface OutputObject {
    communityId: bigint;
    creator: string;
  }
  export type Event = TypedContractEvent<InputTuple, OutputTuple, OutputObject>;
  export type Filter = TypedDeferredTopicFilter<Event>;
  export type Log = TypedEventLog<Event>;
  export type LogDescription = TypedLogDescription<Event>;
}

export namespace CommunityDepositEvent {
  export type InputTuple = [
    sender: AddressLike,
    amount: BigNumberish,
    communityId: BigNumberish
  ];
  export type OutputTuple = [
    sender: string,
    amount: bigint,
    communityId: bigint
  ];
  export interface OutputObject {
    sender: string;
    amount: bigint;
    communityId: bigint;
  }
  export type Event = TypedContractEvent<InputTuple, OutputTuple, OutputObject>;
  export type Filter = TypedDeferredTopicFilter<Event>;
  export type Log = TypedEventLog<Event>;
  export type LogDescription = TypedLogDescription<Event>;
}

export namespace CommunityDisabledEvent {
  export type InputTuple = [communityId: BigNumberish];
  export type OutputTuple = [communityId: bigint];
  export interface OutputObject {
    communityId: bigint;
  }
  export type Event = TypedContractEvent<InputTuple, OutputTuple, OutputObject>;
  export type Filter = TypedDeferredTopicFilter<Event>;
  export type Log = TypedEventLog<Event>;
  export type LogDescription = TypedLogDescription<Event>;
}

export namespace CommunityEnabledEvent {
  export type InputTuple = [communityId: BigNumberish];
  export type OutputTuple = [communityId: bigint];
  export interface OutputObject {
    communityId: bigint;
  }
  export type Event = TypedContractEvent<InputTuple, OutputTuple, OutputObject>;
  export type Filter = TypedDeferredTopicFilter<Event>;
  export type Log = TypedEventLog<Event>;
  export type LogDescription = TypedLogDescription<Event>;
}

export namespace CreateCommunityPriceSetEvent {
  export type InputTuple = [price: BigNumberish];
  export type OutputTuple = [price: bigint];
  export interface OutputObject {
    price: bigint;
  }
  export type Event = TypedContractEvent<InputTuple, OutputTuple, OutputObject>;
  export type Filter = TypedDeferredTopicFilter<Event>;
  export type Log = TypedEventLog<Event>;
  export type LogDescription = TypedLogDescription<Event>;
}

export namespace CreateElectionPermissionSetEvent {
  export type InputTuple = [
    communityId: BigNumberish,
    createElectionPermission: BigNumberish
  ];
  export type OutputTuple = [
    communityId: bigint,
    createElectionPermission: bigint
  ];
  export interface OutputObject {
    communityId: bigint;
    createElectionPermission: bigint;
  }
  export type Event = TypedContractEvent<InputTuple, OutputTuple, OutputObject>;
  export type Filter = TypedDeferredTopicFilter<Event>;
  export type Log = TypedEventLog<Event>;
  export type LogDescription = TypedLogDescription<Event>;
}

export namespace DepositEvent {
  export type InputTuple = [sender: AddressLike, amount: BigNumberish];
  export type OutputTuple = [sender: string, amount: bigint];
  export interface OutputObject {
    sender: string;
    amount: bigint;
  }
  export type Event = TypedContractEvent<InputTuple, OutputTuple, OutputObject>;
  export type Filter = TypedDeferredTopicFilter<Event>;
  export type Log = TypedEventLog<Event>;
  export type LogDescription = TypedLogDescription<Event>;
}

export namespace GuardianAddedEvent {
  export type InputTuple = [communityId: BigNumberish, guardian: BigNumberish];
  export type OutputTuple = [communityId: bigint, guardian: bigint];
  export interface OutputObject {
    communityId: bigint;
    guardian: bigint;
  }
  export type Event = TypedContractEvent<InputTuple, OutputTuple, OutputObject>;
  export type Filter = TypedDeferredTopicFilter<Event>;
  export type Log = TypedEventLog<Event>;
  export type LogDescription = TypedLogDescription<Event>;
}

export namespace GuardianRemovedEvent {
  export type InputTuple = [communityId: BigNumberish, guardian: BigNumberish];
  export type OutputTuple = [communityId: bigint, guardian: bigint];
  export interface OutputObject {
    communityId: bigint;
    guardian: bigint;
  }
  export type Event = TypedContractEvent<InputTuple, OutputTuple, OutputObject>;
  export type Filter = TypedDeferredTopicFilter<Event>;
  export type Log = TypedEventLog<Event>;
  export type LogDescription = TypedLogDescription<Event>;
}

export namespace MetadataSetEvent {
  export type InputTuple = [
    communityId: BigNumberish,
    metadata: ICommunityHub.CommunityMetadataStruct
  ];
  export type OutputTuple = [
    communityId: bigint,
    metadata: ICommunityHub.CommunityMetadataStructOutput
  ];
  export interface OutputObject {
    communityId: bigint;
    metadata: ICommunityHub.CommunityMetadataStructOutput;
  }
  export type Event = TypedContractEvent<InputTuple, OutputTuple, OutputObject>;
  export type Filter = TypedDeferredTopicFilter<Event>;
  export type Log = TypedEventLog<Event>;
  export type LogDescription = TypedLogDescription<Event>;
}

export namespace NotifiableElectionsSetEvent {
  export type InputTuple = [
    communityId: BigNumberish,
    notifiableElections: boolean
  ];
  export type OutputTuple = [communityId: bigint, notifiableElections: boolean];
  export interface OutputObject {
    communityId: bigint;
    notifiableElections: boolean;
  }
  export type Event = TypedContractEvent<InputTuple, OutputTuple, OutputObject>;
  export type Filter = TypedDeferredTopicFilter<Event>;
  export type Log = TypedEventLog<Event>;
  export type LogDescription = TypedLogDescription<Event>;
}

export namespace PricePerElectionSetEvent {
  export type InputTuple = [price: BigNumberish];
  export type OutputTuple = [price: bigint];
  export interface OutputObject {
    price: bigint;
  }
  export type Event = TypedContractEvent<InputTuple, OutputTuple, OutputObject>;
  export type Filter = TypedDeferredTopicFilter<Event>;
  export type Log = TypedEventLog<Event>;
  export type LogDescription = TypedLogDescription<Event>;
}

export namespace WithdrawalEvent {
  export type InputTuple = [amount: BigNumberish, to: AddressLike];
  export type OutputTuple = [amount: bigint, to: string];
  export interface OutputObject {
    amount: bigint;
    to: string;
  }
  export type Event = TypedContractEvent<InputTuple, OutputTuple, OutputObject>;
  export type Filter = TypedDeferredTopicFilter<Event>;
  export type Log = TypedEventLog<Event>;
  export type LogDescription = TypedLogDescription<Event>;
}

export interface ICommunityHub extends BaseContract {
  connect(runner?: ContractRunner | null): ICommunityHub;
  waitForDeployment(): Promise<this>;

  interface: ICommunityHubInterface;

  queryFilter<TCEvent extends TypedContractEvent>(
    event: TCEvent,
    fromBlockOrBlockhash?: string | number | undefined,
    toBlock?: string | number | undefined
  ): Promise<Array<TypedEventLog<TCEvent>>>;
  queryFilter<TCEvent extends TypedContractEvent>(
    filter: TypedDeferredTopicFilter<TCEvent>,
    fromBlockOrBlockhash?: string | number | undefined,
    toBlock?: string | number | undefined
  ): Promise<Array<TypedEventLog<TCEvent>>>;

  on<TCEvent extends TypedContractEvent>(
    event: TCEvent,
    listener: TypedListener<TCEvent>
  ): Promise<this>;
  on<TCEvent extends TypedContractEvent>(
    filter: TypedDeferredTopicFilter<TCEvent>,
    listener: TypedListener<TCEvent>
  ): Promise<this>;

  once<TCEvent extends TypedContractEvent>(
    event: TCEvent,
    listener: TypedListener<TCEvent>
  ): Promise<this>;
  once<TCEvent extends TypedContractEvent>(
    filter: TypedDeferredTopicFilter<TCEvent>,
    listener: TypedListener<TCEvent>
  ): Promise<this>;

  listeners<TCEvent extends TypedContractEvent>(
    event: TCEvent
  ): Promise<Array<TypedListener<TCEvent>>>;
  listeners(eventName?: string): Promise<Array<Listener>>;
  removeAllListeners<TCEvent extends TypedContractEvent>(
    event?: TCEvent
  ): Promise<this>;

  addGuardian: TypedContractMethod<
    [_communityId: BigNumberish, _guardian: BigNumberish],
    [void],
    "nonpayable"
  >;

  adminManageCommunity: TypedContractMethod<
    [
      _communityId: BigNumberish,
      _metadata: ICommunityHub.CommunityMetadataStruct,
      _census: ICommunityHub.CensusStruct,
      _guardians: BigNumberish[],
      _createElectionPermission: BigNumberish,
      _disabled: boolean
    ],
    [void],
    "nonpayable"
  >;

  adminSetCommunityPrice: TypedContractMethod<
    [_price: BigNumberish],
    [void],
    "nonpayable"
  >;

  adminSetPricePerElection: TypedContractMethod<
    [_price: BigNumberish],
    [void],
    "nonpayable"
  >;

  createCommunity: TypedContractMethod<
    [
      _metadata: ICommunityHub.CommunityMetadataStruct,
      _census: ICommunityHub.CensusStruct,
      _guardians: BigNumberish[],
      _createElectionPermission: BigNumberish
    ],
    [void],
    "payable"
  >;

  deposit: TypedContractMethod<[_communityId: BigNumberish], [void], "payable">;

  getCommunity: TypedContractMethod<
    [_communityId: BigNumberish],
    [ICommunityHub.CommunityStructOutput],
    "view"
  >;

  getCreateCommunityPrice: TypedContractMethod<[], [bigint], "view">;

  getNextCommunityId: TypedContractMethod<[], [bigint], "view">;

  getPricePerElection: TypedContractMethod<[], [bigint], "view">;

  removeGuardian: TypedContractMethod<
    [_communityId: BigNumberish, _guardian: BigNumberish],
    [void],
    "nonpayable"
  >;

  setCensus: TypedContractMethod<
    [_communityId: BigNumberish, _census: ICommunityHub.CensusStruct],
    [void],
    "nonpayable"
  >;

  setCreateElectionPermission: TypedContractMethod<
    [_communityId: BigNumberish, _createElectionPermission: BigNumberish],
    [void],
    "nonpayable"
  >;

  setMetadata: TypedContractMethod<
    [
      _communityId: BigNumberish,
      _metadata: ICommunityHub.CommunityMetadataStruct
    ],
    [void],
    "nonpayable"
  >;

  setNotifiableElections: TypedContractMethod<
    [_communityId: BigNumberish, _notifiableElections: boolean],
    [void],
    "nonpayable"
  >;

  withdraw: TypedContractMethod<[], [void], "nonpayable">;

  getFunction<T extends ContractMethod = ContractMethod>(
    key: string | FunctionFragment
  ): T;

  getFunction(
    nameOrSignature: "addGuardian"
  ): TypedContractMethod<
    [_communityId: BigNumberish, _guardian: BigNumberish],
    [void],
    "nonpayable"
  >;
  getFunction(
    nameOrSignature: "adminManageCommunity"
  ): TypedContractMethod<
    [
      _communityId: BigNumberish,
      _metadata: ICommunityHub.CommunityMetadataStruct,
      _census: ICommunityHub.CensusStruct,
      _guardians: BigNumberish[],
      _createElectionPermission: BigNumberish,
      _disabled: boolean
    ],
    [void],
    "nonpayable"
  >;
  getFunction(
    nameOrSignature: "adminSetCommunityPrice"
  ): TypedContractMethod<[_price: BigNumberish], [void], "nonpayable">;
  getFunction(
    nameOrSignature: "adminSetPricePerElection"
  ): TypedContractMethod<[_price: BigNumberish], [void], "nonpayable">;
  getFunction(
    nameOrSignature: "createCommunity"
  ): TypedContractMethod<
    [
      _metadata: ICommunityHub.CommunityMetadataStruct,
      _census: ICommunityHub.CensusStruct,
      _guardians: BigNumberish[],
      _createElectionPermission: BigNumberish
    ],
    [void],
    "payable"
  >;
  getFunction(
    nameOrSignature: "deposit"
  ): TypedContractMethod<[_communityId: BigNumberish], [void], "payable">;
  getFunction(
    nameOrSignature: "getCommunity"
  ): TypedContractMethod<
    [_communityId: BigNumberish],
    [ICommunityHub.CommunityStructOutput],
    "view"
  >;
  getFunction(
    nameOrSignature: "getCreateCommunityPrice"
  ): TypedContractMethod<[], [bigint], "view">;
  getFunction(
    nameOrSignature: "getNextCommunityId"
  ): TypedContractMethod<[], [bigint], "view">;
  getFunction(
    nameOrSignature: "getPricePerElection"
  ): TypedContractMethod<[], [bigint], "view">;
  getFunction(
    nameOrSignature: "removeGuardian"
  ): TypedContractMethod<
    [_communityId: BigNumberish, _guardian: BigNumberish],
    [void],
    "nonpayable"
  >;
  getFunction(
    nameOrSignature: "setCensus"
  ): TypedContractMethod<
    [_communityId: BigNumberish, _census: ICommunityHub.CensusStruct],
    [void],
    "nonpayable"
  >;
  getFunction(
    nameOrSignature: "setCreateElectionPermission"
  ): TypedContractMethod<
    [_communityId: BigNumberish, _createElectionPermission: BigNumberish],
    [void],
    "nonpayable"
  >;
  getFunction(
    nameOrSignature: "setMetadata"
  ): TypedContractMethod<
    [
      _communityId: BigNumberish,
      _metadata: ICommunityHub.CommunityMetadataStruct
    ],
    [void],
    "nonpayable"
  >;
  getFunction(
    nameOrSignature: "setNotifiableElections"
  ): TypedContractMethod<
    [_communityId: BigNumberish, _notifiableElections: boolean],
    [void],
    "nonpayable"
  >;
  getFunction(
    nameOrSignature: "withdraw"
  ): TypedContractMethod<[], [void], "nonpayable">;

  getEvent(
    key: "AdminCommunityManaged"
  ): TypedContractEvent<
    AdminCommunityManagedEvent.InputTuple,
    AdminCommunityManagedEvent.OutputTuple,
    AdminCommunityManagedEvent.OutputObject
  >;
  getEvent(
    key: "CensusSet"
  ): TypedContractEvent<
    CensusSetEvent.InputTuple,
    CensusSetEvent.OutputTuple,
    CensusSetEvent.OutputObject
  >;
  getEvent(
    key: "CommunityCreated"
  ): TypedContractEvent<
    CommunityCreatedEvent.InputTuple,
    CommunityCreatedEvent.OutputTuple,
    CommunityCreatedEvent.OutputObject
  >;
  getEvent(
    key: "CommunityDeposit"
  ): TypedContractEvent<
    CommunityDepositEvent.InputTuple,
    CommunityDepositEvent.OutputTuple,
    CommunityDepositEvent.OutputObject
  >;
  getEvent(
    key: "CommunityDisabled"
  ): TypedContractEvent<
    CommunityDisabledEvent.InputTuple,
    CommunityDisabledEvent.OutputTuple,
    CommunityDisabledEvent.OutputObject
  >;
  getEvent(
    key: "CommunityEnabled"
  ): TypedContractEvent<
    CommunityEnabledEvent.InputTuple,
    CommunityEnabledEvent.OutputTuple,
    CommunityEnabledEvent.OutputObject
  >;
  getEvent(
    key: "CreateCommunityPriceSet"
  ): TypedContractEvent<
    CreateCommunityPriceSetEvent.InputTuple,
    CreateCommunityPriceSetEvent.OutputTuple,
    CreateCommunityPriceSetEvent.OutputObject
  >;
  getEvent(
    key: "CreateElectionPermissionSet"
  ): TypedContractEvent<
    CreateElectionPermissionSetEvent.InputTuple,
    CreateElectionPermissionSetEvent.OutputTuple,
    CreateElectionPermissionSetEvent.OutputObject
  >;
  getEvent(
    key: "Deposit"
  ): TypedContractEvent<
    DepositEvent.InputTuple,
    DepositEvent.OutputTuple,
    DepositEvent.OutputObject
  >;
  getEvent(
    key: "GuardianAdded"
  ): TypedContractEvent<
    GuardianAddedEvent.InputTuple,
    GuardianAddedEvent.OutputTuple,
    GuardianAddedEvent.OutputObject
  >;
  getEvent(
    key: "GuardianRemoved"
  ): TypedContractEvent<
    GuardianRemovedEvent.InputTuple,
    GuardianRemovedEvent.OutputTuple,
    GuardianRemovedEvent.OutputObject
  >;
  getEvent(
    key: "MetadataSet"
  ): TypedContractEvent<
    MetadataSetEvent.InputTuple,
    MetadataSetEvent.OutputTuple,
    MetadataSetEvent.OutputObject
  >;
  getEvent(
    key: "NotifiableElectionsSet"
  ): TypedContractEvent<
    NotifiableElectionsSetEvent.InputTuple,
    NotifiableElectionsSetEvent.OutputTuple,
    NotifiableElectionsSetEvent.OutputObject
  >;
  getEvent(
    key: "PricePerElectionSet"
  ): TypedContractEvent<
    PricePerElectionSetEvent.InputTuple,
    PricePerElectionSetEvent.OutputTuple,
    PricePerElectionSetEvent.OutputObject
  >;
  getEvent(
    key: "Withdrawal"
  ): TypedContractEvent<
    WithdrawalEvent.InputTuple,
    WithdrawalEvent.OutputTuple,
    WithdrawalEvent.OutputObject
  >;

  filters: {
    "AdminCommunityManaged(uint256)": TypedContractEvent<
      AdminCommunityManagedEvent.InputTuple,
      AdminCommunityManagedEvent.OutputTuple,
      AdminCommunityManagedEvent.OutputObject
    >;
    AdminCommunityManaged: TypedContractEvent<
      AdminCommunityManagedEvent.InputTuple,
      AdminCommunityManagedEvent.OutputTuple,
      AdminCommunityManagedEvent.OutputObject
    >;

    "CensusSet(uint256,tuple)": TypedContractEvent<
      CensusSetEvent.InputTuple,
      CensusSetEvent.OutputTuple,
      CensusSetEvent.OutputObject
    >;
    CensusSet: TypedContractEvent<
      CensusSetEvent.InputTuple,
      CensusSetEvent.OutputTuple,
      CensusSetEvent.OutputObject
    >;

    "CommunityCreated(uint256,address)": TypedContractEvent<
      CommunityCreatedEvent.InputTuple,
      CommunityCreatedEvent.OutputTuple,
      CommunityCreatedEvent.OutputObject
    >;
    CommunityCreated: TypedContractEvent<
      CommunityCreatedEvent.InputTuple,
      CommunityCreatedEvent.OutputTuple,
      CommunityCreatedEvent.OutputObject
    >;

    "CommunityDeposit(address,uint256,uint256)": TypedContractEvent<
      CommunityDepositEvent.InputTuple,
      CommunityDepositEvent.OutputTuple,
      CommunityDepositEvent.OutputObject
    >;
    CommunityDeposit: TypedContractEvent<
      CommunityDepositEvent.InputTuple,
      CommunityDepositEvent.OutputTuple,
      CommunityDepositEvent.OutputObject
    >;

    "CommunityDisabled(uint256)": TypedContractEvent<
      CommunityDisabledEvent.InputTuple,
      CommunityDisabledEvent.OutputTuple,
      CommunityDisabledEvent.OutputObject
    >;
    CommunityDisabled: TypedContractEvent<
      CommunityDisabledEvent.InputTuple,
      CommunityDisabledEvent.OutputTuple,
      CommunityDisabledEvent.OutputObject
    >;

    "CommunityEnabled(uint256)": TypedContractEvent<
      CommunityEnabledEvent.InputTuple,
      CommunityEnabledEvent.OutputTuple,
      CommunityEnabledEvent.OutputObject
    >;
    CommunityEnabled: TypedContractEvent<
      CommunityEnabledEvent.InputTuple,
      CommunityEnabledEvent.OutputTuple,
      CommunityEnabledEvent.OutputObject
    >;

    "CreateCommunityPriceSet(uint256)": TypedContractEvent<
      CreateCommunityPriceSetEvent.InputTuple,
      CreateCommunityPriceSetEvent.OutputTuple,
      CreateCommunityPriceSetEvent.OutputObject
    >;
    CreateCommunityPriceSet: TypedContractEvent<
      CreateCommunityPriceSetEvent.InputTuple,
      CreateCommunityPriceSetEvent.OutputTuple,
      CreateCommunityPriceSetEvent.OutputObject
    >;

    "CreateElectionPermissionSet(uint256,uint8)": TypedContractEvent<
      CreateElectionPermissionSetEvent.InputTuple,
      CreateElectionPermissionSetEvent.OutputTuple,
      CreateElectionPermissionSetEvent.OutputObject
    >;
    CreateElectionPermissionSet: TypedContractEvent<
      CreateElectionPermissionSetEvent.InputTuple,
      CreateElectionPermissionSetEvent.OutputTuple,
      CreateElectionPermissionSetEvent.OutputObject
    >;

    "Deposit(address,uint256)": TypedContractEvent<
      DepositEvent.InputTuple,
      DepositEvent.OutputTuple,
      DepositEvent.OutputObject
    >;
    Deposit: TypedContractEvent<
      DepositEvent.InputTuple,
      DepositEvent.OutputTuple,
      DepositEvent.OutputObject
    >;

    "GuardianAdded(uint256,uint256)": TypedContractEvent<
      GuardianAddedEvent.InputTuple,
      GuardianAddedEvent.OutputTuple,
      GuardianAddedEvent.OutputObject
    >;
    GuardianAdded: TypedContractEvent<
      GuardianAddedEvent.InputTuple,
      GuardianAddedEvent.OutputTuple,
      GuardianAddedEvent.OutputObject
    >;

    "GuardianRemoved(uint256,uint256)": TypedContractEvent<
      GuardianRemovedEvent.InputTuple,
      GuardianRemovedEvent.OutputTuple,
      GuardianRemovedEvent.OutputObject
    >;
    GuardianRemoved: TypedContractEvent<
      GuardianRemovedEvent.InputTuple,
      GuardianRemovedEvent.OutputTuple,
      GuardianRemovedEvent.OutputObject
    >;

    "MetadataSet(uint256,tuple)": TypedContractEvent<
      MetadataSetEvent.InputTuple,
      MetadataSetEvent.OutputTuple,
      MetadataSetEvent.OutputObject
    >;
    MetadataSet: TypedContractEvent<
      MetadataSetEvent.InputTuple,
      MetadataSetEvent.OutputTuple,
      MetadataSetEvent.OutputObject
    >;

    "NotifiableElectionsSet(uint256,bool)": TypedContractEvent<
      NotifiableElectionsSetEvent.InputTuple,
      NotifiableElectionsSetEvent.OutputTuple,
      NotifiableElectionsSetEvent.OutputObject
    >;
    NotifiableElectionsSet: TypedContractEvent<
      NotifiableElectionsSetEvent.InputTuple,
      NotifiableElectionsSetEvent.OutputTuple,
      NotifiableElectionsSetEvent.OutputObject
    >;

    "PricePerElectionSet(uint256)": TypedContractEvent<
      PricePerElectionSetEvent.InputTuple,
      PricePerElectionSetEvent.OutputTuple,
      PricePerElectionSetEvent.OutputObject
    >;
    PricePerElectionSet: TypedContractEvent<
      PricePerElectionSetEvent.InputTuple,
      PricePerElectionSetEvent.OutputTuple,
      PricePerElectionSetEvent.OutputObject
    >;

    "Withdrawal(uint256,address)": TypedContractEvent<
      WithdrawalEvent.InputTuple,
      WithdrawalEvent.OutputTuple,
      WithdrawalEvent.OutputObject
    >;
    Withdrawal: TypedContractEvent<
      WithdrawalEvent.InputTuple,
      WithdrawalEvent.OutputTuple,
      WithdrawalEvent.OutputObject
    >;
  };
}