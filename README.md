# Vocdoni Frame Vote for Farcaster

> see https://onvote.app for advanced capabilities on web3 voting

The Vocdoni Frame Vote for Farcaster is a framework designed to enable integrated polling on Farcaster, 
leveraging the decentralized, verifiable, and censorship-resistant Vocdoni infrastructure. 

This repository is home to the Go code necessary for constructing the server node. 
It features a web frontend that facilitates the creation of polls and oversees the communication 
with the Farcaster client adhering to the frame specification.

The operation of the server polling node is centered around processing the signed message packet that 
originates from the Farcaster user upon engaging the vote button. 
This process involves extracting the public key from the signed message and packaging the signature 
into a Vocdoni vote transaction. 

Following the submission of the transaction, the Vocdoni blockchain undertakes the verification process to ensure: 

1. The public key is recognized as a valid Farcaster public key and is listed in the census.
2. The signature correctly corresponds to the public key.
3. The selected button accurately reflects the voting choice.

To assure the presence of a user's public key within the Farcaster protocol, the system employs the Vocdoni census3, 
a service which persistently scans the Optimism network for Farcaster registrations. [Census3 GitHub Repository](https://github.com/vocdoni/census3)

## Instructions

To deploy the server node, Docker Compose is utilized:

```sh
docker compose build
docker compose up -d
```

After startup, navigate to http://localhost:8888 to access the web frontend.

Configuration and environmental variables can be specified within the `.env` file.

### Developing

For those operating within a Go-ready environment, the following command reveals additional options:

```sh
go run . -h
```

