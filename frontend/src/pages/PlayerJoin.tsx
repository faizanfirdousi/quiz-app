import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { motion } from 'framer-motion';
import { joinByPIN } from '../api/sessions';
import { useGameStore } from '../store/gameStore';

export default function PlayerJoin() {
  const [pin, setPin] = useState('');
  const [nickname, setNickname] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const navigate = useNavigate();
  const { setSession, setRole, setNickname: setStoreNickname, setGameStatus } = useGameStore();

  const handleJoin = async (e: React.FormEvent) => {
    e.preventDefault();
    if (pin.length !== 6) return setError('PIN must be 6 digits');
    if (!nickname.trim()) return setError('Enter a nickname');
    if (nickname.length > 20) return setError('Nickname max 20 characters');

    setLoading(true);
    setError('');

    try {
      const result = await joinByPIN(pin, nickname);
      setSession(result.sessionId);
      setRole('PLAYER');
      setStoreNickname(nickname);
      setGameStatus('LOBBY');
      navigate(`/lobby/${result.sessionId}`);
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : 'Failed to join';
      setError(message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-mesh flex items-center justify-center px-4">
      <motion.div
        initial={{ opacity: 0, scale: 0.95 }}
        animate={{ opacity: 1, scale: 1 }}
        className="glass-card p-8 w-full max-w-sm text-center"
      >
        <h1 className="text-3xl font-display font-extrabold mb-2 bg-gradient-to-r from-purple-400 to-blue-400 bg-clip-text text-transparent">
          KahootClone
        </h1>
        <p className="text-white/50 mb-6">Enter the game PIN</p>

        {error && (
          <div className="mb-4 p-3 rounded-xl bg-red-500/10 border border-red-500/20 text-red-400 text-sm">
            {error}
          </div>
        )}

        <form onSubmit={handleJoin} className="space-y-4">
          <input
            type="text"
            value={pin}
            onChange={(e) => setPin(e.target.value.replace(/\D/g, '').slice(0, 6))}
            placeholder="Game PIN"
            className="input-field text-center text-3xl font-display font-bold tracking-[0.3em] placeholder:tracking-normal placeholder:text-lg"
            maxLength={6}
            inputMode="numeric"
          />

          <input
            type="text"
            value={nickname}
            onChange={(e) => setNickname(e.target.value)}
            placeholder="Your nickname"
            className="input-field text-center text-lg"
            maxLength={20}
          />

          <button
            type="submit"
            disabled={loading || pin.length !== 6 || !nickname.trim()}
            className="btn-primary w-full text-lg py-4 disabled:opacity-30"
          >
            {loading ? 'Joining...' : "Let's Go! 🎮"}
          </button>
        </form>
      </motion.div>
    </div>
  );
}
