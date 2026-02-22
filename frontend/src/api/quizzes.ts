import apiClient from './client';
import { Quiz, Question, ApiResponse } from '../types';

export async function createQuiz(data: {
  title: string;
  description: string;
  questions: Question[];
}): Promise<Quiz> {
  const response = await apiClient.post<ApiResponse<Quiz>>('/quizzes', data);
  return response.data.data;
}

export async function getQuiz(quizId: string): Promise<Quiz> {
  const response = await apiClient.get<ApiResponse<Quiz>>(`/quizzes/${quizId}`);
  return response.data.data;
}

export async function listMyQuizzes(): Promise<Quiz[]> {
  const response = await apiClient.get<ApiResponse<Quiz[]>>('/quizzes');
  return response.data.data;
}
