using UnityEngine;
using System.Collections.Generic;
using MMORPG.Map;

namespace MMORPG.Game
{
    public abstract class GridEntity : MonoBehaviour
    {
        [SerializeField] private float _moveSpeed = 5f;

        public string PlayerId { get; protected set; }
        public string Username { get; protected set; }
        public int GridX { get; protected set; }
        public int GridY { get; protected set; }

        private Queue<Vector3> _waypoints = new Queue<Vector3>();
        private Vector3 _currentTarget;
        private bool _isMoving;

        private const float SnapDistance = 0.01f;
        private const int DefaultSpriteSize = 32;
        private const int DefaultPpu = 32;

        public void MoveTo(int x, int y, IsometricMap map)
        {
            GridX = x;
            GridY = y;
            _waypoints.Clear();
            _currentTarget = map.GridToWorld(x, y);
            _isMoving = true;
        }

        public void MoveToPath(List<global::Game.PathNode> pathNodes, IsometricMap map)
        {
            if (pathNodes == null || pathNodes.Count == 0) return;

            _waypoints.Clear();

            int start = 0;
            if (pathNodes[0].X == GridX && pathNodes[0].Y == GridY)
                start = 1;

            for (int i = start; i < pathNodes.Count; i++)
                _waypoints.Enqueue(map.GridToWorld(pathNodes[i].X, pathNodes[i].Y));

            var last = pathNodes[pathNodes.Count - 1];
            GridX = last.X;
            GridY = last.Y;

            if (_waypoints.Count > 0)
            {
                _currentTarget = _waypoints.Dequeue();
                _isMoving = true;
            }
        }

        protected void SnapToGrid(int x, int y, IsometricMap map)
        {
            GridX = x;
            GridY = y;
            _waypoints.Clear();
            _currentTarget = map.GridToWorld(x, y);
            transform.position = _currentTarget;
            _isMoving = false;
        }

        protected SpriteRenderer EnsureSprite(Color color, int sortingOrder = 10)
        {
            var sr = GetComponent<SpriteRenderer>();
            if (sr == null) sr = gameObject.AddComponent<SpriteRenderer>();
            if (sr.sprite == null) sr.sprite = CreateProceduralSprite(color);
            sr.sortingOrder = sortingOrder;
            return sr;
        }

        public static Sprite CreateProceduralSprite(Color color)
        {
            var tex = new Texture2D(DefaultSpriteSize, DefaultSpriteSize);
            var px = new Color[DefaultSpriteSize * DefaultSpriteSize];
            for (int i = 0; i < px.Length; i++) px[i] = color;
            tex.SetPixels(px);
            tex.Apply();
            return Sprite.Create(tex, new Rect(0, 0, DefaultSpriteSize, DefaultSpriteSize), new Vector2(0.5f, 0.5f), DefaultPpu);
        }

        private void Update()
        {
            if (!_isMoving) return;
            transform.position = Vector3.MoveTowards(transform.position, _currentTarget, _moveSpeed * Time.deltaTime);
            if (Vector3.Distance(transform.position, _currentTarget) < SnapDistance)
            {
                transform.position = _currentTarget;
                if (_waypoints.Count > 0)
                    _currentTarget = _waypoints.Dequeue();
                else
                {
                    _isMoving = false;
                    OnMovementFinished();
                }
            }
        }

        protected virtual void OnMovementFinished() { }
    }
}
