import { useParams, useNavigate } from 'react-router-dom';
import { useEffect } from 'react';
import { motion } from 'framer-motion';
import { useGameStore } from '../store/gameStore';
import { useLeaderboardStore } from '../store/leaderboardStore';

export default function AnswerResult() {
  const { sessionId } = useParams<{ sessionId: string }>();
  const navigate = useNavigate();
  const { lastAnswerResult, currentQuestion, gameStatus } = useGameStore();
  const { myRank, myScore } = useLeaderboardStore();

  useEffect(() => {
    if (gameStatus === 'FINISHED') {
      navigate(`/leaderboard/${sessionId}`);
    }
  }, [gameStatus, sessionId, navigate]);

  // Auto-navigate to leaderboard after showing result
  useEffect(() => {
    const timer = setTimeout(() => {
      navigate(`/leaderboard/${sessionId}`);
    }, 5000);
    return () => clearTimeout(timer);
  }, [sessionId, navigate]);

  if (!lastAnswerResult) {
    return (
      <div className="min-h-screen bg-mesh flex items-center justify-center">
        <div className="text-white/50">Loading result...</div>
      </div>
    );
  }

  const { isCorrect, pointsEarned } = lastAnswerResult;

  return (
    <div className="min-h-screen bg-mesh flex items-center justify-center px-4">
      <motion.div
        initial={{ opacity: 0, scale: 0.8 }}
        animate={{ opacity: 1, scale: 1 }}
        transition={{ type: 'spring', stiffness: 200 }}
        className="text-center"
      >
        {/* Result icon */}
        <motion.div
          initial={{ scale: 0 }}
          animate={{ scale: 1 }}
          transition={{ type: 'spring', stiffness: 300, delay: 0.2 }}
          className="text-8xl mb-6"
        >
          {isCorrect ? '🎉' : '😔'}
        </motion.div>

        <motion.h1
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.3 }}
          className={`text-4xl font-display font-extrabold mb-4 ${
            isCorrect ? 'text-green-400' : 'text-red-400'
          }`}
        >
          {isCorrect ? 'Correct!' : 'Wrong!'}
        </motion.h1>

        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.5 }}
          className="glass-card p-6 inline-block"
        >
          <div className="text-3xl font-display font-extrabold bg-gradient-to-r from-purple-400 to-blue-400 bg-clip-text text-transparent">
            +{pointsEarned}
          </div>
          <p className="text-white/50 text-sm">points</p>
        </motion.div>

        <motion.div
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          transition={{ delay: 0.7 }}
          className="mt-6 flex justify-center gap-6"
        >
          <div className="text-center">
            <div className="text-2xl font-display font-bold text-white">{myScore}</div>
            <div className="text-white/40 text-xs">Total Score</div>
          </div>
          <div className="text-center">
            <div className="text-2xl font-display font-bold text-yellow-400">#{myRank || '?'}</div>
            <div className="text-white/40 text-xs">Rank</div>
          </div>
        </motion.div>

        {!isCorrect && currentQuestion && (
          <motion.p
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            transition={{ delay: 1 }}
            className="mt-6 text-white/40 text-sm"
          >
            Correct answer was highlighted
          </motion.p>
        )}
      </motion.div>
    </div>
  );
}
