### Setup and configuration

Available flags:
- `--help` to print help information and exit
- `--version` to print version / build information and exit
- `--no-update` to disable automatic update checks
- `--console` to enable logging to console
- `--debug` to add verbose logging and logging to console
- `-l` or `--logs` to set the logs directory
- `-c` or `--config` to set the config file

You will then need to create an `.env` file and put it as the `--config` file:

```env
TWITCHSPEAK_API_PORT=
TWITCHSPEAK_FRONTEND_URL=
TWITCHSPEAK_BACKEND_URL=
TWITCHSPEAK_SECRET_KEY=
TWITCHSPEAK_TWITCH_CLIENT_ID=
TWITCHSPEAK_TWITCH_CLIENT_SECRET=
TWITCHSPEAK_TWITCH_REDIRECT_URI=
TWITCHSPEAK_POSTGRES_HOST=
TWITCHSPEAK_POSTGRES_PORT=
TWITCHSPEAK_POSTGRES_USER=
TWITCHSPEAK_POSTGRES_PASSWORD=
TWITCHSPEAK_POSTGRES_DB=
TWITCHSPEAK_REDIS_HOST=
TWITCHSPEAK_REDIS_PORT=
TWITCHSPEAK_REDIS_PASSWORD=
TWITCHSPEAK_REDIS_DB=
TWITCHSPEAK_TEAMSPEAK_HOST=
TWITCHSPEAK_TEAMSPEAK_QUERYPORT=
TWITCHSPEAK_TEAMSPEAK_PORT=
TWITCHSPEAK_TEAMSPEAK_USER=
TWITCHSPEAK_TEAMSPEAK_PASSWORD=
TWITCHSPEAK_TEAMSPEAK_NICKNAME=
```

Please take note you will need a working [TeamSpeak 3 server](https://teamspeak.com) with opened ports and queryports, a working [Postgres instance](https://www.postgresql.org/) and a working [redis instance](https://redis.io/).