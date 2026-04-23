using Grpc.Core;
using Grpc.Net.Client;
using Grpc.Net.Client.Web;
using UnityEngine;
using System;
using System.Net.Http;
using System.Threading.Tasks;

namespace MMORPG.Network
{
    public class GrpcClient : MonoBehaviour
    {
        public static GrpcClient Instance { get; private set; }

        private GrpcChannel _channel;
        private global::Auth.AuthServiceClient _authClient;
        private global::Game.GameServiceClient _gameClient;
        private global::Combat.CombatServiceClient _combatClient;

        public string AuthToken { get; private set; }
        public bool IsConnected => _channel != null;

        public event Action<string> OnAuthError;
        public event Action OnLoginSuccess;

        private static readonly TimeSpan _timeout = TimeSpan.FromSeconds(5);

        private void Awake()
        {
            if (Instance != null && Instance != this) { Destroy(gameObject); return; }
            Instance = this;
            DontDestroyOnLoad(gameObject);
        }

        public Task ConnectToServer(string serverAddress = "localhost:50051")
        {
            try
            {
                var address = serverAddress.StartsWith("http") ? serverAddress : $"http://{serverAddress}";
                var httpHandler = new HttpClientHandler();
                var handler = new GrpcWebHandler(GrpcWebMode.GrpcWeb, httpHandler);
                _channel = GrpcChannel.ForAddress(address, new GrpcChannelOptions
                {
                    HttpHandler = handler,
                    MaxReceiveMessageSize = 4 * 1024 * 1024
                });
                var invoker = _channel.CreateCallInvoker();
                _authClient = new global::Auth.AuthServiceClient(invoker);
                _gameClient = new global::Game.GameServiceClient(invoker);
                _combatClient = new global::Combat.CombatServiceClient(invoker);
                Debug.Log($"gRPC channel created for {address}");
            }
            catch (Exception e)
            {
                Debug.LogError($"Failed to create channel: {e.Message}");
                OnAuthError?.Invoke($"Connection failed: {e.Message}");
            }
            return Task.CompletedTask;
        }

        public async Task<bool> Register(string username, string email, string password)
        {
            try
            {
                var response = await _authClient.RegisterAsync(new global::Auth.RegisterRequest
                {
                    Username = username, Email = email, Password = password
                }, deadline: DateTime.UtcNow.Add(_timeout));
                if (response.Success) return true;
                OnAuthError?.Invoke(response.Error);
                return false;
            }
            catch (Exception e)
            {
                Debug.LogError($"Register exception: {e}");
                OnAuthError?.Invoke($"Registration failed: {e.Message}");
                return false;
            }
        }

        public async Task<bool> Login(string username, string password)
        {
            try
            {
                var response = await _authClient.LoginAsync(new global::Auth.LoginRequest
                {
                    Username = username, Password = password
                }, deadline: DateTime.UtcNow.Add(_timeout));
                if (response.Success)
                {
                    AuthToken = response.Token;
                    Debug.Log("Login successful");
                    OnLoginSuccess?.Invoke();
                    return true;
                }
                OnAuthError?.Invoke(response.Error);
                return false;
            }
            catch (Exception e)
            {
                Debug.LogError($"Login exception: {e}");
                OnAuthError?.Invoke($"Login failed: {e.Message}");
                return false;
            }
        }

        public Metadata GetAuthMetadata()
        {
            var metadata = new Metadata();
            if (!string.IsNullOrEmpty(AuthToken))
                metadata.Add("authorization", $"Bearer {AuthToken}");
            return metadata;
        }

        public global::Game.GameServiceClient GameClient => _gameClient;
        public global::Combat.CombatServiceClient CombatClient => _combatClient;

        private async void OnDestroy()
        {
            if (_channel != null) await _channel.ShutdownAsync();
        }
    }
}
