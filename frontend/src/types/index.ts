// Shared TypeScript types for the KahootClone frontend

// --- Quiz Types ---

export interface Option {
  id: string;
  text: string;
}

export interface Question {
  questionId: string;
  text: string;
  options: Option[];
  correctOptionId: string;
  timeLimitSeconds: number;
  points: number;
}

export interface Quiz {
  quizId: string;
  hostUserId: string;
  title: string;
  description: string;
  questions: Question[];
  createdAt: string;
  updatedAt: string;
}

// --- Session Types ---

export type SessionStatus = 'LOBBY' | 'ACTIVE' | 'FINISHED';

export interface Session {
  sessionId: string;
  pin: string;
  quizId: string;
  hostUserId: string;
  status: SessionStatus;
  currentQuestionIndex: number;
  startedAt?: string;
  endedAt?: string;
  createdAt: string;
}

// --- Player Types ---

export type PlayerRole = 'HOST' | 'PLAYER';

export interface PlayerScore {
  userId: string;
  nickname: string;
  score: number;
  rank: number;
}

// --- API Response Types ---

export interface ApiResponse<T> {
  success: boolean;
  data: T;
  requestId: string;
  timestamp: string;
}

export interface ApiError {
  success: false;
  error: {
    code: string;
    message: string;
  };
  requestId: string;
  timestamp: string;
}

// --- Game State ---

export type GameStatus = 'IDLE' | 'LOBBY' | 'ACTIVE' | 'FINISHED';

export type ConnectionStatus = 'connecting' | 'connected' | 'disconnected' | 'error';
