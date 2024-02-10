package main

var frameMain = `
<html lang="en">
      <head>
        <meta property="fc:frame" content="vNext" />
        <meta property="fc:frame:image" content="https://images.unsplash.com/photo-1604065985083-86231f74c233" />
        <meta property="fc:frame:post_url" content="https://celoni.vocdoni.net/router/{processID}" />
	
		<meta property="fc:frame:button:1" content="Results" />
    	<meta property="fc:frame:button:1:action" content="post" />
    	<meta property="fc:frame:button:1:target" content="https://celoni.vocdoni.net/poll/results/{processID}" />


		<meta property="fc:frame:button:2" content="Vote" />
		<meta property="fc:frame:button:2:action" content="post" />
    	<meta property="fc:frame:button:2:target" content="https://celoni.vocdoni.net/poll/{processID}" />

          <title>Vocdoni vote frame</title>
      </head>
      <body>
        <h1>Hello Farcaster! this is <a href="https://vocdoni.io">Vocdoni</a></h1>
      </body>
</html>
`

var frameVote = `
<html lang="en">
      <head>
        <meta property="fc:frame" content="vNext" />
		<meta property="fc:frame:image" content="data:image/png;base64,{image}" />
        <meta property="fc:frame:post_url" content="https://celoni.vocdoni.net/vote/{processID}" />
        <meta property="fc:frame:button:1" content="{option0}" />
        <meta property="fc:frame:button:2" content="{option1}" />
        <meta property="fc:frame:button:3" content="{option2}" />
        <title>Vocdoni Frame</title>
      </head>
      <body>
        <h1>Hello Farcaster! this is <a href="https://vocdoni.io">Vocdoni</a></h1>
      </body>
</html>
`

var frameAfterVote = `
    <!DOCTYPE html>
    <html>
      <head>
				<meta property="fc:frame" content="vNext" />
				<meta property="fc:frame:image" content="https://black-glamorous-rabbit-362.mypinata.cloud/ipfs/QmVyhAuvdLQgWZ7xog2WtXP88B7TswChCqZdKxVUR5rDUq" />
        		<meta property="fc:frame:post_url" content="https://celoni.vocdoni.net/poll/results/{processID}" />
				<meta property="fc:frame:button:1" content="Results" />
				<meta property="fc:frame:button:2" content="Verify on explorer" />
  			  	<meta property="fc:frame:button:2:action" content="link" />
    			<meta property="fc:frame:button:2:target" content="https://dev.explorer.vote/verify/#/{nullifier}" />
	</head>
      <body>
        <h1>Hello Farcaster! this is <a href="https://vocdoni.io">Vocdoni</a></h1>
      </body>
    </html>
`
var frameResults = `
    <!DOCTYPE html>
    <html>
      <head>
				<meta property="fc:frame" content="vNext" />
				<meta property="fc:frame:image" content="data:image/png;base64,{image}" />
        		<meta property="fc:frame:post_url" content="https://celoni.vocdoni.net/" />
        		<meta property="fc:frame:button:1" content="Back" />
      </head>
      <body>
        <h1>Hello Farcaster! this is <a href="https://vocdoni.io">Vocdoni</a></h1>
      </body>
    </html>
`

var frameAlreadyVoted = `
    <!DOCTYPE html>
    <html>
      <head>
				<meta property="fc:frame" content="vNext" />
				<meta property="fc:frame:image" content="data:image/png;base64,{image}" />
        		<meta property="fc:frame:post_url" content="https://celoni.vocdoni.net/" />
        		<meta property="fc:frame:button:1" content="Back" />
				<meta property="fc:frame:button:2" content="Verify on explorer" />
  			  	<meta property="fc:frame:button:2:action" content="link" />
    			<meta property="fc:frame:button:2:target" content="https://dev.explorer.vote/verify/#/{nullifier}" />
		</head>
      <body>
        <h1>Hello Farcaster! this is <a href="https://vocdoni.io">Vocdoni</a></h1>
      </body>
    </html>
`

var frameNotElegible = `
    <!DOCTYPE html>
    <html>
      <head>
				<meta property="fc:frame" content="vNext" />
				<meta property="fc:frame:image" content="data:image/png;base64,{image}" />
        		<meta property="fc:frame:post_url" content="https://celoni.vocdoni.net/" />
        		<meta property="fc:frame:button:1" content="Back" />
		</head>
      <body>
        <h1>Hello Farcaster! this is <a href="https://vocdoni.io">Vocdoni</a></h1>
      </body>
    </html>
`

var frameError = `
    <!DOCTYPE html>
    <html>
      <head>
				<meta property="fc:frame" content="vNext" />
				<meta property="fc:frame:image" content="data:image/png;base64,{image}" />
        		<meta property="fc:frame:post_url" content="https://celoni.vocdoni.net/" />
        		<meta property="fc:frame:button:1" content="Back" />
		</head>
      <body>
        <h1>Hello Farcaster! this is <a href="https://vocdoni.io">Vocdoni</a></h1>
      </body>
    </html>
`
