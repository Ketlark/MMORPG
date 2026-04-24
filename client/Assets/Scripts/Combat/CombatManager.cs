using UnityEngine;
using System.Collections.Generic;
using System.Threading;
using Grpc.Core;
using MMORPG.Network;
using MMORPG.Map;

namespace MMORPG.Combat
{
    public class CombatManager : MonoBehaviour
    {
        public static CombatManager Instance { get; private set; }

        [SerializeField] private IsometricMap _combatMap;
        [SerializeField] private GameObject _combatPlayerPrefab;
        [SerializeField] private UnityEngine.UI.Text _apText;
        [SerializeField] private UnityEngine.UI.Text _mpText;
        [SerializeField] private UnityEngine.UI.Text _turnText;
        [SerializeField] private GameObject _combatUI;

        private Dictionary<string, CombatPlayerEntity> _fighters = new();
        private CancellationTokenSource _cts;
        private string _currentTurnPlayerId;
        private bool _isMyTurn;
        private string _currentCombatId;

        public bool InCombat => _combatUI.activeSelf;
        public bool IsMyTurn => _isMyTurn;

        private void Awake()
        {
            if (Instance != null && Instance != this) { Destroy(gameObject); return; }
            Instance = this;
        }

        private void OnDestroy()
        {
            _cts?.Cancel();
            if (Instance == this) Instance = null;
        }

        public async void JoinCombat(string combatId)
        {
            _cts = new CancellationTokenSource();
            _combatUI.SetActive(true);

            try
            {
                var client = GrpcClient.Instance.CombatClient;
                var metadata = GrpcClient.Instance.GetAuthMetadata();

                using var stream = client.JoinCombat(new global::Combat.JoinCombatRequest
                {
                    CombatId = combatId
                }, metadata);

                while (await stream.ResponseStream.MoveNext(_cts.Token))
                    ProcessCombatEvent(stream.ResponseStream.Current);
            }
            catch (RpcException e) when (e.StatusCode == StatusCode.Cancelled)
            {
                Debug.Log("[CombatManager] Left combat.");
            }
            catch (System.Exception e)
            {
                Debug.LogError($"[CombatManager] Stream error: {e.Message}");
            }
        }

        private void ProcessCombatEvent(global::Combat.CombatEvent evt)
        {
            switch (evt.EventCase)
            {
                case global::Combat.CombatEvent.EventOneofCase.CombatStarted:
                    HandleCombatStarted(evt.CombatStarted);
                    break;
                case global::Combat.CombatEvent.EventOneofCase.TurnStarted:
                    HandleTurnStarted(evt.TurnStarted);
                    break;
                case global::Combat.CombatEvent.EventOneofCase.PlayerAction:
                    HandlePlayerAction(evt.PlayerAction);
                    break;
                case global::Combat.CombatEvent.EventOneofCase.TurnEnded:
                    HandleTurnEnded(evt.TurnEnded);
                    break;
                case global::Combat.CombatEvent.EventOneofCase.CombatEnded:
                    HandleCombatEnded(evt.CombatEnded);
                    break;
            }
        }

        private void HandleCombatStarted(global::Combat.CombatStarted started)
        {
            _currentCombatId = started.CombatId;
            _fighters.Clear();
            foreach (var f in started.Fighters)
            {
                var go = Instantiate(_combatPlayerPrefab, _combatMap.transform);
                var entity = go.GetComponent<CombatPlayerEntity>();
                entity.Initialize(f.PlayerId, f.Username, f.X, f.Y, f.Health, f.MaxHealth, _combatMap);
                _fighters[f.PlayerId] = entity;
            }

            var terrain = new int[started.Cells.Count];
            var walkable = new bool[started.Cells.Count];
            for (int i = 0; i < started.Cells.Count; i++)
            {
                terrain[i] = 0;
                walkable[i] = started.Cells[i].Walkable;
            }
            var combatWidth = started.Width > 0 ? (int)started.Width : 14;
            var combatHeight = started.Height > 0 ? (int)started.Height : 18;
            _combatMap.LoadMapFromProto(combatWidth, combatHeight, terrain, walkable);
        }

        private void HandleTurnStarted(global::Combat.TurnStarted turn)
        {
            _currentTurnPlayerId = turn.PlayerId;
            _isMyTurn = turn.PlayerId == Game.GameManager.Instance.LocalPlayer?.PlayerId;

            if (_isMyTurn)
            {
                _apText.text = $"PA: {turn.ActionPoints}";
                _mpText.text = $"PM: {turn.MovementPoints}";
                _turnText.text = "Votre tour !";
            }
            else
            {
                _turnText.text = "Tour de l'adversaire";
            }
        }

        private void HandlePlayerAction(global::Combat.PlayerAction action)
        {
            if (!_fighters.TryGetValue(action.PlayerId, out var entity)) return;

            switch (action.ActionResultCase)
            {
                case global::Combat.PlayerAction.ActionResultOneofCase.Move:
                    entity.MoveTo(action.Move.ToX, action.Move.ToY, _combatMap);
                    break;
                case global::Combat.PlayerAction.ActionResultOneofCase.Spell:
                    foreach (var dmg in action.Spell.Damages)
                        if (_fighters.TryGetValue(dmg.TargetId, out var target))
                            target.TakeDamage(dmg.Damage, dmg.HealthRemaining);
                    break;
            }
        }

        private void HandleTurnEnded(global::Combat.TurnEnded turn) => _isMyTurn = false;

        private void HandleCombatEnded(global::Combat.CombatEnded ended)
        {
            _combatUI.SetActive(false);
            _fighters.Clear();
        }

        public async void PerformMove(int targetX, int targetY)
        {
            if (!_isMyTurn) return;
            try
            {
                var result = await GrpcClient.Instance.CombatClient.PerformActionAsync(
                    new global::Combat.CombatAction
                    {
                        CombatId = _currentCombatId,
                        Move = new global::Combat.MoveAction { TargetX = targetX, TargetY = targetY }
                    }, GrpcClient.Instance.GetAuthMetadata());
                if (result.Success)
                {
                    _apText.text = $"PA: {result.ActionPointsRemaining}";
                    _mpText.text = $"PM: {result.MovementPointsRemaining}";
                }
            }
            catch (System.Exception e) { Debug.LogError($"Combat action failed: {e.Message}"); }
        }

        public async void PerformSpell(string spellId, int targetX, int targetY)
        {
            if (!_isMyTurn) return;
            try
            {
                var result = await GrpcClient.Instance.CombatClient.PerformActionAsync(
                    new global::Combat.CombatAction
                    {
                        CombatId = _currentCombatId,
                        Spell = new global::Combat.SpellAction { SpellId = spellId, TargetX = targetX, TargetY = targetY }
                    }, GrpcClient.Instance.GetAuthMetadata());
                if (result.Success)
                {
                    _apText.text = $"PA: {result.ActionPointsRemaining}";
                    _mpText.text = $"PM: {result.MovementPointsRemaining}";
                }
            }
            catch (System.Exception e) { Debug.LogError($"Spell action failed: {e.Message}"); }
        }

        public async void EndTurn()
        {
            try
            {
                await GrpcClient.Instance.CombatClient.EndTurnAsync(
                    new global::Combat.EndTurnRequest { CombatId = _currentCombatId },
                    GrpcClient.Instance.GetAuthMetadata());
            }
            catch (System.Exception e) { Debug.LogError($"End turn failed: {e.Message}"); }
        }
    }
}
