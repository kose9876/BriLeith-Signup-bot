# bri

## Setup

1. Copy `config.example.json` to `config.json`.
2. Fill in Discord bot settings in `config.json`.
3. Run `go run .`.

## Runtime JSON files

The app will auto-create these files on first run if they are missing:

- `profiles.json`
- `signups.json`
- `test_signups.json`
- `admin_state.json`
- `signup_schedule_state.json`

All generated JSON files use UTF-8 and two-space indentation.
