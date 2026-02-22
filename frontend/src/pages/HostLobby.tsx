import { useParams } from 'react-router-dom';
import { motion } from 'framer-motion';
import { useGameStore } from '../store/gameStore';
import { useGameSocket } from '../websocket/useGameSocket';
import { WS_ACTIONS } from '../websocket/eventTypes';
import PinDisplay from '../components/PinDisplay';

export default function HostLobby() {
  const { sessionId } = useParams<{ sessionId: string }>();
  const { playerCount, gameStatus } = useGameStore();
  const { sendMessage, connectionStatus } = useGameSocket(sessionId || null, 'HOST');

  const handleStartGame = () => {
    if (sessionId) {
      sendMessage(WS_ACTIONS.START_GAME, { sessionId });
    }
  };

  return (
    <div className="min-h-screen bg-mesh flex flex-col items-center justify-center px-4">
      <motion.div
        initial={{ opacity: 0, y: -20 }}
        animate={{ opacity: 1, y: 0 }}
        className="text-center max-w-lg w-full"
      >
        <h1 className="text-3xl font-display font-bold mb-2">Waiting Room</h1>
        <p className="text-white/50 mb-8">Share this PIN with your players</p>

        <PinDisplay pin={sessionId?.slice(0, 6).toUpperCase() || '------'} />

        <div className="glass-card p-6 mt-8 mb-6">
          <div className="flex items-center justify-center gap-2 mb-2">
            <div className={`w-2 h-2 rounded-full ${connectionStatus === 'connected' ? 'bg-green-400' : 'bg-yellow-400'} animate-pulse`} />
            <span className="text-white/60 text-sm">
              {connectionStatus === 'connected' ? 'Connected' : 'Connecting...'}
            </span>
          </div>

          <motion.div
            key={playerCount}
            initial={{ scale: 1.2 }}
            animate={{ scale: 1 }}
            className="text-5xl font-display font-extrabold bg-gradient-to-r from-purple-400 to-blue-400 bg-clip-text text-transparent"
          >
            {playerCount}
          </motion.div>
          <p className="text-white/50 text-sm mt-1">
            {playerCount === 1 ? 'player joined' : 'players joined'}
          </p>
        </div>

        <button
          onClick={handleStartGame}
          disabled={playerCount === 0 || gameStatus !== 'LOBBY'}
          className="btn-primary text-lg px-10 py-4 disabled:opacity-30"
        >
          🚀 Start Game
        </button>
      </motion.div>
    </div>
  );
}
