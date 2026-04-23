using UnityEngine;
using UnityEngine.UI;
using MMORPG.Game;

namespace MMORPG.UI
{
    public class GameHUD : MonoBehaviour
    {
        [Header("Player Info")]
        [SerializeField] private Text _playerNameText;
        [SerializeField] private Slider _healthBar;
        [SerializeField] private Text _healthText;
        [SerializeField] private Text _apText;
        [SerializeField] private Text _mpText;

        [Header("Combat")]
        [SerializeField] private Button _endTurnButton;
        [SerializeField] private GameObject _combatPanel;

        [Header("Chat")]
        [SerializeField] private InputField _chatInput;
        [SerializeField] private Text _chatLog;
        [SerializeField] private Button _chatSendButton;

        private void Awake()
        {
            EventSystemUtility.EnsureExists();
        }

        private void OnEnable()
        {
            GameManager.OnChatMessageReceived += AddChatMessage;
        }

        private void OnDisable()
        {
            GameManager.OnChatMessageReceived -= AddChatMessage;
        }

        private void Start()
        {
            _endTurnButton.onClick.AddListener(OnEndTurn);
            _chatSendButton.onClick.AddListener(OnChatSend);
            _chatInput.onSubmit.AddListener(_ => OnChatSend());
            _combatPanel.SetActive(false);
        }

        public void UpdatePlayerHUD(string name, int health, int maxHealth, int ap, int mp)
        {
            _playerNameText.text = name;
            _healthBar.maxValue = maxHealth;
            _healthBar.value = health;
            _healthText.text = $"{health}/{maxHealth}";
            _apText.text = $"PA: {ap}";
            _mpText.text = $"PM: {mp}";
        }

        public void ShowCombatUI(bool show) => _combatPanel.SetActive(show);

        public void AddChatMessage(string username, string message)
        {
            _chatLog.text += $"[{username}] {message}\n";
        }

        private void OnEndTurn() => Combat.CombatManager.Instance?.EndTurn();

        private void OnChatSend()
        {
            if (string.IsNullOrWhiteSpace(_chatInput.text)) return;
            var text = _chatInput.text;
            _chatInput.text = "";
            GameManager.Instance?.RequestChat(text);
        }

    }
}
