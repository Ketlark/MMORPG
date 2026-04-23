using UnityEngine;
using UnityEditor;
using UnityEditor.SceneManagement;
using UnityEngine.UI;
using UnityEngine.Tilemaps;

public static class SceneSetup
{
    [MenuItem("MMORPG/Setup Login Scene")]
    public static void SetupLoginScene()
    {
        // Save current scene if any
        if (EditorSceneManager.GetActiveScene().isDirty)
            EditorSceneManager.SaveCurrentModifiedScenesIfUserWantsTo();

        var scene = EditorSceneManager.NewScene(NewSceneSetup.EmptyScene, NewSceneMode.Single);

        BuildLoginScene(scene);

        EditorSceneManager.SaveScene(scene, "Assets/Scenes/LoginScene.unity");
        Debug.Log("[MMORPG] LoginScene created!");
    }

    [MenuItem("MMORPG/Setup Game Scene")]
    public static void SetupGameScene()
    {
        if (EditorSceneManager.GetActiveScene().isDirty)
            EditorSceneManager.SaveCurrentModifiedScenesIfUserWantsTo();

        var scene = EditorSceneManager.NewScene(NewSceneSetup.EmptyScene, NewSceneMode.Single);

        BuildGameScene(scene);

        EditorSceneManager.SaveScene(scene, "Assets/Scenes/GameScene.unity");
        Debug.Log("[MMORPG] GameScene created!");
    }

    static void BuildLoginScene(UnityEngine.SceneManagement.Scene scene)
    {
        var camObj = new GameObject("Main Camera");
        camObj.transform.position = new Vector3(0, 0, -10);
        var cam = camObj.AddComponent<Camera>();
        cam.orthographic = true;
        cam.orthographicSize = 5;
        cam.clearFlags = CameraClearFlags.SolidColor;
        cam.backgroundColor = new Color(0.1f, 0.1f, 0.15f);
        camObj.tag = "MainCamera";

        var canvasObj = new GameObject("Canvas");
        var canvas = canvasObj.AddComponent<Canvas>();
        canvas.renderMode = RenderMode.ScreenSpaceOverlay;
        canvasObj.AddComponent<CanvasScaler>();
        canvasObj.AddComponent<GraphicRaycaster>();

        var eventSysObj = new GameObject("EventSystem");
        eventSysObj.AddComponent<UnityEngine.EventSystems.EventSystem>();
        eventSysObj.AddComponent<UnityEngine.InputSystem.UI.InputSystemUIInputModule>();

        var grpcObj = new GameObject("GrpcClient");
        grpcObj.AddComponent<global::MMORPG.Network.GrpcClient>();

        // Login Panel
        var loginPanel = CreatePanel(canvasObj.transform, "LoginPanel", new Color(0.12f, 0.12f, 0.18f, 0.95f));
        CreateText(loginPanel.transform, "Title", "MMORPG - Login", 32, new Color(1f, 0.85f, 0.3f), new Vector2(0, 120));
        var usernameInput = CreateInputField(loginPanel.transform, "UsernameInput", "Username", new Vector2(0, 40));
        var passwordInput = CreateInputField(loginPanel.transform, "PasswordInput", "Password", new Vector2(0, -10));
        passwordInput.inputType = InputField.InputType.Password;
        var loginBtn = CreateButton(loginPanel.transform, "LoginButton", "Login", new Vector2(0, -70));
        var switchRegBtn = CreateButton(loginPanel.transform, "SwitchToRegister", "Create Account", new Vector2(0, -120), new Color(0.3f, 0.3f, 0.5f));

        // Register Panel
        var registerPanel = CreatePanel(canvasObj.transform, "RegisterPanel", new Color(0.12f, 0.12f, 0.18f, 0.95f));
        registerPanel.SetActive(false);
        CreateText(registerPanel.transform, "Title", "MMORPG - Register", 32, new Color(1f, 0.85f, 0.3f), new Vector2(0, 140));
        var regUsername = CreateInputField(registerPanel.transform, "RegUsernameInput", "Username", new Vector2(0, 60));
        var regEmail = CreateInputField(registerPanel.transform, "RegEmailInput", "Email", new Vector2(0, 10));
        var regPassword = CreateInputField(registerPanel.transform, "RegPasswordInput", "Password", new Vector2(0, -40));
        regPassword.inputType = InputField.InputType.Password;
        var registerBtn = CreateButton(registerPanel.transform, "RegisterButton", "Register", new Vector2(0, -100));
        var switchLoginBtn = CreateButton(registerPanel.transform, "SwitchToLogin", "Back to Login", new Vector2(0, -150), new Color(0.3f, 0.3f, 0.5f));

        var statusText = CreateText(canvasObj.transform, "StatusText", "", 18, Color.white, new Vector2(0, -250));

        var loginUIObj = new GameObject("LoginUI");
        loginUIObj.transform.SetParent(canvasObj.transform);
        var loginUI = loginUIObj.AddComponent<global::MMORPG.UI.LoginUI>();

        var so = new SerializedObject(loginUI);
        so.FindProperty("_loginPanel").objectReferenceValue = loginPanel;
        so.FindProperty("_usernameInput").objectReferenceValue = usernameInput;
        so.FindProperty("_passwordInput").objectReferenceValue = passwordInput;
        so.FindProperty("_loginButton").objectReferenceValue = loginBtn.GetComponent<Button>();
        so.FindProperty("_switchToRegisterButton").objectReferenceValue = switchRegBtn.GetComponent<Button>();
        so.FindProperty("_registerPanel").objectReferenceValue = registerPanel;
        so.FindProperty("_regUsernameInput").objectReferenceValue = regUsername;
        so.FindProperty("_regEmailInput").objectReferenceValue = regEmail;
        so.FindProperty("_regPasswordInput").objectReferenceValue = regPassword;
        so.FindProperty("_registerButton").objectReferenceValue = registerBtn.GetComponent<Button>();
        so.FindProperty("_switchToLoginButton").objectReferenceValue = switchLoginBtn.GetComponent<Button>();
        so.FindProperty("_statusText").objectReferenceValue = statusText;
        so.ApplyModifiedProperties();
    }

    static void BuildGameScene(UnityEngine.SceneManagement.Scene scene)
    {
        // Camera
        var camObj = new GameObject("Main Camera");
        camObj.transform.position = new Vector3(0, 0, -10);
        var cam = camObj.AddComponent<Camera>();
        cam.orthographic = true;
        cam.orthographicSize = 6;
        cam.clearFlags = CameraClearFlags.SolidColor;
        cam.backgroundColor = new Color(0.15f, 0.2f, 0.15f);
        camObj.tag = "MainCamera";
        camObj.tag = "MainCamera";

        var eventSysObj = new GameObject("EventSystem");
        eventSysObj.AddComponent<UnityEngine.EventSystems.EventSystem>();
        eventSysObj.AddComponent<UnityEngine.InputSystem.UI.InputSystemUIInputModule>();

        // Grid + Tilemaps
        var gridObj = new GameObject("Grid");
        var grid = gridObj.AddComponent<Grid>();
        grid.cellLayout = GridLayout.CellLayout.Isometric;
        grid.cellSize = new Vector3(1f, 0.5f, 1f);

        var groundObj = new GameObject("Ground");
        groundObj.transform.SetParent(gridObj.transform);
        var groundTm = groundObj.AddComponent<Tilemap>();
        groundObj.AddComponent<TilemapRenderer>();

        var overlayObj = new GameObject("Overlay");
        overlayObj.transform.SetParent(gridObj.transform);
        var overlayTm = overlayObj.AddComponent<Tilemap>();
        var overlayRend = overlayObj.AddComponent<TilemapRenderer>();
        overlayRend.sortingOrder = 1;

        // Create tiles as assets
        var grassTile = MakeColorTile(new Color(0.4f, 0.7f, 0.3f), "GrassTile");
        var waterTile = MakeColorTile(new Color(0.2f, 0.4f, 0.8f), "WaterTile");
        var wallTile = MakeColorTile(new Color(0.5f, 0.4f, 0.3f), "WallTile");
        var roadTile = MakeColorTile(new Color(0.7f, 0.65f, 0.5f), "RoadTile");
        var highlightTile = MakeColorTile(new Color(1f, 1f, 0.5f, 0.5f), "HighlightTile");

        // Player prefab
        var playerObj = new GameObject("Player");
        var sr = playerObj.AddComponent<SpriteRenderer>();
        sr.sprite = MakeSprite(new Color(0.2f, 0.6f, 1f), 32);
        playerObj.AddComponent<global::MMORPG.Game.PlayerEntity>();
        PrefabUtility.SaveAsPrefabAsset(playerObj, "Assets/Prefabs/Player.prefab");
        Object.DestroyImmediate(playerObj);

        var playerPrefab = AssetDatabase.LoadAssetAtPath<GameObject>("Assets/Prefabs/Player.prefab");

        // IsometricMap
        var mapObj = new GameObject("IsometricMap");
        var isoMap = mapObj.AddComponent<global::MMORPG.Map.IsometricMap>();
        var soMap = new SerializedObject(isoMap);
        soMap.FindProperty("_groundTilemap").objectReferenceValue = groundTm;
        soMap.FindProperty("_overlayTilemap").objectReferenceValue = overlayTm;
        soMap.FindProperty("_grassTile").objectReferenceValue = grassTile;
        soMap.FindProperty("_waterTile").objectReferenceValue = waterTile;
        soMap.FindProperty("_wallTile").objectReferenceValue = wallTile;
        soMap.FindProperty("_roadTile").objectReferenceValue = roadTile;
        soMap.FindProperty("_highlightTile").objectReferenceValue = highlightTile;
        soMap.ApplyModifiedProperties();

        // GameManager
        var gmObj = new GameObject("GameManager");
        var gm = gmObj.AddComponent<global::MMORPG.Game.GameManager>();
        var soGm = new SerializedObject(gm);
        soGm.FindProperty("_map").objectReferenceValue = isoMap;
        soGm.FindProperty("_playerPrefab").objectReferenceValue = playerPrefab;
        soGm.ApplyModifiedProperties();

        // HUD Canvas
        var canvasObj = new GameObject("Canvas");
        var canvas = canvasObj.AddComponent<Canvas>();
        canvas.renderMode = RenderMode.ScreenSpaceOverlay;
        canvasObj.AddComponent<CanvasScaler>();
        canvasObj.AddComponent<GraphicRaycaster>();

        // Player info panel (top-left)
        var infoPanel = MakeUIPanel(canvasObj.transform, "PlayerInfo", new Vector2(-250, 230), new Vector2(180, 130), new Color(0, 0, 0, 0.7f));
        var playerName = CreateText(infoPanel.transform, "PlayerName", "Player", 18, Color.white, new Vector2(0, 45));
        var healthBar = MakeSlider(infoPanel.transform, "HealthBar", new Vector2(0, 15));
        var healthText = CreateText(infoPanel.transform, "HealthText", "100/100", 14, Color.white, new Vector2(0, -10));
        var apText = CreateText(infoPanel.transform, "APText", "PA: 6", 16, new Color(0.3f, 0.7f, 1f), new Vector2(-35, -38));
        var mpText = CreateText(infoPanel.transform, "MPText", "PM: 3", 16, new Color(0.3f, 1f, 0.5f), new Vector2(35, -38));

        // Combat panel (bottom center, hidden)
        var combatPanel = MakeUIPanel(canvasObj.transform, "CombatPanel", new Vector2(0, -220), new Vector2(300, 80), new Color(0.4f, 0.1f, 0.1f, 0.8f));
        combatPanel.SetActive(false);
        var turnText = CreateText(combatPanel.transform, "TurnText", "", 20, Color.white, new Vector2(0, 15));
        var endTurnBtn = CreateButton(combatPanel.transform, "EndTurnButton", "End Turn", new Vector2(0, -15));

        // Chat panel (bottom-left)
        var chatPanel = MakeUIPanel(canvasObj.transform, "ChatPanel", new Vector2(-250, -200), new Vector2(180, 150), new Color(0, 0, 0, 0.7f));
        var chatLog = CreateText(chatPanel.transform, "ChatLog", "", 12, Color.white, new Vector2(0, 25));
        chatLog.rectTransform.sizeDelta = new Vector2(160, 100);
        var chatInput = CreateInputField(chatPanel.transform, "ChatInput", "Message...", new Vector2(-15, -50), 120);
        var chatSendBtn = CreateButton(chatPanel.transform, "SendBtn", ">", new Vector2(70, -50), new Color(0.3f, 0.5f, 0.3f));
        chatSendBtn.GetComponent<RectTransform>().sizeDelta = new Vector2(35, 30);

        // GameHUD
        var hudObj = new GameObject("GameHUD");
        hudObj.transform.SetParent(canvasObj.transform);
        var hud = hudObj.AddComponent<global::MMORPG.UI.GameHUD>();
        var soHud = new SerializedObject(hud);
        soHud.FindProperty("_playerNameText").objectReferenceValue = playerName;
        soHud.FindProperty("_healthBar").objectReferenceValue = healthBar;
        soHud.FindProperty("_healthText").objectReferenceValue = healthText;
        soHud.FindProperty("_apText").objectReferenceValue = apText;
        soHud.FindProperty("_mpText").objectReferenceValue = mpText;
        soHud.FindProperty("_endTurnButton").objectReferenceValue = endTurnBtn.GetComponent<Button>();
        soHud.FindProperty("_combatPanel").objectReferenceValue = combatPanel;
        soHud.FindProperty("_chatInput").objectReferenceValue = chatInput;
        soHud.FindProperty("_chatLog").objectReferenceValue = chatLog;
        soHud.FindProperty("_chatSendButton").objectReferenceValue = chatSendBtn.GetComponent<Button>();
        soHud.ApplyModifiedProperties();

        AssetDatabase.SaveAssets();
    }

    // --- UI Helpers ---

    static GameObject CreatePanel(Transform parent, string name, Color color)
    {
        var obj = new GameObject(name);
        obj.transform.SetParent(parent, false);
        var img = obj.AddComponent<Image>();
        img.color = color;
        var rt = obj.GetComponent<RectTransform>();
        rt.anchorMin = Vector2.zero;
        rt.anchorMax = Vector2.one;
        rt.sizeDelta = Vector2.zero;
        return obj;
    }

    static GameObject MakeUIPanel(Transform parent, string name, Vector2 pos, Vector2 size, Color color)
    {
        var obj = new GameObject(name);
        obj.transform.SetParent(parent, false);
        var img = obj.AddComponent<Image>();
        img.color = color;
        var rt = obj.GetComponent<RectTransform>();
        rt.anchorMin = new Vector2(0.5f, 0.5f);
        rt.anchorMax = new Vector2(0.5f, 0.5f);
        rt.sizeDelta = size;
        rt.anchoredPosition = pos;
        return obj;
    }

    static Text CreateText(Transform parent, string name, string text, int size, Color color, Vector2 pos)
    {
        var obj = new GameObject(name);
        obj.transform.SetParent(parent, false);
        var txt = obj.AddComponent<Text>();
        txt.text = text;
        txt.font = Resources.GetBuiltinResource<Font>("LegacyRuntime.ttf");
        txt.fontSize = size;
        txt.color = color;
        txt.alignment = TextAnchor.MiddleCenter;
        var rt = obj.GetComponent<RectTransform>();
        rt.anchorMin = new Vector2(0.5f, 0.5f);
        rt.anchorMax = new Vector2(0.5f, 0.5f);
        rt.sizeDelta = new Vector2(250, 40);
        rt.anchoredPosition = pos;
        return txt;
    }

    static InputField CreateInputField(Transform parent, string name, string placeholder, Vector2 pos, float width = 200)
    {
        var obj = new GameObject(name);
        obj.transform.SetParent(parent, false);
        obj.AddComponent<Image>().color = new Color(0.9f, 0.9f, 0.9f);
        var rt = obj.GetComponent<RectTransform>();
        rt.anchorMin = new Vector2(0.5f, 0.5f);
        rt.anchorMax = new Vector2(0.5f, 0.5f);
        rt.sizeDelta = new Vector2(width, 35);
        rt.anchoredPosition = pos;

        var input = obj.AddComponent<InputField>();

        var textObj = new GameObject("Text");
        textObj.transform.SetParent(obj.transform, false);
        var t = textObj.AddComponent<Text>();
        t.font = Resources.GetBuiltinResource<Font>("LegacyRuntime.ttf");
        t.fontSize = 16; t.color = Color.black;
        var trt = textObj.GetComponent<RectTransform>();
        trt.anchorMin = Vector2.zero; trt.anchorMax = Vector2.one;
        trt.offsetMin = new Vector2(5, 0); trt.offsetMax = new Vector2(-5, 0);

        var phObj = new GameObject("Placeholder");
        phObj.transform.SetParent(obj.transform, false);
        var p = phObj.AddComponent<Text>();
        p.font = Resources.GetBuiltinResource<Font>("LegacyRuntime.ttf");
        p.fontSize = 14; p.color = new Color(0.5f, 0.5f, 0.5f); p.text = placeholder;
        var prrt = phObj.GetComponent<RectTransform>();
        prrt.anchorMin = Vector2.zero; prrt.anchorMax = Vector2.one;
        prrt.offsetMin = new Vector2(5, 0); prrt.offsetMax = new Vector2(-5, 0);

        input.textComponent = t;
        input.placeholder = p;
        return input;
    }

    static GameObject CreateButton(Transform parent, string name, string label, Vector2 pos, Color? bg = null)
    {
        var obj = new GameObject(name);
        obj.transform.SetParent(parent, false);
        obj.AddComponent<Image>().color = bg ?? new Color(0.3f, 0.4f, 0.7f);
        obj.AddComponent<Button>();
        var tObj = new GameObject("Text");
        tObj.transform.SetParent(obj.transform, false);
        var t = tObj.AddComponent<Text>();
        t.text = label; t.font = Resources.GetBuiltinResource<Font>("LegacyRuntime.ttf");
        t.fontSize = 18; t.color = Color.white; t.alignment = TextAnchor.MiddleCenter;
        var trt = tObj.GetComponent<RectTransform>();
        trt.anchorMin = Vector2.zero; trt.anchorMax = Vector2.one; trt.sizeDelta = Vector2.zero;
        var rt = obj.GetComponent<RectTransform>();
        rt.anchorMin = new Vector2(0.5f, 0.5f); rt.anchorMax = new Vector2(0.5f, 0.5f);
        rt.sizeDelta = new Vector2(200, 40); rt.anchoredPosition = pos;
        return obj;
    }

    static Slider MakeSlider(Transform parent, string name, Vector2 pos)
    {
        var obj = new GameObject(name);
        obj.transform.SetParent(parent, false);
        var slider = obj.AddComponent<Slider>();

        var bg = new GameObject("BG");
        bg.transform.SetParent(obj.transform, false);
        var bgImg = bg.AddComponent<Image>();
        bgImg.color = new Color(0.3f, 0.3f, 0.3f);
        var bgRt = bg.GetComponent<RectTransform>();
        bgRt.anchorMin = Vector2.zero; bgRt.anchorMax = Vector2.one; bgRt.sizeDelta = Vector2.zero;

        var fillArea = new GameObject("Fill Area");
        fillArea.transform.SetParent(obj.transform, false);
        var fll = new GameObject("Fill");
        fll.transform.SetParent(fillArea.transform, false);
        var fllImg = fll.AddComponent<Image>();
        fllImg.color = Color.green;
        var fllRt = fll.GetComponent<RectTransform>();
        fllRt.anchorMin = Vector2.zero; fllRt.anchorMax = new Vector2(1, 1); fllRt.sizeDelta = Vector2.zero;

        slider.targetGraphic = bgImg;
        slider.fillRect = fllRt;
        slider.value = 1;

        var rt = obj.GetComponent<RectTransform>();
        rt.anchorMin = new Vector2(0.5f, 0.5f); rt.anchorMax = new Vector2(0.5f, 0.5f);
        rt.sizeDelta = new Vector2(150, 15); rt.anchoredPosition = pos;
        return slider;
    }

    static Tile MakeColorTile(Color color, string name)
    {
        var sprite = MakeSprite(color, 64);
        var tile = ScriptableObject.CreateInstance<Tile>();
        tile.sprite = sprite;
        tile.name = name;
        var path = $"Assets/Sprites/{name}.asset";
        AssetDatabase.DeleteAsset(path);
        AssetDatabase.CreateAsset(tile, path);
        return tile;
    }

    static Sprite MakeSprite(Color color, int size)
    {
        var tex = new Texture2D(size, size);
        var px = new Color[size * size];
        for (int i = 0; i < px.Length; i++) px[i] = color;
        tex.SetPixels(px);
        tex.Apply();
        tex.filterMode = FilterMode.Point;
        return Sprite.Create(tex, new Rect(0, 0, size, size), new Vector2(0.5f, 0.5f), size);
    }
}
