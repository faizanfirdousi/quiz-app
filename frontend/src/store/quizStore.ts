import { create } from 'zustand';
import { Question } from '../types';

interface QuizState {
  title: string;
  description: string;
  questions: Question[];

  setTitle: (title: string) => void;
  setDescription: (description: string) => void;
  addQuestion: (question: Question) => void;
  updateQuestion: (index: number, question: Question) => void;
  removeQuestion: (index: number) => void;
  reorderQuestions: (fromIndex: number, toIndex: number) => void;
  reset: () => void;
}

export const useQuizStore = create<QuizState>((set, get) => ({
  title: '',
  description: '',
  questions: [],

  setTitle: (title) => set({ title }),
  setDescription: (description) => set({ description }),

  addQuestion: (question) => set((state) => ({
    questions: [...state.questions, question],
  })),

  updateQuestion: (index, question) => set((state) => {
    const newQuestions = [...state.questions];
    newQuestions[index] = question;
    return { questions: newQuestions };
  }),

  removeQuestion: (index) => set((state) => ({
    questions: state.questions.filter((_, i) => i !== index),
  })),

  reorderQuestions: (fromIndex, toIndex) => set((state) => {
    const newQuestions = [...state.questions];
    const [removed] = newQuestions.splice(fromIndex, 1);
    newQuestions.splice(toIndex, 0, removed);
    return { questions: newQuestions };
  }),

  reset: () => set({ title: '', description: '', questions: [] }),
}));
