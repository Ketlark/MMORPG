using System;
using System.Threading;
using System.Threading.Tasks;
using Grpc.Core;

namespace Auth
{
    public static class AuthServiceMethods
    {
        public static readonly string ServiceName = "auth.AuthService";
        public static readonly Method<RegisterRequest, RegisterResponse> RegisterMethod =
            new Method<RegisterRequest, RegisterResponse>(
                MethodType.Unary, ServiceName, "Register",
                ProtoSerializers.MakeMarshaller<RegisterRequest>(),
                ProtoSerializers.MakeMarshaller<RegisterResponse>());
        public static readonly Method<LoginRequest, LoginResponse> LoginMethod =
            new Method<LoginRequest, LoginResponse>(
                MethodType.Unary, ServiceName, "Login",
                ProtoSerializers.MakeMarshaller<LoginRequest>(),
                ProtoSerializers.MakeMarshaller<LoginResponse>());
    }

    public class AuthServiceClient
    {
        private readonly CallInvoker _invoker;
        public AuthServiceClient(CallInvoker invoker) { _invoker = invoker; }
        public AsyncUnaryCall<RegisterResponse> RegisterAsync(RegisterRequest request, Metadata headers = null, DateTime? deadline = null, CancellationToken cancellationToken = default)
            => _invoker.AsyncUnaryCall(AuthServiceMethods.RegisterMethod, null, new CallOptions(headers, deadline, cancellationToken), request);
        public AsyncUnaryCall<LoginResponse> LoginAsync(LoginRequest request, Metadata headers = null, DateTime? deadline = null, CancellationToken cancellationToken = default)
            => _invoker.AsyncUnaryCall(AuthServiceMethods.LoginMethod, null, new CallOptions(headers, deadline, cancellationToken), request);
    }
}

namespace Game
{
    public static class GameServiceMethods
    {
        public static readonly string ServiceName = "game.GameService";
        public static readonly Method<ConnectRequest, GameEvent> ConnectMethod =
            new Method<ConnectRequest, GameEvent>(
                MethodType.ServerStreaming, ServiceName, "Connect",
                ProtoSerializers.MakeMarshaller<ConnectRequest>(),
                ProtoSerializers.MakeMarshaller<GameEvent>());
        public static readonly Method<MoveRequest, MoveResponse> MoveMethod =
            new Method<MoveRequest, MoveResponse>(
                MethodType.Unary, ServiceName, "Move",
                ProtoSerializers.MakeMarshaller<MoveRequest>(),
                ProtoSerializers.MakeMarshaller<MoveResponse>());
        public static readonly Method<ChatRequest, Google.Protobuf.WellKnownTypes.Empty> ChatMethod =
            new Method<ChatRequest, Google.Protobuf.WellKnownTypes.Empty>(
                MethodType.Unary, ServiceName, "Chat",
                ProtoSerializers.MakeMarshaller<ChatRequest>(),
                ProtoSerializers.MakeMarshaller<Google.Protobuf.WellKnownTypes.Empty>());
    }

    public class GameServiceClient
    {
        private readonly CallInvoker _invoker;
        public GameServiceClient(CallInvoker invoker) { _invoker = invoker; }
        public AsyncServerStreamingCall<GameEvent> Connect(ConnectRequest request, Metadata headers = null, DateTime? deadline = null, CancellationToken cancellationToken = default)
            => _invoker.AsyncServerStreamingCall(GameServiceMethods.ConnectMethod, null, new CallOptions(headers, deadline, cancellationToken), request);
        public AsyncUnaryCall<MoveResponse> MoveAsync(MoveRequest request, Metadata headers = null, DateTime? deadline = null, CancellationToken cancellationToken = default)
            => _invoker.AsyncUnaryCall(GameServiceMethods.MoveMethod, null, new CallOptions(headers, deadline, cancellationToken), request);
        public AsyncUnaryCall<Google.Protobuf.WellKnownTypes.Empty> ChatAsync(ChatRequest request, Metadata headers = null, DateTime? deadline = null, CancellationToken cancellationToken = default)
            => _invoker.AsyncUnaryCall(GameServiceMethods.ChatMethod, null, new CallOptions(headers, deadline, cancellationToken), request);
    }
}

namespace Combat
{
    public static class CombatServiceMethods
    {
        public static readonly string ServiceName = "combat.CombatService";
        public static readonly Method<JoinCombatRequest, CombatEvent> JoinCombatMethod =
            new Method<JoinCombatRequest, CombatEvent>(
                MethodType.ServerStreaming, ServiceName, "JoinCombat",
                ProtoSerializers.MakeMarshaller<JoinCombatRequest>(),
                ProtoSerializers.MakeMarshaller<CombatEvent>());
        public static readonly Method<CombatAction, ActionResult> PerformActionMethod =
            new Method<CombatAction, ActionResult>(
                MethodType.Unary, ServiceName, "PerformAction",
                ProtoSerializers.MakeMarshaller<CombatAction>(),
                ProtoSerializers.MakeMarshaller<ActionResult>());
        public static readonly Method<EndTurnRequest, EndTurnResponse> EndTurnMethod =
            new Method<EndTurnRequest, EndTurnResponse>(
                MethodType.Unary, ServiceName, "EndTurn",
                ProtoSerializers.MakeMarshaller<EndTurnRequest>(),
                ProtoSerializers.MakeMarshaller<EndTurnResponse>());
    }

    public class CombatServiceClient
    {
        private readonly CallInvoker _invoker;
        public CombatServiceClient(CallInvoker invoker) { _invoker = invoker; }
        public AsyncServerStreamingCall<CombatEvent> JoinCombat(JoinCombatRequest request, Metadata headers = null, DateTime? deadline = null, CancellationToken cancellationToken = default)
            => _invoker.AsyncServerStreamingCall(CombatServiceMethods.JoinCombatMethod, null, new CallOptions(headers, deadline, cancellationToken), request);
        public AsyncUnaryCall<ActionResult> PerformActionAsync(CombatAction request, Metadata headers = null, DateTime? deadline = null, CancellationToken cancellationToken = default)
            => _invoker.AsyncUnaryCall(CombatServiceMethods.PerformActionMethod, null, new CallOptions(headers, deadline, cancellationToken), request);
        public AsyncUnaryCall<EndTurnResponse> EndTurnAsync(EndTurnRequest request, Metadata headers = null, DateTime? deadline = null, CancellationToken cancellationToken = default)
            => _invoker.AsyncUnaryCall(CombatServiceMethods.EndTurnMethod, null, new CallOptions(headers, deadline, cancellationToken), request);
    }
}

internal static class ProtoSerializers
{
    public static Marshaller<T> MakeMarshaller<T>() where T : class, Google.Protobuf.IMessage<T>, new()
    {
        return new Marshaller<T>(
            msg => Google.Protobuf.MessageExtensions.ToByteArray(msg),
            data => { var parser = new Google.Protobuf.MessageParser<T>(() => new T()); return parser.ParseFrom(data); });
    }
}
