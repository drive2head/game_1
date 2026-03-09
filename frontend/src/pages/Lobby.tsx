import { useParams, useNavigate } from "react-router-dom";
import { useGameState, useGameDispatch } from "../context/GameContext";
import { useGameSocket } from "../hooks/useGameSocket";
import PlayerList from "../components/PlayerList";
import { useCallback } from "react";

export default function Lobby() {
  const { code } = useParams<{ code: string }>();
  const navigate = useNavigate();
  const { playerId, lobby, error } = useGameState();
  const dispatch = useGameDispatch();

  const handleGameStarted = useCallback(() => {
    navigate(`/game/${code}`);
  }, [navigate, code]);

  const { sendCommand, connected, reconnecting, disconnect } = useGameSocket({
    code: code ?? "",
    playerId,
    onGameStarted: handleGameStarted,
  });

  if (!playerId || !code) {
    navigate("/");
    return null;
  }

  const currentPlayer = lobby?.players.find((p) => p.id === playerId);
  const isAdmin = currentPlayer?.isAdmin ?? false;
  const isReady = currentPlayer?.ready ?? false;

  const allReady = lobby
    ? lobby.players.every((p) => p.ready)
    : false;
  const enoughPlayers = lobby
    ? lobby.players.length >= lobby.settings.minPlayers
    : false;
  const canStart = isAdmin && allReady && enoughPlayers;

  function handleToggleReady() {
    sendCommand("player_ready", { ready: !isReady });
  }

  function handleLeave() {
    sendCommand("leave_lobby");
    disconnect();
    dispatch({ type: "RESET" });
    navigate("/");
  }

  function handleStartGame() {
    sendCommand("start_game");
  }

  if (!connected && !reconnecting && !lobby) {
    return (
      <div className="page">
        <p>Подключение...</p>
      </div>
    );
  }

  return (
    <div className="page lobby-page">
      <div className="lobby-header">
        <div>
          <h2>Комната</h2>
          <span className="room-code">{code}</span>
        </div>
        <button type="button" className="btn btn-danger" onClick={handleLeave}>
          Выйти
        </button>
      </div>

      {reconnecting && (
        <p className="warning-message">Переподключение...</p>
      )}

      {error && <p className="error-message">{error}</p>}

      {lobby && (
        <>
          <h3>
            Игроки ({lobby.players.length}/{lobby.settings.maxPlayers})
          </h3>
          <PlayerList players={lobby.players} currentPlayerId={playerId} />

          <div className="lobby-actions">
            <button
              type="button"
              className={`btn ${isReady ? "btn-secondary" : "btn-primary"}`}
              onClick={handleToggleReady}
            >
              {isReady ? "Не готов" : "Готов"}
            </button>

            {isAdmin && (
              <button
                type="button"
                className="btn btn-primary"
                disabled={!canStart}
                onClick={handleStartGame}
                title={
                  !enoughPlayers
                    ? `Нужно минимум ${lobby.settings.minPlayers} игрока`
                    : !allReady
                      ? "Не все игроки готовы"
                      : undefined
                }
              >
                Начать игру
              </button>
            )}
          </div>
        </>
      )}
    </div>
  );
}
