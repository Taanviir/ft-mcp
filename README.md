# ft-mcp

An MCP server for the [42 API](https://api.intra.42.fr/apidoc), built with the official [Go MCP SDK](https://github.com/modelcontextprotocol/go-sdk).

Supports both **stdio** (for local use with Claude Code) and **HTTP** (for remote use with Claude.ai, ChatGPT, etc.).

## Setup

### 1. Get API credentials

Create an application at https://profile.intra.42.fr/oauth/applications/new.
You'll need the **Client UID** and **Client Secret**.

### 2. Configure credentials

```bash
cp .env.example .env
# Edit .env and fill in your credentials
```

### 3. Build

```bash
go build -o ft-mcp .
```

## Usage

### Stdio (Claude Code)

Run directly — reads from stdin, writes to stdout:

```bash
./ft-mcp
```

Add to Claude Code's MCP config (`~/.claude/settings.json`):

```json
{
  "mcpServers": {
    "42": {
      "command": "/path/to/ft-mcp",
      "env": {
        "FT_CLIENT_ID": "your_client_uid",
        "FT_CLIENT_SECRET": "your_client_secret"
      }
    }
  }
}
```

### HTTP (Claude.ai, ChatGPT, etc.)

```bash
./ft-mcp --transport http --port 8080
```

The server listens on `http://localhost:8080/mcp`. Point your MCP client at that URL.

To test locally with MCP Inspector:

```bash
npx @modelcontextprotocol/inspector http://localhost:8080/mcp
```

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
| `list_project_submissions` | Get all submissions for a project, filtered by campus, validation status, and date range |
| `list_events` | List events, optionally filtered by campus |

## API docs

See [`docs/42API.md`](docs/42API.md) for a full reference of the 42 API — auth flows, rate limits, pagination, all resource types, and common query patterns.
