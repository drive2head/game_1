import { useState, type FormEvent } from "react";
import { useNavigate } from "react-router-dom";
import { useGameDispatch } from "../context/GameContext";
import { createLobby, joinLobby, humanizeError } from "../utils/api";

export default function Home() {
  const navigate = useNavigate();
  const dispatch = useGameDispatch();

  const [nickname, setNickname] = useState("");
  const [roomCode, setRoomCode] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  async function handleCreate(e: FormEvent) {
    e.preventDefault();
    setError(null);
    setLoading(true);
    try {
      const { lobby, playerId } = await createLobby(nickname);
      dispatch({ type: "SET_PLAYER_ID", playerId });
      dispatch({ type: "LOBBY_UPDATED", lobby });
      sessionStorage.setItem("lobbyCode", lobby.code);
      navigate(`/lobby/${lobby.code}`);
    } catch (err) {
      setError(humanizeError((err as Error).message));
    } finally {
      setLoading(false);
    }
  }

  async function handleJoin(e: FormEvent) {
    e.preventDefault();
    if (!roomCode.trim()) {
      setError("Введите код комнаты");
      return;
    }
    setError(null);
    setLoading(true);
    try {
      const { lobby, playerId } = await joinLobby(roomCode, nickname);
      dispatch({ type: "SET_PLAYER_ID", playerId });
      dispatch({ type: "LOBBY_UPDATED", lobby });
      sessionStorage.setItem("lobbyCode", lobby.code);
      navigate(`/lobby/${lobby.code}`);
    } catch (err) {
      setError(humanizeError((err as Error).message));
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="page home-page">
      <h1 className="title">Находка шпиона</h1>
      <p className="subtitle">Определите, кто среди вас шпион</p>

      <div className="card">
        <form onSubmit={handleCreate}>
          <label htmlFor="nickname">Никнейм</label>
          <input
            id="nickname"
            type="text"
            placeholder="Ваше имя"
            maxLength={20}
            value={nickname}
            onChange={(e) => setNickname(e.target.value)}
            required
          />
          <button type="submit" className="btn btn-primary" disabled={loading}>
            Создать игру
          </button>
        </form>

        <div className="divider">
          <span>или</span>
        </div>

        <form onSubmit={handleJoin}>
          <label htmlFor="roomCode">Код комнаты</label>
          <input
            id="roomCode"
            type="text"
            placeholder="ABCXYZ"
            maxLength={6}
            value={roomCode}
            onChange={(e) => setRoomCode(e.target.value.toUpperCase())}
          />
          <button type="submit" className="btn btn-secondary" disabled={loading}>
            Присоединиться
          </button>
        </form>

        {error && <p className="error-message">{error}</p>}
      </div>
    </div>
  );
}
