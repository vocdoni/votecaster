package main

import "strings"

func frame(template string) string {
	return strings.ReplaceAll(template, "{server}", serverURL)
}

var frameMain = `
<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <meta name="fc:frame" content="vNext" />
    <meta name="fc:frame:image" content="data:image/png;base64,{image}" />
    <meta name="og:image" content="https://black-glamorous-rabbit-362.mypinata.cloud/ipfs/QmVyhAuvdLQgWZ7xog2WtXP88B7TswChCqZdKxVUR5rDUq" />
    <meta name="fc:frame:image:aspect_ratio" content="1.91:1" />
    <meta name="fc:frame:post_url" content="{server}/router/{processID}" />

    <meta name="fc:frame:button:1" content="Results" />
    <meta name="fc:frame:button:1:action" content="post" />
    <meta name="fc:frame:button:1:target" content="{server}/poll/results/{processID}" />

    <meta name="fc:frame:button:2" content="Vote" />
    <meta name="fc:frame:button:2:action" content="post" />
    <meta name="fc:frame:button:2:target" content="{server}/poll/{processID}" />

    <meta name="fc:frame:button:3" content="Create new" />
    <meta name="fc:frame:button:3:action" content="link" />
    <meta name="fc:frame:button:3:target" content="{server}" />

    <meta name="fc:frame:button:4" content="onvote.app" />
    <meta name="fc:frame:button:4:action" content="link" />
    <meta name="fc:frame:button:4:target" content="https://onvote.app" />

    <meta http-equiv="refresh" content="0; url={server}" />
    <title>Vocdoni vote frame</title>
  </head>
  <body>
    <h1>Hello Farcaster! This is <a href="{server}">Vocdoni</a>.</h1>
  </body>
</html>
`

var frameVote = `
<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <meta property="fc:frame" content="vNext" />
    <meta property="fc:frame:image" content="data:image/png;base64,{image}" />
    <meta name="og:image" content="https://black-glamorous-rabbit-362.mypinata.cloud/ipfs/QmVyhAuvdLQgWZ7xog2WtXP88B7TswChCqZdKxVUR5rDUq" />
    <meta name="fc:frame:image:aspect_ratio" content="1.91:1" />
    <meta property="fc:frame:post_url" content="{server}/vote/{processID}" />
    <meta property="fc:frame:button:1" content="{option0}" />
    <meta property="fc:frame:button:2" content="{option1}" />
    <meta property="fc:frame:button:3" content="{option2}" />
    <meta property="fc:frame:button:4" content="{option3}" />
    <meta http-equiv="refresh" content="0; url={server}" />
    <title>Vocdoni Frame</title>
  </head>
  <body>
    <h1>Hello Farcaster! this is <a href="{server}">Vocdoni</a></h1>
  </body>
</html>
`

var frameAfterVote = `
<!DOCTYPE html>
<html>
  <head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <meta property="fc:frame" content="vNext" />
    <meta name="og:image" content="https://black-glamorous-rabbit-362.mypinata.cloud/ipfs/QmVyhAuvdLQgWZ7xog2WtXP88B7TswChCqZdKxVUR5rDUq" />
    <meta name="fc:frame:image:aspect_ratio" content="1.91:1" />
    <meta property="fc:frame:image" content="https://black-glamorous-rabbit-362.mypinata.cloud/ipfs/QmVyhAuvdLQgWZ7xog2WtXP88B7TswChCqZdKxVUR5rDUq" />
    <meta property="fc:frame:post_url" content="{server}/poll/results/{processID}" />
    <meta property="fc:frame:button:1" content="Results" />
    <meta property="fc:frame:button:2" content="Verify on explorer" />
    <meta property="fc:frame:button:2:action" content="link" />
    <meta property="fc:frame:button:2:target" content="{explorer}/verify/#/{nullifier}" />
    <meta http-equiv="refresh" content="0; url={server}" />
  </head>
  <body>
    <h1>Hello Farcaster! this is <a href="{server}">Vocdoni</a></h1>
  </body>
</html>
`

var frameResults = `
<!DOCTYPE html>
<html>
  <head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <meta property="fc:frame" content="vNext" />
    <meta name="og:image" content="https://black-glamorous-rabbit-362.mypinata.cloud/ipfs/QmVyhAuvdLQgWZ7xog2WtXP88B7TswChCqZdKxVUR5rDUq" />
    <meta name="fc:frame:image:aspect_ratio" content="1.91:1" />
    <meta property="fc:frame:image" content="data:image/png;base64,{image}" />
    <meta property="fc:frame:post_url" content="{server}/main/{processID}" />
    <meta property="fc:frame:button:1" content="Back" />
    <meta http-equiv="refresh" content="0; url={server}" />
  </head>
  <body>
    <h1>Hello Farcaster! this is <a href="{server}">Vocdoni</a></h1>
  </body>
</html>
`

var frameAlreadyVoted = `
<!DOCTYPE html>
<html>
  <head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <meta property="fc:frame" content="vNext" />
    <meta property="fc:frame:image" content="data:image/png;base64,{image}" />
    <meta name="og:image" content="https://black-glamorous-rabbit-362.mypinata.cloud/ipfs/QmVyhAuvdLQgWZ7xog2WtXP88B7TswChCqZdKxVUR5rDUq" />
    <meta name="fc:frame:image:aspect_ratio" content="1.91:1" />
    <meta property="fc:frame:post_url" content="{server}/main/{processID}" />
    <meta property="fc:frame:button:1" content="Back" />
    <meta property="fc:frame:button:2" content="Verify on explorer" />
    <meta property="fc:frame:button:2:action" content="link" />
    <meta property="fc:frame:button:2:target" content="{explorer}/verify/#/{nullifier}" />
    <meta http-equiv="refresh" content="0; url={server}" />
  </head>
  <body>
    <h1>Hello Farcaster! this is <a href="{server}">Vocdoni</a></h1>
  </body>
</html>
`

var frameNotElegible = `
<!DOCTYPE html>
<html>
  <head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <meta property="fc:frame" content="vNext" />
    <meta property="fc:frame:image" content="data:image/png;base64,{image}" />
    <meta name="og:image" content="https://black-glamorous-rabbit-362.mypinata.cloud/ipfs/QmVyhAuvdLQgWZ7xog2WtXP88B7TswChCqZdKxVUR5rDUq" />
    <meta name="fc:frame:image:aspect_ratio" content="1.91:1" />
    <meta property="fc:frame:post_url" content="{server}/main/{processID}" />
    <meta property="fc:frame:button:1" content="Back" />
    <meta http-equiv="refresh" content="0; url={server}" />
  </head>
  <body>
    <h1>Hello Farcaster! this is <a href="{server}">Vocdoni</a></h1>
  </body>
</html>
`

var frameError = `
<!DOCTYPE html>
<html>
  <head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <meta property="fc:frame" content="vNext" />
    <meta property="fc:frame:image" content="data:image/png;base64,{image}" />
    <meta name="og:image" content="https://black-glamorous-rabbit-362.mypinata.cloud/ipfs/QmVyhAuvdLQgWZ7xog2WtXP88B7TswChCqZdKxVUR5rDUq" />
    <meta name="fc:frame:image:aspect_ratio" content="1.91:1" />
    <meta property="fc:frame:post_url" content="{server}/main/{processID}" />
    <meta property="fc:frame:button:1" content="Back" />
    <meta http-equiv="refresh" content="0; url={server}" />
  </head>
  <body>
    <h1>Hello Farcaster! this is <a href="{server}">Vocdoni</a></h1>
  </body>
</html>
`
var testImageHTML = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Image Display Test</title>
</head>
<body>
    <img src="data:image/png;base64,{image}" alt="Image">
</body>
</html>
`
