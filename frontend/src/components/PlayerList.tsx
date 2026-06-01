import type { Player } from "../types/game";

interface Props {
  players: Player[];
  currentPlayerId?: string | null;
}

export default function PlayerList({ players, currentPlayerId }: Props) {
  return (
    <ul className="player-list">
      {players.map((p) => (
        <li key={p.id} className={`player-item ${!p.online ? "offline" : ""}`}>
          <span className={`online-dot ${p.online ? "on" : "off"}`} />
          <span className="player-name">
            {p.nickname}
            {p.id === currentPlayerId && " (вы)"}
          </span>
          {p.isAdmin && <span className="badge admin-badge">admin</span>}
          <span className={`ready-status ${p.ready ? "ready" : "not-ready"}`}>
            {p.ready ? "Готов" : "Не готов"}
          </span>
        </li>
      ))}
    </ul>
  );
}
