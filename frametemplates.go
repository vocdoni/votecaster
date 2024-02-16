package main

import "strings"

func frame(template string) string {
	template = strings.ReplaceAll(template, "{server}", serverURL)
	template = strings.ReplaceAll(template, "{explorer}", explorerURL)
	template = strings.ReplaceAll(template, "{onvote}", onvoteURL)
	return template
}

var commonHeaders = `
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <link rel="apple-touch-icon" sizes="72x72" href="/app/apple-touch-icon.png">
    <link rel="icon" type="image/png" sizes="32x32" href="/app/favicon-32x32.png">
    <link rel="icon" type="image/png" sizes="16x16" href="/app/favicon-16x16.png">
    <link rel="manifest" href="/app/site.webmanifest">
    <link rel="mask-icon" href="/app/safari-pinned-tab.svg" color="#5bbad5">
    <meta name="msapplication-TileColor" content="#da532c">
    <meta name="theme-color" content="#ffffff">
    <meta name="og:image" content="https://black-glamorous-rabbit-362.mypinata.cloud/ipfs/QmVyhAuvdLQgWZ7xog2WtXP88B7TswChCqZdKxVUR5rDUq" />
    <title>farcaster.vote — Farcaster Polls by Vocdoni</title>
`

var frameMain = `
<!DOCTYPE html>
<html lang="en">
  <head>` +
	commonHeaders + `
    <meta name="fc:frame" content="vNext" />
    <meta name="fc:frame:image" content="data:image/png;base64,{image}" />
    <meta name="fc:frame:image:aspect_ratio" content="1.91:1" />
    <meta name="fc:frame:post_url" content="{server}/router/{processID}" />

    <meta name="fc:frame:button:1" content="Results" />
    <meta name="fc:frame:button:1:action" content="post" />
    <meta name="fc:frame:button:1:target" content="{server}/poll/results/{processID}" />

    <meta name="fc:frame:button:2" content="Vote" />
    <meta name="fc:frame:button:2:action" content="post" />
    <meta name="fc:frame:button:2:target" content="{server}/poll/{processID}" />

    <meta name="fc:frame:button:3" content="Info" />
    <meta name="fc:frame:button:3:action" content="post" />
    <meta name="fc:frame:button:3:target" content="{server}/info/{processID}" />

    <meta name="fc:frame:button:4" content="Create new" />
    <meta name="fc:frame:button:4:action" content="link" />
    <meta name="fc:frame:button:4:target" content="{server}" />

    <meta http-equiv="refresh" content="0; url={server}" />
  </head>
  <body>
    <h1>Hello Farcaster! This is <a href="{server}">Vocdoni</a>.</h1>
  </body>
</html>
`

var frameVote = `
<!DOCTYPE html>
<html lang="en">
  <head>` +
	commonHeaders + `
    <meta property="fc:frame" content="vNext" />
    <meta property="fc:frame:image" content="data:image/png;base64,{image}" />
    <meta name="fc:frame:image:aspect_ratio" content="1.91:1" />
    <meta property="fc:frame:post_url" content="{server}/vote/{processID}" />
    <meta property="fc:frame:button:1" content="{option0}" />
    <meta property="fc:frame:button:2" content="{option1}" />
    <meta property="fc:frame:button:3" content="{option2}" />
    <meta property="fc:frame:button:4" content="{option3}" />
    <meta http-equiv="refresh" content="0; url={server}" />
  </head>
  <body>
    <h1>Hello Farcaster! this is <a href="{server}">Vocdoni</a></h1>
  </body>
</html>
`

var frameAfterVote = `
<!DOCTYPE html>
<html lang="en">
  <head>` +
	commonHeaders + `
    <meta property="fc:frame" content="vNext" />
    <meta name="fc:frame:image:aspect_ratio" content="1.91:1" />
    <meta property="fc:frame:image" content="data:image/png;base64,{image}" />
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
<html lang="en">
  <head>` +
	commonHeaders + `
    <meta property="fc:frame" content="vNext" />
    <meta name="fc:frame:image:aspect_ratio" content="1.91:1" />
    <meta property="fc:frame:image" content="data:image/png;base64,{image}" />

    <meta property="fc:frame:post_url" content="{server}/{processID}" />
    <meta property="fc:frame:button:1" content="Back" />

    <meta property="fc:frame:button:2" content="Check at onvote.app" />
    <meta property="fc:frame:button:2:action" content="link" />
    <meta property="fc:frame:button:2:target" content="{onvote}/processes/{processID}" />

    <meta http-equiv="refresh" content="0; url={server}" />
  </head>
  <body>
    <h1>Hello Farcaster! this is <a href="{server}">Vocdoni</a></h1>
  </body>
</html>
`

var frameInfo = `
<!DOCTYPE html>
<html lang="en">
  <head>` +
	commonHeaders + `
    <meta property="fc:frame" content="vNext" />
    <meta name="fc:frame:image:aspect_ratio" content="1.91:1" />
    <meta property="fc:frame:image" content="data:image/png;base64,{image}" />
    <meta property="fc:frame:post_url" content="{server}/{processID}" />
    <meta property="fc:frame:button:1" content="Back" />

    <meta property="fc:frame:button:2" content="Onvote" />
    <meta property="fc:frame:button:2:action" content="link" />
    <meta property="fc:frame:button:2:target" content="{onvote}/processes/{processID}" />

    <meta property="fc:frame:button:3" content="Explorer" />
    <meta property="fc:frame:button:3:action" content="link" />
    <meta property="fc:frame:button:3:target" content="{explorer}/processes/show/#/{processID}" />

    <meta property="fc:frame:button:4" content="About us" />
    <meta property="fc:frame:button:4:action" content="link" />
    <meta property="fc:frame:button:4:target" content="https://warpcast.com/vocdoni" />

    <meta http-equiv="refresh" content="0; url={server}" />
  </head>
  <body>
    <h1>Hello Farcaster! this is <a href="{server}">Vocdoni</a></h1>
  </body>
</html>
`

var frameAlreadyVoted = `
<!DOCTYPE html>
<html lang="en">
  <head>` +
	commonHeaders + `
    <meta property="fc:frame" content="vNext" />
    <meta property="fc:frame:image" content="data:image/png;base64,{image}" />
    <meta name="fc:frame:image:aspect_ratio" content="1.91:1" />
    <meta property="fc:frame:post_url" content="{server}/{processID}" />
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
<html lang="en">
  <head>` +
	commonHeaders + `
    <meta property="fc:frame" content="vNext" />
    <meta property="fc:frame:image" content="data:image/png;base64,{image}" />
    <meta name="fc:frame:image:aspect_ratio" content="1.91:1" />
    <meta property="fc:frame:post_url" content="{server}/{processID}" />
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
<html lang="en">
  <head>` +
	commonHeaders + `
    <meta property="fc:frame" content="vNext" />
    <meta property="fc:frame:image" content="data:image/png;base64,{image}" />
    <meta name="fc:frame:image:aspect_ratio" content="1.91:1" />
    <meta property="fc:frame:post_url" content="{server}/{processID}" />
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
  <head>` +
	commonHeaders + `
  </head>
  <body>
      <img src="data:image/png;base64,{image}" alt="Image" />
  </body>
</html>
`
