package main

import "strings"

func frame(template string) string {
	template = strings.ReplaceAll(template, "{server}", serverURL)
	template = strings.ReplaceAll(template, "{explorer}", explorerURL)
	template = strings.ReplaceAll(template, "{onvote}", onvoteURL)
	return template
}

var header = `
<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <link rel="apple-touch-icon" sizes="180x180" href="/app/apple-touch-icon.png?v=1">
    <link rel="icon" type="image/png" sizes="32x32" href="/app/favicon-32x32.png?v=1">
    <link rel="icon" type="image/png" sizes="16x16" href="/app/favicon-16x16.png?v=1">
    <link rel="manifest" href="/app/site.webmanifest?v=1">
    <link rel="mask-icon" href="/app/safari-pinned-tab.svg?v=1" color="#5bbad5">
    <link rel="shortcut icon" href="/app/favicon.ico?v=1">
    <meta name="msapplication-TileColor" content="#603cba">
    <meta name="msapplication-config" content="/app/browserconfig.xml?v=1">
    <meta name="theme-color" content="#ffffff">
    <!-- Open Graph Meta Tags -->
    <meta property="og:type" content="website" />
    <meta property="og:title" content="Votecaster - Farcaster polls">
    <meta property="og:url" content="https://farcaster.vote" />
    <meta property="og:description" content="Run quick polls and participate in Farcaster communities with e2e verifiable voting within a Frame. Built by Vocdoni." />
    <meta property="og:image" content="/app/opengraph.png" />
    <meta property="og:image:width" content="1200" />
    <meta property="og:image:height" content="630" />
    <meta property="og:image:alt" content="Votecaster presentation image. Votecaster. The Farcaster governance client. Run quick polls. Manage your community. Vote within a Frame." />
    <!-- Twitter Card Meta Tags -->
    <meta name="twitter:card" content="summary_large_image">
    <meta name="twitter:title" content="Votecaster - Farcaster polls">
    <meta name="twitter:description" content="Run quick polls and participate in Farcaster communities with e2e verifiable voting within a Frame. Built by Vocdoni.">
    <meta name="twitter:image" content="/app/opengraph.png">
    <meta name="twitter:image:alt" content="Votecaster presentation image. Votecaster. The Farcaster governance client. Run quick polls. Manage your community. Vote within a Frame.">

    <title>Votecaster - Farcaster polls</title>
    <link rel="preconnect" href="https://fonts.googleapis.com">
    <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
    <link href="https://fonts.googleapis.com/css2?family=Inter:wght@100..800&display=swap" rel="stylesheet">
    <style>
    * {
      font-family: "Inter", sans-serif;
    }
    </style>
`

var body = `
  </head>
  <body>
    <div style="margin: 0 auto; max-width: 100%; width: 600px;">
      <p><img src="{image}" alt="{title} poll image" style="max-width: 100%" /> </p>
      <h1>{title}</h1>
      <p>Create your own secure and decentralized polls with <a href="{server}">Votecaster</a>.</p>
    </div>
  </body>
</html>`

var frameMain = header + `
    <meta name="fc:frame" content="vNext" />
    <meta name="fc:frame:image" content="{image}" />
    <meta name="fc:frame:image:aspect_ratio" content="1:1" />
    <meta name="fc:frame:post_url" content="{server}/router/{processID}" />

    <meta name="fc:frame:button:1" content="🗳️ Vote" />
    <meta name="fc:frame:button:1:action" content="post" />
    <meta name="fc:frame:button:1:target" content="{server}/poll/{processID}" />

    <meta name="fc:frame:button:2" content="👀 Results" />
    <meta name="fc:frame:button:2:action" content="post" />
    <meta name="fc:frame:button:2:target" content="{server}/poll/results/{processID}" />

    <meta name="fc:frame:button:3" content="🔎 Info" />
    <meta name="fc:frame:button:3:action" content="post" />
    <meta name="fc:frame:button:3:target" content="{server}/info/{processID}" />

    <meta name="fc:frame:button:4" content="📝 New" />
    <meta name="fc:frame:button:4:action" content="link" />
    <meta name="fc:frame:button:4:target" content="{server}" />

    <meta http-equiv="refresh" content="0;url={server}/app/#poll/{processID}" />
` + body

var frameVote = header + `
    <meta property="fc:frame" content="vNext" />
    <meta property="fc:frame:image" content="{image}" />
    <meta name="fc:frame:image:aspect_ratio" content="1:1" />
    <meta property="fc:frame:post_url" content="{server}/vote/{processID}" />
    <meta property="fc:frame:button:1" content="{option0}" />
    <meta property="fc:frame:button:2" content="{option1}" />
    <meta property="fc:frame:button:3" content="{option2}" />
    <meta property="fc:frame:button:4" content="{option3}" />
    <meta property="fc:frame:state" content='{state}' />
` + body

var frameAfterVote = header + `
    <meta property="fc:frame" content="vNext" />
    <meta name="fc:frame:image:aspect_ratio" content="1:1" />
    <meta property="fc:frame:image" content="{image}" />
    <meta property="fc:frame:post_url" content="{server}/poll/results/{processID}" />
    <meta property="fc:frame:button:1" content="📋 Results" />
    <meta property="fc:frame:button:2" content="🔎 Verify on explorer" />
    <meta property="fc:frame:button:2:action" content="link" />
    <meta property="fc:frame:button:2:target" content="{explorer}/verify/#/{nullifier}" />
` + body

var frameResults = header + `
    <meta property="fc:frame" content="vNext" />
    <meta name="fc:frame:image:aspect_ratio" content="1:1" />
    <meta property="fc:frame:image" content="{image}" />

    <meta property="fc:frame:post_url" content="{server}/{processID}" />
    <meta property="fc:frame:button:1" content="⬅️ Back" />

    <meta property="fc:frame:button:2" content="🔎 Explorer" />
    <meta property="fc:frame:button:2:action" content="link" />
    <meta property="fc:frame:button:2:target" content="{explorer}/processes/show/#/{processID}" />

    <meta property="fc:frame:button:3" content="📋 Participants" />
    <meta property="fc:frame:button:3:action" content="link" />
    <meta property="fc:frame:button:3:target" content="{server}/app/#poll/{processID}" />
` + body

var frameFinalResults = header + `
    <meta property="fc:frame" content="vNext" />
    <meta name="fc:frame:image:aspect_ratio" content="1:1" />
    <meta property="fc:frame:image" content="{image}" />
    <meta http-equiv="refresh" content="0;url={server}/app/#poll/{processID}" />
` + body

var frameInfo = header + `
    <meta property="fc:frame" content="vNext" />
    <meta name="fc:frame:image:aspect_ratio" content="1:1" />
    <meta property="fc:frame:image" content="{image}" />
    <meta property="fc:frame:post_url" content="{server}/{processID}" />
    <meta property="fc:frame:button:1" content="️⬅️ Back" />

    <meta property="fc:frame:button:2" content="🔎 Explorer" />
    <meta property="fc:frame:button:2:action" content="link" />
    <meta property="fc:frame:button:2:target" content="{explorer}/processes/show/#/{processID}" />

    <meta property="fc:frame:button:3" content="😊 About us" />
    <meta property="fc:frame:button:3:action" content="link" />
    <meta property="fc:frame:button:3:target" content="https://warpcast.com/vocdoni" />
` + body

var frameAlreadyVoted = header + `
    <meta property="fc:frame" content="vNext" />
    <meta property="fc:frame:image" content="{image}" />
    <meta name="fc:frame:image:aspect_ratio" content="1:1" />
    <meta property="fc:frame:post_url" content="{server}/{processID}" />
    <meta property="fc:frame:button:1" content="⬅️ Back" />
    <meta property="fc:frame:button:2" content="🔍 Verify on explorer" />
    <meta property="fc:frame:button:2:action" content="link" />
    <meta property="fc:frame:button:2:target" content="{explorer}/verify/#/{nullifier}" />
` + body

var frameNotElegible = header + `
    <meta property="fc:frame" content="vNext" />
    <meta property="fc:frame:image" content="{image}" />
    <meta name="fc:frame:image:aspect_ratio" content="1:1" />
    <meta property="fc:frame:post_url" content="{server}/{processID}" />
    <meta property="fc:frame:button:1" content="Back" />
` + body

var frameDelegatedVote = header + `
    <meta property="fc:frame" content="vNext" />
    <meta property="fc:frame:image" content="{image}" />
    <meta name="fc:frame:image:aspect_ratio" content="1:1" />
    <meta property="fc:frame:post_url" content="{server}/{processID}" />
    <meta property="fc:frame:button:1" content="Back" />
` + body

var frameError = header + `
    <meta property="fc:frame" content="vNext" />
    <meta property="fc:frame:image" content="{image}" />
    <meta name="fc:frame:image:aspect_ratio" content="1:1" />
    <meta property="fc:frame:post_url" content="{server}/{processID}" />
    <meta property="fc:frame:button:1" content="⬅️ Back" />
` + body

var frameNotifications = header + `
    <meta property="fc:frame" content="vNext" />
    <meta property="fc:frame:image" content="{image}" />
    <meta name="fc:frame:image:aspect_ratio" content="1:1" />
    <meta property="fc:frame:post_url" content="{server}/notifications/set" />
    <meta property="fc:frame:button:1" content="✅ Allow" />
    <meta property="fc:frame:button:2" content="❌ Disable" />
    <meta property="fc:frame:button:3" content="🔍 Mute a user" />
` + body

var frameNotificationsResponse = header + `
    <meta property="fc:frame" content="vNext" />
    <meta property="fc:frame:image" content="{image}" />
    <meta name="fc:frame:image:aspect_ratio" content="1:1" />
    <meta property="fc:frame:post_url" content="{server}/notifications" />
    <meta property="fc:frame:button:1" content="⬅️ Back" />
` + body

var frameNotificationsManager = header + `
    <meta property="fc:frame" content="vNext" />
    <meta property="fc:frame:image" content="{image}" />
    <meta name="fc:frame:image:aspect_ratio" content="1:1" />
    <meta property="fc:frame:post_url" content="{server}/notifications/filter" />
    <meta property="fc:frame:input:text" content="User handle" />
    <meta property="fc:frame:button:1" content="✅ Allow" />
    <meta property="fc:frame:button:2" content="🤐 Mute" />
` + body
