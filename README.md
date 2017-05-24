# Github auto stats

Automatically retrieve important data of Pull Request and fill in the corresponding column in Google Spreadsheet

## Update secret information

- `private/chatwork_secret.json`: Chatwork BOT access token and room ID number
- `private/chatwork_notice.tmpl`: Chatwork notice message template
- `private/spread_sheets.json`: List spreadsheet ID & columns information
- `private/google_secret.json`: Download from [Credentials tab](https://console.developers.google.com/start/api?id=sheets.googleapis.com) in the Google Developers Console. See more [detail](https://developers.google.com/sheets/api/quickstart/go)
- `private/github_oauth.json`: You'll need to [register your application](https://github.com/settings/applications/new). See more [detail](https://developer.github.com/v3/guides/basics-of-authentication/)
- `private/google_oauth.json`: This credential file is automatically generated
- `private/github_oauth.json`: This credential file is automatically generated

## Setup Development Environment

### Install Docker & Docker Compose

- Install [Docker](https://docs.docker.com/engine/installation/)
- Install [Docker Composer](https://docs.docker.com/compose/install/)

### Start and run project

```sh
$ cd /path/to/git-auto-stats
$ docker-compose up
$ docker-compose exec app /bin/bash
$ go get -u github.com/Masterminds/glide
$ glide install
```