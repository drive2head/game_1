import { useEffect, useRef, useCallback, useState } from "react";
import { useGameDispatch } from "../context/GameContext";
import type {
  WsMessage,
  Lobby,
  GameStartedPayload,
  RoleAssignedPayload,
} from "../types/game";

const RECONNECT_INTERVAL_MS = 2000;
const MAX_RECONNECT_INTERVAL_MS = 16000;

interface UseGameSocketOptions {
  code: string;
  playerId: string | null;
  onGameStarted?: () => void;
}

export function useGameSocket({
  code,
  playerId,
  onGameStarted,
}: UseGameSocketOptions) {
  const dispatch = useGameDispatch();
  const wsRef = useRef<WebSocket | null>(null);
  const reconnectTimerRef = useRef<ReturnType<typeof setTimeout> | undefined>(undefined);
  const reconnectIntervalRef = useRef(RECONNECT_INTERVAL_MS);
  const [connected, setConnected] = useState(false);
  const [reconnecting, setReconnecting] = useState(false);
  const onGameStartedRef = useRef(onGameStarted);
  onGameStartedRef.current = onGameStarted;

  const getWsUrl = useCallback(() => {
    const proto = window.location.protocol === "https:" ? "wss:" : "ws:";
    const host = window.location.host;
    return `${proto}//${host}/api/lobbies/${code}/ws?playerId=${playerId}`;
  }, [code, playerId]);

  const sendCommand = useCallback(
    (type: string, payload: Record<string, unknown> = {}) => {
      if (wsRef.current?.readyState === WebSocket.OPEN) {
        wsRef.current.send(JSON.stringify({ type, payload }));
      }
    },
    [],
  );

  const connect = useCallback(() => {
    if (!playerId || !code) return;

    const ws = new WebSocket(getWsUrl());
    wsRef.current = ws;

    ws.onopen = () => {
      setConnected(true);
      setReconnecting(false);
      reconnectIntervalRef.current = RECONNECT_INTERVAL_MS;
    };

    ws.onmessage = (event) => {
      const msg: WsMessage = JSON.parse(event.data);

      switch (msg.type) {
        case "lobby_updated": {
          const data = msg.payload as { lobby: Lobby };
          dispatch({ type: "LOBBY_UPDATED", lobby: data.lobby });
          break;
        }
        case "game_started": {
          const data = msg.payload as GameStartedPayload;
          dispatch({ type: "GAME_STARTED", payload: data });
          onGameStartedRef.current?.();
          break;
        }
        case "role_assigned": {
          const data = msg.payload as RoleAssignedPayload;
          dispatch({ type: "ROLE_ASSIGNED", payload: data });
          break;
        }
        case "error": {
          const data = msg.payload as { error: string; message: string };
          dispatch({ type: "SET_ERROR", error: data.message || data.error });
          break;
        }
        case "player_disconnected":
        case "player_reconnected":
        case "player_left":
          break;
      }
    };

    ws.onclose = () => {
      setConnected(false);
      setReconnecting(true);

      reconnectTimerRef.current = setTimeout(() => {
        reconnectIntervalRef.current = Math.min(
          reconnectIntervalRef.current * 2,
          MAX_RECONNECT_INTERVAL_MS,
        );
        connect();
      }, reconnectIntervalRef.current);
    };

    ws.onerror = () => {
      ws.close();
    };
  }, [playerId, code, getWsUrl, dispatch]);

  useEffect(() => {
    connect();

    return () => {
      clearTimeout(reconnectTimerRef.current);
      if (wsRef.current) {
        wsRef.current.onclose = null;
        wsRef.current.close();
      }
    };
  }, [connect]);

  const disconnect = useCallback(() => {
    clearTimeout(reconnectTimerRef.current);
    if (wsRef.current) {
      wsRef.current.onclose = null;
      wsRef.current.close();
    }
    setConnected(false);
    setReconnecting(false);
  }, []);

  return { sendCommand, connected, reconnecting, disconnect };
}
