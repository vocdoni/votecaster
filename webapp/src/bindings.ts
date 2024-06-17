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
    inputs: [
      { name: '_communityId', internalType: 'uint256', type: 'uint256' },
    ],
    name: 'deposit',
    outputs: [],
    stateMutability: 'payable',
  },
  {
    type: 'function',
    inputs: [
      { name: '_communityId', internalType: 'uint256', type: 'uint256' },
    ],
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

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// ICommunityHub
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

export const iCommunityHubAbi = [
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
    inputs: [
      { name: '_communityId', internalType: 'uint256', type: 'uint256' },
    ],
    name: 'deposit',
    outputs: [],
    stateMutability: 'payable',
  },
  {
    type: 'function',
    inputs: [
      { name: '_communityId', internalType: 'uint256', type: 'uint256' },
    ],
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
      { name: '_guardian', internalType: 'uint256', type: 'uint256' },
    ],
    name: 'removeGuardian',
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
  { type: 'error', inputs: [], name: 'ZeroAmount' },
] as const

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// IERC165
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

export const ierc165Abi = [
  {
    type: 'function',
    inputs: [{ name: 'interfaceID', internalType: 'bytes4', type: 'bytes4' }],
    name: 'supportsInterface',
    outputs: [{ name: '', internalType: 'bool', type: 'bool' }],
    stateMutability: 'view',
  },
] as const

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// IERC20
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

export const ierc20Abi = [
  {
    type: 'function',
    inputs: [
      { name: 'owner', internalType: 'address', type: 'address' },
      { name: 'spender', internalType: 'address', type: 'address' },
    ],
    name: 'allowance',
    outputs: [{ name: '', internalType: 'uint256', type: 'uint256' }],
    stateMutability: 'view',
  },
  {
    type: 'function',
    inputs: [
      { name: 'spender', internalType: 'address', type: 'address' },
      { name: 'amount', internalType: 'uint256', type: 'uint256' },
    ],
    name: 'approve',
    outputs: [{ name: '', internalType: 'bool', type: 'bool' }],
    stateMutability: 'nonpayable',
  },
  {
    type: 'function',
    inputs: [{ name: 'account', internalType: 'address', type: 'address' }],
    name: 'balanceOf',
    outputs: [{ name: '', internalType: 'uint256', type: 'uint256' }],
    stateMutability: 'view',
  },
  {
    type: 'function',
    inputs: [],
    name: 'decimals',
    outputs: [{ name: '', internalType: 'uint8', type: 'uint8' }],
    stateMutability: 'view',
  },
  {
    type: 'function',
    inputs: [],
    name: 'name',
    outputs: [{ name: '', internalType: 'string', type: 'string' }],
    stateMutability: 'view',
  },
  {
    type: 'function',
    inputs: [],
    name: 'symbol',
    outputs: [{ name: '', internalType: 'string', type: 'string' }],
    stateMutability: 'view',
  },
  {
    type: 'function',
    inputs: [],
    name: 'totalSupply',
    outputs: [{ name: '', internalType: 'uint256', type: 'uint256' }],
    stateMutability: 'view',
  },
  {
    type: 'function',
    inputs: [
      { name: 'to', internalType: 'address', type: 'address' },
      { name: 'amount', internalType: 'uint256', type: 'uint256' },
    ],
    name: 'transfer',
    outputs: [{ name: '', internalType: 'bool', type: 'bool' }],
    stateMutability: 'nonpayable',
  },
  {
    type: 'function',
    inputs: [
      { name: 'from', internalType: 'address', type: 'address' },
      { name: 'to', internalType: 'address', type: 'address' },
      { name: 'amount', internalType: 'uint256', type: 'uint256' },
    ],
    name: 'transferFrom',
    outputs: [{ name: '', internalType: 'bool', type: 'bool' }],
    stateMutability: 'nonpayable',
  },
  {
    type: 'event',
    anonymous: false,
    inputs: [
      {
        name: 'owner',
        internalType: 'address',
        type: 'address',
        indexed: true,
      },
      {
        name: 'spender',
        internalType: 'address',
        type: 'address',
        indexed: true,
      },
      {
        name: 'value',
        internalType: 'uint256',
        type: 'uint256',
        indexed: false,
      },
    ],
    name: 'Approval',
  },
  {
    type: 'event',
    anonymous: false,
    inputs: [
      { name: 'from', internalType: 'address', type: 'address', indexed: true },
      { name: 'to', internalType: 'address', type: 'address', indexed: true },
      {
        name: 'value',
        internalType: 'uint256',
        type: 'uint256',
        indexed: false,
      },
    ],
    name: 'Transfer',
  },
] as const

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// IERC721
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

export const ierc721Abi = [
  {
    type: 'function',
    inputs: [
      { name: '_approved', internalType: 'address', type: 'address' },
      { name: '_tokenId', internalType: 'uint256', type: 'uint256' },
    ],
    name: 'approve',
    outputs: [],
    stateMutability: 'payable',
  },
  {
    type: 'function',
    inputs: [{ name: '_owner', internalType: 'address', type: 'address' }],
    name: 'balanceOf',
    outputs: [{ name: '', internalType: 'uint256', type: 'uint256' }],
    stateMutability: 'view',
  },
  {
    type: 'function',
    inputs: [{ name: '_tokenId', internalType: 'uint256', type: 'uint256' }],
    name: 'getApproved',
    outputs: [{ name: '', internalType: 'address', type: 'address' }],
    stateMutability: 'view',
  },
  {
    type: 'function',
    inputs: [
      { name: '_owner', internalType: 'address', type: 'address' },
      { name: '_operator', internalType: 'address', type: 'address' },
    ],
    name: 'isApprovedForAll',
    outputs: [{ name: '', internalType: 'bool', type: 'bool' }],
    stateMutability: 'view',
  },
  {
    type: 'function',
    inputs: [{ name: '_tokenId', internalType: 'uint256', type: 'uint256' }],
    name: 'ownerOf',
    outputs: [{ name: '', internalType: 'address', type: 'address' }],
    stateMutability: 'view',
  },
  {
    type: 'function',
    inputs: [
      { name: '_from', internalType: 'address', type: 'address' },
      { name: '_to', internalType: 'address', type: 'address' },
      { name: '_tokenId', internalType: 'uint256', type: 'uint256' },
    ],
    name: 'safeTransferFrom',
    outputs: [],
    stateMutability: 'payable',
  },
  {
    type: 'function',
    inputs: [
      { name: '_from', internalType: 'address', type: 'address' },
      { name: '_to', internalType: 'address', type: 'address' },
      { name: '_tokenId', internalType: 'uint256', type: 'uint256' },
      { name: 'data', internalType: 'bytes', type: 'bytes' },
    ],
    name: 'safeTransferFrom',
    outputs: [],
    stateMutability: 'payable',
  },
  {
    type: 'function',
    inputs: [
      { name: '_operator', internalType: 'address', type: 'address' },
      { name: '_approved', internalType: 'bool', type: 'bool' },
    ],
    name: 'setApprovalForAll',
    outputs: [],
    stateMutability: 'nonpayable',
  },
  {
    type: 'function',
    inputs: [{ name: 'interfaceID', internalType: 'bytes4', type: 'bytes4' }],
    name: 'supportsInterface',
    outputs: [{ name: '', internalType: 'bool', type: 'bool' }],
    stateMutability: 'view',
  },
  {
    type: 'function',
    inputs: [
      { name: '_from', internalType: 'address', type: 'address' },
      { name: '_to', internalType: 'address', type: 'address' },
      { name: '_tokenId', internalType: 'uint256', type: 'uint256' },
    ],
    name: 'transferFrom',
    outputs: [],
    stateMutability: 'payable',
  },
  {
    type: 'event',
    anonymous: false,
    inputs: [
      {
        name: '_owner',
        internalType: 'address',
        type: 'address',
        indexed: true,
      },
      {
        name: '_approved',
        internalType: 'address',
        type: 'address',
        indexed: true,
      },
      {
        name: '_tokenId',
        internalType: 'uint256',
        type: 'uint256',
        indexed: true,
      },
    ],
    name: 'Approval',
  },
  {
    type: 'event',
    anonymous: false,
    inputs: [
      {
        name: '_owner',
        internalType: 'address',
        type: 'address',
        indexed: true,
      },
      {
        name: '_operator',
        internalType: 'address',
        type: 'address',
        indexed: true,
      },
      { name: '_approved', internalType: 'bool', type: 'bool', indexed: false },
    ],
    name: 'ApprovalForAll',
  },
  {
    type: 'event',
    anonymous: false,
    inputs: [
      {
        name: '_from',
        internalType: 'address',
        type: 'address',
        indexed: true,
      },
      { name: '_to', internalType: 'address', type: 'address', indexed: true },
      {
        name: '_tokenId',
        internalType: 'uint256',
        type: 'uint256',
        indexed: true,
      },
    ],
    name: 'Transfer',
  },
] as const

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// IERC721Enumerable
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

export const ierc721EnumerableAbi = [
  {
    type: 'function',
    inputs: [
      { name: '_approved', internalType: 'address', type: 'address' },
      { name: '_tokenId', internalType: 'uint256', type: 'uint256' },
    ],
    name: 'approve',
    outputs: [],
    stateMutability: 'payable',
  },
  {
    type: 'function',
    inputs: [{ name: '_owner', internalType: 'address', type: 'address' }],
    name: 'balanceOf',
    outputs: [{ name: '', internalType: 'uint256', type: 'uint256' }],
    stateMutability: 'view',
  },
  {
    type: 'function',
    inputs: [{ name: '_tokenId', internalType: 'uint256', type: 'uint256' }],
    name: 'getApproved',
    outputs: [{ name: '', internalType: 'address', type: 'address' }],
    stateMutability: 'view',
  },
  {
    type: 'function',
    inputs: [
      { name: '_owner', internalType: 'address', type: 'address' },
      { name: '_operator', internalType: 'address', type: 'address' },
    ],
    name: 'isApprovedForAll',
    outputs: [{ name: '', internalType: 'bool', type: 'bool' }],
    stateMutability: 'view',
  },
  {
    type: 'function',
    inputs: [{ name: '_tokenId', internalType: 'uint256', type: 'uint256' }],
    name: 'ownerOf',
    outputs: [{ name: '', internalType: 'address', type: 'address' }],
    stateMutability: 'view',
  },
  {
    type: 'function',
    inputs: [
      { name: '_from', internalType: 'address', type: 'address' },
      { name: '_to', internalType: 'address', type: 'address' },
      { name: '_tokenId', internalType: 'uint256', type: 'uint256' },
    ],
    name: 'safeTransferFrom',
    outputs: [],
    stateMutability: 'payable',
  },
  {
    type: 'function',
    inputs: [
      { name: '_from', internalType: 'address', type: 'address' },
      { name: '_to', internalType: 'address', type: 'address' },
      { name: '_tokenId', internalType: 'uint256', type: 'uint256' },
      { name: 'data', internalType: 'bytes', type: 'bytes' },
    ],
    name: 'safeTransferFrom',
    outputs: [],
    stateMutability: 'payable',
  },
  {
    type: 'function',
    inputs: [
      { name: '_operator', internalType: 'address', type: 'address' },
      { name: '_approved', internalType: 'bool', type: 'bool' },
    ],
    name: 'setApprovalForAll',
    outputs: [],
    stateMutability: 'nonpayable',
  },
  {
    type: 'function',
    inputs: [{ name: 'interfaceID', internalType: 'bytes4', type: 'bytes4' }],
    name: 'supportsInterface',
    outputs: [{ name: '', internalType: 'bool', type: 'bool' }],
    stateMutability: 'view',
  },
  {
    type: 'function',
    inputs: [{ name: '_index', internalType: 'uint256', type: 'uint256' }],
    name: 'tokenByIndex',
    outputs: [{ name: '', internalType: 'uint256', type: 'uint256' }],
    stateMutability: 'view',
  },
  {
    type: 'function',
    inputs: [
      { name: '_owner', internalType: 'address', type: 'address' },
      { name: '_index', internalType: 'uint256', type: 'uint256' },
    ],
    name: 'tokenOfOwnerByIndex',
    outputs: [{ name: '', internalType: 'uint256', type: 'uint256' }],
    stateMutability: 'view',
  },
  {
    type: 'function',
    inputs: [],
    name: 'totalSupply',
    outputs: [{ name: '', internalType: 'uint256', type: 'uint256' }],
    stateMutability: 'view',
  },
  {
    type: 'function',
    inputs: [
      { name: '_from', internalType: 'address', type: 'address' },
      { name: '_to', internalType: 'address', type: 'address' },
      { name: '_tokenId', internalType: 'uint256', type: 'uint256' },
    ],
    name: 'transferFrom',
    outputs: [],
    stateMutability: 'payable',
  },
  {
    type: 'event',
    anonymous: false,
    inputs: [
      {
        name: '_owner',
        internalType: 'address',
        type: 'address',
        indexed: true,
      },
      {
        name: '_approved',
        internalType: 'address',
        type: 'address',
        indexed: true,
      },
      {
        name: '_tokenId',
        internalType: 'uint256',
        type: 'uint256',
        indexed: true,
      },
    ],
    name: 'Approval',
  },
  {
    type: 'event',
    anonymous: false,
    inputs: [
      {
        name: '_owner',
        internalType: 'address',
        type: 'address',
        indexed: true,
      },
      {
        name: '_operator',
        internalType: 'address',
        type: 'address',
        indexed: true,
      },
      { name: '_approved', internalType: 'bool', type: 'bool', indexed: false },
    ],
    name: 'ApprovalForAll',
  },
  {
    type: 'event',
    anonymous: false,
    inputs: [
      {
        name: '_from',
        internalType: 'address',
        type: 'address',
        indexed: true,
      },
      { name: '_to', internalType: 'address', type: 'address', indexed: true },
      {
        name: '_tokenId',
        internalType: 'uint256',
        type: 'uint256',
        indexed: true,
      },
    ],
    name: 'Transfer',
  },
] as const

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// IERC721Metadata
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

export const ierc721MetadataAbi = [
  {
    type: 'function',
    inputs: [
      { name: '_approved', internalType: 'address', type: 'address' },
      { name: '_tokenId', internalType: 'uint256', type: 'uint256' },
    ],
    name: 'approve',
    outputs: [],
    stateMutability: 'payable',
  },
  {
    type: 'function',
    inputs: [{ name: '_owner', internalType: 'address', type: 'address' }],
    name: 'balanceOf',
    outputs: [{ name: '', internalType: 'uint256', type: 'uint256' }],
    stateMutability: 'view',
  },
  {
    type: 'function',
    inputs: [{ name: '_tokenId', internalType: 'uint256', type: 'uint256' }],
    name: 'getApproved',
    outputs: [{ name: '', internalType: 'address', type: 'address' }],
    stateMutability: 'view',
  },
  {
    type: 'function',
    inputs: [
      { name: '_owner', internalType: 'address', type: 'address' },
      { name: '_operator', internalType: 'address', type: 'address' },
    ],
    name: 'isApprovedForAll',
    outputs: [{ name: '', internalType: 'bool', type: 'bool' }],
    stateMutability: 'view',
  },
  {
    type: 'function',
    inputs: [],
    name: 'name',
    outputs: [{ name: '_name', internalType: 'string', type: 'string' }],
    stateMutability: 'view',
  },
  {
    type: 'function',
    inputs: [{ name: '_tokenId', internalType: 'uint256', type: 'uint256' }],
    name: 'ownerOf',
    outputs: [{ name: '', internalType: 'address', type: 'address' }],
    stateMutability: 'view',
  },
  {
    type: 'function',
    inputs: [
      { name: '_from', internalType: 'address', type: 'address' },
      { name: '_to', internalType: 'address', type: 'address' },
      { name: '_tokenId', internalType: 'uint256', type: 'uint256' },
    ],
    name: 'safeTransferFrom',
    outputs: [],
    stateMutability: 'payable',
  },
  {
    type: 'function',
    inputs: [
      { name: '_from', internalType: 'address', type: 'address' },
      { name: '_to', internalType: 'address', type: 'address' },
      { name: '_tokenId', internalType: 'uint256', type: 'uint256' },
      { name: 'data', internalType: 'bytes', type: 'bytes' },
    ],
    name: 'safeTransferFrom',
    outputs: [],
    stateMutability: 'payable',
  },
  {
    type: 'function',
    inputs: [
      { name: '_operator', internalType: 'address', type: 'address' },
      { name: '_approved', internalType: 'bool', type: 'bool' },
    ],
    name: 'setApprovalForAll',
    outputs: [],
    stateMutability: 'nonpayable',
  },
  {
    type: 'function',
    inputs: [{ name: 'interfaceID', internalType: 'bytes4', type: 'bytes4' }],
    name: 'supportsInterface',
    outputs: [{ name: '', internalType: 'bool', type: 'bool' }],
    stateMutability: 'view',
  },
  {
    type: 'function',
    inputs: [],
    name: 'symbol',
    outputs: [{ name: '_symbol', internalType: 'string', type: 'string' }],
    stateMutability: 'view',
  },
  {
    type: 'function',
    inputs: [{ name: '_tokenId', internalType: 'uint256', type: 'uint256' }],
    name: 'tokenURI',
    outputs: [{ name: '', internalType: 'string', type: 'string' }],
    stateMutability: 'view',
  },
  {
    type: 'function',
    inputs: [
      { name: '_from', internalType: 'address', type: 'address' },
      { name: '_to', internalType: 'address', type: 'address' },
      { name: '_tokenId', internalType: 'uint256', type: 'uint256' },
    ],
    name: 'transferFrom',
    outputs: [],
    stateMutability: 'payable',
  },
  {
    type: 'event',
    anonymous: false,
    inputs: [
      {
        name: '_owner',
        internalType: 'address',
        type: 'address',
        indexed: true,
      },
      {
        name: '_approved',
        internalType: 'address',
        type: 'address',
        indexed: true,
      },
      {
        name: '_tokenId',
        internalType: 'uint256',
        type: 'uint256',
        indexed: true,
      },
    ],
    name: 'Approval',
  },
  {
    type: 'event',
    anonymous: false,
    inputs: [
      {
        name: '_owner',
        internalType: 'address',
        type: 'address',
        indexed: true,
      },
      {
        name: '_operator',
        internalType: 'address',
        type: 'address',
        indexed: true,
      },
      { name: '_approved', internalType: 'bool', type: 'bool', indexed: false },
    ],
    name: 'ApprovalForAll',
  },
  {
    type: 'event',
    anonymous: false,
    inputs: [
      {
        name: '_from',
        internalType: 'address',
        type: 'address',
        indexed: true,
      },
      { name: '_to', internalType: 'address', type: 'address', indexed: true },
      {
        name: '_tokenId',
        internalType: 'uint256',
        type: 'uint256',
        indexed: true,
      },
    ],
    name: 'Transfer',
  },
] as const

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// IERC721TokenReceiver
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

export const ierc721TokenReceiverAbi = [
  {
    type: 'function',
    inputs: [
      { name: '_operator', internalType: 'address', type: 'address' },
      { name: '_from', internalType: 'address', type: 'address' },
      { name: '_tokenId', internalType: 'uint256', type: 'uint256' },
      { name: '_data', internalType: 'bytes', type: 'bytes' },
    ],
    name: 'onERC721Received',
    outputs: [{ name: '', internalType: 'bytes4', type: 'bytes4' }],
    stateMutability: 'nonpayable',
  },
] as const

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// IElectionResults
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

export const iElectionResultsAbi = [
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
] as const

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// IMulticall3
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

export const iMulticall3Abi = [
  {
    type: 'function',
    inputs: [
      {
        name: 'calls',
        internalType: 'struct IMulticall3.Call[]',
        type: 'tuple[]',
        components: [
          { name: 'target', internalType: 'address', type: 'address' },
          { name: 'callData', internalType: 'bytes', type: 'bytes' },
        ],
      },
    ],
    name: 'aggregate',
    outputs: [
      { name: 'blockNumber', internalType: 'uint256', type: 'uint256' },
      { name: 'returnData', internalType: 'bytes[]', type: 'bytes[]' },
    ],
    stateMutability: 'payable',
  },
  {
    type: 'function',
    inputs: [
      {
        name: 'calls',
        internalType: 'struct IMulticall3.Call3[]',
        type: 'tuple[]',
        components: [
          { name: 'target', internalType: 'address', type: 'address' },
          { name: 'allowFailure', internalType: 'bool', type: 'bool' },
          { name: 'callData', internalType: 'bytes', type: 'bytes' },
        ],
      },
    ],
    name: 'aggregate3',
    outputs: [
      {
        name: 'returnData',
        internalType: 'struct IMulticall3.Result[]',
        type: 'tuple[]',
        components: [
          { name: 'success', internalType: 'bool', type: 'bool' },
          { name: 'returnData', internalType: 'bytes', type: 'bytes' },
        ],
      },
    ],
    stateMutability: 'payable',
  },
  {
    type: 'function',
    inputs: [
      {
        name: 'calls',
        internalType: 'struct IMulticall3.Call3Value[]',
        type: 'tuple[]',
        components: [
          { name: 'target', internalType: 'address', type: 'address' },
          { name: 'allowFailure', internalType: 'bool', type: 'bool' },
          { name: 'value', internalType: 'uint256', type: 'uint256' },
          { name: 'callData', internalType: 'bytes', type: 'bytes' },
        ],
      },
    ],
    name: 'aggregate3Value',
    outputs: [
      {
        name: 'returnData',
        internalType: 'struct IMulticall3.Result[]',
        type: 'tuple[]',
        components: [
          { name: 'success', internalType: 'bool', type: 'bool' },
          { name: 'returnData', internalType: 'bytes', type: 'bytes' },
        ],
      },
    ],
    stateMutability: 'payable',
  },
  {
    type: 'function',
    inputs: [
      {
        name: 'calls',
        internalType: 'struct IMulticall3.Call[]',
        type: 'tuple[]',
        components: [
          { name: 'target', internalType: 'address', type: 'address' },
          { name: 'callData', internalType: 'bytes', type: 'bytes' },
        ],
      },
    ],
    name: 'blockAndAggregate',
    outputs: [
      { name: 'blockNumber', internalType: 'uint256', type: 'uint256' },
      { name: 'blockHash', internalType: 'bytes32', type: 'bytes32' },
      {
        name: 'returnData',
        internalType: 'struct IMulticall3.Result[]',
        type: 'tuple[]',
        components: [
          { name: 'success', internalType: 'bool', type: 'bool' },
          { name: 'returnData', internalType: 'bytes', type: 'bytes' },
        ],
      },
    ],
    stateMutability: 'payable',
  },
  {
    type: 'function',
    inputs: [],
    name: 'getBasefee',
    outputs: [{ name: 'basefee', internalType: 'uint256', type: 'uint256' }],
    stateMutability: 'view',
  },
  {
    type: 'function',
    inputs: [{ name: 'blockNumber', internalType: 'uint256', type: 'uint256' }],
    name: 'getBlockHash',
    outputs: [{ name: 'blockHash', internalType: 'bytes32', type: 'bytes32' }],
    stateMutability: 'view',
  },
  {
    type: 'function',
    inputs: [],
    name: 'getBlockNumber',
    outputs: [
      { name: 'blockNumber', internalType: 'uint256', type: 'uint256' },
    ],
    stateMutability: 'view',
  },
  {
    type: 'function',
    inputs: [],
    name: 'getChainId',
    outputs: [{ name: 'chainid', internalType: 'uint256', type: 'uint256' }],
    stateMutability: 'view',
  },
  {
    type: 'function',
    inputs: [],
    name: 'getCurrentBlockCoinbase',
    outputs: [{ name: 'coinbase', internalType: 'address', type: 'address' }],
    stateMutability: 'view',
  },
  {
    type: 'function',
    inputs: [],
    name: 'getCurrentBlockDifficulty',
    outputs: [{ name: 'difficulty', internalType: 'uint256', type: 'uint256' }],
    stateMutability: 'view',
  },
  {
    type: 'function',
    inputs: [],
    name: 'getCurrentBlockGasLimit',
    outputs: [{ name: 'gaslimit', internalType: 'uint256', type: 'uint256' }],
    stateMutability: 'view',
  },
  {
    type: 'function',
    inputs: [],
    name: 'getCurrentBlockTimestamp',
    outputs: [{ name: 'timestamp', internalType: 'uint256', type: 'uint256' }],
    stateMutability: 'view',
  },
  {
    type: 'function',
    inputs: [{ name: 'addr', internalType: 'address', type: 'address' }],
    name: 'getEthBalance',
    outputs: [{ name: 'balance', internalType: 'uint256', type: 'uint256' }],
    stateMutability: 'view',
  },
  {
    type: 'function',
    inputs: [],
    name: 'getLastBlockHash',
    outputs: [{ name: 'blockHash', internalType: 'bytes32', type: 'bytes32' }],
    stateMutability: 'view',
  },
  {
    type: 'function',
    inputs: [
      { name: 'requireSuccess', internalType: 'bool', type: 'bool' },
      {
        name: 'calls',
        internalType: 'struct IMulticall3.Call[]',
        type: 'tuple[]',
        components: [
          { name: 'target', internalType: 'address', type: 'address' },
          { name: 'callData', internalType: 'bytes', type: 'bytes' },
        ],
      },
    ],
    name: 'tryAggregate',
    outputs: [
      {
        name: 'returnData',
        internalType: 'struct IMulticall3.Result[]',
        type: 'tuple[]',
        components: [
          { name: 'success', internalType: 'bool', type: 'bool' },
          { name: 'returnData', internalType: 'bytes', type: 'bytes' },
        ],
      },
    ],
    stateMutability: 'payable',
  },
  {
    type: 'function',
    inputs: [
      { name: 'requireSuccess', internalType: 'bool', type: 'bool' },
      {
        name: 'calls',
        internalType: 'struct IMulticall3.Call[]',
        type: 'tuple[]',
        components: [
          { name: 'target', internalType: 'address', type: 'address' },
          { name: 'callData', internalType: 'bytes', type: 'bytes' },
        ],
      },
    ],
    name: 'tryBlockAndAggregate',
    outputs: [
      { name: 'blockNumber', internalType: 'uint256', type: 'uint256' },
      { name: 'blockHash', internalType: 'bytes32', type: 'bytes32' },
      {
        name: 'returnData',
        internalType: 'struct IMulticall3.Result[]',
        type: 'tuple[]',
        components: [
          { name: 'success', internalType: 'bool', type: 'bool' },
          { name: 'returnData', internalType: 'bytes', type: 'bytes' },
        ],
      },
    ],
    stateMutability: 'payable',
  },
] as const

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Ownable
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

export const ownableAbi = [
  {
    type: 'function',
    inputs: [],
    name: 'owner',
    outputs: [{ name: '', internalType: 'address', type: 'address' }],
    stateMutability: 'view',
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
    inputs: [{ name: 'newOwner', internalType: 'address', type: 'address' }],
    name: 'transferOwnership',
    outputs: [],
    stateMutability: 'nonpayable',
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
    type: 'error',
    inputs: [{ name: 'owner', internalType: 'address', type: 'address' }],
    name: 'OwnableInvalidOwner',
  },
  {
    type: 'error',
    inputs: [{ name: 'account', internalType: 'address', type: 'address' }],
    name: 'OwnableUnauthorizedAccount',
  },
] as const
