import { useParams, useNavigate } from 'react-router-dom';
import { useEffect, useCallback } from 'react';
import { motion } from 'framer-motion';
import { useGameStore } from '../store/gameStore';
import { useGameSocket } from '../websocket/useGameSocket';
import { WS_ACTIONS } from '../websocket/eventTypes';
import CountdownTimer from '../components/CountdownTimer';
import AnswerButton from '../components/AnswerButton';

const ANSWER_COLORS: Array<'red' | 'blue' | 'yellow' | 'green'> = ['red', 'blue', 'yellow', 'green'];
const ANSWER_SHAPES = ['▲', '◆', '●', '■'];

export default function Question() {
  const { sessionId } = useParams<{ sessionId: string }>();
  const navigate = useNavigate();
  const {
    currentQuestion,
    questionIndex,
    totalQuestions,
    timeRemaining,
    hasAnswered,
    lastAnswerResult,
    gameStatus,
  } = useGameStore();
  const { sendMessage } = useGameSocket(sessionId || null, 'PLAYER');

  useEffect(() => {
    if (lastAnswerResult) {
      navigate(`/result/${sessionId}`);
    }
  }, [lastAnswerResult, sessionId, navigate]);

  useEffect(() => {
    if (gameStatus === 'FINISHED') {
      navigate(`/leaderboard/${sessionId}`);
    }
  }, [gameStatus, sessionId, navigate]);

  const handleAnswer = useCallback((optionId: string) => {
    if (hasAnswered || !currentQuestion) return;

    const timeTaken = (currentQuestion.timeLimitMs) - timeRemaining;
    sendMessage(WS_ACTIONS.SUBMIT_ANSWER, {
      questionId: currentQuestion.options[0]?.id ? currentQuestion.questionIndex.toString() : '',
      selectedOptionId: optionId,
      timeTakenMs: Math.max(0, timeTaken),
    });

    useGameStore.getState().setHasAnswered(true);
  }, [hasAnswered, currentQuestion, timeRemaining, sendMessage]);

  // Auto-submit null when timer runs out
  useEffect(() => {
    if (timeRemaining <= 0 && !hasAnswered && currentQuestion) {
      handleAnswer('');
    }
  }, [timeRemaining, hasAnswered, currentQuestion, handleAnswer]);

  if (!currentQuestion) {
    return (
      <div className="min-h-screen bg-mesh flex items-center justify-center">
        <div className="text-white/50 text-xl">Waiting for question...</div>
      </div>
    );
  }

  const progressPercent = currentQuestion.timeLimitMs > 0
    ? (timeRemaining / currentQuestion.timeLimitMs) * 100
    : 0;

  return (
    <div className="min-h-screen bg-mesh flex flex-col">
      {/* Top bar */}
      <div className="px-6 py-3 flex items-center justify-between">
        <span className="text-white/40 text-sm font-display">
          {questionIndex + 1} / {totalQuestions}
        </span>
        <CountdownTimer
          timeRemaining={timeRemaining}
          totalTime={currentQuestion.timeLimitMs}
        />
        <span className="text-white/40 text-sm">
          {currentQuestion.points} pts
        </span>
      </div>

      {/* Progress bar */}
      <div className="h-1 bg-white/5">
        <motion.div
          className="h-full bg-gradient-to-r from-kahoot-purple to-kahoot-blue"
          style={{ width: `${progressPercent}%` }}
          transition={{ duration: 0.1 }}
        />
      </div>

      {/* Question */}
      <div className="flex-1 flex flex-col items-center justify-center px-6 py-8">
        <motion.h2
          initial={{ opacity: 0, y: -10 }}
          animate={{ opacity: 1, y: 0 }}
          className="text-2xl md:text-4xl font-display font-bold text-center mb-8 max-w-2xl"
        >
          {currentQuestion.text}
        </motion.h2>

        {hasAnswered ? (
          <motion.div
            initial={{ opacity: 0, scale: 0.9 }}
            animate={{ opacity: 1, scale: 1 }}
            className="glass-card p-8 text-center"
          >
            <div className="text-3xl mb-2">⏳</div>
            <p className="text-white/60">Answer submitted! Waiting for results...</p>
          </motion.div>
        ) : (
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-3 w-full max-w-2xl">
            {currentQuestion.options.map((option, i) => (
              <AnswerButton
                key={option.id}
                label={option.text}
                shape={ANSWER_SHAPES[i]}
                color={ANSWER_COLORS[i]}
                onClick={() => handleAnswer(option.id)}
                disabled={hasAnswered}
                index={i}
              />
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
