using UnityEngine;
using UnityEngine.EventSystems;

namespace MMORPG.UI
{
    public static class EventSystemUtility
    {
        public static void EnsureExists()
        {
            if (Object.FindAnyObjectByType<EventSystem>() != null) return;
            var go = new GameObject("EventSystem");
            go.AddComponent<EventSystem>();
            go.AddComponent<UnityEngine.InputSystem.UI.InputSystemUIInputModule>();
        }
    }
}
