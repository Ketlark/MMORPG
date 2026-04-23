using UnityEngine;
using MMORPG.Map;

namespace MMORPG.Game
{
    public class PlayerEntity : GridEntity
    {
        public int Health { get; private set; }
        public int MaxHealth { get; private set; }
        public int ActionPoints { get; private set; }
        public int MovementPoints { get; private set; }

        public void Initialize(string playerId, string username, int x, int y, IsometricMap map)
        {
            PlayerId = playerId;
            Username = username;
            SnapToGrid(x, y, map);
            EnsureSprite(Color.magenta);
        }

        public void UpdateStats(int health, int maxHealth, int ap, int mp)
        {
            Health = health;
            MaxHealth = maxHealth;
            ActionPoints = ap;
            MovementPoints = mp;
        }
    }
}
