import {
  createContext,
  useContext,
  useReducer,
  type ReactNode,
  type Dispatch,
} from "react";
import type {
  Lobby,
  RoleInfo,
  GameStartedPayload,
  RoleAssignedPayload,
} from "../types/game";

interface GameState {
  playerId: string | null;
  lobby: Lobby | null;
  roleInfo: RoleInfo | null;
  sessionMeta: GameStartedPayload | null;
  error: string | null;
}

type GameAction =
  | { type: "SET_PLAYER_ID"; playerId: string }
  | { type: "LOBBY_UPDATED"; lobby: Lobby }
  | { type: "GAME_STARTED"; payload: GameStartedPayload }
  | { type: "ROLE_ASSIGNED"; payload: RoleAssignedPayload }
  | { type: "SET_ERROR"; error: string }
  | { type: "CLEAR_ERROR" }
  | { type: "RESET" };

const initialState: GameState = {
  playerId: sessionStorage.getItem("playerId"),
  lobby: null,
  roleInfo: null,
  sessionMeta: null,
  error: null,
};

function gameReducer(state: GameState, action: GameAction): GameState {
  switch (action.type) {
    case "SET_PLAYER_ID":
      sessionStorage.setItem("playerId", action.playerId);
      return { ...state, playerId: action.playerId, error: null };

    case "LOBBY_UPDATED":
      return { ...state, lobby: action.lobby };

    case "GAME_STARTED":
      return { ...state, sessionMeta: action.payload };

    case "ROLE_ASSIGNED":
      return {
        ...state,
        roleInfo: {
          role: action.payload.role,
          location: action.payload.location,
        },
      };

    case "SET_ERROR":
      return { ...state, error: action.error };

    case "CLEAR_ERROR":
      return { ...state, error: null };

    case "RESET":
      sessionStorage.removeItem("playerId");
      sessionStorage.removeItem("lobbyCode");
      return { ...initialState, playerId: null };

    default:
      return state;
  }
}

const GameStateContext = createContext<GameState>(initialState);
const GameDispatchContext = createContext<Dispatch<GameAction>>(() => {});

export function GameProvider({ children }: { children: ReactNode }) {
  const [state, dispatch] = useReducer(gameReducer, initialState);

  return (
    <GameStateContext.Provider value={state}>
      <GameDispatchContext.Provider value={dispatch}>
        {children}
      </GameDispatchContext.Provider>
    </GameStateContext.Provider>
  );
}

export function useGameState() {
  return useContext(GameStateContext);
}

export function useGameDispatch() {
  return useContext(GameDispatchContext);
}
