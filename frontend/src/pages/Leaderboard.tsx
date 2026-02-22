import { useParams, Link } from 'react-router-dom';
import { motion } from 'framer-motion';
import { useLeaderboardStore } from '../store/leaderboardStore';
import { useGameStore } from '../store/gameStore';
import LeaderboardRow from '../components/LeaderboardRow';

const PODIUM_EMOJIS = ['🥇', '🥈', '🥉'];

export default function Leaderboard() {
  const { sessionId } = useParams<{ sessionId: string }>();
  const { entries } = useLeaderboardStore();
  const { gameStatus, questionIndex, totalQuestions, role } = useGameStore();
  const isFinal = gameStatus === 'FINISHED';

  return (
    <div className="min-h-screen bg-mesh flex flex-col">
      {/* Header */}
      <div className="px-6 py-4 flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-display font-bold">
            {isFinal ? '🏆 Final Results' : '📊 Leaderboard'}
          </h1>
          {!isFinal && (
            <p className="text-white/40 text-sm">
              After question {questionIndex + 1} of {totalQuestions}
            </p>
          )}
        </div>
        {isFinal && (
          <Link to="/" className="btn-primary text-sm">
            🏠 Home
          </Link>
        )}
      </div>

      {/* Podium for top 3 (on final screen) */}
      {isFinal && entries.length >= 3 && (
        <div className="flex justify-center items-end gap-4 px-6 py-8">
          {[1, 0, 2].map((idx) => {
            const player = entries[idx];
            if (!player) return null;
            const heights = ['h-32', 'h-24', 'h-20'];
            return (
              <motion.div
                key={player.userId}
                initial={{ opacity: 0, y: 30 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ delay: idx * 0.2 }}
                className="text-center"
              >
                <div className="text-3xl mb-1">{PODIUM_EMOJIS[idx]}</div>
                <div className="font-display font-bold text-sm mb-1 truncate max-w-[100px]">
                  {player.nickname}
                </div>
                <div className="text-xs text-white/60 mb-2">{Math.round(player.score)}</div>
                <div
                  className={`${heights[idx]} w-20 rounded-t-xl ${
                    idx === 0 ? 'bg-gradient-to-t from-yellow-600 to-yellow-400' :
                    idx === 1 ? 'bg-gradient-to-t from-gray-500 to-gray-300' :
                    'bg-gradient-to-t from-amber-700 to-amber-500'
                  }`}
                />
              </motion.div>
            );
          })}
        </div>
      )}

      {/* Full leaderboard list */}
      <div className="flex-1 px-6 pb-8">
        <div className="max-w-xl mx-auto space-y-2">
          {entries.length === 0 ? (
            <div className="text-center py-12 text-white/50">
              No scores yet
            </div>
          ) : (
            entries.map((entry, i) => (
              <motion.div
                key={entry.userId}
                initial={{ opacity: 0, x: -20 }}
                animate={{ opacity: 1, x: 0 }}
                transition={{ delay: i * 0.05 }}
              >
                <LeaderboardRow
                  rank={i + 1}
                  nickname={entry.nickname}
                  score={Math.round(entry.score)}
                />
              </motion.div>
            ))
          )}
        </div>
      </div>

      {/* Next question button (host only, non-final) */}
      {role === 'HOST' && !isFinal && (
        <div className="px-6 py-4 border-t border-white/5">
          <div className="max-w-xl mx-auto">
            <button
              onClick={() => {
                // This would send next_question via WS
              }}
              className="btn-primary w-full text-lg"
            >
              Next Question →
            </button>
          </div>
        </div>
      )}
    </div>
  );
}
