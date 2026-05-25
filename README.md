# ft-mcp

An MCP server for the [42 API](https://api.intra.42.fr/apidoc), built with the official [Go MCP SDK](https://github.com/modelcontextprotocol/go-sdk).

Each user connects with their own 42 API credentials, so your rate limits and permissions are entirely your own.

**Live server:** `https://ft-mcp.tanvirahmedanas.com/mcp`

---

## Setup

### 1. Get 42 API credentials

Create an application at [profile.intra.42.fr/oauth/applications/new](https://profile.intra.42.fr/oauth/applications/new). You'll need the **Client UID** and **Client Secret**.

### 2. Get a bearer token

**Browser:** visit **[ft-mcp.tanvirahmedanas.com/token](https://ft-mcp.tanvirahmedanas.com/token)**, enter your credentials, and copy the token.

**CLI:**
```bash
curl -s -X POST https://ft-mcp.tanvirahmedanas.com/token \
  -d "grant_type=client_credentials" \
  -d "client_id=YOUR_CLIENT_UID" \
  -d "client_secret=YOUR_CLIENT_SECRET" \
  | jq -r .access_token
```

Tokens are valid for **24 hours**. When one expires, requests return `401` — just re-run the command above to get a fresh token and update your client config.

> Claude Code handles authentication automatically via OAuth — skip this step for that client.

---

## Connecting clients

<details>
<summary>Claude.ai</summary>

Go to **Settings → Connectors → Add custom connector**:

| Field | Value |
|-------|-------|
| MCP Server URL | `https://ft-mcp.tanvirahmedanas.com/mcp` |
| Authorization Header | `Bearer YOUR_TOKEN` |

Get your token from [ft-mcp.tanvirahmedanas.com/token](https://ft-mcp.tanvirahmedanas.com/token).

</details>

<details>
<summary>Claude Code</summary>

**Remote (recommended):**
```bash
claude mcp add --transport http ft-mcp https://ft-mcp.tanvirahmedanas.com/mcp
```
A browser window will open to complete the OAuth flow.

**Local (stdio):** build the binary and add to `~/.claude/settings.json`:
```json
{
  "mcpServers": {
    "ft-mcp": {
      "command": "/path/to/ft-mcp",
      "env": {
        "FT_CLIENT_ID": "your_client_uid",
        "FT_CLIENT_SECRET": "your_client_secret"
      }
    }
  }
}
```

</details>

<details>
<summary>Claude Desktop</summary>

Edit `claude_desktop_config.json` (build the binary first):
```json
{
  "mcpServers": {
    "ft-mcp": {
      "command": "/path/to/ft-mcp",
      "env": {
        "FT_CLIENT_ID": "your_client_uid",
        "FT_CLIENT_SECRET": "your_client_secret"
      }
    }
  }
}
```

</details>

<details>
<summary>Cursor</summary>

Add to `~/.cursor/mcp.json`:
```json
{
  "mcpServers": {
    "ft-mcp": {
      "url": "https://ft-mcp.tanvirahmedanas.com/mcp",
      "headers": {
        "Authorization": "Bearer YOUR_TOKEN"
      }
    }
  }
}
```

</details>

<details>
<summary>VS Code</summary>

Add to `.vscode/mcp.json`:
```json
{
  "servers": {
    "ft-mcp": {
      "type": "http",
      "url": "https://ft-mcp.tanvirahmedanas.com/mcp",
      "headers": {
        "Authorization": "Bearer YOUR_TOKEN"
      }
    }
  }
}
```

</details>

<details>
<summary>Windsurf</summary>

Add to `~/.codeium/windsurf/mcp_config.json`:
```json
{
  "mcpServers": {
    "ft-mcp": {
      "serverUrl": "https://ft-mcp.tanvirahmedanas.com/mcp",
      "headers": {
        "Authorization": "Bearer YOUR_TOKEN"
      }
    }
  }
}
```

</details>

<details>
<summary>OpenCode</summary>

Add to your OpenCode configuration file. See [OpenCode MCP docs](https://opencode.ai/docs/mcp-servers) for more info.

**Remote (recommended):**
```json
"mcp": {
  "ft-mcp": {
    "type": "remote",
    "url": "https://ft-mcp.tanvirahmedanas.com/mcp",
    "headers": {
      "Authorization": "Bearer YOUR_TOKEN"
    },
    "enabled": true
  }
}
```

Get your token from [ft-mcp.tanvirahmedanas.com/token](https://ft-mcp.tanvirahmedanas.com/token).

**Local (stdio):** build the binary first, then:
```json
{
  "mcp": {
    "ft-mcp": {
      "type": "local",
      "command": ["/path/to/ft-mcp"],
      "environment": {
        "FT_CLIENT_ID": "your_client_uid",
        "FT_CLIENT_SECRET": "your_client_secret"
      },
      "enabled": true
    }
  }
}
```

</details>

<details>
<summary>Other clients</summary>

Any MCP client that supports stdio can connect via [mcp-remote](https://www.npmjs.com/package/mcp-remote):
```json
{
  "mcpServers": {
    "ft-mcp": {
      "command": "npx",
      "args": ["mcp-remote", "https://ft-mcp.tanvirahmedanas.com/mcp"]
    }
  }
}
```

For clients that support custom HTTP headers, use the bearer token from [/token](https://ft-mcp.tanvirahmedanas.com/token) and add `Authorization: Bearer YOUR_TOKEN` to your headers config.

</details>

---

## Tools

| Tool | Description |
|------|-------------|
| `get_user` | Get a user profile by login or numeric ID |
| `list_users` | List users, optionally filtered by campus |
| `get_user_cursus` | Get a user's cursus info (level, grade, skills) |
| `get_user_projects` | Get a user's project submissions and status |
| `get_user_achievements` | Get a user's achievements |
| `list_campus` | List all 42 campuses |
| `get_campus_users` | List users at a specific campus |
| `get_locations` | Get active locations (who's logged into a computer) |
| `list_cursus` | List all cursus (curricula) |
| `list_projects` | List projects, optionally filtered by cursus |
| `search_projects` | Search projects by name to get their numeric ID |
| `list_project_submissions` | Get all submissions for a project, filtered by campus and date range. Response includes total count. Each result has a `validated?` field for client-side filtering. |
| `count_project_submissions` | Get the total count of submissions for a project matching given filters — without fetching the records |
| `list_events` | List events, optionally filtered by campus |

---

## Self-hosting

```bash
go build -o ft-mcp .
./ft-mcp --transport http --port 8080
```

## API reference

See [`docs/42API.md`](docs/42API.md) for the full 42 API reference — auth flows, rate limits, pagination, and all resource types.
