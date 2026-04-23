using UnityEngine;
using System.Collections.Generic;
using System.Threading;
using Grpc.Core;
using MMORPG.Network;
using MMORPG.Map;

namespace MMORPG.Game
{
    public class GameManager : MonoBehaviour
    {
        public static GameManager Instance { get; private set; }

        public static event System.Action<string, string> OnChatMessageReceived;

        [SerializeField] private IsometricMap _map;
        [SerializeField] private GameObject _playerPrefab;

        private Dictionary<string, PlayerEntity> _players = new();
        private PlayerEntity _localPlayer;
        private CancellationTokenSource _cts;
        private string _localPlayerId;
        private string _characterName;
        private int _moveSeq;

        public PlayerEntity LocalPlayer => _localPlayer;
        public IsometricMap Map => _map;

        private void Awake()
        {
            if (Instance != null && Instance != this) { Destroy(gameObject); return; }
            Instance = this;
        }

        private void Start()
        {
            _characterName = PlayerPrefs.GetString("CharacterName", "Player");
            ConnectToWorld(_characterName);
        }

        private void OnDestroy()
        {
            _cts?.Cancel();
            if (_map != null) _map.OnCellClicked -= RequestMove;
        }

        public async void ConnectToWorld(string characterName)
        {
            _characterName = characterName;
            _cts = new CancellationTokenSource();
            _map.OnCellClicked += RequestMove;

            try
            {
                var grpc = GrpcClient.Instance;
                if (grpc?.GameClient == null)
                {
                    Debug.LogError("[GameManager] GrpcClient or GameClient is null.");
                    return;
                }

                using var stream = grpc.GameClient.Connect(
                    new global::Game.ConnectRequest { CharacterName = characterName },
                    grpc.GetAuthMetadata());

                while (await stream.ResponseStream.MoveNext(_cts.Token))
                    ProcessGameEvent(stream.ResponseStream.Current);
            }
            catch (RpcException e) when (e.StatusCode == StatusCode.Cancelled)
            {
                Debug.Log("[GameManager] Disconnected from server.");
            }
            catch (System.Exception e)
            {
                Debug.LogError($"[GameManager] Stream error: {e.Message}");
            }
        }

        private void ProcessGameEvent(global::Game.GameEvent evt)
        {
            switch (evt.EventCase)
            {
                case global::Game.GameEvent.EventOneofCase.MapData:
                    HandleMapData(evt.MapData);
                    break;
                case global::Game.GameEvent.EventOneofCase.PlayerConnected:
                    HandlePlayerConnected(evt.PlayerConnected);
                    break;
                case global::Game.GameEvent.EventOneofCase.PlayerDisconnected:
                    HandlePlayerDisconnected(evt.PlayerDisconnected);
                    break;
                case global::Game.GameEvent.EventOneofCase.PlayerMoved:
                    HandlePlayerMoved(evt.PlayerMoved);
                    break;
                case global::Game.GameEvent.EventOneofCase.ChatMessage:
                    HandleChatMessage(evt.ChatMessage);
                    break;
            }
        }

        private void HandleMapData(global::Game.MapData mapData)
        {
            var terrainTypes = new int[mapData.Cells.Count];
            var walkable = new bool[mapData.Cells.Count];
            for (int i = 0; i < mapData.Cells.Count; i++)
            {
                terrainTypes[i] = mapData.Cells[i].TerrainType;
                walkable[i] = mapData.Cells[i].Walkable;
            }
            _map.LoadMapFromProto(mapData.Width, mapData.Height, terrainTypes, walkable);
        }

        private void HandlePlayerConnected(global::Game.PlayerConnected player)
        {
            if (_players.ContainsKey(player.PlayerId)) return;

            var go = CreatePlayerObject();
            var entity = go.GetComponent<PlayerEntity>();
            if (entity == null) entity = go.AddComponent<PlayerEntity>();
            entity.Initialize(player.PlayerId, player.Username, player.X, player.Y, _map);
            _players[player.PlayerId] = entity;

            if (_localPlayer == null || player.PlayerId == _localPlayerId)
            {
                _localPlayerId = player.PlayerId;
                _localPlayer = entity;
            }
        }

        private void HandlePlayerDisconnected(global::Game.PlayerDisconnected player)
        {
            if (_players.TryGetValue(player.PlayerId, out var entity))
            {
                Destroy(entity.gameObject);
                _players.Remove(player.PlayerId);
            }
        }

        private void HandlePlayerMoved(global::Game.PlayerMoved moved)
        {
            if (!_players.TryGetValue(moved.PlayerId, out var entity)) return;
            if (moved.Path.Count > 0)
                entity.MoveToPath(new List<global::Game.PathNode>(moved.Path), _map);
            else
                entity.MoveTo(moved.X, moved.Y, _map);
        }

        private void HandleChatMessage(global::Game.ChatMessage msg)
        {
            OnChatMessageReceived?.Invoke(msg.Username, msg.Message);
        }

        public async void RequestMove(int targetX, int targetY)
        {
            if (_localPlayer == null) return;
            int seq = ++_moveSeq;
            try
            {
                var response = await GrpcClient.Instance.GameClient.MoveAsync(
                    new global::Game.MoveRequest { TargetX = targetX, TargetY = targetY },
                    GrpcClient.Instance.GetAuthMetadata());
                if (seq != _moveSeq) return;
                if (response.Success)
                {
                    if (response.Path.Count > 0)
                        _localPlayer.MoveToPath(new List<global::Game.PathNode>(response.Path), _map);
                    else
                        _localPlayer.MoveTo(response.X, response.Y, _map);
                }
            }
            catch (System.Exception e)
            {
                Debug.LogError($"Move failed: {e.Message}");
            }
        }

        public async void RequestChat(string message)
        {
            try
            {
                await GrpcClient.Instance.GameClient.ChatAsync(
                    new global::Game.ChatRequest { Message = message },
                    GrpcClient.Instance.GetAuthMetadata());
            }
            catch (System.Exception e)
            {
                Debug.LogError($"Chat failed: {e.Message}");
            }
        }

        private GameObject CreatePlayerObject()
        {
            if (_playerPrefab != null) return Instantiate(_playerPrefab, transform);
            var go = new GameObject("Player");
            go.transform.parent = transform;
            go.AddComponent<SpriteRenderer>();
            go.AddComponent<PlayerEntity>();
            return go;
        }
    }
}
