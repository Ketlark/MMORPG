# CLAUDE.md - MMORPG Unity Client

## Project Overview

Client Unity d'un MMORPG 3-tier : **Unity Client** (C#) - **Go Server** (gRPC) - **Protocol Buffers**.

- Unity 6000.5.0b4 (Unity 6)
- Universal Render Pipeline (URP)
- Communication serveur via gRPC (Grpc.Net.Client.Web)
- Isometric tilemap, combat au tour par tour

## Architecture

```
LoginScene → LoginUI (auth) → GameScene → GameManager
                                         ├── GrpcClient (singleton, gRPC streams)
                                         ├── IsometricMap (tilemap, mouvement)
                                         ├── CombatManager (tour par tour)
                                         └── GameHUD (UI)
```

## Namespaces

- `MMORPG.Game` - Logique jeu (GameManager, PlayerEntity)
- `MMORPG.Network` - gRPC client (GrpcClient, GrpcStubs)
- `MMORPG.Map` - Carte isométrique (IsometricMap)
- `MMORPG.UI` - Interface (LoginUI, GameHUD)
- `MMORPG.Combat` - Combat (CombatManager, CombatPlayerEntity)

## Key Files

- `Assets/Scripts/Game/GameManager.cs` - Controller principal (Singleton)
- `Assets/Scripts/Network/GrpcClient.cs` - Client gRPC (Singleton)
- `Assets/Scripts/Map/IsometricMap.cs` - Rendu et logique de la carte
- `Assets/Scripts/Combat/CombatManager.cs` - Systeme de combat
- `Assets/Protos/` - Classes C# generees depuis les .proto
- `Assets/Plugins/` - DLLs gRPC/Protobuf (Grpc.Core, Google.Protobuf, etc.)

## Conventions

- Singletons pour GameManager et GrpcClient
- Evenements C# pour communication inter-systemes
- Les proto generes sont dans `Assets/Protos/` - NE PAS modifier manuellement
- Les DLLs natives sont dans `Assets/Plugins/` - versions specifiques pour chaque plateforme
- Scripts organisés par domaine fonctionnel (Combat/, Game/, Map/, Network/, UI/)

## Network Flow

1. `GrpcClient.ConnectToServer()` - connexion
2. `AuthService` - auth JWT
3. `GameService` - streaming bi-directionnel (positions, chat, events)
4. `CombatService` - actions de combat

## Combat System

- Tour par tour avec PA (Points d'Action) et PM (Points de Mouvement)
- Sorts avec calcul de degats
- Conditions de victoire par equipe

## Unity MCP Plugin

Le plugin `com.ivanmurzak.unity.mcp` est installé via Git URL.
Apres ouverture du projet dans Unity Editor :

1. Ouvrir `Window/AI Game Developer (Unity-MCP)`
2. Le port est calculé automatiquement (hash SHA256 du path projet)
3. Configurer Claude Code :
```bash
claude mcp add ai-game-developer "/Users/dehoux/dev/mmorpg/client/Library/mcp-server/osx-arm64/unity-mcp-server" port=<port> client-transport=stdio
```

## Server Repo

Le serveur Go est dans `/Users/dehoux/dev/mmorpg/server/` avec :
- PostgreSQL + Redis
- gRPC Gateway
- Services : Auth, Game, Combat
