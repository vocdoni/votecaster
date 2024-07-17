//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// CommunityHub
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

export const communityHubAbi = [
  { type: 'constructor', inputs: [], stateMutability: 'nonpayable' },
  { type: 'receive', stateMutability: 'payable' },
  {
    type: 'function',
    inputs: [
      { name: '_communityId', internalType: 'uint256', type: 'uint256' },
      { name: '_guardian', internalType: 'uint256', type: 'uint256' },
    ],
    name: 'addGuardian',
    outputs: [],
    stateMutability: 'nonpayable',
  },
  {
    type: 'function',
    inputs: [
      { name: '_communityId', internalType: 'uint256', type: 'uint256' },
      {
        name: '_metadata',
        internalType: 'struct ICommunityHub.CommunityMetadata',
        type: 'tuple',
        components: [
          { name: 'name', internalType: 'string', type: 'string' },
          { name: 'imageURI', internalType: 'string', type: 'string' },
          { name: 'groupChatURL', internalType: 'string', type: 'string' },
          { name: 'channels', internalType: 'string[]', type: 'string[]' },
          { name: 'notifications', internalType: 'bool', type: 'bool' },
        ],
      },
      {
        name: '_census',
        internalType: 'struct ICommunityHub.Census',
        type: 'tuple',
        components: [
          {
            name: 'censusType',
            internalType: 'enum ICommunityHub.CensusType',
            type: 'uint8',
          },
          {
            name: 'tokens',
            internalType: 'struct ICommunityHub.Token[]',
            type: 'tuple[]',
            components: [
              { name: 'blockchain', internalType: 'string', type: 'string' },
              {
                name: 'contractAddress',
                internalType: 'address',
                type: 'address',
              },
            ],
          },
          { name: 'channel', internalType: 'string', type: 'string' },
        ],
      },
      { name: '_guardians', internalType: 'uint256[]', type: 'uint256[]' },
      {
        name: '_createElectionPermission',
        internalType: 'enum ICommunityHub.CreateElectionPermission',
        type: 'uint8',
      },
      { name: '_disabled', internalType: 'bool', type: 'bool' },
    ],
    name: 'adminManageCommunity',
    outputs: [],
    stateMutability: 'nonpayable',
  },
  {
    type: 'function',
    inputs: [{ name: '_price', internalType: 'uint256', type: 'uint256' }],
    name: 'adminSetCommunityPrice',
    outputs: [],
    stateMutability: 'nonpayable',
  },
  {
    type: 'function',
    inputs: [{ name: '_price', internalType: 'uint256', type: 'uint256' }],
    name: 'adminSetPricePerElection',
    outputs: [],
    stateMutability: 'nonpayable',
  },
  {
    type: 'function',
    inputs: [
      {
        name: '_metadata',
        internalType: 'struct ICommunityHub.CommunityMetadata',
        type: 'tuple',
        components: [
          { name: 'name', internalType: 'string', type: 'string' },
          { name: 'imageURI', internalType: 'string', type: 'string' },
          { name: 'groupChatURL', internalType: 'string', type: 'string' },
          { name: 'channels', internalType: 'string[]', type: 'string[]' },
          { name: 'notifications', internalType: 'bool', type: 'bool' },
        ],
      },
      {
        name: '_census',
        internalType: 'struct ICommunityHub.Census',
        type: 'tuple',
        components: [
          {
            name: 'censusType',
            internalType: 'enum ICommunityHub.CensusType',
            type: 'uint8',
          },
          {
            name: 'tokens',
            internalType: 'struct ICommunityHub.Token[]',
            type: 'tuple[]',
            components: [
              { name: 'blockchain', internalType: 'string', type: 'string' },
              {
                name: 'contractAddress',
                internalType: 'address',
                type: 'address',
              },
            ],
          },
          { name: 'channel', internalType: 'string', type: 'string' },
        ],
      },
      { name: '_guardians', internalType: 'uint256[]', type: 'uint256[]' },
      {
        name: '_createElectionPermission',
        internalType: 'enum ICommunityHub.CreateElectionPermission',
        type: 'uint8',
      },
    ],
    name: 'createCommunity',
    outputs: [],
    stateMutability: 'payable',
  },
  {
    type: 'function',
    inputs: [{ name: '_communityId', internalType: 'uint256', type: 'uint256' }],
    name: 'deposit',
    outputs: [],
    stateMutability: 'payable',
  },
  {
    type: 'function',
    inputs: [{ name: '_communityId', internalType: 'uint256', type: 'uint256' }],
    name: 'getCommunity',
    outputs: [
      {
        name: '',
        internalType: 'struct ICommunityHub.Community',
        type: 'tuple',
        components: [
          {
            name: 'metadata',
            internalType: 'struct ICommunityHub.CommunityMetadata',
            type: 'tuple',
            components: [
              { name: 'name', internalType: 'string', type: 'string' },
              { name: 'imageURI', internalType: 'string', type: 'string' },
              { name: 'groupChatURL', internalType: 'string', type: 'string' },
              { name: 'channels', internalType: 'string[]', type: 'string[]' },
              { name: 'notifications', internalType: 'bool', type: 'bool' },
            ],
          },
          {
            name: 'census',
            internalType: 'struct ICommunityHub.Census',
            type: 'tuple',
            components: [
              {
                name: 'censusType',
                internalType: 'enum ICommunityHub.CensusType',
                type: 'uint8',
              },
              {
                name: 'tokens',
                internalType: 'struct ICommunityHub.Token[]',
                type: 'tuple[]',
                components: [
                  {
                    name: 'blockchain',
                    internalType: 'string',
                    type: 'string',
                  },
                  {
                    name: 'contractAddress',
                    internalType: 'address',
                    type: 'address',
                  },
                ],
              },
              { name: 'channel', internalType: 'string', type: 'string' },
            ],
          },
          { name: 'guardians', internalType: 'uint256[]', type: 'uint256[]' },
          {
            name: 'createElectionPermission',
            internalType: 'enum ICommunityHub.CreateElectionPermission',
            type: 'uint8',
          },
          { name: 'disabled', internalType: 'bool', type: 'bool' },
          { name: 'funds', internalType: 'uint256', type: 'uint256' },
        ],
      },
    ],
    stateMutability: 'view',
  },
  {
    type: 'function',
    inputs: [],
    name: 'getCreateCommunityPrice',
    outputs: [{ name: '', internalType: 'uint256', type: 'uint256' }],
    stateMutability: 'view',
  },
  {
    type: 'function',
    inputs: [],
    name: 'getNextCommunityId',
    outputs: [{ name: '', internalType: 'uint256', type: 'uint256' }],
    stateMutability: 'view',
  },
  {
    type: 'function',
    inputs: [],
    name: 'getPricePerElection',
    outputs: [{ name: '', internalType: 'uint256', type: 'uint256' }],
    stateMutability: 'view',
  },
  {
    type: 'function',
    inputs: [
      { name: '_communityId', internalType: 'uint256', type: 'uint256' },
      { name: '_electionId', internalType: 'bytes32', type: 'bytes32' },
    ],
    name: 'getResult',
    outputs: [
      {
        name: 'result',
        internalType: 'struct IResult.Result',
        type: 'tuple',
        components: [
          { name: 'question', internalType: 'string', type: 'string' },
          { name: 'options', internalType: 'string[]', type: 'string[]' },
          { name: 'date', internalType: 'string', type: 'string' },
          { name: 'tally', internalType: 'uint256[][]', type: 'uint256[][]' },
          { name: 'turnout', internalType: 'uint256', type: 'uint256' },
          {
            name: 'totalVotingPower',
            internalType: 'uint256',
            type: 'uint256',
          },
          {
            name: 'participants',
            internalType: 'uint256[]',
            type: 'uint256[]',
          },
          { name: 'censusRoot', internalType: 'bytes32', type: 'bytes32' },
          { name: 'censusURI', internalType: 'string', type: 'string' },
        ],
      },
    ],
    stateMutability: 'view',
  },
  {
    type: 'function',
    inputs: [],
    name: 'owner',
    outputs: [{ name: '', internalType: 'address', type: 'address' }],
    stateMutability: 'view',
  },
  {
    type: 'function',
    inputs: [
      { name: '_communityId', internalType: 'uint256', type: 'uint256' },
      { name: '_guardian', internalType: 'uint256', type: 'uint256' },
    ],
    name: 'removeGuardian',
    outputs: [],
    stateMutability: 'nonpayable',
  },
  {
    type: 'function',
    inputs: [],
    name: 'renounceOwnership',
    outputs: [],
    stateMutability: 'nonpayable',
  },
  {
    type: 'function',
    inputs: [
      { name: '_communityId', internalType: 'uint256', type: 'uint256' },
      {
        name: '_census',
        internalType: 'struct ICommunityHub.Census',
        type: 'tuple',
        components: [
          {
            name: 'censusType',
            internalType: 'enum ICommunityHub.CensusType',
            type: 'uint8',
          },
          {
            name: 'tokens',
            internalType: 'struct ICommunityHub.Token[]',
            type: 'tuple[]',
            components: [
              { name: 'blockchain', internalType: 'string', type: 'string' },
              {
                name: 'contractAddress',
                internalType: 'address',
                type: 'address',
              },
            ],
          },
          { name: 'channel', internalType: 'string', type: 'string' },
        ],
      },
    ],
    name: 'setCensus',
    outputs: [],
    stateMutability: 'nonpayable',
  },
  {
    type: 'function',
    inputs: [
      { name: '_communityId', internalType: 'uint256', type: 'uint256' },
      {
        name: '_createElectionPermission',
        internalType: 'enum ICommunityHub.CreateElectionPermission',
        type: 'uint8',
      },
    ],
    name: 'setCreateElectionPermission',
    outputs: [],
    stateMutability: 'nonpayable',
  },
  {
    type: 'function',
    inputs: [
      { name: '_communityId', internalType: 'uint256', type: 'uint256' },
      {
        name: '_metadata',
        internalType: 'struct ICommunityHub.CommunityMetadata',
        type: 'tuple',
        components: [
          { name: 'name', internalType: 'string', type: 'string' },
          { name: 'imageURI', internalType: 'string', type: 'string' },
          { name: 'groupChatURL', internalType: 'string', type: 'string' },
          { name: 'channels', internalType: 'string[]', type: 'string[]' },
          { name: 'notifications', internalType: 'bool', type: 'bool' },
        ],
      },
    ],
    name: 'setMetadata',
    outputs: [],
    stateMutability: 'nonpayable',
  },
  {
    type: 'function',
    inputs: [
      { name: '_communityId', internalType: 'uint256', type: 'uint256' },
      { name: '_notifiableElections', internalType: 'bool', type: 'bool' },
    ],
    name: 'setNotifiableElections',
    outputs: [],
    stateMutability: 'nonpayable',
  },
  {
    type: 'function',
    inputs: [
      { name: '_communityId', internalType: 'uint256', type: 'uint256' },
      { name: '_electionId', internalType: 'bytes32', type: 'bytes32' },
      {
        name: '_result',
        internalType: 'struct IResult.Result',
        type: 'tuple',
        components: [
          { name: 'question', internalType: 'string', type: 'string' },
          { name: 'options', internalType: 'string[]', type: 'string[]' },
          { name: 'date', internalType: 'string', type: 'string' },
          { name: 'tally', internalType: 'uint256[][]', type: 'uint256[][]' },
          { name: 'turnout', internalType: 'uint256', type: 'uint256' },
          {
            name: 'totalVotingPower',
            internalType: 'uint256',
            type: 'uint256',
          },
          {
            name: 'participants',
            internalType: 'uint256[]',
            type: 'uint256[]',
          },
          { name: 'censusRoot', internalType: 'bytes32', type: 'bytes32' },
          { name: 'censusURI', internalType: 'string', type: 'string' },
        ],
      },
    ],
    name: 'setResult',
    outputs: [],
    stateMutability: 'nonpayable',
  },
  {
    type: 'function',
    inputs: [{ name: 'newOwner', internalType: 'address', type: 'address' }],
    name: 'transferOwnership',
    outputs: [],
    stateMutability: 'nonpayable',
  },
  {
    type: 'function',
    inputs: [],
    name: 'withdraw',
    outputs: [],
    stateMutability: 'nonpayable',
  },
  {
    type: 'event',
    anonymous: false,
    inputs: [
      {
        name: 'communityId',
        internalType: 'uint256',
        type: 'uint256',
        indexed: false,
      },
    ],
    name: 'AdminCommunityManaged',
  },
  {
    type: 'event',
    anonymous: false,
    inputs: [
      {
        name: 'communityId',
        internalType: 'uint256',
        type: 'uint256',
        indexed: false,
      },
      {
        name: 'census',
        internalType: 'struct ICommunityHub.Census',
        type: 'tuple',
        components: [
          {
            name: 'censusType',
            internalType: 'enum ICommunityHub.CensusType',
            type: 'uint8',
          },
          {
            name: 'tokens',
            internalType: 'struct ICommunityHub.Token[]',
            type: 'tuple[]',
            components: [
              { name: 'blockchain', internalType: 'string', type: 'string' },
              {
                name: 'contractAddress',
                internalType: 'address',
                type: 'address',
              },
            ],
          },
          { name: 'channel', internalType: 'string', type: 'string' },
        ],
        indexed: false,
      },
    ],
    name: 'CensusSet',
  },
  {
    type: 'event',
    anonymous: false,
    inputs: [
      {
        name: 'communityId',
        internalType: 'uint256',
        type: 'uint256',
        indexed: false,
      },
      {
        name: 'creator',
        internalType: 'address',
        type: 'address',
        indexed: false,
      },
    ],
    name: 'CommunityCreated',
  },
  {
    type: 'event',
    anonymous: false,
    inputs: [
      {
        name: 'sender',
        internalType: 'address',
        type: 'address',
        indexed: false,
      },
      {
        name: 'amount',
        internalType: 'uint256',
        type: 'uint256',
        indexed: false,
      },
      {
        name: 'communityId',
        internalType: 'uint256',
        type: 'uint256',
        indexed: false,
      },
    ],
    name: 'CommunityDeposit',
  },
  {
    type: 'event',
    anonymous: false,
    inputs: [
      {
        name: 'communityId',
        internalType: 'uint256',
        type: 'uint256',
        indexed: false,
      },
    ],
    name: 'CommunityDisabled',
  },
  {
    type: 'event',
    anonymous: false,
    inputs: [
      {
        name: 'communityId',
        internalType: 'uint256',
        type: 'uint256',
        indexed: false,
      },
    ],
    name: 'CommunityEnabled',
  },
  {
    type: 'event',
    anonymous: false,
    inputs: [
      {
        name: 'price',
        internalType: 'uint256',
        type: 'uint256',
        indexed: false,
      },
    ],
    name: 'CreateCommunityPriceSet',
  },
  {
    type: 'event',
    anonymous: false,
    inputs: [
      {
        name: 'communityId',
        internalType: 'uint256',
        type: 'uint256',
        indexed: false,
      },
      {
        name: 'createElectionPermission',
        internalType: 'enum ICommunityHub.CreateElectionPermission',
        type: 'uint8',
        indexed: false,
      },
    ],
    name: 'CreateElectionPermissionSet',
  },
  {
    type: 'event',
    anonymous: false,
    inputs: [
      {
        name: 'sender',
        internalType: 'address',
        type: 'address',
        indexed: false,
      },
      {
        name: 'amount',
        internalType: 'uint256',
        type: 'uint256',
        indexed: false,
      },
    ],
    name: 'Deposit',
  },
  {
    type: 'event',
    anonymous: false,
    inputs: [
      {
        name: 'communityId',
        internalType: 'uint256',
        type: 'uint256',
        indexed: false,
      },
      {
        name: 'guardian',
        internalType: 'uint256',
        type: 'uint256',
        indexed: false,
      },
    ],
    name: 'GuardianAdded',
  },
  {
    type: 'event',
    anonymous: false,
    inputs: [
      {
        name: 'communityId',
        internalType: 'uint256',
        type: 'uint256',
        indexed: false,
      },
      {
        name: 'guardian',
        internalType: 'uint256',
        type: 'uint256',
        indexed: false,
      },
    ],
    name: 'GuardianRemoved',
  },
  {
    type: 'event',
    anonymous: false,
    inputs: [
      {
        name: 'communityId',
        internalType: 'uint256',
        type: 'uint256',
        indexed: false,
      },
      {
        name: 'metadata',
        internalType: 'struct ICommunityHub.CommunityMetadata',
        type: 'tuple',
        components: [
          { name: 'name', internalType: 'string', type: 'string' },
          { name: 'imageURI', internalType: 'string', type: 'string' },
          { name: 'groupChatURL', internalType: 'string', type: 'string' },
          { name: 'channels', internalType: 'string[]', type: 'string[]' },
          { name: 'notifications', internalType: 'bool', type: 'bool' },
        ],
        indexed: false,
      },
    ],
    name: 'MetadataSet',
  },
  {
    type: 'event',
    anonymous: false,
    inputs: [
      {
        name: 'communityId',
        internalType: 'uint256',
        type: 'uint256',
        indexed: false,
      },
      {
        name: 'notifiableElections',
        internalType: 'bool',
        type: 'bool',
        indexed: false,
      },
    ],
    name: 'NotifiableElectionsSet',
  },
  {
    type: 'event',
    anonymous: false,
    inputs: [
      {
        name: 'previousOwner',
        internalType: 'address',
        type: 'address',
        indexed: true,
      },
      {
        name: 'newOwner',
        internalType: 'address',
        type: 'address',
        indexed: true,
      },
    ],
    name: 'OwnershipTransferred',
  },
  {
    type: 'event',
    anonymous: false,
    inputs: [
      {
        name: 'price',
        internalType: 'uint256',
        type: 'uint256',
        indexed: false,
      },
    ],
    name: 'PricePerElectionSet',
  },
  {
    type: 'event',
    anonymous: false,
    inputs: [
      {
        name: 'communityId',
        internalType: 'uint256',
        type: 'uint256',
        indexed: false,
      },
      {
        name: 'electionId',
        internalType: 'bytes32',
        type: 'bytes32',
        indexed: false,
      },
    ],
    name: 'ResultsSet',
  },
  {
    type: 'event',
    anonymous: false,
    inputs: [
      {
        name: 'amount',
        internalType: 'uint256',
        type: 'uint256',
        indexed: false,
      },
      { name: 'to', internalType: 'address', type: 'address', indexed: false },
    ],
    name: 'Withdrawal',
  },
  {
    type: 'error',
    inputs: [
      { name: 'expected', internalType: 'uint256', type: 'uint256' },
      { name: 'actual', internalType: 'uint256', type: 'uint256' },
    ],
    name: 'AmountMismatch',
  },
  {
    type: 'error',
    inputs: [{ name: 'guardian', internalType: 'uint256', type: 'uint256' }],
    name: 'GuardianNotFound',
  },
  { type: 'error', inputs: [], name: 'InvalidCreateElectionPermission' },
  {
    type: 'error',
    inputs: [{ name: 'owner', internalType: 'address', type: 'address' }],
    name: 'OwnableInvalidOwner',
  },
  {
    type: 'error',
    inputs: [{ name: 'account', internalType: 'address', type: 'address' }],
    name: 'OwnableUnauthorizedAccount',
  },
  { type: 'error', inputs: [], name: 'ZeroAmount' },
] as const
