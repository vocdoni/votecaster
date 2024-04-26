<p align="center" width="100%">
    <img src="https://developer.vocdoni.io/img/vocdoni_logotype_full_white.svg" />
</p>

<p align="center" width="100%">
    <a href="https://github.com/vocdoni/REPO_NAME/commits/main/"><img src="https://img.shields.io/github/commit-activity/m/vocdoni/REPO_NAME" /></a>
    <a href="https://github.com/vocdoni/REPO_NAME/issues"><img src="https://img.shields.io/github/issues/vocdoni/REPO_NAME" /></a>
    <a href="https://discord.gg/xFTh8Np2ga"><img src="https://img.shields.io/badge/discord-join%20chat-blue.svg" /></a>
    <a href="https://twitter.com/vocdoni"><img src="https://img.shields.io/twitter/follow/vocdoni.svg?style=social&label=Follow" /></a>
</p>

  <div align="center">
    Vocdoni is the first universally verifiable, censorship-resistant, anonymous, and self-sovereign governance protocol. <br />
    Our main aim is a trustless voting system where anyone can speak their voice and where everything is auditable. <br />
    We are engineering building blocks for a permissionless, private and censorship resistant democracy.
    <br />
    <a href="https://developer.vocdoni.io/"><strong>Explore the developer portal »</strong></a>
    <br />
    <h3>More About Us</h3>
    <a href="https://vocdoni.io">Vocdoni Website</a>
    |
    <a href="https://vocdoni.app">Web Application</a>
    |
    <a href="https://explorer.vote/">Blockchain Explorer</a>
    |
    <a href="https://law.mit.edu/pub/remotevotingintheageofcryptography/release/1">MIT Law Publication</a>
    |
    <a href="https://chat.vocdoni.io">Contact Us</a>
    <br />
    <h3>Key Repositories</h3>
    <a href="https://github.com/vocdoni/vocdoni-node">Vocdoni Node</a>
    |
    <a href="https://github.com/vocdoni/vocdoni-sdk/">Vocdoni SDK</a>
    |
    <a href="https://github.com/vocdoni/ui-components">UI Components</a>
    |
    <a href="https://github.com/vocdoni/ui-scaffold">Application UI</a>
    |
    <a href="https://github.com/vocdoni/census3">Census3</a>
  </div>

# vote-frame

This repository contains a Farcaster frame for voting with the Vocdoni protocol. It is designed to enable integrated polling on Farcaster, leveraging Vocdoni's decentralized, verifiable, and censorship-resistant infrastructure. 

This frame is deployed at https://farcaster.vote/app

This repository contains the Go code necessary for constructing the server node. It features a web frontend that facilitates the creation of polls and oversees the communication with the Farcaster client, adhering to the frame specification.

The operation of the server polling node is centered around processing the signed message packet that originates from the Farcaster user upon engaging the vote button. This process involves extracting the public key from the signed message and packaging the signature into a Vocdoni vote transaction. 

Following the submission of the transaction, the Vocdoni blockchain undertakes the verification process to ensure: 

1. The public key is recognized as a valid Farcaster public key and is listed in the census.
2. The signature correctly corresponds to the public key.
3. The selected button accurately reflects the voting choice.

To assure the presence of a user's public key within the Farcaster protocol, the system employs the Vocdoni census3, 
a service which persistently scans the Optimism network for Farcaster registrations. [Census3 GitHub Repository](https://github.com/vocdoni/census3)

### Table of Contents
- [Getting Started](#getting-started)
- [Reference](#reference)
- [Examples](#examples)
- [Preview](#preview)
- [Disclaimer](#disclaimer)
- [Contributing](#contributing)
- [License](#license)


## Getting Started

To deploy the server node, Docker Compose is utilized:

```sh
cp .env.example .env
docker compose build
docker compose up -d
```

See the `.env` file; you need to define at least a Neynar API key.

After startup, navigate to http://localhost:8888 to access the web frontend.

Configuration and environment variables can be specified within the `.env` file.

For those operating within a Go-ready environment, you can also run the go code directly with several options:

```sh
go run . --mongoURL=mongodb://root:vocdoni@127.0.0.1:27017 --logLevel=debug --neynarAPIKey=<your_key> --web3=https://mainnet.optimism.io,https://optimism.llamarpc.com,https://optimism-mainnet.public.blastapi.io,https://rpc.ankr.com/optimism,https://optimism.drpc.org --indexer=false --apiEndpoint=https://api-dev.vocdoni.net/v2
```

## Reference

### Authentication

The Authentication API provides a set of endpoints for managing user authentication via Warpcast. 
This includes creating authentication links, verifying authentication status, and checking authentication tokens.

The authentication token must be set as a Bearer HTTP header `authorization: Bearer <token>`

The web application might store the bearer token in the local storage, so the user does not need to reauthenticate on each access.

The token expires after 15 days of non activity. Multiple tokens for the same user are allowed.

#### 1. Create Authentication Link

The authantication link is a Warpcast deep link that should be returned to the user (usually showing a QR code to scan with the smartphone camera).

- **Endpoint:** `/auth`
- **Method:** `GET`
- **Access:** Public
- **Description:** Creates a new authentication channel and returns a URL and an ID for the client to initiate the authentication process.
- **Returns:**
  - **HTTP 200 OK** on success with JSON payload containing:
    - `url`: The URL to which the user should be directed to complete the authentication process.
    - `id`: The unique identifier for the authentication request.
  - **HTTP 500 Internal Server Error** on failure with an error message.

```sh
curl -X GET "http://localhost:8888/auth"
```

#### 2. Verify Authentication

Once the user verifies on Warpcast, this endpoint can be used to fetch the Bearer token.
Note that this endpoint can only by called once (then it is removed and 404 will be returned).

- **Endpoint:** `/auth/{id}`
- **Method:** `GET`
- **Access:** Public
- **Description:** Verifies the status of an authentication channel using the ID provided when the channel was created. Returns an authentication token upon successful authentication.
- **URL Parameters:**
  - `id`: The unique identifier for the authentication request.
- **Returns:**
  - **HTTP 200 OK** on successful authentication with JSON payload containing:
    - `authToken`: The authentication token.
  - **HTTP 204 No Content** if authentication is still pending.
  - **HTTP 404 Not Found** if the specified ID does not correspond to an existing authentication channel.
  - **HTTP 500 Internal Server Error** on other errors with an error message.


```sh
curl -X GET "http://localhost:8888/auth/123e4567"
```

```json
{
    "authToken": "17c512b4-d55c-a2fe-144a-8fad0daa357c",
    "profile": {
        "fid": 237855,
        "username": "foo",
        "displayName": "bar 🎩 ⛓️‍💥",
        "bio": "I do stuff",
        "pfpUrl": "https://i.imgur.com/jFrkJ0KO.gif",
        "custody": "0xabcde....",
        "verifications": [
            "0xabcde..."
        ]
    },
    "reputation": 12,
    "reputationData": {
        "followersCount": 200,
        "electionsCreated": 20,
        "castedVotes": 54,
        "participationAchievement": 114
    }
}
```

#### 3. Check Authentication Token

Every time a user access the web application it is expected to check this endpoint. 
This action will update the expiration date.

- **Endpoint:** `/auth/check`
- **Method:** `GET`
- **Access:** Public
- **Description:** Checks if the provided authentication token is valid and updates the activity time. This endpoint should be used to verify token validity and refresh token activity.
- **Headers:**
  - `AuthToken`: The authentication token to be validated.
- **Returns:**
  - **HTTP 200 OK** if the token is valid.
  - **HTTP 404 Not Found** if the token is invalid or expired, with an error message.
  - **HTTP 400 Bad Request** if the `AuthToken` header is missing.


```sh
curl -X GET "http://example.com/auth/check" -H "authorization: Bearer {your_auth_token}"
```

```json
{
    "reputation": 12,
    "reputationData": {
        "followersCount": 200,
        "electionsCreated": 20,
        "castedVotes": 54,
        "participationAchievement": 114
    }
}
```

## Preview

This frame is live at https://farcaster.vote/app


## Contributing 

While we welcome contributions from the community, we do not track all of our issues on Github and we may not have the resources to onboard developers and review complex pull requests. That being said, there are multiple ways you can get involved with the project. 

Please review our [development guidelines](https://developer.vocdoni.io/development-guidelines).

## License

This repository is licensed under the [GNU Affero General Public License v3.0.](./LICENSE)


    Vocdoni Farcaster Vote Frame
    Copyright (C) 2024 Vocdoni Association

    This program is free software: you can redistribute it and/or modify
    it under the terms of the GNU Affero General Public License as published by
    the Free Software Foundation, either version 3 of the License, or
    (at your option) any later version.

    This program is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU Affero General Public License for more details.

    You should have received a copy of the GNU Affero General Public License
    along with this program.  If not, see <https://www.gnu.org/licenses/>.

[![License: AGPL v3](https://img.shields.io/badge/License-AGPL%20v3-blue.svg)](https://www.gnu.org/licenses/agpl-3.0)