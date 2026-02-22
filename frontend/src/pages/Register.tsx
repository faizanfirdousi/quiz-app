import { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { motion } from 'framer-motion';
import { useAuthContext } from '../auth/AuthContext';

export default function Register() {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [nickname, setNickname] = useState('');
  const [confirmCode, setConfirmCode] = useState('');
  const [needsConfirmation, setNeedsConfirmation] = useState(false);
  const { register, confirmRegistration, isLoading, error } = useAuthContext();
  const navigate = useNavigate();

  const handleRegister = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      await register(email, password, nickname);
      setNeedsConfirmation(true);
    } catch {
      // Error handled by hook
    }
  };

  const handleConfirm = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      await confirmRegistration(email, confirmCode);
      navigate('/login');
    } catch {
      // Error handled by hook
    }
  };

  return (
    <div className="min-h-screen bg-mesh flex items-center justify-center px-4">
      <motion.div
        initial={{ opacity: 0, scale: 0.95 }}
        animate={{ opacity: 1, scale: 1 }}
        className="glass-card p-8 w-full max-w-md"
      >
        <Link to="/" className="block text-center mb-6">
          <h1 className="text-2xl font-display font-bold bg-gradient-to-r from-purple-400 to-blue-400 bg-clip-text text-transparent">
            KahootClone
          </h1>
        </Link>

        <h2 className="text-2xl font-display font-bold text-center mb-6">
          {needsConfirmation ? 'Verify Email' : 'Create Account'}
        </h2>

        {error && (
          <div className="mb-4 p-3 rounded-xl bg-red-500/10 border border-red-500/20 text-red-400 text-sm">
            {error}
          </div>
        )}

        {!needsConfirmation ? (
          <form onSubmit={handleRegister} className="space-y-4">
            <div>
              <label className="block text-sm text-white/60 mb-1">Nickname</label>
              <input
                type="text"
                value={nickname}
                onChange={(e) => setNickname(e.target.value)}
                className="input-field"
                placeholder="QuizMaster"
                maxLength={20}
                required
              />
            </div>
            <div>
              <label className="block text-sm text-white/60 mb-1">Email</label>
              <input
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                className="input-field"
                placeholder="you@example.com"
                required
              />
            </div>
            <div>
              <label className="block text-sm text-white/60 mb-1">Password</label>
              <input
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                className="input-field"
                placeholder="••••••••"
                minLength={8}
                required
              />
              <p className="mt-1 text-xs text-white/40">Minimum 8 characters</p>
            </div>
            <button type="submit" disabled={isLoading} className="btn-primary w-full disabled:opacity-50">
              {isLoading ? 'Creating account...' : 'Sign Up'}
            </button>
          </form>
        ) : (
          <form onSubmit={handleConfirm} className="space-y-4">
            <p className="text-white/60 text-sm text-center mb-4">
              We sent a verification code to <span className="text-white">{email}</span>
            </p>
            <div>
              <label className="block text-sm text-white/60 mb-1">Verification Code</label>
              <input
                type="text"
                value={confirmCode}
                onChange={(e) => setConfirmCode(e.target.value)}
                className="input-field text-center tracking-widest text-xl"
                placeholder="123456"
                maxLength={6}
                required
              />
            </div>
            <button type="submit" disabled={isLoading} className="btn-primary w-full disabled:opacity-50">
              {isLoading ? 'Verifying...' : 'Verify Email'}
            </button>
          </form>
        )}

        <p className="mt-6 text-center text-white/50 text-sm">
          Already have an account?{' '}
          <Link to="/login" className="text-purple-400 hover:text-purple-300 transition-colors">Sign In</Link>
        </p>
      </motion.div>
    </div>
  );
}
