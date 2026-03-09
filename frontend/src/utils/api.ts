import type { Lobby } from "../types/game";

interface LobbyResponse {
  lobby: Lobby;
  playerId: string;
}

interface ApiError {
  error: string;
}

const API_BASE = "/api";

async function request<T>(
  path: string,
  options?: RequestInit,
): Promise<T> {
  const res = await fetch(`${API_BASE}${path}`, {
    headers: { "Content-Type": "application/json" },
    ...options,
  });

  const body = await res.json();

  if (!res.ok) {
    const err = body as ApiError;
    throw new Error(err.error || `HTTP ${res.status}`);
  }

  return body as T;
}

export async function createLobby(nickname: string): Promise<LobbyResponse> {
  return request<LobbyResponse>("/lobbies", {
    method: "POST",
    body: JSON.stringify({ nickname }),
  });
}

export async function joinLobby(
  code: string,
  nickname: string,
): Promise<LobbyResponse> {
  return request<LobbyResponse>(`/lobbies/${code.toUpperCase()}/join`, {
    method: "POST",
    body: JSON.stringify({ nickname }),
  });
}

const ERROR_MESSAGES: Record<string, string> = {
  invalid_nickname: "Никнейм должен быть от 1 до 20 символов",
  lobby_not_found: "Лобби не найдено",
  lobby_full: "Лобби заполнено",
  game_in_progress: "Игра уже началась",
  nickname_taken: "Этот никнейм уже занят",
};

export function humanizeError(error: string): string {
  return ERROR_MESSAGES[error] || error;
}
