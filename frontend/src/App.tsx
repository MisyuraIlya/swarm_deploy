import { useEffect, useState, type FormEvent } from "react";

const API_BASE = `http://${window.location.hostname}:8080`; 

type Item = {
  id: number;
  title: string;
  created_at: string; 
};

function App() {
  const [items, setItems] = useState<Item[]>([]);
  const [title, setTitle] = useState("");
  const [loading, setLoading] = useState(false);
  const [creating, setCreating] = useState(false);
  const [error, setError] = useState("");

  async function loadItems(): Promise<void> {
    try {
      setLoading(true);
      setError("");
      const res = await fetch(`${API_BASE}/api/items`);
      if (!res.ok) throw new Error("Failed to load items");
      const data: Item[] = await res.json();
      setItems(data);
    } catch (err) {
      console.error(err);
      setError("Could not load items");
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    void loadItems();
  }, []);

  async function handleSubmit(e: FormEvent<HTMLFormElement>): Promise<void> {
    e.preventDefault();
    if (!title.trim()) return;

    try {
      setCreating(true);
      setError("");
      const res = await fetch(`${API_BASE}/api/items`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ title }),
      });
      if (!res.ok) throw new Error("Failed to create item");
      const newItem: Item = await res.json();
      setItems((prev) => [newItem, ...prev]);
      setTitle("");
    } catch (err) {
      console.error(err);
      setError("Could not create item");
    } finally {
      setCreating(false);
    }
  }

  return (
    <div
      style={{
        maxWidth: "600px",
        margin: "2rem auto",
        fontFamily: "system-ui, -apple-system, BlinkMacSystemFont, sans-serif",
      }}
    >
      <h1>Docker Swarm Demo – Items</h1>
      <p style={{ color: "#555" }}>
        Backend: Go + Postgres · Frontend: React · Deployed with Docker Swarm
      </p>

      <form onSubmit={handleSubmit} style={{ marginBottom: "1.5rem" }}>
        <label style={{ display: "block", marginBottom: "0.5rem" }}>
          New item
        </label>
        <div style={{ display: "flex", gap: "0.5rem" }}>
          <input
            type="text"
            value={title}
            onChange={(e) => setTitle(e.target.value)}
            placeholder="e.g. Learn Docker Swarm"
            style={{
              flex: 1,
              padding: "0.5rem 0.75rem",
              borderRadius: "4px",
              border: "1px solid #ccc",
            }}
          />
          <button
            type="submit"
            disabled={creating || !title.trim()}
            style={{
              padding: "0.5rem 1rem",
              borderRadius: "4px",
              border: "none",
              backgroundColor: creating ? "#888" : "#2563eb",
              color: "white",
              cursor: creating ? "default" : "pointer",
            }}
          >
            {creating ? "Adding..." : "Add"}
          </button>
        </div>
      </form>

      <button
        onClick={loadItems}
        disabled={loading}
        style={{
          marginBottom: "1rem",
          padding: "0.4rem 0.9rem",
          borderRadius: "4px",
          border: "1px solid #ccc",
          background: "#f9fafb",
          cursor: loading ? "default" : "pointer",
        }}
      >
        {loading ? "Refreshing..." : "Refresh"}
      </button>

      {error && (
        <div
          style={{
            marginBottom: "1rem",
            padding: "0.75rem",
            borderRadius: "4px",
            background: "#fee2e2",
            color: "#b91c1c",
          }}
        >
          {error}
        </div>
      )}

      {items.length === 0 && !loading && (
        <p style={{ color: "#666" }}>No items yet. Add your first one!</p>
      )}

      <ul style={{ listStyle: "none", padding: 0 }}>
        {items.map((item) => (
          <li
            key={item.id}
            style={{
              padding: "0.75rem 1rem",
              borderRadius: "6px",
              border: "1px solid #e5e7eb",
              marginBottom: "0.5rem",
              display: "flex",
              justifyContent: "space-between",
              alignItems: "center",
            }}
          >
            <span>{item.title}</span>
            <span style={{ fontSize: "0.8rem", color: "#6b7280" }}>
              {new Date(item.created_at).toLocaleString()}
            </span>
          </li>
        ))}
      </ul>
    </div>
  );
}

export default App;
