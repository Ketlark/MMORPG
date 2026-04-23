using UnityEngine;
using UnityEngine.Tilemaps;

namespace MMORPG.Map
{
    public enum TerrainType
    {
        Grass = 0,
        Water = 1,
        Wall = 2,
        Road = 3
    }

    public class IsometricMap : MonoBehaviour
    {
        [Header("Tilemap References")]
        [SerializeField] private Tilemap _groundTilemap;
        [SerializeField] private Tilemap _overlayTilemap;

        [Header("Tiles")]
        [SerializeField] private TileBase _grassTile;
        [SerializeField] private TileBase _waterTile;
        [SerializeField] private TileBase _wallTile;
        [SerializeField] private TileBase _roadTile;
        [SerializeField] private TileBase _highlightTile;

        [Header("Map Settings")]
        [SerializeField] private int _mapWidth = 20;
        [SerializeField] private int _mapHeight = 15;

        private const int TileWidth = 64;
        private const int TileHeight = 32;
        private const int Ppu = 64;
        private const float BorderThickness = 0.9f;

        private int[,] _terrainGrid;
        private bool[,] _walkableGrid;
        private Vector3Int _lastHighlightedCell = new Vector3Int(-1, -1, -1);
        private Camera _cachedCamera;

        public event System.Action<int, int> OnCellClicked;

        public int MapWidth => _mapWidth;
        public int MapHeight => _mapHeight;

        private void Awake()
        {
            _cachedCamera = Camera.main;
        }

        private void OnValidate()
        {
            if (_groundTilemap != null && _overlayTilemap != null) return;
            var tilemaps = GetComponentsInChildren<Tilemap>();
            if (tilemaps.Length >= 1 && _groundTilemap == null) _groundTilemap = tilemaps[0];
            if (tilemaps.Length >= 2 && _overlayTilemap == null) _overlayTilemap = tilemaps[1];
        }

        private void Start()
        {
            EnsureTiles();
            GenerateTestMap();
        }

        private void EnsureTiles()
        {
            if (_grassTile == null || (_grassTile as Tile)?.sprite == null)
                _grassTile = MakeIsometricTile(new Color(0.4f, 0.7f, 0.3f));
            if (_waterTile == null || (_waterTile as Tile)?.sprite == null)
                _waterTile = MakeIsometricTile(new Color(0.2f, 0.4f, 0.8f));
            if (_wallTile == null || (_wallTile as Tile)?.sprite == null)
                _wallTile = MakeIsometricTile(new Color(0.5f, 0.4f, 0.3f));
            if (_roadTile == null || (_roadTile as Tile)?.sprite == null)
                _roadTile = MakeIsometricTile(new Color(0.7f, 0.65f, 0.5f));
            if (_highlightTile == null || (_highlightTile as Tile)?.sprite == null)
                _highlightTile = MakeIsometricTile(new Color(1f, 1f, 0.5f, 0.5f));
        }

        private static Tile MakeIsometricTile(Color color, bool addBorder = true)
        {
            var tex = new Texture2D(TileWidth, TileHeight);
            var px = new Color[TileWidth * TileHeight];
            float halfW = TileWidth / 2f;
            float halfH = TileHeight / 2f;

            for (int y = 0; y < TileHeight; y++)
            {
                for (int x = 0; x < TileWidth; x++)
                {
                    float dx = Mathf.Abs(x - halfW + 0.5f) / halfW;
                    float dy = Mathf.Abs(y - halfH + 0.5f) / halfH;
                    float d = dx + dy;

                    if (d <= 1f)
                    {
                        if (addBorder && d > BorderThickness)
                            px[y * TileWidth + x] = new Color(0f, 0f, 0f, 0.3f);
                        else
                            px[y * TileWidth + x] = color;
                    }
                }
            }

            tex.SetPixels(px);
            tex.Apply();
            tex.filterMode = FilterMode.Point;
            var sprite = Sprite.Create(tex, new Rect(0, 0, TileWidth, TileHeight), new Vector2(0.5f, 0.5f), Ppu);
            var tile = ScriptableObject.CreateInstance<Tile>();
            tile.sprite = sprite;
            return tile;
        }

        public void GenerateTestMap()
        {
            _terrainGrid = new int[_mapWidth, _mapHeight];
            _walkableGrid = new bool[_mapWidth, _mapHeight];

            for (int x = 0; x < _mapWidth; x++)
                for (int y = 0; y < _mapHeight; y++)
                {
                    _terrainGrid[x, y] = (int)TerrainType.Grass;
                    _walkableGrid[x, y] = true;
                }

            for (int x = 0; x < _mapWidth; x++)
                SetTerrain(x, _mapHeight / 2, TerrainType.Road);
            for (int y = 0; y < _mapHeight; y++)
                SetTerrain(_mapWidth / 2, y, TerrainType.Road);

            for (int x = 14; x < 18; x++)
                for (int y = 10; y < 14; y++)
                    SetTerrain(x, y, TerrainType.Water);

            SetTerrain(3, 3, TerrainType.Wall); SetTerrain(3, 4, TerrainType.Wall);
            SetTerrain(4, 3, TerrainType.Wall); SetTerrain(4, 4, TerrainType.Wall);
            SetTerrain(7, 3, TerrainType.Wall); SetTerrain(7, 4, TerrainType.Wall);
            SetTerrain(8, 3, TerrainType.Wall); SetTerrain(8, 4, TerrainType.Wall);

            RenderMap();
        }

        private void SetTerrain(int x, int y, TerrainType type)
        {
            _terrainGrid[x, y] = (int)type;
            _walkableGrid[x, y] = type != TerrainType.Water && type != TerrainType.Wall;
        }

        public void LoadMapFromProto(int width, int height, int[] terrainTypes, bool[] walkable)
        {
            _mapWidth = width;
            _mapHeight = height;
            _terrainGrid = new int[width, height];
            _walkableGrid = new bool[width, height];

            for (int x = 0; x < width; x++)
                for (int y = 0; y < height; y++)
                {
                    int idx = y * width + x;
                    _terrainGrid[x, y] = terrainTypes[idx];
                    _walkableGrid[x, y] = walkable[idx];
                }
            RenderMap();
        }

        private void RenderMap()
        {
            _groundTilemap.ClearAllTiles();
            for (int x = 0; x < _mapWidth; x++)
                for (int y = 0; y < _mapHeight; y++)
                {
                    TileBase tile = _terrainGrid[x, y] switch
                    {
                        (int)TerrainType.Water => _waterTile,
                        (int)TerrainType.Wall => _wallTile,
                        (int)TerrainType.Road => _roadTile,
                        _ => _grassTile
                    };
                    _groundTilemap.SetTile(new Vector3Int(x, y, 0), tile);
                }
            CenterCamera();
        }

        public void CenterCamera()
        {
            var centerCell = new Vector3Int(_mapWidth / 2, _mapHeight / 2, 0);
            var centerWorld = _groundTilemap.GetCellCenterWorld(centerCell);
            var cam = _cachedCamera;
            if (cam != null)
                cam.transform.position = new Vector3(centerWorld.x, centerWorld.y, cam.transform.position.z);
        }

        public bool IsWalkable(int x, int y)
        {
            if (x < 0 || x >= _mapWidth || y < 0 || y >= _mapHeight) return false;
            return _walkableGrid[x, y];
        }

        public Vector3 GridToWorld(int x, int y)
            => _groundTilemap.GetCellCenterWorld(new Vector3Int(x, y, 0));

        public Vector3Int WorldToGrid(Vector3 worldPos)
            => _groundTilemap.WorldToCell(worldPos);

        public void HighlightCell(int x, int y)
        {
            if (_lastHighlightedCell.x >= 0)
                _overlayTilemap.SetTile(_lastHighlightedCell, null);
            var cell = new Vector3Int(x, y, 0);
            _overlayTilemap.SetTile(cell, _highlightTile);
            _lastHighlightedCell = cell;
        }

        public void ClearHighlight()
        {
            if (_lastHighlightedCell.x >= 0)
                _overlayTilemap.SetTile(_lastHighlightedCell, null);
        }

        private void Update()
        {
            var mouse = UnityEngine.InputSystem.Mouse.current;
            if (mouse == null) return;

            if (mouse.leftButton.wasPressedThisFrame)
            {
                Vector3 mouseWorldPos = _cachedCamera.ScreenToWorldPoint(mouse.position.ReadValue());
                Vector3Int cell = WorldToGrid(mouseWorldPos);
                if (cell.x >= 0 && cell.x < _mapWidth && cell.y >= 0 && cell.y < _mapHeight)
                    OnCellClicked?.Invoke(cell.x, cell.y);
            }

            Vector3 hoverWorldPos = _cachedCamera.ScreenToWorldPoint(mouse.position.ReadValue());
            Vector3Int hoverCell = WorldToGrid(hoverWorldPos);
            if (hoverCell.x >= 0 && hoverCell.x < _mapWidth && hoverCell.y >= 0 && hoverCell.y < _mapHeight)
                HighlightCell(hoverCell.x, hoverCell.y);
        }
    }
}
