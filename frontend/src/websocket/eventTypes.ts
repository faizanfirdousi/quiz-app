// WebSocket event type definitions

import { Option, PlayerScore } from '../types';

// --- Inbound (client → server) ---

export interface WSInbound<T = unknown> {
  action: string;
  data: T;
}

export interface JoinSessionPayload {
  sessionId: string;
  nickname: string;
}

export interface SubmitAnswerPayload {
  questionId: string;
  selectedOptionId: string;
  timeTakenMs: number;
}

export interface StartGamePayload {
  sessionId: string;
}

export interface NextQuestionPayload {
  sessionId: string;
}

export interface EndGamePayload {
  sessionId: string;
}

// --- Outbound (server → client) ---

export interface WSOutbound<T = unknown> {
  type: string;
  payload: T;
}

export interface PlayerJoinedPayload {
  nickname: string;
  playerCount: number;
}

export interface GameStartedPayload {
  totalQuestions: number;
}

export interface QuestionPayload {
  questionIndex: number;
  totalQuestions: number;
  text: string;
  options: Option[];
  timeLimitMs: number;
  points: number;
}

export interface AnswerResultPayload {
  isCorrect: boolean;
  pointsEarned: number;
  totalScore: number;
  rank: number;
  correctOptionId: string;
}

export interface QuestionEndedPayload {
  correctOptionId: string;
  leaderboard: PlayerScore[];
}

export interface LeaderboardUpdatePayload {
  leaderboard: PlayerScore[];
}

export interface GameOverPayload {
  finalLeaderboard: PlayerScore[];
}

export interface ErrorPayload {
  code: string;
  message: string;
}

// Event type constants
export const WS_TYPES = {
  PLAYER_JOINED: 'player_joined',
  GAME_STARTED: 'game_started',
  QUESTION: 'question',
  ANSWER_RESULT: 'answer_result',
  QUESTION_ENDED: 'question_ended',
  LEADERBOARD_UPDATE: 'leaderboard_update',
  GAME_OVER: 'game_over',
  ERROR: 'error',
} as const;

export const WS_ACTIONS = {
  JOIN_SESSION: 'join_session',
  SUBMIT_ANSWER: 'submit_answer',
  START_GAME: 'start_game',
  NEXT_QUESTION: 'next_question',
  END_GAME: 'end_game',
} as const;
