<div align="center">

# MMORPG

**A turn-based isometric MMORPG inspired by Dofus**

Built with Unity 6 & Go — real-time gRPC-Web multiplayer, A* pathfinding, turn-based combat

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go](https://img.shields.io/badge/Go-1.25-00ADD8?logo=go)](https://go.dev/)
[![Unity](https://img.shields.io/badge/Unity-6-000000?logo=unity)](https://unity.com/)

</div>

---

## Overview

MMORPG is a full-stack multiplayer game featuring an isometric 2D world with server-authoritative pathfinding and Dofus-style turn-based combat. Players explore a tile-based map, move using A*-computed paths, and engage in tactical fights with action points, movement points, and spell casting.

<table>
<tr><td>

**World**
- Isometric diamond-shaped tilemap (64×32)
- Procedurally generated terrain
- 8-directional A* pathfinding with anti corner-cutting

</td><td>

**Combat**
- Turn-based tactical system
- Action points (AP) & movement points (MP)
- Spell casting with damage & range
- Turn queue with end-turn flow

</td><td>

**Multiplayer**
- gRPC-Web server streaming
- Live position & chat broadcast
- JWT authentication with bcrypt
- Per-IP rate limiting

</td></tr>
</table>

---

## Architecture

```
┌──────────────────────────────────────────────────────────┐
│                       Unity 6 Client                     │
│                                                          │
│   IsometricMap    GridEntity      CombatManager   UI     │
│   (Tile Gen)      (Waypoints)     (Turn System)   (HUD)  │
│         ╲            │              │              ╱      │
│          └───────────┼──────────────┘────────────┘       │
│                  GrpcClient (gRPC-Web)                    │
└──────────────────────────┬───────────────────────────────┘
                           │  HTTP/1.1 + gRPC-Web
┌──────────────────────────┼───────────────────────────────┐
│                     Go Server                             │
│                                                           │
│   ┌──────────┬───────────────────┬──────────────────┐    │
│   │   Auth   │       Game        │     Combat       │    │
│   │ Register │  Connect (stream) │  JoinCombat      │    │
│   │ Login    │  Move → A* Path   │  PerformAction   │    │
│   │ JWT      │  Chat (broadcast) │  EndTurn         │    │
│   └────┬─────┴─────────┬─────────┴────────┬─────────┘    │
│        │               │                  │              │
│   PostgreSQL       Redis             In-Memory           │
│   Accounts         Positions          Combat State       │
│   Characters       Sessions                              │
└──────────────────────────────────────────────────────────┘
```

<details>
<summary><strong>Tech Stack</strong></summary>

| Layer     | Technology                                  |
|-----------|---------------------------------------------|
| Client    | Unity 6, URP, C#, gRPC-Web, Input System   |
| Server    | Go 1.25, gRPC, grpc-web proxy              |
| Database  | PostgreSQL 16                               |
| Cache     | Redis 7                                     |
| Transport | Protocol Buffers 3, gRPC-Web over HTTP/1.1 |
| Infra     | Docker Compose, multi-stage Alpine builds   |

</details>

---

## Getting Started

### Prerequisites

- [Docker](https://docs.docker.com/get-docker/) & Docker Compose
- [Unity 6](https://unity.com/) (6000.0+)
- [Go 1.25+](https://go.dev/dl/) _(optional, for local server dev)_
- [protoc](https://grpc.io/docs/protoc-installation/) _(optional, for proto changes)_

### Quick Start

```bash
git clone git@github.com:Ketlark/MMORPG.git
cd MMORPG

# Start the full stack (PostgreSQL + Redis + Go server)
docker compose up --build
```

Then open `client/` in Unity Hub and press Play in the `GameScene`.

The server listens on **http://localhost:50051**.

<details>
<summary><strong>Configuration</strong></summary>

Set via environment variables or `docker-compose.yml`:

| Variable          | Default              | Description                          |
|-------------------|----------------------|--------------------------------------|
| `DB_HOST`         | `localhost`          | PostgreSQL host                      |
| `DB_PORT`         | `5432`               | PostgreSQL port                      |
| `DB_USER`         | `postgres`           | Database user                        |
| `DB_PASSWORD`     | `postgres`           | Database password                    |
| `DB_NAME`         | `mmorpg`             | Database name                        |
| `REDIS_HOST`      | `localhost`          | Redis host                           |
| `REDIS_PORT`      | `6379`               | Redis port                           |
| `JWT_SECRET`      | **required**         | Min 32 chars — server panics if unset |
| `SERVER_PORT`     | `:50051`             | Server listen address                |
| `ALLOWED_ORIGINS` | `http://localhost:*` | Comma-separated CORS origins         |

</details>

---

## API Reference

Proto definitions live in `server/api/proto/`.

### AuthService — Registration & Login

| Method     | Request                      | Response         |
|------------|------------------------------|------------------|
| `Register` | `username, email, password`  | `success, error` |
| `Login`    | `username, password`         | `success, token` |

### GameService — World Exploration

| Method    | Request              | Response                                          |
|-----------|----------------------|---------------------------------------------------|
| `Connect` | `character_name`     | `stream GameEvent` — real-time event stream       |
| `Move`    | `target_x, target_y` | `success, path[], path_cost, x, y, ap, mp`       |
| `Chat`    | `message`            | `Empty` — broadcasted to all connected players    |

> `GameEvent` is a `oneof`: `PlayerConnected`, `PlayerDisconnected`, `PlayerMoved`, `ChatMessage`, `MapData`, `CombatStart`

### CombatService — Turn-Based Fights

| Method           | Request                   | Response                              |
|------------------|---------------------------|---------------------------------------|
| `JoinCombat`     | `combat_id`               | `stream CombatEvent`                  |
| `PerformAction`  | `move / spell / skip`     | `success, ap_remaining, mp_remaining` |
| `EndTurn`        | `combat_id`               | `success, next_player_id`             |

<details>
<summary><strong>Proto Message Schemas</strong></summary>

**Movement**

```
PathNode { x, y }
MoveResponse { success, x, y, action_points, movement_points, path[], path_cost }
PlayerMoved { player_id, x, y, path[] }
```

**Combat**

```
CombatFighter { player_id, username, team, x, y, health, max_health, action_points, movement_points }
SpellAction { spell_id, target_x, target_y }
DamageResult { target_id, damage, health_remaining }
CombatReward { player_id, experience, gold }
```

</details>

---

## Project Structure

```
.
├── client/                              # Unity 6 project
│   └── Assets/
│       ├── Protos/                      # Generated C# protobuf classes
│       └── Scripts/
│           ├── Combat/                  # CombatManager, CombatPlayerEntity
│           ├── Game/                    # GameManager, GridEntity, PlayerEntity
│           ├── Map/                     # IsometricMap (diamond tile generation)
│           ├── Network/                 # GrpcClient (singleton, gRPC-Web)
│           └── UI/                      # LoginUI, GameHUD, EventSystemUtility
│
├── server/                              # Go server
│   ├── api/proto/                       # Protobuf definitions
│   │   ├── auth.proto
│   │   ├── combat.proto
│   │   ├── game.proto
│   │   └── gen/                         # Generated Go gRPC stubs
│   ├── cmd/server/                      # Entry point, middleware, wiring
│   ├── internal/
│   │   ├── auth/                        # JWT auth, register/login handlers
│   │   ├── combat/                      # Turn-based combat engine
│   │   ├── config/                      # Env config with wildcard CORS
│   │   ├── database/                    # PostgreSQL migrations + Redis
│   │   ├── game/                        # Game service, A* pathfinding, map
│   │   └── models/                      # Domain models
│   ├── Dockerfile                       # Multi-stage Alpine build
│   └── go.mod
│
├── proto/                               # Shared proto source files
├── docker-compose.yml                   # Full stack orchestration
├── LICENSE                              # MIT
└── .gitignore
```

---

## Development

<details>
<summary><strong>Running locally without Docker</strong></summary>

```bash
# Start infrastructure only
docker compose up postgres redis

# Run the server
cd server
export JWT_SECRET="dev-secret-key-change-in-production-minimum-32-chars!"
export DB_HOST=localhost
export REDIS_HOST=localhost
go run ./cmd/server
```

</details>

<details>
<summary><strong>Regenerating protobuf code</strong></summary>

```bash
cd server

# Install protoc plugins
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Go server stubs
protoc --go_out=./api/proto/gen --go_opt=paths=source_relative \
       --go-grpc_out=./api/proto/gen --go-grpc_opt=paths=source_relative \
       -I api/proto api/proto/*.proto

# C# client classes
protoc --csharp_out=../client/Assets/Protos \
       -I api/proto api/proto/*.proto
```

</details>

<details>
<summary><strong>Design Decisions</strong></summary>

**Server-authoritative movement**
The client sends a target cell. The server computes the A* path, validates walkability, and broadcasts waypoints to all connected players. The client only animates — it never decides the route.

**Anti corner-cutting**
Diagonal movement requires both adjacent cardinal cells to be walkable. This prevents visually clipping through wall corners.

**No `Find()` in Unity**
All component references use `SerializeField` + Inspector assignment, static events (`GameManager.OnChatMessageReceived`), singletons (`GrpcClient.Instance`, `CombatManager.Instance`), and `OnValidate` for editor-time auto-wiring.

**GridEntity base class**
`PlayerEntity` and `CombatPlayerEntity` share movement animation, sprite creation, and grid snapping via a common `GridEntity` base class with a waypoint queue.

**Middleware chain**
The Go server chains interceptors: logging → rate limiting → JWT auth. Stream interceptors apply the same logging + auth pattern.

</details>

---

## License

This project is licensed under the [MIT License](LICENSE).
