# ft-mcp

An MCP server for the [42 API](https://api.intra.42.fr/apidoc), built with the official [Go MCP SDK](https://github.com/modelcontextprotocol/go-sdk).

Use your own 42 API credentials, so rate limits and permissions come from your own 42 application.

## Setup

### 1. Get 42 API credentials

Create an application at [profile.intra.42.fr/oauth/applications/new](https://profile.intra.42.fr/oauth/applications/new). You need the **Client UID** and **Client Secret**.

### 2. Add ft-mcp to your MCP client

Most clients can run the server with `npx`:

```json
{
  "mcpServers": {
    "ft-mcp": {
      "command": "npx",
      "args": ["-y", "ft-mcp"],
      "env": {
        "FT_CLIENT_ID": "your_client_uid",
        "FT_CLIENT_SECRET": "your_client_secret"
      }
    }
  }
}
```

The npm package installs the matching native `ft-mcp` binary for your OS from GitHub Releases.

## Claude Desktop

Edit `claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "ft-mcp": {
      "command": "npx",
      "args": ["-y", "ft-mcp"],
      "env": {
        "FT_CLIENT_ID": "your_client_uid",
        "FT_CLIENT_SECRET": "your_client_secret"
      }
    }
  }
}
```

## Codex App

Add the server with the Codex CLI:

```bash
codex mcp add ft-mcp \
  --env FT_CLIENT_ID="your_client_uid" \
  --env FT_CLIENT_SECRET="your_client_secret" \
  -- npx -y ft-mcp
```

## Tools

| Tool | Description |
|------|-------------|
| `get_user` | Get a user profile by login or numeric ID |
| `list_users` | List users, optionally filtered by campus |
| `get_user_cursus` | Get a user's cursus info, level, grade, and skills |
| `get_user_projects` | Get a user's project submissions and status |
| `get_user_achievements` | Get a user's achievements |
| `list_campus` | List all 42 campuses |
| `get_campus_users` | List users at a specific campus |
| `get_locations` | Get active locations |
| `list_cursus` | List all cursus |
| `list_projects` | List projects, optionally filtered by cursus |
| `search_projects` | Search projects by name to get their numeric ID |
| `list_project_submissions` | Get submissions for a project, filtered by campus and date range |
| `count_project_submissions` | Get the total count of submissions for a project without fetching every record |
| `list_events` | List events, optionally filtered by campus |

## From source

```bash
go build -o ft-mcp .
FT_CLIENT_ID="your_client_uid" FT_CLIENT_SECRET="your_client_secret" ./ft-mcp
```

HTTP mode is also available:

```bash
go build -o ft-mcp .
./ft-mcp --transport http --port 8080
```

## API reference

See [`docs/42API.md`](docs/42API.md) for the full 42 API reference, including auth flows, rate limits, pagination, and resource types.
