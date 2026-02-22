import apiClient from './client';
import { Session, PlayerScore, ApiResponse } from '../types';

export async function createSession(quizId: string): Promise<Session> {
  const response = await apiClient.post<ApiResponse<Session>>('/sessions', { quizId });
  return response.data.data;
}

export async function joinSession(
  sessionId: string,
  nickname: string,
  pin?: string
): Promise<{ sessionId: string; pin: string; nickname: string }> {
  const response = await apiClient.post<ApiResponse<{ sessionId: string; pin: string; nickname: string }>>(
    `/sessions/${sessionId}/join`,
    { nickname, pin }
  );
  return response.data.data;
}

export async function joinByPIN(
  pin: string,
  nickname: string
): Promise<{ sessionId: string; pin: string; nickname: string }> {
  const response = await apiClient.post<ApiResponse<{ sessionId: string; pin: string; nickname: string }>>(
    '/sessions/_/join',
    { nickname, pin }
  );
  return response.data.data;
}

export async function getLeaderboard(sessionId: string): Promise<{
  sessionId: string;
  leaderboard: PlayerScore[];
}> {
  const response = await apiClient.get<ApiResponse<{
    sessionId: string;
    leaderboard: PlayerScore[];
  }>>(`/sessions/${sessionId}/leaderboard`);
  return response.data.data;
}
