import { useParams, useNavigate } from "react-router-dom";
import { useGameState } from "../context/GameContext";
import { useGameSocket } from "../hooks/useGameSocket";
import RoleCard from "../components/RoleCard";

export default function Game() {
  const { code } = useParams<{ code: string }>();
  const navigate = useNavigate();
  const { playerId, lobby, roleInfo, sessionMeta } = useGameState();

  useGameSocket({
    code: code ?? "",
    playerId,
  });

  if (!playerId || !code) {
    navigate("/");
    return null;
  }

  if (!roleInfo || !sessionMeta) {
    return (
      <div className="page">
        <p>Загрузка данных сессии...</p>
      </div>
    );
  }

  const players = lobby?.players ?? [];

  const playerMap = new Map(players.map((p) => [p.id, p]));

  return (
    <div className="page game-page">
      <div className="game-header">
        <h2>Сессия: {code}</h2>
      </div>

      <RoleCard roleInfo={roleInfo} />

      {roleInfo.role === "regular" && (
        <p className="reminder">Не называйте локацию прямо!</p>
      )}

      <div className="turn-order">
        <h3>Порядок ходов</h3>
        <ol>
          {sessionMeta.turnOrder.map((pid, idx) => {
            const p = playerMap.get(pid);
            const isCurrent = idx === sessionMeta.currentTurnIndex;
            const isMe = pid === playerId;
            return (
              <li
                key={pid}
                className={`turn-item ${isCurrent ? "active-turn" : ""}`}
              >
                {isCurrent && <span className="turn-arrow">&#9654; </span>}
                {p?.nickname ?? pid}
                {isMe && " (вы)"}
                {isCurrent && isMe && " — ваш ход"}
              </li>
            );
          })}
        </ol>
      </div>

      <div className="game-players">
        <h3>Игроки</h3>
        <div className="player-chips">
          {players.map((p) => (
            <span
              key={p.id}
              className={`player-chip ${p.online ? "online" : "offline"}`}
            >
              {p.nickname}
              <span className={`dot ${p.online ? "on" : "off"}`} />
            </span>
          ))}
        </div>
      </div>
    </div>
  );
}
