export interface Player {
  id: string;
  nickname: string;
  isAdmin: boolean;
  ready: boolean;
  online: boolean;
}

export interface LobbySettings {
  minPlayers: number;
  maxPlayers: number;
  locations: string[];
  voteThreshold: number;
  votingTimerSeconds: number;
}

export type LobbyState = "waiting" | "in_game";

export interface Session {
  id: string;
  location: string;
  spyId: string;
  turnOrder: string[];
  currentTurnIndex: number;
  phase: SessionPhase;
}

export type SessionPhase =
  | "question"
  | "vote_proposal"
  | "voting"
  | "spy_guess"
  | "finished";

export interface Lobby {
  code: string;
  players: Player[];
  settings: LobbySettings;
  state: LobbyState;
  currentSession: Session | null;
}

export type PlayerRole = "spy" | "regular";

export interface RoleInfo {
  role: PlayerRole;
  location: string | null;
}

export interface WsMessage<T = unknown> {
  type: string;
  payload: T;
}

export interface GameStartedPayload {
  sessionId: string;
  turnOrder: string[];
  phase: SessionPhase;
  currentTurnIndex: number;
}

export interface RoleAssignedPayload {
  role: PlayerRole;
  location: string | null;
}

export interface ErrorPayload {
  error: string;
  message: string;
}
