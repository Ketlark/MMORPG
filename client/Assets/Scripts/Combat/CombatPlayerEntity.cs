using UnityEngine;
using MMORPG.Map;
using MMORPG.Game;

namespace MMORPG.Combat
{
    public class CombatPlayerEntity : GridEntity
    {
        private SpriteRenderer _spriteRenderer;
        private Color _originalColor;

        public int Health { get; private set; }
        public int MaxHealth { get; private set; }

        public void Initialize(string playerId, string username, int x, int y,
            int health, int maxHealth, IsometricMap map)
        {
            PlayerId = playerId;
            Username = username;
            Health = health;
            MaxHealth = maxHealth;
            SnapToGrid(x, y, map);
            _spriteRenderer = EnsureSprite(new Color(0.2f, 0.6f, 1f));
            _originalColor = _spriteRenderer.color;
        }

        public void TakeDamage(int damage, int healthRemaining)
        {
            Health = healthRemaining;
            StartCoroutine(FlashRed());
        }

        private System.Collections.IEnumerator FlashRed()
        {
            _spriteRenderer.color = Color.red;
            yield return new WaitForSeconds(0.2f);
            _spriteRenderer.color = _originalColor;
        }
    }
}
