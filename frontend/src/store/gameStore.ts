import { create } from 'zustand';
import { GameStatus, PlayerRole } from '../types';
import { QuestionPayload, AnswerResultPayload } from '../websocket/eventTypes';

interface GameState {
  sessionId: string | null;
  role: PlayerRole | null;
  gameStatus: GameStatus;
  currentQuestion: QuestionPayload | null;
  questionIndex: number;
  totalQuestions: number;
  timeRemaining: number; // in ms
  hasAnswered: boolean;
  lastAnswerResult: AnswerResultPayload | null;
  playerCount: number;
  nickname: string | null;
  countdownInterval: ReturnType<typeof setInterval> | null;

  // Actions
  setSession: (sessionId: string) => void;
  setRole: (role: PlayerRole) => void;
  setGameStatus: (status: GameStatus) => void;
  setQuestion: (question: QuestionPayload) => void;
  startCountdown: (durationMs: number) => void;
  setHasAnswered: (answered: boolean) => void;
  setAnswerResult: (result: AnswerResultPayload) => void;
  setPlayerCount: (count: number) => void;
  setTotalQuestions: (total: number) => void;
  setNickname: (nickname: string) => void;
  resetGame: () => void;
}

export const useGameStore = create<GameState>((set, get) => ({
  sessionId: null,
  role: null,
  gameStatus: 'IDLE',
  currentQuestion: null,
  questionIndex: 0,
  totalQuestions: 0,
  timeRemaining: 0,
  hasAnswered: false,
  lastAnswerResult: null,
  playerCount: 0,
  nickname: null,
  countdownInterval: null,

  setSession: (sessionId) => set({ sessionId }),
  setRole: (role) => set({ role }),
  setGameStatus: (status) => set({ gameStatus: status }),

  setQuestion: (question) => {
    // Clear previous countdown
    const prev = get().countdownInterval;
    if (prev) clearInterval(prev);

    set({
      currentQuestion: question,
      questionIndex: question.questionIndex,
      totalQuestions: question.totalQuestions,
      timeRemaining: question.timeLimitMs,
      hasAnswered: false,
      lastAnswerResult: null,
      countdownInterval: null,
    });
  },

  startCountdown: (durationMs) => {
    const prev = get().countdownInterval;
    if (prev) clearInterval(prev);

    set({ timeRemaining: durationMs });

    const interval = setInterval(() => {
      const remaining = get().timeRemaining;
      if (remaining <= 0) {
        clearInterval(interval);
        set({ timeRemaining: 0, countdownInterval: null });
        return;
      }
      set({ timeRemaining: remaining - 100 });
    }, 100);

    set({ countdownInterval: interval });
  },

  setHasAnswered: (answered) => set({ hasAnswered: answered }),

  setAnswerResult: (result) => set({
    lastAnswerResult: result,
    hasAnswered: true,
  }),

  setPlayerCount: (count) => set({ playerCount: count }),
  setTotalQuestions: (total) => set({ totalQuestions: total }),
  setNickname: (nickname) => set({ nickname }),

  resetGame: () => {
    const interval = get().countdownInterval;
    if (interval) clearInterval(interval);

    set({
      sessionId: null,
      role: null,
      gameStatus: 'IDLE',
      currentQuestion: null,
      questionIndex: 0,
      totalQuestions: 0,
      timeRemaining: 0,
      hasAnswered: false,
      lastAnswerResult: null,
      playerCount: 0,
      nickname: null,
      countdownInterval: null,
    });
  },
}));
