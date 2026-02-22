import { create } from 'zustand';
import { PlayerScore } from '../types';

interface LeaderboardState {
  entries: PlayerScore[];
  myRank: number | null;
  myScore: number;

  setLeaderboard: (entries: PlayerScore[]) => void;
  updateMyScore: (score: number, rank: number) => void;
  reset: () => void;
}

export const useLeaderboardStore = create<LeaderboardState>((set) => ({
  entries: [],
  myRank: null,
  myScore: 0,

  setLeaderboard: (entries) => set({ entries }),

  updateMyScore: (score, rank) => set({ myScore: score, myRank: rank }),

  reset: () => set({ entries: [], myRank: null, myScore: 0 }),
}));
