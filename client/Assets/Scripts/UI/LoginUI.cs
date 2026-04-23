using UnityEngine;
using UnityEngine.UI;
using UnityEngine.SceneManagement;
using System.Threading.Tasks;
using MMORPG.Network;

namespace MMORPG.UI
{
    public class LoginUI : MonoBehaviour
    {
        [Header("Login Panel")]
        [SerializeField] private GameObject _loginPanel;
        [SerializeField] private InputField _usernameInput;
        [SerializeField] private InputField _passwordInput;
        [SerializeField] private Button _loginButton;
        [SerializeField] private Button _switchToRegisterButton;

        [Header("Register Panel")]
        [SerializeField] private GameObject _registerPanel;
        [SerializeField] private InputField _regUsernameInput;
        [SerializeField] private InputField _regEmailInput;
        [SerializeField] private InputField _regPasswordInput;
        [SerializeField] private Button _registerButton;
        [SerializeField] private Button _switchToLoginButton;

        [Header("Status")]
        [SerializeField] private Text _statusText;
        [SerializeField] private string _serverAddress = "localhost:50051";

        private void Awake()
        {
            EventSystemUtility.EnsureExists();
            ValidateReferences();
        }

        private async void Start()
        {
            _loginPanel.SetActive(true);
            _registerPanel.SetActive(false);
            _statusText.text = "Connecting to server...";

            _loginButton.onClick.AddListener(OnLoginClicked);
            _registerButton.onClick.AddListener(OnRegisterClicked);
            _switchToRegisterButton.onClick.AddListener(() =>
            {
                _loginPanel.SetActive(false);
                _registerPanel.SetActive(true);
            });
            _switchToLoginButton.onClick.AddListener(() =>
            {
                _registerPanel.SetActive(false);
                _loginPanel.SetActive(true);
            });

            if (GrpcClient.Instance == null)
            {
                _statusText.text = "Error: GrpcClient not found.";
                return;
            }
            GrpcClient.Instance.OnAuthError += OnAuthError;
            await GrpcClient.Instance.ConnectToServer(_serverAddress);
            _statusText.text = "Connected. Please login.";
        }

        private async void OnLoginClicked()
        {
            _statusText.text = "Logging in...";
            _loginButton.interactable = false;

            bool success = await GrpcClient.Instance.Login(_usernameInput.text, _passwordInput.text);
            _loginButton.interactable = true;

            if (success)
            {
                _statusText.text = "Login successful!";
                PlayerPrefs.SetString("CharacterName", _usernameInput.text);
                SceneManager.LoadScene("GameScene");
            }
        }

        private async void OnRegisterClicked()
        {
            _statusText.text = "Registering...";
            _registerButton.interactable = false;

            bool success = await GrpcClient.Instance.Register(
                _regUsernameInput.text, _regEmailInput.text, _regPasswordInput.text);
            _registerButton.interactable = true;

            if (success)
            {
                _statusText.text = "Registration successful! Please login.";
                _registerPanel.SetActive(false);
                _loginPanel.SetActive(true);
            }
        }

        private void OnAuthError(string error) => _statusText.text = error;

        private void OnDestroy()
        {
            if (GrpcClient.Instance != null)
                GrpcClient.Instance.OnAuthError -= OnAuthError;
        }

        private void ValidateReferences()
        {
            if (_loginPanel == null) Debug.LogError("[LoginUI] LoginPanel not assigned.", this);
            if (_registerPanel == null) Debug.LogError("[LoginUI] RegisterPanel not assigned.", this);
            if (_usernameInput == null) Debug.LogError("[LoginUI] UsernameInput not assigned.", this);
            if (_passwordInput == null) Debug.LogError("[LoginUI] PasswordInput not assigned.", this);
            if (_loginButton == null) Debug.LogError("[LoginUI] LoginButton not assigned.", this);
            if (_statusText == null) Debug.LogError("[LoginUI] StatusText not assigned.", this);
        }
    }
}
