import { useParams, useNavigate } from 'react-router-dom';
import { useEffect } from 'react';
import { motion } from 'framer-motion';
import { useGameStore } from '../store/gameStore';
import { useGameSocket } from '../websocket/useGameSocket';

export default function PlayerLobby() {
  const { sessionId } = useParams<{ sessionId: string }>();
  const navigate = useNavigate();
  const { nickname, gameStatus } = useGameStore();
  const { connectionStatus } = useGameSocket(sessionId || null, 'PLAYER');

  useEffect(() => {
    if (gameStatus === 'ACTIVE') {
      navigate(`/play/${sessionId}`);
    }
  }, [gameStatus, sessionId, navigate]);

  return (
    <div className="min-h-screen bg-mesh flex items-center justify-center px-4">
      <motion.div
        initial={{ opacity: 0 }}
        animate={{ opacity: 1 }}
        className="text-center"
      >
        <motion.div
          animate={{ rotate: 360 }}
          transition={{ duration: 3, repeat: Infinity, ease: 'linear' }}
          className="text-6xl mb-6 inline-block"
        >
          🎯
        </motion.div>

        <h1 className="text-3xl font-display font-bold mb-2">You're In!</h1>
        <p className="text-xl text-purple-400 font-display font-bold mb-4">{nickname}</p>

        <div className="glass-card p-6 inline-block">
          <div className="flex items-center gap-2">
            <div className={`w-2 h-2 rounded-full ${connectionStatus === 'connected' ? 'bg-green-400' : 'bg-yellow-400'} animate-pulse`} />
            <span className="text-white/60">Waiting for host to start the game...</span>
          </div>
        </div>

        <motion.div
          animate={{ opacity: [0.3, 1, 0.3] }}
          transition={{ duration: 2, repeat: Infinity }}
          className="mt-8 text-white/30 text-sm"
        >
          Get ready!
        </motion.div>
      </motion.div>
    </div>
  );
}
