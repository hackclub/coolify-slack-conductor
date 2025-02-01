# Coolify Slack Conductor

Coolify supports sending notifications to Slack. However, it sends all notifications to a single Slack channel. This
means that all app's deployment notifications end up in one noisy channel rather than each project's respective
Slack channel.

<details>

<summary>Coolify's Notifications include... (expand me)</summary>

- Deployment success
- Deployment failure
- Container status changes
- Backup Success
- Backup Failure
- Scheduled task success
- Scheduled task failure
- Docker cleanup success
- Docker cleanup failure
- Server disk usage
- Server reachable
- Server unreachable

</details>

This simple Go app solves this problem by serving as a reverse proxy. It mirrors 🪞 notifications to their respective
project channels.

## 🫵 The Conductor

In order to route notification to the right Slack channel, The Conductor requires a list of **Destinations**. Each
Destination consists of:

- A Slack channel webhook URL
- A (or multiple) regular expression(s)

When a new notification webhook is sent to The Conductor, it builds a list of relevant Destinations and mirrors the
notification to them. Pretty simple!

So, how does The Conductor determine whether a Destination is relevant for a given notification?

It uses the Destination's regex(es). If any of the Destination's regexes matches any part of the notification webhook's
payload (body), then the notification will be sent to the Destination's Slack channel.

## Notification webhook payload

A typical Slack notification looks like this

![Image](https://github.com/user-attachments/assets/ce3839ea-027b-4a4f-ab04-85084bbba7e0)

However, it's webhook payload is JSON (built using Slack's [Block Kit](https://api.slack.com/block-kit)).

This means that when writing regular expressions for Destinations, you must consider the JSON payload itself. Here's an
example of one of those payloads:

```json
{
  "blocks": [
    {
      "type": "section",
      "text": {
        "type": "plain_text",
        "text": "Coolify Notification"
      }
    }
  ],
  "attachments": [
    {
      "color": "#00ff00",
      "blocks": [
        {
          "type": "header",
          "text": {
            "type": "plain_text",
            "text": "New version successfully deployed"
          }
        },
        {
          "type": "section",
          "text": {
            "type": "mrkdwn",
            "text": "New version successfully deployed for hackclub\/mfa:main\nApplication URL: https:\/\/mfa.hackclub.com\n\n**Project:** gary@mfa\n**Environment:** production\n**Deployment Logs:** https:\/\/app.coolify.io\/project\/ik0w8s404gg88ww0o4wgg048\/production\/application\/vogokcg8s4c4ok40880ssko8\/deployment\/oowo484co8go84ss0kso0gwc"
          }
        }
      ]
    }
  ]
}
```

but without the nice formatting (no new lines or indentation).

## Configuring Destinations

To begin, configurations are located in the [`config.yml`](config.yml) file.

As mentioned above, each Destination requires a Slack webhook URL and one (or multiple) regexes.

### 1. Obtaining a Slack Webhook URL

1. Visit the **Coolify** Slack app's Incoming Webhooks
   settings: https://api.slack.com/apps/A08AQL7JLT1/incoming-webhooks
2. Click "Add New Webhook to Workspace"
3. Choose your project's Slack channel
4. Click "Allow"
5. Copy the **Webhook URL** for your channel

Hang on to that URL.

### 2. Writing the regex

Keep in mind you may need to escape certain special characters.

For examples, I'd recommend reading the [`config.yml`](config.yml) file. However, here are a few things to know:

- Project-related notifications (e.g. Deployment success) include its project name.
  ```
  **Project:** gary@coolify-slack-conductor
  ```
  So, you can use that to match all notification related to your project.

- It may be helpful to look at existing notification messages to see what to match for. You can find them in
  the [`#coolify-notifs`](https://hackclub.slack.com/archives/C08AQL0DLF9) channel. Since coolify is open source, you
  can also find their notification templates [here](https://github.com/coollabsio/coolify/pull/4264).

### 3. Making the configuration

Inside [`config.yml`](config.yml), add a new array item under the `destinations` key. That item should have the
following keys:

- `name`: Must be all caps. Should be similar to your Slack channel's name.
- `regex`: Any array of regular expressions

Here's an example:

```yaml
  - name: HCB_ENGR_NOTIFS
    regex:
      - \*\*Project:\*\* gary@mfa\\n
      - \*\*Project:\*\* gary@g-verify\\n
      - \*\*Project:\*\* ian@bank-shields\\n
```

This Destination is named `HCB_ENGR_NOTIFS`. The regex will match notifications for the
projects `gary@mfa`, `gary@g-verify`, and `ian@bank-shields`.

You might notice that you're still hanging onto that Slack channel webhook URL. So, where does that go?

Since Slack webhook URLs should be treated like secrets, they're not stored in this codebase. Instead, the app will look
for them in its environment variables. In the example above, since the Destination's name is `HCB_ENGR_NOTIFS`, the app
expects an environment variable called `WEBHOOK_HCB_ENGR_NOTIFS_URL` to be set. It MUST be set, otherwise, the app
refuses the boot.

### 4. Deploying your changes

1. PR your changes
2. If you added a new Destination, add its respective environment variable to the Coolify deployment
    1. Go to Coolify
    2. Find the `gary@coolify-slack-conductor` app
    3. Find the environment variable section
    4. Add the new variable. If you destination is named `MY_CHANNEL`, then you should be
       adding `WEBHOOK_MY_CHANNEL_URL`. The URL should look something
       like `https://hooks.slack.com/services/T0230FmGR/B083YH3Sn8W/gTZwlvOe4gsrO6JeO3N3ghjk`.
3. Merge the PR.

The environment variable **_MUST_** be added before merging your changes!

## Main channel

The Conductor will always send all notification to the main channel. This is currently set
as [`#coolify-notifs`](https://hackclub.slack.com/archives/C08AQL0DLF9) and is configured via the `WEBHOOK_MAIN_URL`.

## Deployment

_via Docker on Coolify._

Set the following environment variables:

- `AUTH_KEY`: any string. This "secures" this reverse proxy and prevents people from spamming our channels
- `WEBHOOK_MAIN_URL`: the Slack webhook url of the main channel. See the [Main channel](#main-channel) section above.

### Setting up Coolify for Slack notification

1. Go to Coolify dashboard
2. On the left side, click "Notifications"
3. Click the "Slack" tab
4. Set the Webhook field
   https://coolify-conductor.hackclub.com?key=WHAT_EVER_YOU_SET_AS_AUTH_KEY
5. Enable it
